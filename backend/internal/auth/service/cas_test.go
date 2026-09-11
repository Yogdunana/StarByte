package service

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type memCASStore struct {
	mu    sync.Mutex
	state map[string]string
	codes map[string][]byte
}

func newMemCASStore() *memCASStore {
	return &memCASStore{state: map[string]string{}, codes: map[string][]byte{}}
}

func (m *memCASStore) PutState(_ context.Context, state, redirect string, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state[state] = redirect
	return nil
}

func (m *memCASStore) TakeState(_ context.Context, state string) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.state[state]
	if ok {
		delete(m.state, state)
	}
	return v, ok, nil
}

func (m *memCASStore) PutCode(_ context.Context, code string, payload []byte, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[code] = payload
	return nil
}

func (m *memCASStore) TakeCode(_ context.Context, code string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.codes[code]
	if !ok {
		return nil, goredis.Nil
	}
	delete(m.codes, code)
	return v, nil
}

type stubValidator struct {
	p   *CASPrincipal
	err error
}

func (s stubValidator) Validate(context.Context, string, string) (*CASPrincipal, error) {
	return s.p, s.err
}

type stubIdentity struct {
	byNo    map[string]uuid.UUID
	ensured []ensuredProfile
}

func (s *stubIdentity) GetByUserID(context.Context, uuid.UUID) (*MemberIdentity, error) {
	return nil, nil
}

func (s *stubIdentity) GetUserIDByStudentNo(_ context.Context, studentNo string) (uuid.UUID, error) {
	if s == nil {
		return uuid.Nil, nil
	}
	return s.byNo[studentNo], nil
}

type ensuredProfile struct {
	userID    uuid.UUID
	studentNo string
	realName  string
}

func (s *stubIdentity) EnsureStudentNo(_ context.Context, userID uuid.UUID, studentNo, realName string) error {
	if s == nil {
		return nil
	}
	s.ensured = append(s.ensured, ensuredProfile{userID: userID, studentNo: studentNo, realName: realName})
	return nil
}

func casTestService(store *memCASStore, validator TicketValidator, users *mockUserRepo) *authService {
	svc, _, authRepo, perm := setupTestService()
	if users != nil {
		svc.userRepo = users
	}
	svc.cas = &config.CASConfig{
		Enabled:            true,
		ServerURL:          "https://authserver.smbu.edu.cn/authserver",
		AllowAutoProvision: true,
		DefaultRole:        "member",
	}
	svc.casStore = store
	svc.casValidator = validator
	_ = authRepo
	_ = perm
	return svc
}

func TestSanitizeRedirect(t *testing.T) {
	assert.Equal(t, "/dashboard", sanitizeRedirect(""))
	assert.Equal(t, "/dashboard", sanitizeRedirect("https://evil.com"))
	assert.Equal(t, "/dashboard", sanitizeRedirect("//evil.com"))
	assert.Equal(t, "/tasks", sanitizeRedirect("/tasks"))
}

func TestSanitizeCASUsername(t *testing.T) {
	assert.Equal(t, "20210001", sanitizeCASUsername(" 20210001 "))
	assert.Equal(t, "", sanitizeCASUsername("a b"))
	assert.Equal(t, "", sanitizeCASUsername("../../../etc"))
}

func TestSanitizeChosenUsername(t *testing.T) {
	assert.Equal(t, "alice_wang", sanitizeChosenUsername(" alice_wang "))
	assert.Equal(t, "", sanitizeChosenUsername("ab"))
	assert.Equal(t, "", sanitizeChosenUsername("alice wang"))
	assert.Equal(t, "", sanitizeChosenUsername("alice-wang"))
}

func TestParseCASXML_Success(t *testing.T) {
	raw := `<?xml version="1.0"?>
<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
  <cas:authenticationSuccess>
    <cas:user>20210001</cas:user>
    <cas:attributes>
      <cas:name>张三</cas:name>
      <cas:mail>zhang@smbu.edu.cn</cas:mail>
    </cas:attributes>
  </cas:authenticationSuccess>
</cas:serviceResponse>`
	p, err := parseCASXML(raw)
	require.NoError(t, err)
	assert.Equal(t, "20210001", p.User)
	assert.Equal(t, "张三", p.Attributes["name"])
	assert.Equal(t, "zhang@smbu.edu.cn", p.Attributes["mail"])
}

func TestParseCASXML_Failure(t *testing.T) {
	raw := `<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
  <cas:authenticationFailure code="INVALID_TICKET">Ticket not recognized</cas:authenticationFailure>
</cas:serviceResponse>`
	_, err := parseCASXML(raw)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Ticket not recognized")
}

func TestParseCASJSON_Success(t *testing.T) {
	raw := `{"serviceResponse":{"authenticationSuccess":{"user":"20210001","attributes":{"cn":"李四","email":"li@smbu.edu.cn"}}}}`
	p, err := parseCASJSON(raw)
	require.NoError(t, err)
	assert.Equal(t, "20210001", p.User)
	assert.Equal(t, "李四", pickRealName(p))
	assert.Equal(t, "li@smbu.edu.cn", pickAttr(p, "email"))
}

func TestBuildCASLoginURL(t *testing.T) {
	store := newMemCASStore()
	svc := casTestService(store, stubValidator{}, nil)
	start, err := svc.BuildCASLoginURL(context.Background(), "/tasks", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, start.Location, "https://authserver.smbu.edu.cn/authserver/login?service=")
	assert.Contains(t, start.Location, "10.0.0.8")
	assert.Equal(t, "http://10.0.0.8/api/v1/auth/cas/callback", start.Service)
	assert.Len(t, store.state, 1)
}

func TestBuildCASLoginURL_Disabled(t *testing.T) {
	svc, _, _, _ := setupTestService()
	_, err := svc.BuildCASLoginURL(context.Background(), "/", "http://10.0.0.8")
	require.Error(t, err)
	assert.Equal(t, response.CodeNotImplemented, err.(*response.AppError).Code)
}

func TestCompleteCASCallback_ExistingUser(t *testing.T) {
	store := newMemCASStore()
	require.NoError(t, store.PutState(context.Background(), "st1", `{"redirect":"/dashboard","origin":"http://10.0.0.8","service":"http://10.0.0.8/api/v1/auth/cas/callback"}`, time.Minute))
	userID := uuid.New()
	users := &mockUserRepo{}
	users.On("GetByIdentity", mock.Anything, identityTypeCAS, "20210001").Return(&model.User{
		ID: userID, Username: "20210001", RealName: "张三", Status: 0,
	}, nil)
	users.On("UpdateLastLogin", mock.Anything, userID, "1.1.1.1").Return(nil)

	svc := casTestService(store, stubValidator{p: &CASPrincipal{User: "20210001", Attributes: map[string]string{"name": "张三"}}}, users)
	authRepo := svc.authRepo.(*mockAuthRepo)
	perm := svc.permCacheSvc.(*mockPermCache)
	authRepo.On("StoreRefreshToken", mock.Anything, "test-refresh-token-uuid", userID.String(), mock.AnythingOfType("string"), mock.Anything).Return(nil)
	authRepo.On("StoreSession", mock.Anything, userID.String(), mock.AnythingOfType("string"), "1.1.1.1", "ua", mock.Anything).Return(nil)
	perm.On("GetUserPermissionsAndSuperAdmin", mock.Anything, userID).Return([]string{"member:read"}, false, nil)
	perm.On("GetUserRoleCodes", mock.Anything, userID).Return([]string{"member"}, nil)

	loc, err := svc.CompleteCASCallback(context.Background(), "ST-1", "st1", "1.1.1.1", "ua", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, loc, "http://10.0.0.8/login/cas?code=")
}

func TestCompleteCASCallback_UnknownNeedsRegistration(t *testing.T) {
	store := newMemCASStore()
	require.NoError(t, store.PutState(context.Background(), "st2", `{"redirect":"/dashboard","origin":"http://10.0.0.8","service":"http://10.0.0.8/api/v1/auth/cas/callback"}`, time.Minute))
	users := &mockUserRepo{}
	users.On("GetByIdentity", mock.Anything, identityTypeCAS, "20219999").Return((*model.User)(nil), nil)

	svc := casTestService(store, stubValidator{p: &CASPrincipal{
		User:       "20219999",
		Attributes: map[string]string{"name": "王五", "mail": "wang@smbu.edu.cn"},
	}}, users)
	svc.identity = &stubIdentity{}

	loc, err := svc.CompleteCASCallback(context.Background(), "ST-2", "st2", "2.2.2.2", "ua", "http://10.0.0.8")
	require.NoError(t, err)
	assert.Contains(t, loc, "/login/cas?code=")
	users.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)

	code := strings.TrimPrefix(loc, "http://10.0.0.8/login/cas?code=")
	out, err := svc.ExchangeCASCode(context.Background(), code)
	require.NoError(t, err)
	assert.True(t, out.NeedsRegistration)
	assert.Equal(t, "20219999", out.StudentNo)
	assert.Equal(t, "王五", out.RealName)
	assert.Equal(t, "wang@smbu.edu.cn", out.Email)
	assert.NotEmpty(t, out.RegistrationToken)
	assert.Empty(t, out.AccessToken)
}

func TestRegisterWithCASToken_CreatesUserAndBinds(t *testing.T) {
	store := newMemCASStore()
	require.NoError(t, store.PutState(context.Background(), "st2", `{"redirect":"/tasks","origin":"http://10.0.0.8","service":"http://10.0.0.8/api/v1/auth/cas/callback"}`, time.Minute))
	users := &mockUserRepo{}
	users.On("GetByIdentity", mock.Anything, identityTypeCAS, "20219999").Return((*model.User)(nil), nil)
	users.On("GetByUsername", mock.Anything, "alice_wang").Return((*model.User)(nil), nil)
	users.On("Create", mock.Anything, (*gorm.DB)(nil), mock.AnythingOfType("*model.User")).Return(nil).Run(func(args mock.Arguments) {
		u := args.Get(2).(*model.User)
		assert.Equal(t, "alice_wang", u.Username)
		assert.Equal(t, "王五", u.RealName)
		assert.NotEqual(t, "20219999", u.Username)
	})
	users.On("CreateIdentity", mock.Anything, mock.AnythingOfType("*model.UserIdentity")).Return(nil).Run(func(args mock.Arguments) {
		ident := args.Get(1).(*model.UserIdentity)
		assert.Equal(t, identityTypeCAS, ident.IdentityType)
		assert.Equal(t, "20219999", ident.IdentityValue)
	})
	users.On("UpdateLastLogin", mock.Anything, mock.Anything, "9.9.9.9").Return(nil)

	ident := &stubIdentity{}
	svc := casTestService(store, stubValidator{p: &CASPrincipal{
		User:       "20219999",
		Attributes: map[string]string{"name": "王五"},
	}}, users)
	svc.identity = ident
	authRepo := svc.authRepo.(*mockAuthRepo)
	perm := svc.permCacheSvc.(*mockPermCache)
	authRepo.On("StoreRefreshToken", mock.Anything, "test-refresh-token-uuid", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(nil)
	authRepo.On("StoreSession", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "9.9.9.9", "ua", mock.Anything).Return(nil)
	perm.On("GetUserPermissionsAndSuperAdmin", mock.Anything, mock.Anything).Return([]string{}, false, nil)
	perm.On("GetUserRoleCodes", mock.Anything, mock.Anything).Return([]string{"member"}, nil)

	loc, err := svc.CompleteCASCallback(context.Background(), "ST-2", "st2", "2.2.2.2", "ua", "http://10.0.0.8")
	require.NoError(t, err)
	code := strings.TrimPrefix(loc, "http://10.0.0.8/login/cas?code=")
	ex, err := svc.ExchangeCASCode(context.Background(), code)
	require.NoError(t, err)
	require.True(t, ex.NeedsRegistration)

	out, err := svc.RegisterWithCASToken(context.Background(), &dto.CASRegisterRequest{
		Token:    ex.RegistrationToken,
		Username: "alice_wang",
		Password: "Passw0rd!",
		RealName: "王五",
	}, "9.9.9.9", "ua")
	require.NoError(t, err)
	assert.NotEmpty(t, out.AccessToken)
	assert.Equal(t, "/tasks", out.Redirect)
	require.Len(t, ident.ensured, 1)
	assert.Equal(t, "20219999", ident.ensured[0].studentNo)
	assert.Equal(t, "王五", ident.ensured[0].realName)
	users.AssertExpectations(t)
}

func TestExchangeCASCode(t *testing.T) {
	store := newMemCASStore()
	payload, _ := json.Marshal(casExchangePayload{
		Login:    dto.LoginResponse{AccessToken: "a", RefreshToken: "r"},
		Redirect: "/tasks",
	})
	require.NoError(t, store.PutCode(context.Background(), "c1", payload, time.Minute))
	svc := casTestService(store, stubValidator{}, nil)
	out, err := svc.ExchangeCASCode(context.Background(), "c1")
	require.NoError(t, err)
	assert.Equal(t, "a", out.AccessToken)
	assert.Equal(t, "/tasks", out.Redirect)
	_, err = svc.ExchangeCASCode(context.Background(), "c1")
	require.Error(t, err)
}

func TestResolveCASURLs_FromIP(t *testing.T) {
	service, origin, err := resolveCASURLs(&config.CASConfig{}, "http://10.0.0.8")
	require.NoError(t, err)
	assert.Equal(t, "http://10.0.0.8", origin)
	assert.Equal(t, "http://10.0.0.8/api/v1/auth/cas/callback", service)
}

func TestSanitizePublicOrigin(t *testing.T) {
	assert.Equal(t, "http://10.0.0.8", sanitizePublicOrigin("http://10.0.0.8/login"))
	assert.Equal(t, "", sanitizePublicOrigin("javascript:alert(1)"))
	assert.Equal(t, "", sanitizePublicOrigin("//evil.example"))
}
