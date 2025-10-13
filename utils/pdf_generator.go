package utils

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/hariomtransport/backend/models"
	"github.com/hariomtransport/backend/repository"
)

// GenerateBiltyPDF fetches the template from the URL specified in TEMPLATE_FILE env variable
// and generates a PDF.
func GenerateBiltyPDF(repo *repository.PDFRepository, biltyID int64) ([]byte, error) {
	// Fetch initial setup
	initial, err := repo.GetInitialForPDF()
	if err != nil {
		return nil, err
	}

	// Fetch bilty
	bilty, err := repo.GetBiltyForPDF(biltyID)
	if err != nil {
		return nil, err
	}
	if bilty == nil {
		return nil, nil
	}

	// Format bilty date safely
	formattedBiltyDate := "-"
	if !bilty.Date.IsZero() {
		formattedBiltyDate = bilty.Date.Format("02-Jan-2006")
	}

	// Prepare contact numbers
	contacts := ""
	for _, m := range initial.Mobile {
		contacts += m.Number + "(" + m.Label + "), "
	}
	if len(contacts) > 2 {
		contacts = contacts[:len(contacts)-2]
	}

	// Copy titles
	copyTitles := []string{"Consignor Copy", "Consignee Copy", "Driver Copy"}

	// --- Fetch template from URL in TEMPLATE_FILE ---
	templateURL := os.Getenv("TEMPLATE_FILE")
	if templateURL == "" {
		return nil, fmt.Errorf("TEMPLATE_FILE environment variable not set")
	}

	resp, err := http.Get(templateURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch template, status code: %d", resp.StatusCode)
	}

	tmplBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read template body: %w", err)
	}

	tmpl, err := template.New("bilty").Parse(string(tmplBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var fullHTML bytes.Buffer
	for _, title := range copyTitles {
		data := models.BiltyPDFData{
			Company:    initial,
			Bilty:      bilty,
			Contacts:   contacts,
			Date:       formattedBiltyDate,
			Total:      bilty.ToPay,
			TotalWords: NumberToCurrencyWords(bilty.ToPay),
			CopyTitle:  title,
			GoodsCount: len(bilty.Goods),
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("failed to execute template: %w", err)
		}

		fullHTML.WriteString("<div class='bilty-copy'>")
		fullHTML.Write(buf.Bytes())
		fullHTML.WriteString("</div>")
	}

	finalHTML := `
	<!DOCTYPE html>
	<html>
	<head>
	<meta charset="UTF-8">
	<style>
	@page { size: A4; margin: 20px; }
	body { font-family: Arial, Helvetica, sans-serif; font-size: 12px; margin:0; padding:0; }
	.bilty-copy { page-break-inside: avoid; border:none; }
	</style>
	</head>
	<body>` + fullHTML.String() + `</body></html>`

	// Create temp HTML file
	tmpFile := fmt.Sprintf("/tmp/bilty_%s.html", time.Now().UTC().Format("20060102150405"))
	if err := os.WriteFile(tmpFile, []byte(finalHTML), 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp HTML: %w", err)
	}
	defer os.Remove(tmpFile)

	fileURL := "file://" + tmpFile

	// Generate PDF with headless Chrome
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var pdfBuf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.Sleep(2*time.Second), // ensure page fully loads
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).  // A4 width
				WithPaperHeight(11.7). // A4 height
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfBuf, nil
}
