package certificates

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

func BuildPDF(item Certificate) []byte {
	lines := []string{
		"Пульс",
		item.Title,
		"Recipient: " + item.UserName,
		fmt.Sprintf("Approved volunteer hours: %.1f", item.TotalHours),
		"Issued at: " + item.IssuedAt.Format("02.01.2006"),
		"Verification code: " + item.VerifyCode,
		"Проверить код можно в системе Пульс.",
	}
	if item.OrganizationName != nil {
		lines = append(lines[:3], append([]string{"Organization: " + *item.OrganizationName}, lines[3:]...)...)
	}
	if item.Description != "" {
		lines = append(lines[:2], append([]string{item.Description}, lines[2:]...)...)
	}

	var content strings.Builder
	content.WriteString("BT\n/F1 24 Tf\n72 760 Td\n(Пульс - сертификат) Tj\n")
	content.WriteString("/F1 13 Tf\n0 -42 Td\n")
	for _, line := range lines {
		content.WriteString("(" + pdfEscape(line) + ") Tj\n0 -24 Td\n")
	}
	content.WriteString("ET\n")

	stream := content.String()
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream),
	}

	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, out.Len())
		out.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", i+1, obj))
	}
	xref := out.Len()
	out.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", len(objects)+1))
	for i := 1; i < len(offsets); i++ {
		out.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	out.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R /Info << /CreationDate (D:%s) >> >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, time.Now().Format("20060102150405"), xref))
	return out.Bytes()
}

func pdfEscape(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "(", `\(`)
	value = strings.ReplaceAll(value, ")", `\)`)
	return value
}
