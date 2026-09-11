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
func (s *authService) BuildCASLoginURL(ctx context.Context, redirect string) (string, error) {
	if !s.casEnabled() {
		return "", response.NewError(response.CodeNotImplemented, "学校统一认证暂未开通")
	}
	if s.casStore == nil || s.cas.ServiceURL == "" || s.cas.ServerURL == "" {
		return "", response.NewError(response.CodeInternalError, "CAS 配置不完整")
	}
	state, err := randomHex(16)
	if err != nil {
		return "", fmt.Errorf("cas state: %w", err)
	}
	if err := s.casStore.PutState(ctx, state, sanitizeRedirect(redirect), casStateTTL); err != nil {
		return "", fmt.Errorf("store cas state: %w", err)
	}
	service := casServiceURL(s.cas.ServiceURL, state)
	login := strings.TrimRight(s.cas.ServerURL, "/") + "/login?service=" + url.QueryEscape(service)
	return login, nil
}

// CompleteCASCallback validates the ST, issues JWT, and returns the frontend exchange URL.
func (s *authService) CompleteCASCallback(ctx context.Context, ticket, state, ip, userAgent string) (string, error) {
	if !s.casEnabled() {
		return s.casFrontendError("disabled"), nil
	}
	ticket = strings.TrimSpace(ticket)
	state = strings.TrimSpace(state)
	if ticket == "" || state == "" {
		return s.casFrontendError("missing_ticket"), nil
	}
	if s.casStore == nil || s.casValidator == nil {
		return s.casFrontendError("not_configured"), nil
	}
	redirect, ok, err := s.casStore.TakeState(ctx, state)
	if err != nil {
		return s.casFrontendError("state_error"), nil
	}
	if !ok {
		return s.casFrontendError("state_expired"), nil
	}
	service := casServiceURL(s.cas.ServiceURL, state)
	principal, err := s.casValidator.Validate(ctx, service, ticket)
	if err != nil || principal == nil || strings.TrimSpace(principal.User) == "" {
		return s.casFrontendError("ticket_invalid"), nil
	}
	user, err := s.resolveOrProvisionCASUser(ctx, principal)
	if err != nil {
		return s.casFrontendError("user_resolve"), nil
	}
	if user.Status == 1 {
		return s.casFrontendError("disabled_user"), nil
	}
	if user.Status == 2 {
		return s.casFrontendError("locked_user"), nil
	}
	tokens, err := s.issueSession(ctx, user, ip, userAgent)
	if err != nil {
		return s.casFrontendError("token"), nil
	}
	code, err := randomHex(16)
	if err != nil {
		return s.casFrontendError("code"), nil
	}
	payload, err := json.Marshal(casExchangePayload{Login: *tokens, Redirect: redirect})
	if err != nil {
		return s.casFrontendError("code"), nil
	}
	if err := s.casStore.PutCode(ctx, code, payload, casExchangeTTL); err != nil {
		return s.casFrontendError("code"), nil
	}
	return s.casFrontendSuccess(code), nil
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

func (s *authService) casFrontendSuccess(code string) string {
	base := strings.TrimRight(s.cas.FrontendURL, "/")
	if base == "" {
		base = "https://starbyte.smbu.edu.cn"
	}
	return base + "/login/cas?code=" + url.QueryEscape(code)
}

func (s *authService) casFrontendError(reason string) string {
	base := strings.TrimRight(s.cas.FrontendURL, "/")
	if base == "" {
		base = "https://starbyte.smbu.edu.cn"
	}
	if reason == "" {
		reason = "unknown"
	}
	return base + "/login?cas_error=" + url.QueryEscape(reason)
}

func casServiceURL(serviceURL, state string) string {
	u := strings.TrimSpace(serviceURL)
	sep := "?"
	if strings.Contains(u, "?") {
		sep = "&"
	}
	return u + sep + "state=" + url.QueryEscape(state)
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
