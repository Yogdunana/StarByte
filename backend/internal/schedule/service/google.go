package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/dto"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/importer"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

// GoogleSettings 来自环境变量，禁止写入仓库密钥。
type GoogleSettings struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	FrontendURL  string
}

func LoadGoogleSettings() GoogleSettings {
	return GoogleSettings{
		ClientID:     os.Getenv("GOOGLE_CALENDAR_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CALENDAR_CLIENT_SECRET"),
		RedirectURI:  firstNonEmpty(os.Getenv("GOOGLE_CALENDAR_REDIRECT_URI"), os.Getenv("GOOGLE_CALENDAR_REDIRECT_URL")),
		FrontendURL:  os.Getenv("CAS_FRONTEND_URL"),
	}
}

func (g GoogleSettings) Configured() bool {
	return strings.TrimSpace(g.ClientID) != "" && strings.TrimSpace(g.ClientSecret) != "" && strings.TrimSpace(g.RedirectURI) != ""
}

func (s *scheduleService) GoogleStatus(ctx context.Context, operator uuid.UUID) (*dto.GoogleStatusResponse, error) {
	out := &dto.GoogleStatusResponse{Configured: s.google.Configured()}
	if !out.Configured {
		return out, nil
	}
	acct, err := s.rows.GetGoogleAccount(ctx, operator)
	if err != nil {
		return nil, fmt.Errorf("google account: %w", err)
	}
	if acct == nil || acct.RefreshToken == "" && acct.AccessToken == "" {
		return out, nil
	}
	out.Connected = true
	out.Email = acct.GoogleEmail
	if acct.CalendarID != nil {
		out.CalendarID = acct.CalendarID.String()
	}
	return out, nil
}

func (s *scheduleService) GoogleConnectURL(_ context.Context, operator uuid.UUID) (*dto.GoogleConnectResponse, error) {
	if !s.google.Configured() {
		return nil, response.NewError(response.CodeScheduleGoogleNotReady, "未配置 GOOGLE_CALENDAR_CLIENT_ID / SECRET / REDIRECT_URI，Google 同步仅预留接口")
	}
	q := url.Values{}
	q.Set("client_id", s.google.ClientID)
	q.Set("redirect_uri", s.google.RedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "https://www.googleapis.com/auth/calendar.readonly https://www.googleapis.com/auth/userinfo.email")
	q.Set("access_type", "offline")
	q.Set("prompt", "consent")
	q.Set("state", s.signGoogleState(operator))
	return &dto.GoogleConnectResponse{
		AuthURL: "https://accounts.google.com/o/oauth2/v2/auth?" + q.Encode(),
	}, nil
}

func (s *scheduleService) GoogleCallback(ctx context.Context, operator uuid.UUID, code, state string, scope *rbacModel.DataScopeCondition) (*dto.GoogleStatusResponse, error) {
	if !s.google.Configured() {
		return nil, response.NewError(response.CodeScheduleGoogleNotReady, "Google 日历未配置")
	}
	if operator == uuid.Nil {
		return nil, response.NewError(response.CodeUnauthorized, "请登录后完成 Google 绑定")
	}
	stateUID, err := s.parseGoogleState(state)
	if err != nil || stateUID == uuid.Nil {
		return nil, response.NewError(response.CodeUnauthorized, "OAuth state 无效或已过期")
	}
	if stateUID != operator {
		return nil, response.NewError(response.CodeForbidden, "OAuth state 与当前用户不一致")
	}
	if strings.TrimSpace(code) == "" {
		return nil, response.NewError(response.CodeScheduleImportInvalid, "缺少授权码")
	}
	tok, err := s.exchangeGoogleCode(ctx, code)
	if err != nil {
		return nil, response.NewError(response.CodeScheduleGoogleNotReady, "兑换 Google 令牌失败: "+err.Error())
	}
	email, _ := s.googleEmail(ctx, tok.AccessToken)
	now := time.Now()
	acct := &model.GoogleAccount{
		ID: uuid.New(), UserID: operator, AccessToken: tok.AccessToken, RefreshToken: tok.RefreshToken,
		GoogleEmail: email, CreatedAt: now, UpdatedAt: now,
	}
	if tok.ExpiresIn > 0 {
		exp := now.Add(time.Duration(tok.ExpiresIn) * time.Second)
		acct.TokenExpiry = &exp
	}
	if err := s.rows.UpsertGoogleAccount(ctx, acct); err != nil {
		return nil, fmt.Errorf("store google account: %w", err)
	}
	if _, err := s.GoogleSync(ctx, operator, scope); err != nil {
		// 授权成功但首拉失败仍视为已连接
		_ = err
	}
	return s.GoogleStatus(ctx, operator)
}

func (s *scheduleService) GoogleDisconnect(ctx context.Context, operator uuid.UUID) error {
	if err := s.rows.DeleteGoogleAccount(ctx, operator); err != nil {
		return fmt.Errorf("disconnect google: %w", err)
	}
	return nil
}

func (s *scheduleService) GoogleSync(ctx context.Context, operator uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.ImportResult, error) {
	if !s.google.Configured() {
		return nil, response.NewError(response.CodeScheduleGoogleNotReady, "Google 日历未配置，接口已预留 source=google")
	}
	acct, err := s.rows.GetGoogleAccount(ctx, operator)
	if err != nil {
		return nil, fmt.Errorf("google account: %w", err)
	}
	if acct == nil || (acct.AccessToken == "" && acct.RefreshToken == "") {
		return nil, response.NewError(response.CodeScheduleGoogleNotReady, "尚未连接 Google 日历")
	}
	token, err := s.ensureGoogleAccess(ctx, acct)
	if err != nil {
		return nil, response.NewError(response.CodeScheduleGoogleNotReady, "刷新 Google 令牌失败: "+err.Error())
	}
	drafts, err := s.pullGoogleEvents(ctx, token)
	if err != nil {
		return nil, response.NewError(response.CodeScheduleImportInvalid, "拉取 Google 事件失败: "+err.Error())
	}
	if len(drafts) == 0 {
		return nil, response.NewError(response.CodeScheduleImportEmpty, "Google 主日历没有可导入的事件")
	}
	name := "Google 日历"
	if acct.GoogleEmail != "" {
		name = "Google · " + acct.GoogleEmail
	}
	cal, replaced, err := s.ensureLayer(ctx, operator, model.SourceGoogle, "google:primary", name, acct.GoogleEmail)
	if err != nil {
		return nil, err
	}
	if err := s.writeDrafts(ctx, operator, cal.ID, model.OriginGoogle, drafts); err != nil {
		return nil, err
	}
	acct.CalendarID = &cal.ID
	acct.UpdatedAt = time.Now()
	if err := s.rows.UpsertGoogleAccount(ctx, acct); err != nil {
		return nil, fmt.Errorf("update google calendar id: %w", err)
	}
	out, err := s.GetCalendar(ctx, operator, cal.ID, scope)
	if err != nil {
		return nil, err
	}
	return &dto.ImportResult{
		CalendarID: cal.ID.String(), Calendar: out, EventCount: len(drafts),
		Replaced: replaced, Source: model.SourceGoogle,
	}, nil
}

func (s *scheduleService) DispatchGoogleSync(ctx context.Context, _ string, logf func(string)) error {
	if logf == nil {
		logf = func(string) {}
	}
	if !s.google.Configured() {
		logf("google calendar sync skipped: credentials not configured")
		return nil
	}
	logf("google calendar sync hook ready (per-user pull via POST /schedules/google/sync)")
	return nil
}

func (s *scheduleService) ParseGoogleState(state string) (uuid.UUID, error) {
	return s.parseGoogleState(state)
}

func (s *scheduleService) FrontendRedirect() string {
	if u := strings.TrimRight(s.google.FrontendURL, "/"); u != "" {
		return u + "/schedule?google=connected"
	}
	return ""
}

func (s *scheduleService) FrontendCallbackRedirect(code, state string) string {
	base := strings.TrimRight(s.google.FrontendURL, "/")
	if base == "" {
		return ""
	}
	u, err := url.Parse(base + "/schedule")
	if err != nil {
		return ""
	}
	q := u.Query()
	q.Set("google", "callback")
	if code != "" {
		q.Set("code", code)
	}
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *scheduleService) signGoogleState(userID uuid.UUID) string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	payload := userID.String() + "." + ts
	mac := hmac.New(sha256.New, []byte(s.google.ClientSecret))
	_, _ = mac.Write([]byte(payload))
	return payload + "." + hex.EncodeToString(mac.Sum(nil))
}

func (s *scheduleService) parseGoogleState(state string) (uuid.UUID, error) {
	parts := strings.Split(state, ".")
	if len(parts) != 3 {
		return uuid.Nil, fmt.Errorf("bad state")
	}
	uid, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, err
	}
	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return uuid.Nil, err
	}
	if time.Since(time.Unix(ts, 0)) > 20*time.Minute {
		return uuid.Nil, fmt.Errorf("state expired")
	}
	mac := hmac.New(sha256.New, []byte(s.google.ClientSecret))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal([]byte(hex.EncodeToString(mac.Sum(nil))), []byte(parts[2])) {
		return uuid.Nil, fmt.Errorf("state mismatch")
	}
	return uid, nil
}

type googleToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func (s *scheduleService) exchangeGoogleCode(ctx context.Context, code string) (*googleToken, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", s.google.ClientID)
	form.Set("client_secret", s.google.ClientSecret)
	form.Set("redirect_uri", s.google.RedirectURI)
	form.Set("grant_type", "authorization_code")
	return s.postGoogleToken(ctx, form)
}

func (s *scheduleService) ensureGoogleAccess(ctx context.Context, acct *model.GoogleAccount) (string, error) {
	if acct.AccessToken != "" && (acct.TokenExpiry == nil || acct.TokenExpiry.After(time.Now().Add(time.Minute))) {
		return acct.AccessToken, nil
	}
	if acct.RefreshToken == "" {
		return "", fmt.Errorf("no refresh token")
	}
	form := url.Values{}
	form.Set("client_id", s.google.ClientID)
	form.Set("client_secret", s.google.ClientSecret)
	form.Set("refresh_token", acct.RefreshToken)
	form.Set("grant_type", "refresh_token")
	tok, err := s.postGoogleToken(ctx, form)
	if err != nil {
		return "", err
	}
	acct.AccessToken = tok.AccessToken
	if tok.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
		acct.TokenExpiry = &exp
	}
	acct.UpdatedAt = time.Now()
	_ = s.rows.UpsertGoogleAccount(ctx, acct)
	return acct.AccessToken, nil
}

func (s *scheduleService) postGoogleToken(ctx context.Context, form url.Values) (*googleToken, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.doHTTP(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("token http %d", resp.StatusCode)
	}
	var tok googleToken
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, err
	}
	if tok.AccessToken == "" {
		return nil, fmt.Errorf("empty access token")
	}
	return &tok, nil
}

func (s *scheduleService) googleEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := s.doHTTP(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var out struct {
		Email string `json:"email"`
	}
	_ = json.Unmarshal(body, &out)
	return out.Email, nil
}

func (s *scheduleService) pullGoogleEvents(ctx context.Context, accessToken string) ([]importer.DraftEvent, error) {
	u, _ := url.Parse("https://www.googleapis.com/calendar/v3/calendars/primary/events")
	q := u.Query()
	q.Set("singleEvents", "true")
	q.Set("maxResults", "2500")
	q.Set("orderBy", "startTime")
	q.Set("timeMin", time.Now().AddDate(-1, 0, 0).UTC().Format(time.RFC3339))
	q.Set("timeMax", time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC3339))
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := s.doHTTP(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("events http %d", resp.StatusCode)
	}
	var payload struct {
		Items []struct {
			ID          string `json:"id"`
			Summary     string `json:"summary"`
			Location    string `json:"location"`
			Description string `json:"description"`
			Start       struct {
				DateTime string `json:"dateTime"`
				Date     string `json:"date"`
			} `json:"start"`
			End struct {
				DateTime string `json:"dateTime"`
				Date     string `json:"date"`
			} `json:"end"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	out := make([]importer.DraftEvent, 0, len(payload.Items))
	for _, it := range payload.Items {
		start, allDay, ok := parseGoogleTime(it.Start.DateTime, it.Start.Date)
		if !ok {
			continue
		}
		end, _, endOK := parseGoogleTime(it.End.DateTime, it.End.Date)
		if !endOK {
			end = start.Add(time.Hour)
		}
		title := strings.TrimSpace(it.Summary)
		if title == "" {
			title = "(无标题)"
		}
		out = append(out, importer.DraftEvent{
			Title: title, Description: it.Description, Location: it.Location,
			StartAt: start, EndAt: end, AllDay: allDay, ExternalUID: "google:" + it.ID,
		})
	}
	return out, nil
}

func parseGoogleTime(dateTime, date string) (time.Time, bool, bool) {
	if dateTime != "" {
		if t, err := time.Parse(time.RFC3339, dateTime); err == nil {
			return t, false, true
		}
	}
	if date != "" {
		if t, err := time.ParseInLocation("2006-01-02", date, locationShanghai()); err == nil {
			return t, true, true
		}
	}
	return time.Time{}, false, false
}

func locationShanghai() *time.Location {
	if tz, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		return tz
	}
	return time.FixedZone("CST", 8*3600)
}

func (s *scheduleService) doHTTP(req *http.Request) (*http.Response, error) {
	if s.httpDo != nil {
		return s.httpDo(req)
	}
	client := &http.Client{Timeout: 20 * time.Second}
	return client.Do(req)
}
