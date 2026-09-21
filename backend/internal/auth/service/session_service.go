package service

import (
	"context"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/internal/auth/model"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *authService) ListSessions(ctx context.Context, keyword, userID string) (*dto.SessionListResponse, error) {
	var (
		sessions []model.Session
		err      error
	)
	if userID != "" {
		sessions, err = s.authRepo.ListSessionsByUser(ctx, userID)
	} else {
		sessions, err = s.authRepo.ListSessions(ctx)
	}
	if err != nil {
		return nil, err
	}
	views := s.toViews(ctx, sessions)
	annotateAnomalies(views)
	if keyword != "" {
		kw := strings.ToLower(strings.TrimSpace(keyword))
		filtered := make([]dto.SessionView, 0, len(views))
		for _, v := range views {
			if sessionMatches(v, kw) {
				filtered = append(filtered, v)
			}
		}
		views = filtered
	}
	if views == nil {
		views = []dto.SessionView{}
	}
	return &dto.SessionListResponse{List: views, Total: int64(len(views))}, nil
}

func (s *authService) GetUserSessions(ctx context.Context, userID string) (*dto.UserSessionsResponse, error) {
	sessions, err := s.authRepo.ListSessionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	views := s.toViews(ctx, sessions)
	annotateAnomalies(views)
	resp := &dto.UserSessionsResponse{
		UserID:   userID,
		Sessions: views,
	}
	if len(views) > 0 {
		resp.Username = views[0].Username
		resp.RealName = views[0].RealName
		resp.MultiDevice = views[0].MultiDevice
		resp.MultiIP = views[0].MultiIP
	} else {
		resp.Sessions = []dto.SessionView{}
		s.fillUserNames(ctx, userID, resp)
	}
	return resp, nil
}

func (s *authService) KickSession(ctx context.Context, tokenID string) error {
	sess, err := s.authRepo.GetSession(ctx, tokenID)
	if err != nil {
		return err
	}
	if sess == nil {
		return response.NewError(response.CodeSessionNotFound, "会话不存在或已失效")
	}
	ttl := time.Until(sess.ExpiresAt)
	if ttl < time.Second {
		ttl = time.Second
	}
	// 写侧 fail-closed：必须先成功写入黑名单，再删除 session / refresh token。
	// 黑名单写失败立即返回错误，且不继续删除该 session —— 保留重试依据。
	if err := s.authRepo.BlacklistToken(ctx, tokenID, ttl); err != nil {
		logger.Error("kick session: blacklist token failed",
			zap.String("token_id", tokenID), zap.String("user_id", sess.UserID), zap.Error(err))
		return response.NewError(response.CodeInternalError, "会话吊销未完全成功，请重试")
	}
	// 黑名单已写入成功，继续删除 session 与 refresh token；任一失败也须上报。
	var failed bool
	if err := s.authRepo.DeleteSession(ctx, tokenID); err != nil {
		logger.Error("kick session: delete session failed",
			zap.String("token_id", tokenID), zap.String("user_id", sess.UserID), zap.Error(err))
		failed = true
	}
	if err := s.authRepo.DeleteRefreshTokensByJTI(ctx, sess.UserID, tokenID); err != nil {
		logger.Error("kick session: delete refresh tokens by jti failed",
			zap.String("token_id", tokenID), zap.String("user_id", sess.UserID), zap.Error(err))
		failed = true
	}
	if failed {
		return response.NewError(response.CodeInternalError, "会话吊销未完全成功，请重试")
	}
	return nil
}

// revokeAllUserSessions 吊销该用户全部在线会话（blacklist access token + 删除 session）
// 以及全部 refresh token。
// 与 KickUserSessions 的区别：本方法在无会话时不报错，供改密等「强制下线」场景调用。
// 写侧 fail-closed：任一 Redis 写操作失败都必须上报错误（不静默吞掉），
// 以免“以为踢了其实没踢”。黑名单写入优先于删除，失败即跳过该会话的删除以保留重试依据。
func (s *authService) revokeAllUserSessions(ctx context.Context, userID string) error {
	sessions, err := s.authRepo.ListSessionsByUser(ctx, userID)
	if err != nil {
		return err
	}
	accessTTL := time.Duration(s.jwtConfig.AccessTokenExp) * time.Second
	var failed int
	for _, sess := range sessions {
		ttl := time.Until(sess.ExpiresAt)
		if ttl < time.Second {
			ttl = accessTTL
		}
		// 先成功写入黑名单，再删除 session；黑名单失败计入并跳过删除。
		if err := s.authRepo.BlacklistToken(ctx, sess.TokenID, ttl); err != nil {
			logger.Error("revoke all sessions: blacklist token failed",
				zap.String("token_id", sess.TokenID), zap.String("user_id", userID), zap.Error(err))
			failed++
			continue
		}
		if err := s.authRepo.DeleteSession(ctx, sess.TokenID); err != nil {
			logger.Error("revoke all sessions: delete session failed",
				zap.String("token_id", sess.TokenID), zap.String("user_id", userID), zap.Error(err))
			failed++
		}
	}
	if err := s.authRepo.DeleteRefreshTokensByUser(ctx, userID); err != nil {
		logger.Error("revoke all sessions: delete refresh tokens by user failed",
			zap.String("user_id", userID), zap.Error(err))
		failed++
	}
	if failed > 0 {
		return response.NewError(response.CodeInternalError, "会话吊销未完全成功，请重试")
	}
	return nil
}

// RevokeAllUserSessions 导出方法：吊销该用户全部在线会话（blacklist access token + 删除 session）
// 以及全部 refresh token。供其它包（如 user service 改密）在「强制下线」场景调用，
// 避免在 auth 包之外再复制一遍吊销逻辑。内部复用 revokeAllUserSessions。
func (s *authService) RevokeAllUserSessions(ctx context.Context, userID string) error {
	return s.revokeAllUserSessions(ctx, userID)
}

// KickUserSessions 主动踢下线：仅在该用户存在在线会话时吊销，否则返回 CodeSessionUserOffline。
func (s *authService) KickUserSessions(ctx context.Context, userID string) error {
	sessions, err := s.authRepo.ListSessionsByUser(ctx, userID)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return response.NewError(response.CodeSessionUserOffline, "该用户当前没有在线会话")
	}
	return s.revokeAllUserSessions(ctx, userID)
}

func (s *authService) toViews(ctx context.Context, sessions []model.Session) []dto.SessionView {
	users := map[string][2]string{}
	out := make([]dto.SessionView, 0, len(sessions))
	for _, sess := range sessions {
		browser, osName, device := ParseUserAgent(sess.UserAgent)
		username, realName := s.lookupUser(ctx, sess.UserID, users)
		out = append(out, dto.SessionView{
			TokenID:   sess.TokenID,
			UserID:    sess.UserID,
			Username:  username,
			RealName:  realName,
			IP:        sess.IP,
			UserAgent: sess.UserAgent,
			Browser:   browser,
			OS:        osName,
			Device:    device,
			LoginAt:   sess.LoginAt,
			ExpiresAt: sess.ExpiresAt,
		})
	}
	return out
}

func (s *authService) lookupUser(ctx context.Context, userID string, cache map[string][2]string) (string, string) {
	if userID == "" {
		return "", ""
	}
	if v, ok := cache[userID]; ok {
		return v[0], v[1]
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		cache[userID] = [2]string{}
		return "", ""
	}
	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil || user == nil {
		cache[userID] = [2]string{}
		return "", ""
	}
	cache[userID] = [2]string{user.Username, user.RealName}
	return user.Username, user.RealName
}

func (s *authService) fillUserNames(ctx context.Context, userID string, resp *dto.UserSessionsResponse) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return
	}
	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil || user == nil {
		return
	}
	resp.Username = user.Username
	resp.RealName = user.RealName
}

func sessionMatches(v dto.SessionView, kw string) bool {
	fields := []string{v.Username, v.RealName, v.IP, v.UserAgent, v.Browser, v.OS, v.Device, v.UserID}
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), kw) {
			return true
		}
	}
	return false
}

func annotateAnomalies(views []dto.SessionView) {
	type acc struct {
		n   int
		ips map[string]struct{}
	}
	byUser := map[string]*acc{}
	for i := range views {
		a := byUser[views[i].UserID]
		if a == nil {
			a = &acc{ips: map[string]struct{}{}}
			byUser[views[i].UserID] = a
		}
		a.n++
		if views[i].IP != "" {
			a.ips[views[i].IP] = struct{}{}
		}
	}
	for i := range views {
		a := byUser[views[i].UserID]
		views[i].MultiDevice = a.n > 1
		views[i].MultiIP = len(a.ips) > 1
	}
}
