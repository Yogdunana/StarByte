package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/internal/backup/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stub struct {
	created  *dto.Record
	preview  *dto.Preview
	drill    *dto.DrillResult
	waited   *dto.Record
	list     []dto.Record
	err      error
	lastReq  *dto.DrillRequest
	lastConf string
}

func (s *stub) List(context.Context, *dto.ListRequest) ([]dto.Record, int64, error) {
	return s.list, int64(len(s.list)), s.err
}
func (s *stub) Get(context.Context, uuid.UUID) (*dto.Record, error) { return s.created, s.err }
func (s *stub) Create(context.Context, uuid.UUID, *dto.CreateRequest) (*dto.Record, error) {
	return s.created, s.err
}
func (s *stub) Delete(context.Context, uuid.UUID) error { return s.err }
func (s *stub) Restore(_ context.Context, _ uuid.UUID, _ uuid.UUID, req *dto.RestoreRequest) (*dto.Record, error) {
	if req != nil {
		s.lastConf = req.Confirmation
	}
	return s.created, s.err
}
func (s *stub) DrillRestore(_ context.Context, _ uuid.UUID, _ uuid.UUID, req *dto.DrillRequest) (*dto.DrillResult, error) {
	s.lastReq = req
	return s.drill, s.err
}
func (s *stub) Wait(context.Context, uuid.UUID) (*dto.Record, error) { return s.waited, s.err }
func (s *stub) GetPolicy(context.Context) (*dto.Policy, error)       { return &dto.Policy{}, s.err }
func (s *stub) UpdatePolicy(context.Context, uuid.UUID, *dto.UpdatePolicyRequest) (*dto.Policy, error) {
	return &dto.Policy{}, s.err
}
func (s *stub) Storage(context.Context) (*dto.StorageStats, error) { return &dto.StorageStats{}, s.err }
func (s *stub) Preview(context.Context, uuid.UUID) (*dto.Preview, error) {
	return s.preview, s.err
}
func (s *stub) RunScheduled(context.Context, string, func(string)) error   { return s.err }
func (s *stub) CleanupExpired(context.Context, string, func(string)) error { return s.err }
func (s *stub) SyncSchedule(context.Context) error                         { return s.err }

func TestExecuteHelpAndUnknown(t *testing.T) {
	var out, errb bytes.Buffer
	assert.Equal(t, 0, Execute(context.Background(), &stub{}, nil, &out, &errb))
	assert.Contains(t, out.String(), "drill")
	out.Reset()
	assert.Equal(t, 2, Execute(context.Background(), &stub{}, []string{"nope"}, &out, &errb))
}

func TestExecuteCreatePreviewDrill(t *testing.T) {
	id := uuid.New()
	svc := &stub{
		created: &dto.Record{ID: id.String(), Status: model.StatusSuccess, Filename: "a.dump.gz"},
		waited:  &dto.Record{ID: id.String(), Status: model.StatusSuccess, Filename: "a.dump.gz"},
		preview: &dto.Preview{ID: id.String(), Ready: true, ChecksumOK: true, DecryptOK: true, GzipOK: true, TOCValid: true},
		drill:   &dto.DrillResult{ID: id.String(), Restored: true, TargetHost: "postgres", TargetDBName: "starbyte_drill"},
	}
	var out, errb bytes.Buffer
	assert.Equal(t, 0, Execute(context.Background(), svc, []string{"create"}, &out, &errb))
	assert.Contains(t, out.String(), id.String())

	out.Reset()
	assert.Equal(t, 0, Execute(context.Background(), svc, []string{"preview", id.String()}, &out, &errb))
	assert.Contains(t, out.String(), "ready=true")

	out.Reset()
	code := Execute(context.Background(), svc, []string{"drill", id.String(), "--dbname", "starbyte_drill", "--confirm", "DRILL"}, &out, &errb)
	require.Equal(t, 0, code)
	require.NotNil(t, svc.lastReq)
	assert.Equal(t, "DRILL", svc.lastReq.Confirmation)
	assert.Equal(t, "starbyte_drill", svc.lastReq.TargetDBName)
	assert.True(t, strings.Contains(out.String(), "restored=true"))
}
