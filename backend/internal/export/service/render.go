package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/export/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/xuri/excelize/v2"
)

func buildCSV(req *dto.TableExportRequest) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return nil, err
	}
	w := csv.NewWriter(&buf)
	w.Comma = csvComma(req.Delimiter)
	if err := w.Write(req.Columns); err != nil {
		return nil, response.NewError(response.CodeInternalError, "生成 CSV 失败")
	}
	n := len(req.Columns)
	for _, row := range req.Rows {
		line := make([]string, n)
		for i := 0; i < n; i++ {
			if i < len(row) {
				line[i] = row[i]
			}
		}
		if err := w.Write(line); err != nil {
			return nil, response.NewError(response.CodeInternalError, "生成 CSV 失败")
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, response.NewError(response.CodeInternalError, "生成 CSV 失败")
	}
	return buf.Bytes(), nil
}

func csvComma(delim string) rune {
	switch strings.TrimSpace(delim) {
	case "\t", "tab":
		return '\t'
	case ";", "semicolon":
		return ';'
	default:
		return ','
	}
}

func buildExcel(req *dto.TableExportRequest) ([]byte, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := sanitizeSheet(req.Sheet)
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		sheet = "Sheet1"
	}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1D4ED8"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	start := 1
	n := len(req.Columns)
	if n == 0 {
		n = 1
	}
	if strings.TrimSpace(req.Title) != "" {
		end, _ := excelize.CoordinatesToCellName(n, 1)
		_ = f.MergeCell(sheet, "A1", end)
		_ = f.SetCellValue(sheet, "A1", req.Title)
		_ = f.SetCellStyle(sheet, "A1", "A1", titleStyle)
		_ = f.SetRowHeight(sheet, 1, 24)
		start = 2
	}
	for i, h := range req.Columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, start)
		_ = f.SetCellValue(sheet, cell, h)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sheet, col, col, 16)
	}
	for r, row := range req.Rows {
		for c := 0; c < n; c++ {
			val := ""
			if c < len(row) {
				val = row[c]
			}
			cell, _ := excelize.CoordinatesToCellName(c+1, start+1+r)
			_ = f.SetCellValue(sheet, cell, val)
		}
	}
	from, _ := excelize.CoordinatesToCellName(1, start)
	to, _ := excelize.CoordinatesToCellName(n, start+len(req.Rows))
	_ = f.AutoFilter(sheet, from+":"+to, nil)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, response.NewError(response.CodeInternalError, "生成 Excel 失败")
	}
	return buf.Bytes(), nil
}

func buildJSON(req *dto.TableExportRequest) ([]byte, error) {
	payload := struct {
		Columns []string   `json:"columns"`
		Rows    [][]string `json:"rows"`
	}{Columns: req.Columns, Rows: req.Rows}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, response.NewError(response.CodeInternalError, "生成 JSON 失败")
	}
	return data, nil
}
