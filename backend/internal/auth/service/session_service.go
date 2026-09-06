package service

import (
	"context"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/internal/auth/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
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
	_ = s.authRepo.BlacklistToken(ctx, tokenID, ttl)
	_ = s.authRepo.DeleteSession(ctx, tokenID)
	_ = s.authRepo.DeleteRefreshTokensByJTI(ctx, sess.UserID, tokenID)
	return nil
}

func (s *authService) KickUserSessions(ctx context.Context, userID string) error {
	sessions, err := s.authRepo.ListSessionsByUser(ctx, userID)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return response.NewError(response.CodeSessionUserOffline, "该用户当前没有在线会话")
	}
	accessTTL := time.Duration(s.jwtConfig.AccessTokenExp) * time.Second
	for _, sess := range sessions {
		ttl := time.Until(sess.ExpiresAt)
		if ttl < time.Second {
			ttl = accessTTL
		}
		_ = s.authRepo.BlacklistToken(ctx, sess.TokenID, ttl)
		_ = s.authRepo.DeleteSession(ctx, sess.TokenID)
	}
	_ = s.authRepo.DeleteRefreshTokensByUser(ctx, userID)
	return nil
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
