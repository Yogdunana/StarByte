package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/Yogdunana/StarByte/backend/internal/audit/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/xuri/excelize/v2"
)

var exportHeaders = []string{"时间", "用户名", "操作类型", "模块", "路径", "IP", "耗时"}

func (s *auditService) Export(ctx context.Context, req *dto.ExportAuditLogRequest) ([]byte, string, error) {
	params := toListParams(req.UserID, req.Username, req.Action, req.Module, req.Keyword, req.IPAddress, req.Method, req.StartTime, req.EndTime)
	count, err := s.auditRepo.Count(ctx, params)
	if err != nil {
		return nil, "", fmt.Errorf("count audit logs: %w", err)
	}
	if count > model.MaxExportRows {
		return nil, "", response.NewError(response.CodeAuditExportLimit,
			fmt.Sprintf("导出数量超限（最大 %d）", model.MaxExportRows))
	}

	var buf bytes.Buffer
	switch req.Format {
	case "csv":
		if err := s.writeCSV(ctx, params, &buf); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "audit_logs.csv", nil
	case "excel":
		if err := s.writeExcel(ctx, params, &buf); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "audit_logs.xlsx", nil
	default:
		return nil, "", response.NewError(response.CodeAuditExportErr, "不支持的导出格式: "+req.Format)
	}
}

func (s *auditService) writeCSV(ctx context.Context, params *repo.ListParams, w io.Writer) error {
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write(exportHeaders); err != nil {
		return err
	}
	err := s.auditRepo.Iterate(ctx, params, model.DefaultIterateBatch, func(logs []model.AuditLog) error {
		for _, log := range logs {
			if err := cw.Write(exportRow(log)); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	})
	if err != nil {
		return fmt.Errorf("stream csv export: %w", err)
	}
	cw.Flush()
	return cw.Error()
}

func (s *auditService) writeExcel(ctx context.Context, params *repo.ListParams, w io.Writer) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheet := "Sheet1"
	sw, err := f.NewStreamWriter(sheet)
	if err != nil {
		return fmt.Errorf("excel stream writer: %w", err)
	}
	header := make([]any, len(exportHeaders))
	for i, h := range exportHeaders {
		header[i] = h
	}
	if err := sw.SetRow("A1", header); err != nil {
		return err
	}
	row := 2
	err = s.auditRepo.Iterate(ctx, params, model.DefaultIterateBatch, func(logs []model.AuditLog) error {
		for _, log := range logs {
			cell, _ := excelize.CoordinatesToCellName(1, row)
			values := exportRow(log)
			iface := make([]any, len(values))
			for i, v := range values {
				iface[i] = v
			}
			if err := sw.SetRow(cell, iface); err != nil {
				return err
			}
			row++
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("stream excel export: %w", err)
	}
	if err := sw.Flush(); err != nil {
		return err
	}
	return f.Write(w)
}

func exportRow(log model.AuditLog) []string {
	return []string{
		log.CreatedAt.Format("2006-01-02 15:04:05"),
		log.Username,
		log.Action,
		log.Module,
		log.Path,
		log.IP,
		fmt.Sprintf("%d", log.DurationMs),
	}
}
