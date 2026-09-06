package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	authmodel "github.com/Yogdunana/StarByte/backend/internal/auth/model"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func sampleSession(userID, tokenID, ip, ua string) authmodel.Session {
	return authmodel.Session{
		UserID:    userID,
		TokenID:   tokenID,
		IP:        ip,
		UserAgent: ua,
		LoginAt:   time.Now().Add(-time.Hour),
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestListSessions_AnomalyAndKeyword(t *testing.T) {
	svc, userRepo, authRepo, _ := setupTestService()
	ctx := context.Background()
	uid := uuid.New()
	chrome := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0.0.0 Safari/537.36"
	sess := []authmodel.Session{
		sampleSession(uid.String(), "j1", "1.1.1.1", chrome),
		sampleSession(uid.String(), "j2", "2.2.2.2", chrome),
	}
	authRepo.On("ListSessions", ctx).Return(sess, nil)
	userRepo.On("GetByID", ctx, uid).Return(&model.User{ID: uid, Username: "alice", RealName: "爱丽丝"}, nil)

	out, err := svc.ListSessions(ctx, "", "")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), out.Total)
	assert.True(t, out.List[0].MultiDevice)
	assert.True(t, out.List[0].MultiIP)
	assert.Equal(t, "Chrome", out.List[0].Browser)
	assert.Equal(t, "alice", out.List[0].Username)

	filtered, err := svc.ListSessions(ctx, "爱丽丝", "")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), filtered.Total)

	none, err := svc.ListSessions(ctx, "nobody", "")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), none.Total)
	assert.NotNil(t, none.List)
}

func TestListSessions_FilterByUser(t *testing.T) {
	svc, userRepo, authRepo, _ := setupTestService()
	ctx := context.Background()
	uid := uuid.New()
	authRepo.On("ListSessionsByUser", ctx, uid.String()).Return([]authmodel.Session{
		sampleSession(uid.String(), "j1", "1.1.1.1", "curl/8.0"),
	}, nil)
	userRepo.On("GetByID", ctx, uid).Return(&model.User{ID: uid, Username: "bob"}, nil)

	out, err := svc.ListSessions(ctx, "", uid.String())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), out.Total)
	assert.False(t, out.List[0].MultiDevice)
	assert.Equal(t, "bob", out.List[0].Username)
}

func TestGetUserSessions_EmptyFillsProfile(t *testing.T) {
	svc, userRepo, authRepo, _ := setupTestService()
	ctx := context.Background()
	uid := uuid.New()
	authRepo.On("ListSessionsByUser", ctx, uid.String()).Return([]authmodel.Session{}, nil)
	userRepo.On("GetByID", ctx, uid).Return(&model.User{ID: uid, Username: "carol", RealName: "李四"}, nil)

	out, err := svc.GetUserSessions(ctx, uid.String())
	assert.NoError(t, err)
	assert.Equal(t, "carol", out.Username)
	assert.Equal(t, "李四", out.RealName)
	assert.Empty(t, out.Sessions)
}

func TestKickSession_NotFound(t *testing.T) {
	svc, _, authRepo, _ := setupTestService()
	ctx := context.Background()
	authRepo.On("GetSession", ctx, "missing").Return(nil, nil)

	err := svc.KickSession(ctx, "missing")
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.CodeSessionNotFound, appErr.Code)
}

func TestKickSession_Success(t *testing.T) {
	svc, _, authRepo, _ := setupTestService()
	ctx := context.Background()
	sess := sampleSession("user-1", "jti-1", "1.1.1.1", "ua")
	authRepo.On("GetSession", ctx, "jti-1").Return(&sess, nil)
	authRepo.On("BlacklistToken", ctx, "jti-1", mock.Anything).Return(nil)
	authRepo.On("DeleteSession", ctx, "jti-1").Return(nil)
	authRepo.On("DeleteRefreshTokensByJTI", ctx, "user-1", "jti-1").Return(nil)

	assert.NoError(t, svc.KickSession(ctx, "jti-1"))
}

func TestKickUserSessions_Offline(t *testing.T) {
	svc, _, authRepo, _ := setupTestService()
	ctx := context.Background()
	authRepo.On("ListSessionsByUser", ctx, "u1").Return([]authmodel.Session{}, nil)

	err := svc.KickUserSessions(ctx, "u1")
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.CodeSessionUserOffline, appErr.Code)
}

func TestKickUserSessions_Success(t *testing.T) {
	svc, _, authRepo, _ := setupTestService()
	ctx := context.Background()
	sess := []authmodel.Session{
		sampleSession("u1", "a", "1.1.1.1", "ua"),
		sampleSession("u1", "b", "1.1.1.1", "ua"),
	}
	authRepo.On("ListSessionsByUser", ctx, "u1").Return(sess, nil)
	authRepo.On("BlacklistToken", ctx, "a", mock.Anything).Return(nil)
	authRepo.On("BlacklistToken", ctx, "b", mock.Anything).Return(nil)
	authRepo.On("DeleteSession", ctx, "a").Return(nil)
	authRepo.On("DeleteSession", ctx, "b").Return(nil)
	authRepo.On("DeleteRefreshTokensByUser", ctx, "u1").Return(nil)

	assert.NoError(t, svc.KickUserSessions(ctx, "u1"))
}

func TestSessionMatchesAndAnnotate(t *testing.T) {
	views := []dto.SessionView{
		{UserID: "u1", IP: "1.1.1.1", Browser: "Chrome", Username: "alice"},
		{UserID: "u1", IP: "1.1.1.1", Browser: "Safari", Username: "alice"},
		{UserID: "u2", IP: "9.9.9.9", Browser: "Firefox", Username: "bob"},
	}
	annotateAnomalies(views)
	assert.True(t, views[0].MultiDevice)
	assert.False(t, views[0].MultiIP)
	assert.False(t, views[2].MultiDevice)
	assert.True(t, sessionMatches(views[0], "chrome"))
	assert.False(t, sessionMatches(views[0], "zzz"))
}
