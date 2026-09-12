package service

import (
	"context"
	"errors"
	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/Yogdunana/StarByte/backend/internal/notification/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

type eventTemplateRepo struct {
	repo.NotificationTemplateRepo
	template *model.NotificationTemplate
	err      error
}

func (r eventTemplateRepo) GetByCode(context.Context, string) (*model.NotificationTemplate, error) {
	return r.template, r.err
}

type captureEventChannel struct{ messages []*NotificationMessage }

func (c *captureEventChannel) Type() string      { return "in_app" }
func (c *captureEventChannel) IsAvailable() bool { return true }
func (c *captureEventChannel) Send(_ context.Context, m *NotificationMessage) error {
	c.messages = append(c.messages, m)
	return nil
}
func TestWorkflowNotificationFallbackHasContextAndAction(t *testing.T) {
	_ = testutil.NewEngine()
	for _, mode := range []string{"missing", "nil", "invalid", "custom"} {
		t.Run(mode, func(t *testing.T) {
			r := eventTemplateRepo{}
			switch mode {
			case "missing":
				r.err = errors.New("missing template")
			case "invalid":
				r.template = &model.NotificationTemplate{TitleTemplate: "{{", Channels: `["in_app"]`}
			case "custom":
				r.template = &model.NotificationTemplate{TitleTemplate: "审核 {{.applicant_name}}", BodyTemplate: "{{.node_name}}", Channels: `["in_app"]`}
			}
			ch := &captureEventChannel{}
			registry := NewChannelRegistry()
			registry.Register(ch)
			listener := NewEventListener(r, NewTemplateEngine(r), registry)
			event := events.TaskCreatedEvent{InstanceID: uuid.New(), TaskID: uuid.New(), AssigneeID: uuid.New(), NodeName: "干事审批", TaskType: "approval", BusinessType: "member_application", BusinessKey: uuid.NewString(), ApplicantName: "张三"}
			require.NoError(t, listener.onTaskCreated(context.Background(), event))
			require.Len(t, ch.messages, 1)
			got := ch.messages[0]
			require.Equal(t, event.AssigneeID, got.UserID)
			require.Equal(t, "member", got.Category)
			require.Equal(t, "/workflow/todo?task_id="+event.TaskID.String(), got.ActionURL)
			if mode == "custom" {
				require.Equal(t, "审核 张三", got.Title)
			} else {
				require.Contains(t, got.Title, "入会申请")
				require.Contains(t, got.Content, "张三")
				require.Contains(t, got.Content, event.BusinessKey)
			}
		})
	}
}
