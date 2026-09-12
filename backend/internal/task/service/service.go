package service

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/google/uuid"
	"gorm.io/gorm"

	filedto "github.com/Yogdunana/StarByte/backend/internal/file/dto"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/repo"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
)

type Notifier interface {
	Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error
}

type FileBridge interface {
	Upload(ctx context.Context, userID uuid.UUID, header *multipart.FileHeader, category string, isPublic bool) (*filedto.FileUploadResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*filedto.FileDetailResponse, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type ObjectDownloader interface {
	Download(ctx context.Context, objectName string) (io.ReadCloser, string, error)
}

type TaskService interface {
	AssignmentRoles(context.Context, string) ([]dto.Person, error)
	SetWorkflowEngine(*engine.FlowEngine)
	GetWorkflow(context.Context, uuid.UUID, uuid.UUID) (*dto.WorkflowResponse, error)
	ActWorkflow(context.Context, uuid.UUID, uuid.UUID, *dto.WorkflowActionRequest) (*dto.WorkflowResponse, error)
	RequestHandover(context.Context, uuid.UUID, uuid.UUID, *dto.HandoverRequest) (*dto.HandoverResponse, error)
	GetHandover(context.Context, uuid.UUID, uuid.UUID) (*dto.HandoverResponse, error)
	GetTransfer(context.Context, uuid.UUID, uuid.UUID) (*dto.HandoverResponse, error)
	DecideHandover(context.Context, uuid.UUID, uuid.UUID, *dto.HandoverDecision) (*dto.HandoverResponse, error)
	DecideTransfer(context.Context, uuid.UUID, uuid.UUID, *dto.HandoverDecision) (*dto.HandoverResponse, error)
	EscalateOverdueWorkflows(context.Context) (int, error)

	ProcessAttachmentDeletions(context.Context) (int, error)
	Candidates(context.Context, string) ([]dto.Person, error)
	Create(ctx context.Context, operator uuid.UUID, req *dto.CreateTaskRequest) (*dto.TaskResponse, error)
	List(ctx context.Context, viewer uuid.UUID, req *dto.ListTaskRequest, scope *rbacModel.DataScopeCondition) ([]*dto.TaskResponse, int64, error)
	Get(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.TaskResponse, error)
	Update(ctx context.Context, id, operator uuid.UUID, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error)
	Delete(ctx context.Context, id, operator uuid.UUID) error

	Assign(ctx context.Context, id, operator uuid.UUID, assigneeID string) (*dto.TaskResponse, error)
	Transfer(ctx context.Context, id, operator uuid.UUID, req *dto.TransferRequest) (*dto.TaskResponse, error)
	ChangeStatus(ctx context.Context, id, operator uuid.UUID, req *dto.StatusRequest) (*dto.TaskResponse, error)
	Urge(ctx context.Context, id, operator uuid.UUID, message string) error
	ListLogs(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.LogResponse, error)

	ListComments(ctx context.Context, viewer, taskID uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.CommentResponse, error)
	AddComment(ctx context.Context, taskID, operator uuid.UUID, req *dto.CommentRequest) (*dto.CommentResponse, error)
	UpdateComment(ctx context.Context, taskID, commentID, operator uuid.UUID, content string) (*dto.CommentResponse, error)
	DeleteComment(ctx context.Context, taskID, commentID, operator uuid.UUID) error

	UploadAttachment(ctx context.Context, taskID, operator uuid.UUID, header *multipart.FileHeader) (*dto.AttachmentResponse, error)
	ListAttachments(ctx context.Context, viewer, taskID uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.AttachmentResponse, error)
	DownloadAttachment(ctx context.Context, viewer, taskID, attachID uuid.UUID, scope *rbacModel.DataScopeCondition) (io.ReadCloser, string, string, error)
	DeleteAttachment(ctx context.Context, taskID, attachID, operator uuid.UUID) error

	ListMy(ctx context.Context, userID uuid.UUID, kind string, req *dto.MyTaskRequest) ([]*dto.TaskResponse, int64, error)
	Stats(ctx context.Context, viewer uuid.UUID, req *dto.StatsRequest, scope *rbacModel.DataScopeCondition) (*dto.StatsResponse, error)
	RemindDueAndOverdue(ctx context.Context) (int, error)
}

type taskService struct {
	transfers   repo.TransferRepo
	assignments repo.AssignmentRepo
	engine      *engine.FlowEngine
	flow        TaskWorkflowRuntime
	cleanup     repo.CleanupRepo
	db          *gorm.DB
	afterCommit *[]func()
	tasks       repo.TaskRepo
	logs        repo.LogRepo
	comments    repo.CommentRepo
	files       repo.AttachmentRepo
	notify      Notifier
	bridge      FileBridge
	store       ObjectDownloader
}

func NewTaskService(
	tasks repo.TaskRepo,
	logs repo.LogRepo,
	comments repo.CommentRepo,
	files repo.AttachmentRepo,
	notify Notifier,
	bridge FileBridge,
	store ObjectDownloader,
	databases ...*gorm.DB,
) TaskService {
	var db *gorm.DB
	if len(databases) > 0 {
		db = databases[0]
	}
	svc := &taskService{db: db, cleanup: repo.NewCleanupRepo(db),
		tasks: tasks, logs: logs, comments: comments, files: files,
		notify: notify, bridge: bridge, store: store,
	}
	if db != nil {
		svc.transfers = repo.NewTransferRepo(db)
		svc.assignments = repo.NewAssignmentRepo(db)
	}
	return svc
}

func (s *taskService) SetWorkflowEngine(flow *engine.FlowEngine) { s.engine = flow; s.flow = flow }
