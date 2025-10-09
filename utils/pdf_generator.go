package utils

import (
	"bytes"
	"context"
	"hariomtransport/models"
	"hariomtransport/repository"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// GenerateBiltyPDF ensures each copy stays whole, but only moves to new page if cut.
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

	// Load HTML template once
	tmpl, err := template.ParseFiles("templates/bilty_template.html")
	if err != nil {
		return nil, err
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
			return nil, err
		}

		// Wrap each copy in a div that avoids breaking across pages
		fullHTML.WriteString("<div class='bilty-copy'>")
		fullHTML.Write(buf.Bytes())
		fullHTML.WriteString("</div>")
	}

	// Final HTML with smart CSS page handling
	finalHTML := `
		<!DOCTYPE html>
		<html>
		<head>
		<meta charset="UTF-8">
		<style>
		@page {
			size: A4;
			margin: 20px;
		}
		body {
			font-family: Arial, Helvetica, sans-serif;
			font-size: 12px;
			margin: 0;
			padding: 0;
		}
		.bilty-copy {
			page-break-inside: avoid; /* Prevent cutting copy in middle */
			// margin-bottom: 15px;
			border: none;
		}
		.bilty-copy:not(:last-child) {
			// margin-bottom: 15px;
		}
		</style>
		</head>
		<body>` + fullHTML.String() + `</body></html>`

	// Create temp HTML file
	tmpDir := os.TempDir()
	tmpHTML := filepath.Join(tmpDir, "bilty_"+time.Now().Format("20060102150405")+".html")
	if err := os.WriteFile(tmpHTML, []byte(finalHTML), 0644); err != nil {
		return nil, err
	}
	defer os.Remove(tmpHTML)

	// Generate PDF with headless Chrome
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var pdfBuf []byte
	fileURL := "file://" + tmpHTML

	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.Sleep(1*time.Second),
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
		return nil, err
	}

	return pdfBuf, nil
}
