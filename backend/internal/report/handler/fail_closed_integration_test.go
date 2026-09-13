package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	reportRepo "github.com/Yogdunana/StarByte/backend/internal/report/repo"
	reportService "github.com/Yogdunana/StarByte/backend/internal/report/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type queryLogger struct {
	mu      sync.Mutex
	queries []string
}

func (l *queryLogger) LogMode(logger.LogLevel) logger.Interface      { return l }
func (l *queryLogger) Info(context.Context, string, ...interface{})  {}
func (l *queryLogger) Warn(context.Context, string, ...interface{})  {}
func (l *queryLogger) Error(context.Context, string, ...interface{}) {}
func (l *queryLogger) Trace(_ context.Context, _ time.Time, sql func() (string, int64), _ error) {
	query, _ := sql()
	l.mu.Lock()
	l.queries = append(l.queries, query)
	l.mu.Unlock()
}

func (l *queryLogger) captured() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.queries...)
}

func TestListChainMissingDataScopeBuildsFailClosedQueries(t *testing.T) {
	log := &queryLogger{}
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=localhost user=unused dbname=unused sslmode=disable",
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
		Logger:               log,
	})
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(auth.ContextKeyUserID, uuid.New().String())
		c.Next()
	})
	h := New(reportService.New(reportRepo.NewReportRepo(db)))
	router.GET("/api/v1/reports", h.List)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil))

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	queries := log.captured()
	require.Len(t, queries, 2)
	for _, query := range queries {
		assert.Contains(t, strings.ToUpper(query), "WHERE 1 = 0")
	}
}
