package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/internal/auth/repo"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/utils"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

const (
	identityTypeCAS     = "cas"
	casStateTTL         = 10 * time.Minute
	casExchangeTTL      = 60 * time.Second
	casUsernameMaxLen   = 50
	casRedirectFallback = "/dashboard"
)

// CASDeps wires optional campus CAS login. All fields may be nil when CAS is off.
type CASDeps struct {
	Config     *config.CASConfig
	Store      repo.CASTicketStore
	Validator  TicketValidator
	AssignRole RoleAssigner
}

func (s *authService) casEnabled() bool {
	return s != nil && s.cas != nil && s.cas.Enabled
}

// CASStatus reports whether campus CAS login is turned on.
func (s *authService) CASStatus() dto.CASStatusResponse {
	return dto.CASStatusResponse{Enabled: s.casEnabled()}
}

// BuildCASLoginURL stores state and returns the CAS /login redirect.
// publicOrigin 是浏览器实际访问地址（http://IP 或后续域名）；漏测无域名时靠它拼 service。
func (s *authService) BuildCASLoginURL(ctx context.Context, redirect, publicOrigin string) (*dto.CASLoginStart, error) {
	if !s.casEnabled() {
		return nil, response.NewError(response.CodeNotImplemented, "学校统一认证暂未开通")
	}
	if s.casStore == nil || s.cas.ServerURL == "" {
		return nil, response.NewError(response.CodeInternalError, "CAS 配置不完整")
	}
	service, origin, err := resolveCASURLs(s.cas, publicOrigin)
	if err != nil {
		return nil, err
	}
	state, err := randomHex(16)
	if err != nil {
		return nil, fmt.Errorf("cas state: %w", err)
	}
	payload, err := json.Marshal(casStateRecord{
		Redirect: sanitizeRedirect(redirect),
		Origin:   origin,
		Service:  service,
	})
	if err != nil {
		return nil, fmt.Errorf("cas state: %w", err)
	}
	if err := s.casStore.PutState(ctx, state, string(payload), casStateTTL); err != nil {
		return nil, fmt.Errorf("store cas state: %w", err)
	}
	login := strings.TrimRight(s.cas.ServerURL, "/") + "/login?service=" + url.QueryEscape(service)
	return &dto.CASLoginStart{Location: login, State: state, Service: service}, nil
}

// CompleteCASCallback validates the ST, issues JWT, and returns the frontend exchange URL.
func (s *authService) CompleteCASCallback(ctx context.Context, ticket, state, ip, userAgent, publicOrigin string) (string, error) {
	_, fallbackOrigin, _ := resolveCASURLs(s.cas, publicOrigin)
	if !s.casEnabled() {
		return casFrontendError(fallbackOrigin, "disabled"), nil
	}
	ticket = strings.TrimSpace(ticket)
	state = strings.TrimSpace(state)
	if ticket == "" || state == "" {
		return casFrontendError(fallbackOrigin, "missing_ticket"), nil
	}
	if s.casStore == nil || s.casValidator == nil {
		return casFrontendError(fallbackOrigin, "not_configured"), nil
	}
	raw, ok, err := s.casStore.TakeState(ctx, state)
	if err != nil {
		return casFrontendError(fallbackOrigin, "state_error"), nil
	}
	if !ok {
		return casFrontendError(fallbackOrigin, "state_expired"), nil
	}
	rec := parseCASState(raw, fallbackOrigin)
	principal, err := s.casValidator.Validate(ctx, rec.Service, ticket)
	if err != nil || principal == nil || strings.TrimSpace(principal.User) == "" {
		return casFrontendError(rec.Origin, "ticket_invalid"), nil
	}
	user, err := s.resolveOrProvisionCASUser(ctx, principal)
	if err != nil {
		return casFrontendError(rec.Origin, "user_resolve"), nil
	}
	if user.Status == 1 {
		return casFrontendError(rec.Origin, "disabled_user"), nil
	}
	if user.Status == 2 {
		return casFrontendError(rec.Origin, "locked_user"), nil
	}
	tokens, err := s.issueSession(ctx, user, ip, userAgent)
	if err != nil {
		return casFrontendError(rec.Origin, "token"), nil
	}
	code, err := randomHex(16)
	if err != nil {
		return casFrontendError(rec.Origin, "code"), nil
	}
	payload, err := json.Marshal(casExchangePayload{Login: *tokens, Redirect: rec.Redirect})
	if err != nil {
		return casFrontendError(rec.Origin, "code"), nil
	}
	if err := s.casStore.PutCode(ctx, code, payload, casExchangeTTL); err != nil {
		return casFrontendError(rec.Origin, "code"), nil
	}
	return casFrontendSuccess(rec.Origin, code), nil
}

// ExchangeCASCode consumes a one-time callback code and returns the JWT pair.
func (s *authService) ExchangeCASCode(ctx context.Context, code string) (*dto.CASExchangeResponse, error) {
	if !s.casEnabled() {
		return nil, response.NewError(response.CodeNotImplemented, "学校统一认证暂未开通")
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, response.NewError(response.CodeBadRequest, "缺少兑换码")
	}
	if s.casStore == nil {
		return nil, response.NewError(response.CodeInternalError, "CAS 未配置")
	}
	raw, err := s.casStore.TakeCode(ctx, code)
	if err == goredis.Nil || len(raw) == 0 {
		return nil, response.NewError(response.CodeTokenInvalid, "统一认证凭证已失效，请重新登录")
	}
	if err != nil {
		return nil, fmt.Errorf("take cas code: %w", err)
	}
	var payload casExchangePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, response.NewError(response.CodeTokenInvalid, "统一认证凭证无效")
	}
	return &dto.CASExchangeResponse{
		LoginResponse: payload.Login,
		Redirect:      sanitizeRedirect(payload.Redirect),
	}, nil
}

type casExchangePayload struct {
	Login    dto.LoginResponse `json:"login"`
	Redirect string            `json:"redirect"`
}

func (s *authService) resolveOrProvisionCASUser(ctx context.Context, p *CASPrincipal) (*model.User, error) {
	casUser := sanitizeCASUsername(p.User)
	if casUser == "" {
		return nil, response.NewError(response.CodeInvalidCredentials, "CAS 未返回有效账号")
	}
	studentNo := pickStudentNo(p, casUser)

	if user, err := s.userRepo.GetByIdentity(ctx, identityTypeCAS, casUser); err != nil {
		return nil, fmt.Errorf("lookup cas identity: %w", err)
	} else if user != nil {
		s.touchCASProfile(ctx, user, p)
		return user, nil
	}
	if studentNo != "" && s.identity != nil {
		uid, err := s.identity.GetUserIDByStudentNo(ctx, studentNo)
		if err != nil {
			return nil, fmt.Errorf("lookup student no: %w", err)
		}
		if uid != uuid.Nil {
			user, err := s.userRepo.GetByID(ctx, uid)
			if err != nil {
				return nil, err
			}
			if user != nil {
				_ = s.bindCASIdentity(ctx, user.ID, casUser)
				s.touchCASProfile(ctx, user, p)
				return user, nil
			}
		}
	}
	if user, err := s.userRepo.GetByUsername(ctx, casUser); err != nil {
		return nil, fmt.Errorf("lookup username: %w", err)
	} else if user != nil {
		_ = s.bindCASIdentity(ctx, user.ID, casUser)
		s.touchCASProfile(ctx, user, p)
		return user, nil
	}

	if s.cas == nil || !s.cas.AllowAutoProvision {
		return nil, response.NewError(response.CodeUserNotFound, "本地没有对应账号，请先完成入会或联系管理员")
	}
	return s.provisionCASUser(ctx, p, casUser)
}

func (s *authService) provisionCASUser(ctx context.Context, p *CASPrincipal, casUser string) (*model.User, error) {
	rawPwd, err := randomHex(24)
	if err != nil {
		return nil, fmt.Errorf("cas password: %w", err)
	}
	hash, err := utils.HashPassword(rawPwd)
	if err != nil {
		return nil, fmt.Errorf("hash cas password: %w", err)
	}
	user := &model.User{
		ID:           uuid.New(),
		Username:     casUser,
		PasswordHash: hash,
		RealName:     pickRealName(p),
		Email:        pickAttr(p, "email", "mail"),
		Status:       0,
	}
	if err := s.userRepo.Create(ctx, nil, user); err != nil {
		if existing, getErr := s.userRepo.GetByUsername(ctx, casUser); getErr == nil && existing != nil {
			_ = s.bindCASIdentity(ctx, existing.ID, casUser)
			return existing, nil
		}
		return nil, fmt.Errorf("create cas user: %w", err)
	}
	_ = s.bindCASIdentity(ctx, user.ID, casUser)
	if s.casRole != nil {
		_ = s.casRole.AssignDefault(ctx, user.ID)
	}
	return user, nil
}

func (s *authService) bindCASIdentity(ctx context.Context, userID uuid.UUID, casUser string) error {
	existing, err := s.userRepo.GetByIdentity(ctx, identityTypeCAS, casUser)
	if err != nil || existing != nil {
		return err
	}
	return s.userRepo.CreateIdentity(ctx, &model.UserIdentity{
		ID:            uuid.New(),
		UserID:        userID,
		IdentityType:  identityTypeCAS,
		IdentityValue: casUser,
		IsPrimary:     true,
	})
}

func (s *authService) touchCASProfile(ctx context.Context, user *model.User, p *CASPrincipal) {
	name := pickRealName(p)
	email := pickAttr(p, "email", "mail")
	changed := false
	if user.RealName == "" && name != "" {
		user.RealName = name
		changed = true
	}
	if user.Email == "" && email != "" {
		user.Email = email
		changed = true
	}
	if changed {
		_ = s.userRepo.Update(ctx, nil, user)
	}
}

type casStateRecord struct {
	Redirect string `json:"redirect"`
	Origin   string `json:"origin"`
	Service  string `json:"service"`
}

func parseCASState(raw, fallbackOrigin string) casStateRecord {
	rec := casStateRecord{Redirect: casRedirectFallback, Origin: sanitizePublicOrigin(fallbackOrigin)}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if rec.Origin != "" {
			rec.Service = rec.Origin + "/api/v1/auth/cas/callback"
		}
		return rec
	}
	if strings.HasPrefix(raw, "{") {
		var parsed casStateRecord
		if json.Unmarshal([]byte(raw), &parsed) == nil {
			if parsed.Redirect != "" {
				rec.Redirect = sanitizeRedirect(parsed.Redirect)
			}
			if origin := sanitizePublicOrigin(parsed.Origin); origin != "" {
				rec.Origin = origin
			}
			if parsed.Service != "" {
				rec.Service = parsed.Service
			}
		}
	} else {
		rec.Redirect = sanitizeRedirect(raw)
	}
	if rec.Service == "" && rec.Origin != "" {
		rec.Service = rec.Origin + "/api/v1/auth/cas/callback"
	}
	return rec
}

func resolveCASURLs(cfg *config.CASConfig, requestOrigin string) (serviceURL, frontendOrigin string, err error) {
	origin := ""
	if cfg != nil {
		origin = sanitizePublicOrigin(cfg.FrontendURL)
	}
	if origin == "" {
		origin = sanitizePublicOrigin(requestOrigin)
	}
	if origin == "" && cfg != nil && strings.TrimSpace(cfg.ServiceURL) != "" {
		if u, perr := url.Parse(strings.TrimSpace(cfg.ServiceURL)); perr == nil {
			origin = sanitizePublicOrigin(u.Scheme + "://" + u.Host)
		}
	}
	if origin == "" {
		return "", "", response.NewError(response.CodeBadRequest, "无法确定访问地址，请用校园网 IP 打开系统")
	}
	service := ""
	if cfg != nil {
		service = strings.TrimSpace(cfg.ServiceURL)
	}
	if service == "" {
		service = origin + "/api/v1/auth/cas/callback"
	}
	return service, origin, nil
}

func sanitizePublicOrigin(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func casFrontendSuccess(origin, code string) string {
	base := sanitizePublicOrigin(origin)
	if base == "" {
		return "/login/cas?code=" + url.QueryEscape(code)
	}
	return base + "/login/cas?code=" + url.QueryEscape(code)
}

func casFrontendError(origin, reason string) string {
	if reason == "" {
		reason = "unknown"
	}
	base := sanitizePublicOrigin(origin)
	if base == "" {
		return "/login?cas_error=" + url.QueryEscape(reason)
	}
	return base + "/login?cas_error=" + url.QueryEscape(reason)
}

func sanitizeRedirect(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || !strings.HasPrefix(p, "/") || strings.HasPrefix(p, "//") || strings.Contains(p, "://") {
		return casRedirectFallback
	}
	return p
}

func sanitizeCASUsername(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > casUsernameMaxLen {
		return ""
	}
	for _, r := range raw {
		if r > unicode.MaxASCII || r <= 32 || strings.ContainsRune(" /\\?&#%=", r) {
			return ""
		}
	}
	return raw
}

func pickStudentNo(p *CASPrincipal, fallback string) string {
	if v := pickAttr(p, "studentno", "student_no", "stuno", "uid"); v != "" {
		return v
	}
	if isMostlyDigits(fallback) {
		return fallback
	}
	return ""
}

func pickRealName(p *CASPrincipal) string {
	return pickAttr(p, "name", "cn", "displayname", "xm", "xingming")
}

func pickAttr(p *CASPrincipal, keys ...string) string {
	if p == nil {
		return ""
	}
	for _, k := range keys {
		if v := strings.TrimSpace(p.Attributes[strings.ToLower(k)]); v != "" {
			return v
		}
	}
	return ""
}

func isMostlyDigits(s string) bool {
	if s == "" {
		return false
	}
	digits := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	return digits >= len(s)*3/4
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
