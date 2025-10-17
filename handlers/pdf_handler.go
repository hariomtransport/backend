package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hariomtransport/backend/repository"
	"github.com/hariomtransport/backend/utils"
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
		writeJSON(w, http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Missing bilty ID",
		})
		return
	}

	biltyID, err := strconv.ParseInt(biltyIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Invalid bilty ID",
		})
		return
	}

	// Fetch bilty record
	bilty, err := h.Repo.BiltyRepo.GetBiltyByID(biltyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to fetch bilty: " + err.Error(),
		})
		return
	}
	if bilty == nil {
		writeJSON(w, http.StatusNotFound, ApiResponse{
			Success: false,
			Message: "Bilty not found",
		})
		return
	}

	// Determine save directory
	saveDir := h.SavePath
	if saveDir == "" {
		saveDir = "./pdfs"
	}
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to create save directory: " + err.Error(),
		})
		return
	}

	// Check if PDF needs regeneration
	shouldGenerate := false
	if bilty.PdfPath == nil ||
		bilty.PdfCreatedAt == nil ||
		(bilty.UpdatedAt != nil && bilty.PdfCreatedAt.Before(*bilty.UpdatedAt)) {
		shouldGenerate = true
	}

	// Reuse existing PDF if still valid
	if !shouldGenerate {
		writeJSON(w, http.StatusOK, ApiResponse{
			Success: true,
			Message: "Existing PDF is up-to-date",
			Data: map[string]interface{}{
				"file": *bilty.PdfPath,
			},
		})
		return
	}

	// Generate new PDF
	pdfBytes, err := utils.GenerateBiltyPDF(h.Repo, biltyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to generate PDF: " + err.Error(),
		})
		return
	}
	if len(pdfBytes) == 0 {
		writeJSON(w, http.StatusNotFound, ApiResponse{
			Success: false,
			Message: "No bilty data found to generate PDF",
		})
		return
	}

	// Save or upload PDF
	filename := fmt.Sprintf("bilty_%d_%d.pdf", biltyID, time.Now().Unix())

	// Upload PDF to Cloudflare R2
	r2URL, err := utils.UploadToR2(pdfBytes, filename)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to upload PDF to R2: " + err.Error(),
		})
		return
	}

	oldPdfPath := bilty.PdfPath

	// Update database
	now := time.Now().UTC()
	if err := h.Repo.BiltyRepo.UpdatePDFInfo(biltyID, r2URL, now); err != nil {
		fmt.Printf("⚠️ Failed to update PDF info for bilty %d: %v\n", biltyID, err)
	}

	// Delete old PDF from R2
	if oldPdfPath != nil && *oldPdfPath != "" {
		if err := utils.DeleteFromR2(*oldPdfPath); err != nil {
			fmt.Printf("⚠️ Failed to delete old PDF from R2 for bilty %d: %v\n", biltyID, err)
		}
	}

	// Success response
	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Bilty PDF generated and uploaded successfully",
		Data: map[string]interface{}{
			"file": r2URL,
		},
	})
}
