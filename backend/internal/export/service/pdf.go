package service

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/Yogdunana/StarByte/backend/internal/export/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/phpdave11/gofpdf"
)

var pdfFontCandidates = []string{
	"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
	"/usr/share/fonts/truetype/droid/DroidSansFallbackFull.ttf",
	"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
}

func buildTablePDF(req *dto.TableExportRequest) ([]byte, error) {
	orient := "P"
	if len(req.Columns) > 6 {
		orient = "L"
	}
	pdf, font := newPDF(orient, "StarByte")
	pdf.SetTitle(req.Title, true)
	pdf.AddPage()
	if strings.TrimSpace(req.Title) != "" {
		pdf.SetFont(font, "", 16)
		pdf.CellFormat(0, 10, req.Title, "", 1, "C", false, 0, "")
		pdf.Ln(2)
	}
	drawPDFTable(pdf, font, req.Columns, req.Rows)
	return outputPDF(pdf)
}

func buildTemplatePDF(htmlBody, title, watermark string) ([]byte, error) {
	if strings.TrimSpace(watermark) == "" {
		watermark = "StarByte"
	}
	pdf, font := newPDF("P", watermark)
	if title == "" {
		title = "打印件"
	}
	pdf.SetTitle(title, true)
	pdf.AddPage()
	pdf.SetFont(font, "", 11)
	for _, line := range parseSimpleHTML(htmlBody) {
		if line.kind == "h1" {
			pdf.SetFont(font, "", 18)
			pdf.MultiCell(0, 10, line.text, "", "C", false)
			pdf.Ln(4)
			pdf.SetFont(font, "", 11)
			continue
		}
		pdf.MultiCell(0, 7, line.text, "", "L", false)
		pdf.Ln(1)
	}
	return outputPDF(pdf)
}

func newPDF(orientation, watermark string) (*gofpdf.Fpdf, string) {
	pdf := gofpdf.New(orientation, "mm", "A4", "")
	font := addPDFFont(pdf)
	mark := watermark
	pdf.SetMargins(14, 18, 14)
	pdf.SetAutoPageBreak(true, 18)
	pdf.SetHeaderFunc(func() {
		drawWatermark(pdf, font, mark)
		pdf.SetFont(font, "", 9)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(0, 6, "StarByte 计协 · 导出报表", "", 0, "L", false, 0, "")
		pdf.Ln(8)
		pdf.SetTextColor(15, 23, 42)
		pdf.SetDrawColor(29, 78, 216)
		pdf.Line(14, 16, 196, 16)
	})
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont(font, "", 8)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(0, 8, fmt.Sprintf("第 %d 页", pdf.PageNo()), "", 0, "C", false, 0, "")
	})
	return pdf, font
}

func addPDFFont(pdf *gofpdf.Fpdf) string {
	for _, path := range pdfFontCandidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		// TTC is not supported by UTF-8 embedding; skip after a failed probe.
		if strings.HasSuffix(strings.ToLower(path), ".ttc") {
			probe := gofpdf.New("P", "mm", "A4", "")
			probe.AddUTF8Font("export", "", path)
			if probe.Err() {
				continue
			}
			pdf.AddUTF8Font("export", "", path)
			if pdf.Err() {
				pdf.ClearError()
				continue
			}
			return "export"
		}
		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 {
			continue
		}
		pdf.AddUTF8FontFromBytes("export", "", data)
		if pdf.Err() {
			pdf.ClearError()
			continue
		}
		return "export"
	}
	return "Helvetica"
}

func drawWatermark(pdf *gofpdf.Fpdf, font, text string) {
	if text == "" {
		text = "StarByte"
	}
	pdf.SetAlpha(0.07, "Normal")
	pdf.TransformBegin()
	pdf.TransformRotate(35, 105, 150)
	pdf.SetFont(font, "", 42)
	pdf.SetTextColor(15, 23, 42)
	pdf.Text(40, 160, text)
	pdf.TransformEnd()
	pdf.SetAlpha(1, "Normal")
	pdf.SetTextColor(15, 23, 42)
}

func drawPDFTable(pdf *gofpdf.Fpdf, font string, headers []string, rows [][]string) {
	n := len(headers)
	if n == 0 {
		return
	}
	pageW, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	colW := (pageW - left - right) / float64(n)
	pdf.SetFont(font, "", 9)
	pdf.SetFillColor(29, 78, 216)
	pdf.SetTextColor(255, 255, 255)
	for _, h := range headers {
		pdf.CellFormat(colW, 8, clipPDF(h, 24), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetTextColor(15, 23, 42)
	fill := false
	for _, row := range rows {
		pdf.SetFillColor(241, 245, 249)
		for i := 0; i < n; i++ {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			pdf.CellFormat(colW, 7, clipPDF(val, 28), "1", 0, "L", fill, 0, "")
		}
		pdf.Ln(-1)
		fill = !fill
	}
}

func clipPDF(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max-1]) + "…"
}

func outputPDF(pdf *gofpdf.Fpdf) ([]byte, error) {
	if pdf.Err() {
		return nil, response.NewError(response.CodeInternalError, "生成 PDF 失败")
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, response.NewError(response.CodeInternalError, "生成 PDF 失败")
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		return nil, response.NewError(response.CodeInternalError, "生成 PDF 失败")
	}
	return buf.Bytes(), nil
}

type htmlLine struct {
	kind string
	text string
}

func parseSimpleHTML(raw string) []htmlLine {
	raw = strings.ReplaceAll(raw, "\r", "")
	var out []htmlLine
	for _, p := range strings.Split(raw, "\n") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kind := "p"
		if strings.Contains(p, "<h1") {
			kind = "h1"
		}
		p = stripTags(p)
		if p == "" {
			continue
		}
		out = append(out, htmlLine{kind: kind, text: p})
	}
	return out
}

func stripTags(s string) string {
	var b strings.Builder
	skip := false
	for _, r := range s {
		if r == '<' {
			skip = true
			continue
		}
		if r == '>' {
			skip = false
			continue
		}
		if !skip {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
