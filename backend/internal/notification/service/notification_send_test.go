package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubNotifRepo struct {
	emails map[uuid.UUID]string
}

func (s *stubNotifRepo) Create(context.Context, *model.Notification) error { return nil }
func (s *stubNotifRepo) BatchCreate(context.Context, []*model.Notification) error {
	return nil
}
func (s *stubNotifRepo) ListByUser(context.Context, uuid.UUID, int, int, string, bool) ([]*model.Notification, int64, error) {
	return nil, 0, nil
}
func (s *stubNotifRepo) GetByID(context.Context, uuid.UUID) (*model.Notification, error) {
	return nil, nil
}
func (s *stubNotifRepo) MarkAsRead(context.Context, uuid.UUID, []uuid.UUID) error { return nil }
func (s *stubNotifRepo) MarkAllAsRead(context.Context, uuid.UUID, string) error   { return nil }
func (s *stubNotifRepo) Delete(context.Context, uuid.UUID, uuid.UUID) error       { return nil }
func (s *stubNotifRepo) GetUnreadCount(context.Context, uuid.UUID) (int64, error) { return 0, nil }
func (s *stubNotifRepo) EmailsByUserIDs(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	out := map[uuid.UUID]string{}
	for _, id := range ids {
		if email := s.emails[id]; email != "" {
			out[id] = email
		}
	}
	return out, nil
}

type stubTplRepo struct{ tpl *model.NotificationTemplate }

func (s stubTplRepo) Create(context.Context, *model.NotificationTemplate) error { return nil }
func (s stubTplRepo) GetByID(context.Context, uuid.UUID) (*model.NotificationTemplate, error) {
	return s.tpl, nil
}
func (s stubTplRepo) GetByCode(context.Context, string) (*model.NotificationTemplate, error) {
	return s.tpl, nil
}
func (s stubTplRepo) List(context.Context, int, int, string) ([]*model.NotificationTemplate, int64, error) {
	return nil, 0, nil
}
func (s stubTplRepo) Update(context.Context, *model.NotificationTemplate) error { return nil }
func (s stubTplRepo) Delete(context.Context, uuid.UUID) error                   { return nil }

type captureEmail struct{ last *NotificationMessage }

func (c *captureEmail) Type() string      { return "email" }
func (c *captureEmail) IsAvailable() bool { return true }
func (c *captureEmail) Send(_ context.Context, msg *NotificationMessage) error {
	c.last = msg
	if msg.Email == "" {
		return assert.AnError
	}
	return nil
}

func TestSendFillsUserEmailForMailChannel(t *testing.T) {
	uid := uuid.New()
	mail := &captureEmail{}
	reg := NewChannelRegistry()
	reg.Register(mail)
	svc := NewNotificationService(
		&stubNotifRepo{emails: map[uuid.UUID]string{uid: "ops@starbyte.test"}},
		stubTplRepo{tpl: &model.NotificationTemplate{Code: "backup_failed", Category: "system"}},
		stubEngine{},
		reg,
	)
	err := svc.Send(context.Background(), &dto.SendNotificationRequest{
		UserIDs:      []uuid.UUID{uid},
		TemplateCode: "backup_failed",
		Channels:     []string{"email"},
	})
	require.NoError(t, err)
	require.NotNil(t, mail.last)
	assert.Equal(t, "ops@starbyte.test", mail.last.Email)
	assert.Equal(t, uid, mail.last.UserID)
}

func TestSendSkipsEmailWhenUserHasNoAddress(t *testing.T) {
	uid := uuid.New()
	mail := &captureEmail{}
	reg := NewChannelRegistry()
	reg.Register(mail)
	svc := NewNotificationService(
		&stubNotifRepo{emails: map[uuid.UUID]string{}},
		stubTplRepo{tpl: &model.NotificationTemplate{Code: "backup_failed"}},
		stubEngine{},
		reg,
	)
	err := svc.Send(context.Background(), &dto.SendNotificationRequest{
		UserIDs:      []uuid.UUID{uid},
		TemplateCode: "backup_failed",
		Channels:     []string{"email"},
	})
	require.NoError(t, err)
	assert.Nil(t, mail.last)
}

func TestWantsEmail(t *testing.T) {
	assert.True(t, wantsEmail([]string{"in_app", "email"}))
	assert.False(t, wantsEmail([]string{"in_app", "websocket"}))
	assert.Equal(t, []string{"in_app"}, dropChannel([]string{"in_app", "email"}, "email"))
}
