package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hariomtransport/repository"
	"github.com/hariomtransport/utils"
)

type PDFHandler struct {
	Repo     *repository.PDFRepository
	SavePath string
}

// BiltyPDF handles the API request to generate and save a Bilty PDF
// BiltyPDF handles the API request to generate and save a Bilty PDF
func (h *PDFHandler) BiltyPDF(w http.ResponseWriter, r *http.Request) {
	// Parse bilty ID
	biltyIDStr := r.URL.Query().Get("id")
	if biltyIDStr == "" {
		http.Error(w, "missing bilty id", http.StatusBadRequest)
		return
	}

	biltyID, err := strconv.ParseInt(biltyIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid bilty id", http.StatusBadRequest)
		return
	}

	// Fetch bilty record
	bilty, err := h.Repo.BiltyRepo.GetBiltyByID(biltyID)
	if err != nil {
		http.Error(w, "failed to fetch bilty: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if bilty == nil {
		http.Error(w, "bilty not found", http.StatusNotFound)
		return
	}

	// Determine save directory
	saveDir := h.SavePath
	if saveDir == "" {
		saveDir = "./pdfs"
	}
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		http.Error(w, "failed to create save directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Decision: Should we reuse or regenerate PDF?
	shouldGenerate := false

	if bilty.PdfPath == nil {
		shouldGenerate = true
	} else if bilty.PdfCreatedAt == nil {
		shouldGenerate = true
	} else if bilty.UpdatedAt != nil && bilty.PdfCreatedAt.Before(*bilty.UpdatedAt) {
		shouldGenerate = true
	}

	// If we already have a valid and up-to-date PDF → reuse it
	if !shouldGenerate {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"success":true,"file":"%s"}`, *bilty.PdfPath)))
		return
	}

	// Generate new PDF
	pdfBytes, err := utils.GenerateBiltyPDF(h.Repo, biltyID)
	if err != nil {
		http.Error(w, "failed to generate PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(pdfBytes) == 0 {
		http.Error(w, "no bilty found", http.StatusNotFound)
		return
	}

	// Save PDF to file
	filename := fmt.Sprintf("bilty_%d_%d.pdf", biltyID, time.Now().Unix())

	// Upload PDF to Cloudflare R2
	r2URL, err := utils.UploadToR2(pdfBytes, filename)
	if err != nil {
		http.Error(w, "failed to upload PDF to R2: "+err.Error(), http.StatusInternalServerError)
		return
	}

	oldPdfPath := bilty.PdfPath

	// Update database with new PDF info
	now := time.Now().UTC()
	if err := h.Repo.BiltyRepo.UpdatePDFInfo(biltyID, r2URL, now); err != nil {
		fmt.Printf("failed to update pdf_path/pdf_created_at for bilty %d: %v\n", biltyID, err)
	}

	// Delete old PDF from R2 (after new is successfully uploaded & DB updated)
	if oldPdfPath != nil && *oldPdfPath != "" {
		fmt.Println("Deleting old PDF from R2:", *oldPdfPath)
		if err := utils.DeleteFromR2(*oldPdfPath); err != nil {
			fmt.Printf("failed to delete old PDF from R2 for bilty %d: %v\n", biltyID, err)
		} else {
			fmt.Println("Old PDF deleted successfully from R2.")
		}
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"success":true,"file":"%s"}`, r2URL)))
}
