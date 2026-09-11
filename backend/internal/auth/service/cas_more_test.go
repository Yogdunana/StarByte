package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCASStatusAndDisabledPaths(t *testing.T) {
	svc, _, _, _ := setupTestService()
	assert.False(t, svc.CASStatus().Enabled)

	_, err := svc.ExchangeCASCode(context.Background(), "x")
	require.Error(t, err)
	assert.Equal(t, response.CodeNotImplemented, err.(*response.AppError).Code)

	loc, err := svc.CompleteCASCallback(context.Background(), "ST", "st", "", "", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, loc, "cas_error=disabled")
}

func TestBuildCASLoginURL_Incomplete(t *testing.T) {
	svc := casTestService(nil, stubValidator{}, nil)
	svc.cas.ServerURL = ""
	_, err := svc.BuildCASLoginURL(context.Background(), "/", "http://10.0.0.8")
	require.Error(t, err)
}

func TestBuildCASLoginURL_NoOrigin(t *testing.T) {
	svc := casTestService(newMemCASStore(), stubValidator{}, nil)
	_, err := svc.BuildCASLoginURL(context.Background(), "/", "")
	require.Error(t, err)
}

func TestCompleteCASCallback_ErrorReasons(t *testing.T) {
	store := newMemCASStore()
	svc := casTestService(store, stubValidator{err: errors.New("bad ticket")}, nil)

	loc, err := svc.CompleteCASCallback(context.Background(), "", "", "", "", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, loc, "missing_ticket")

	loc, err = svc.CompleteCASCallback(context.Background(), "ST", "gone", "", "", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, loc, "state_expired")

	require.NoError(t, store.PutState(context.Background(), "st", `{"origin":"http://10.0.0.8","service":"http://10.0.0.8/api/v1/auth/cas/callback"}`, time.Minute))
	loc, err = svc.CompleteCASCallback(context.Background(), "ST", "st", "", "", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, loc, "ticket_invalid")
}

func TestCompleteCASCallback_DisabledUser(t *testing.T) {
	store := newMemCASStore()
	require.NoError(t, store.PutState(context.Background(), "st", `{"origin":"http://10.0.0.8","service":"http://10.0.0.8/cb"}`, time.Minute))
	users := &mockUserRepo{}
	users.On("GetByIdentity", mock.Anything, identityTypeCAS, "u1").Return(&model.User{ID: uuid.New(), Username: "u1", Status: 1}, nil)
	svc := casTestService(store, stubValidator{p: &CASPrincipal{User: "u1"}}, users)
	loc, err := svc.CompleteCASCallback(context.Background(), "ST", "st", "", "", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, loc, "disabled_user")
}

func TestCompleteCASCallback_LockedUser(t *testing.T) {
	store := newMemCASStore()
	require.NoError(t, store.PutState(context.Background(), "st", `{"origin":"http://10.0.0.8","service":"http://10.0.0.8/cb"}`, time.Minute))
	users := &mockUserRepo{}
	users.On("GetByIdentity", mock.Anything, identityTypeCAS, "u2").Return(&model.User{ID: uuid.New(), Username: "u2", Status: 2}, nil)
	svc := casTestService(store, stubValidator{p: &CASPrincipal{User: "u2"}}, users)
	loc, err := svc.CompleteCASCallback(context.Background(), "ST", "st", "", "", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, loc, "locked_user")
}

func TestCompleteCASCallback_DoesNotBindByUsername(t *testing.T) {
	store := newMemCASStore()
	require.NoError(t, store.PutState(context.Background(), "st", `{"origin":"http://10.0.0.8","service":"http://10.0.0.8/cb"}`, time.Minute))
	users := &mockUserRepo{}
	users.On("GetByIdentity", mock.Anything, identityTypeCAS, "alice").Return((*model.User)(nil), nil)
	users.On("Create", mock.Anything, (*gorm.DB)(nil), mock.AnythingOfType("*model.User")).Return(fmt.Errorf("duplicate username"))

	svc := casTestService(store, stubValidator{p: &CASPrincipal{User: "alice"}}, users)
	svc.identity = &stubIdentity{}
	loc, err := svc.CompleteCASCallback(context.Background(), "ST", "st", "9.9.9.9", "ua", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, loc, "cas_error=user_resolve")
	users.AssertNotCalled(t, "CreateIdentity", mock.Anything, mock.Anything)
}

func TestCompleteCASCallback_BindStudentNo(t *testing.T) {
	store := newMemCASStore()
	require.NoError(t, store.PutState(context.Background(), "st", `{"origin":"http://10.0.0.8","service":"http://10.0.0.8/cb"}`, time.Minute))
	userID := uuid.New()
	users := &mockUserRepo{}
	users.On("GetByIdentity", mock.Anything, identityTypeCAS, "cas-alice").Return((*model.User)(nil), nil).Once()
	users.On("GetByID", mock.Anything, userID).Return(&model.User{ID: userID, Username: "old"}, nil)
	users.On("GetByIdentity", mock.Anything, identityTypeCAS, "cas-alice").Return((*model.User)(nil), nil)
	users.On("CreateIdentity", mock.Anything, mock.AnythingOfType("*model.UserIdentity")).Return(nil)
	users.On("UpdateLastLogin", mock.Anything, userID, "1.2.3.4").Return(nil)

	svc := casTestService(store, stubValidator{p: &CASPrincipal{
		User:       "cas-alice",
		Attributes: map[string]string{"studentno": "20210001"},
	}}, users)
	svc.identity = &stubIdentity{byNo: map[string]uuid.UUID{"20210001": userID}}
	authRepo := svc.authRepo.(*mockAuthRepo)
	perm := svc.permCacheSvc.(*mockPermCache)
	authRepo.On("StoreRefreshToken", mock.Anything, "test-refresh-token-uuid", userID.String(), mock.AnythingOfType("string"), mock.Anything).Return(nil)
	authRepo.On("StoreSession", mock.Anything, userID.String(), mock.AnythingOfType("string"), "1.2.3.4", "ua", mock.Anything).Return(nil)
	perm.On("GetUserPermissionsAndSuperAdmin", mock.Anything, userID).Return([]string{}, false, nil)
	perm.On("GetUserRoleCodes", mock.Anything, userID).Return([]string{"member"}, nil)

	loc, err := svc.CompleteCASCallback(context.Background(), "ST", "st", "1.2.3.4", "ua", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, loc, "/login/cas?code=")
}

func TestCompleteCASCallback_NoAutoProvision(t *testing.T) {
	store := newMemCASStore()
	require.NoError(t, store.PutState(context.Background(), "st", `{"origin":"http://10.0.0.8","service":"http://10.0.0.8/cb"}`, time.Minute))
	users := &mockUserRepo{}
	users.On("GetByIdentity", mock.Anything, identityTypeCAS, "nobody").Return((*model.User)(nil), nil)
	svc := casTestService(store, stubValidator{p: &CASPrincipal{User: "nobody"}}, users)
	svc.cas.AllowAutoProvision = false
	svc.identity = &stubIdentity{}
	loc, err := svc.CompleteCASCallback(context.Background(), "ST", "st", "", "", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, loc, "/login/cas?code=")
	users.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)

	code := loc[len("http://10.0.0.8/login/cas?code="):]
	out, err := svc.ExchangeCASCode(context.Background(), code)
	require.NoError(t, err)
	assert.True(t, out.NeedsRegistration)
}

func TestRegisterWithCASToken_UsernameTakenKeepsToken(t *testing.T) {
	store := newMemCASStore()
	pending, _ := json.Marshal(casExchangePayload{
		Kind: casExchangeKindRegister,
		Pending: &casPendingIdentity{
			CASUser:   "20218888",
			StudentNo: "20218888",
			RealName:  "赵六",
		},
		Redirect: "/dashboard",
	})
	require.NoError(t, store.PutCode(context.Background(), "reg1", pending, time.Minute))
	users := &mockUserRepo{}
	users.On("GetByIdentity", mock.Anything, identityTypeCAS, "20218888").Return((*model.User)(nil), nil)
	users.On("GetByUsername", mock.Anything, "taken_user").Return(&model.User{Username: "taken_user"}, nil)
	svc := casTestService(store, stubValidator{}, users)

	_, err := svc.RegisterWithCASToken(context.Background(), &dto.CASRegisterRequest{
		Token:    "reg1",
		Username: "taken_user",
		Password: "Passw0rd!",
	}, "", "")
	require.Error(t, err)
	assert.Equal(t, response.CodeUserExists, err.(*response.AppError).Code)

	out, err := svc.RegisterWithCASToken(context.Background(), &dto.CASRegisterRequest{
		Token:    "reg1",
		Username: "ok_user",
		Password: "weak",
	}, "", "")
	require.Error(t, err)
	assert.Equal(t, response.CodePasswordTooWeak, err.(*response.AppError).Code)
	assert.Nil(t, out)
}

func TestWouldSetUsernameToStudentNo(t *testing.T) {
	assert.True(t, wouldSetUsernameToStudentNo("20210001", "20210001"))
	assert.True(t, wouldSetUsernameToStudentNo("20210001", ""))
	assert.False(t, wouldSetUsernameToStudentNo("alice", ""))
	assert.False(t, wouldSetUsernameToStudentNo("alice", "20210001"))
}

func TestExchangeCASCode_Errors(t *testing.T) {
	svc := casTestService(newMemCASStore(), stubValidator{}, nil)
	_, err := svc.ExchangeCASCode(context.Background(), "  ")
	require.Error(t, err)
	assert.Equal(t, response.CodeBadRequest, err.(*response.AppError).Code)

	_, err = svc.ExchangeCASCode(context.Background(), "missing")
	require.Error(t, err)
	assert.Equal(t, response.CodeTokenInvalid, err.(*response.AppError).Code)

	svc.casStore = nil
	_, err = svc.ExchangeCASCode(context.Background(), "x")
	require.Error(t, err)
	assert.Equal(t, response.CodeInternalError, err.(*response.AppError).Code)
}

func TestExchangeCASCode_InvalidJSON(t *testing.T) {
	store := newMemCASStore()
	require.NoError(t, store.PutCode(context.Background(), "bad", []byte("not-json"), time.Minute))
	svc := casTestService(store, stubValidator{}, nil)
	_, err := svc.ExchangeCASCode(context.Background(), "bad")
	require.Error(t, err)
	assert.Equal(t, response.CodeTokenInvalid, err.(*response.AppError).Code)
}

func TestParseCASStateAndFrontendHelpers(t *testing.T) {
	rec := parseCASState("", "http://10.0.0.8")
	assert.Equal(t, "/dashboard", rec.Redirect)
	assert.Equal(t, "http://10.0.0.8/api/v1/auth/cas/callback", rec.Service)

	rec = parseCASState("/tasks", "http://10.0.0.8")
	assert.Equal(t, "/tasks", rec.Redirect)

	rec = parseCASState(`{"redirect":"/ok","origin":"http://1.2.3.4","service":"http://1.2.3.4/cb"}`, "http://10.0.0.8")
	assert.Equal(t, "/ok", rec.Redirect)
	assert.Equal(t, "http://1.2.3.4", rec.Origin)
	assert.Equal(t, "http://1.2.3.4/cb", rec.Service)

	assert.Equal(t, "/login/cas?code=abc", casFrontendSuccess("", "abc"))
	assert.Equal(t, "/login?cas_error=unknown", casFrontendError("", ""))
}

func TestResolveCASURLs_FromServiceURL(t *testing.T) {
	service, origin, err := resolveCASURLs(&config.CASConfig{
		ServiceURL: "https://starbyte.smbu.edu.cn/api/v1/auth/cas/callback",
	}, "")
	require.NoError(t, err)
	assert.Equal(t, "https://starbyte.smbu.edu.cn", origin)
	assert.Equal(t, "https://starbyte.smbu.edu.cn/api/v1/auth/cas/callback", service)
}

func TestPickStudentNoAndDigits(t *testing.T) {
	assert.Equal(t, "2021", pickStudentNo(&CASPrincipal{Attributes: map[string]string{"uid": "2021"}}, "x"))
	assert.Equal(t, "20210001", pickStudentNo(&CASPrincipal{}, "20210001"))
	assert.Equal(t, "", pickStudentNo(&CASPrincipal{}, "alice"))
	assert.False(t, isMostlyDigits(""))
	assert.True(t, isMostlyDigits("123a"))
}

func TestParseCASJSON_Failure(t *testing.T) {
	_, err := parseCASJSON(`{"serviceResponse":{"authenticationFailure":{"code":"INVALID_TICKET"}}}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "INVALID_TICKET")

	_, err = parseCASJSON(`{"serviceResponse":{}}`)
	require.Error(t, err)
}

func TestParseCASXML_MissingUser(t *testing.T) {
	_, err := parseCASXML(`<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas"><cas:authenticationSuccess></cas:authenticationSuccess></cas:serviceResponse>`)
	require.Error(t, err)
}

func TestHTTPTicketValidator_JSONThenXML(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		body, _ := io.ReadAll(r.Body)
		_ = body
		if strings.Contains(r.URL.Path, "/p3/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
  <cas:authenticationSuccess><cas:user>s1</cas:user></cas:authenticationSuccess>
</cas:serviceResponse>`))
	}))
	defer srv.Close()

	v := NewHTTPTicketValidator(srv.URL, srv.Client())
	p, err := v.Validate(context.Background(), "http://app/cb", "ST-1")
	require.NoError(t, err)
	assert.Equal(t, "s1", p.User)
	assert.GreaterOrEqual(t, hits, 2)
}

func TestHTTPTicketValidator_JSONSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"serviceResponse":{"authenticationSuccess":{"user":"s2","attributes":{"name":["王","五"],"age":21}}}}`))
	}))
	defer srv.Close()

	v := NewHTTPTicketValidator(srv.URL, srv.Client())
	p, err := v.Validate(context.Background(), "http://app/cb", "ST-2")
	require.NoError(t, err)
	assert.Equal(t, "s2", p.User)
	assert.Equal(t, "王", p.Attributes["name"])
	assert.Equal(t, "21", p.Attributes["age"])
}

func TestHTTPTicketValidator_P3FailureDoesNotFallback(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"serviceResponse":{"authenticationFailure":{"code":"INVALID_TICKET"}}}`))
	}))
	defer srv.Close()

	v := NewHTTPTicketValidator(srv.URL, srv.Client())
	_, err := v.Validate(context.Background(), "http://app/cb", "ST-used")
	require.Error(t, err)
	assert.Equal(t, 1, hits)
}

func TestHTTPTicketValidator_EmptyURL(t *testing.T) {
	v := NewHTTPTicketValidator("", nil)
	_, err := v.Validate(context.Background(), "s", "t")
	require.Error(t, err)
}

func TestRoleAssigner_NilSafe(t *testing.T) {
	a := NewRoleAssigner(nil, nil, "")
	require.NoError(t, a.AssignDefault(context.Background(), uuid.New()))
	require.NoError(t, a.AssignDefault(context.Background(), uuid.Nil))
}

func TestSanitizeCASUsername_TooLong(t *testing.T) {
	assert.Equal(t, "", sanitizeCASUsername(strings.Repeat("a", 51)))
}

func TestCASFrontendSuccessWithOrigin(t *testing.T) {
	assert.Equal(t, "http://10.0.0.8/login/cas?code=z", casFrontendSuccess("http://10.0.0.8/x", "z"))
}

func TestExchangePayloadRedirectSanitize(t *testing.T) {
	store := newMemCASStore()
	payload, _ := json.Marshal(casExchangePayload{
		Login:    dto.LoginResponse{AccessToken: "a"},
		Redirect: "https://evil.com",
	})
	require.NoError(t, store.PutCode(context.Background(), "c2", payload, time.Minute))
	svc := casTestService(store, stubValidator{}, nil)
	out, err := svc.ExchangeCASCode(context.Background(), "c2")
	require.NoError(t, err)
	assert.Equal(t, "/dashboard", out.Redirect)
}
