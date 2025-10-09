package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"hariomtransport/repository"
	"hariomtransport/utils"
)

type PDFHandler struct {
	Repo     *repository.PDFRepository
	SavePath string
}

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

	// Ensure save directory exists
	saveDir := h.SavePath
	if saveDir == "" {
		saveDir = "./pdfs"
	}
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		http.Error(w, "failed to create save directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate PDF bytes
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
	savePath := filepath.Join(saveDir, filename)

	if err := os.WriteFile(savePath, pdfBytes, 0644); err != nil {
		http.Error(w, "failed to save PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update pdf_created_at in bilty table
	if err := h.Repo.BiltyRepo.UpdatePDFCreatedAt(biltyID, time.Now()); err != nil {
		// Log the error but don't block the response
		fmt.Printf("failed to update pdf_created_at for bilty %d: %v\n", biltyID, err)
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"success":true,"file":"%s"}`, filename)))
}
