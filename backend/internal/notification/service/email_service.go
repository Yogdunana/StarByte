package service

import (
	"context"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/Yogdunana/StarByte/backend/internal/notification/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

type EmailService interface {
	Send(ctx context.Context, req *dto.SendEmailRequest, operator uuid.UUID) (*dto.SendEmailResponse, error)
	SendBatch(ctx context.Context, req *dto.BatchSendEmailRequest, operator uuid.UUID) (*dto.BatchSendEmailResponse, error)
	ListLogs(ctx context.Context, req *dto.ListEmailLogsRequest) ([]*dto.EmailLogResponse, int64, error)
}

type emailService struct {
	dispatch MailDispatcher
	engine   TemplateEngine
	logs     repo.EmailLogRepo
}

func NewEmailService(dispatch MailDispatcher, engine TemplateEngine, logs repo.EmailLogRepo) EmailService {
	return &emailService{dispatch: dispatch, engine: engine, logs: logs}
}

func (s *emailService) Send(ctx context.Context, req *dto.SendEmailRequest, operator uuid.UUID) (*dto.SendEmailResponse, error) {
	ids, err := parseAttachmentIDs(req.AttachmentIDs)
	if err != nil {
		return nil, response.NewError(response.CodeNotificationEmailInvalid, "邮件参数无效")
	}
	nid, err := dto.ParseOptionalUUID(stringOrEmpty(req.NotificationID))
	if err != nil {
		return nil, response.NewError(response.CodeNotificationEmailInvalid, "邮件参数无效")
	}
	job := MailJob{
		To: req.To, CC: req.CC, Subject: req.Subject, Body: req.Body, IsHTML: req.IsHTML,
		AttachmentIDs: ids, NotificationID: nid, TemplateCode: strings.TrimSpace(req.TemplateCode), UserID: &operator,
	}
	if job.TemplateCode != "" && job.Body == "" {
		rendered, rerr := s.engine.Render(ctx, job.TemplateCode, map[string]any{})
		if rerr != nil {
			return nil, rerr
		}
		job.Subject = rendered.Title
		job.Body = rendered.Content
	}
	id, err := s.dispatch.Enqueue(ctx, job)
	if err != nil {
		if err == errEmailQueueFull {
			return nil, response.NewError(response.CodeNotificationEmailRate, "邮件队列已满")
		}
		return nil, err
	}
	return &dto.SendEmailResponse{MessageID: id.String(), Status: "queued", RecipientCount: len(job.To)}, nil
}

func (s *emailService) SendBatch(ctx context.Context, req *dto.BatchSendEmailRequest, operator uuid.UUID) (*dto.BatchSendEmailResponse, error) {
	if len(req.Recipients) > emailPerMinute {
		return nil, response.NewError(response.CodeNotificationEmailRate, "批量发送超出每分钟 50 封限制")
	}
	out := &dto.BatchSendEmailResponse{Total: len(req.Recipients)}
	for _, r := range req.Recipients {
		subject, body := req.Subject, req.Body
		if strings.TrimSpace(req.TemplateCode) != "" {
			vars := map[string]any{}
			for k, v := range r.Data {
				vars[k] = v
			}
			rendered, err := s.engine.Render(ctx, req.TemplateCode, vars)
			if err != nil {
				out.Failed++
				continue
			}
			subject, body = rendered.Title, rendered.Content
		}
		if strings.TrimSpace(subject) == "" || strings.TrimSpace(body) == "" {
			out.Failed++
			continue
		}
		job := MailJob{
			To: []string{r.Email}, Subject: subject, Body: body, IsHTML: req.IsHTML,
			TemplateCode: req.TemplateCode, UserID: &operator,
		}
		if _, err := s.dispatch.Enqueue(ctx, job); err != nil {
			out.Failed++
			continue
		}
		out.Queued++
	}
	return out, nil
}

func (s *emailService) ListLogs(ctx context.Context, req *dto.ListEmailLogsRequest) ([]*dto.EmailLogResponse, int64, error) {
	status, ok := dto.ParseLogStatus(req.Status)
	if !ok {
		return nil, 0, response.NewError(response.CodeNotificationEmailInvalid, "邮件参数无效")
	}
	start, err := parseLogDate(req.StartDate, false)
	if err != nil {
		return nil, 0, response.NewError(response.CodeNotificationEmailInvalid, "邮件参数无效")
	}
	end, err := parseLogDate(req.EndDate, true)
	if err != nil {
		return nil, 0, response.NewError(response.CodeNotificationEmailInvalid, "邮件参数无效")
	}
	list, total, err := s.logs.List(ctx, status, start, end, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*dto.EmailLogResponse, 0, len(list))
	for _, row := range list {
		out = append(out, &dto.EmailLogResponse{
			ID: row.ID.String(), To: row.ToAddress, CC: row.CC, Subject: row.Subject,
			TemplateCode: row.TemplateCode, Status: model.EmailStatusName(row.Status),
			ErrorMessage: row.ErrorMessage, RetryCount: row.RetryCount, SentAt: row.SentAt, CreatedAt: row.CreatedAt,
		})
	}
	return out, total, nil
}

func parseAttachmentIDs(raw []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		id, err := uuid.Parse(strings.TrimSpace(s))
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func stringOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func parseLogDate(raw string, endOfDay bool) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Nanosecond)
	}
	return &t, nil
}
