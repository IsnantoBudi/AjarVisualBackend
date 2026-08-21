package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"ajarvisual-backend/config"
	"ajarvisual-backend/models"
	"ajarvisual-backend/services"
)

type GenerateRequest struct {
	Topik       string `json:"topik"`
	Kelas       int    `json:"kelas"`
	JumlahSoal  int    `json:"jumlah_soal"`
	TipeSoal    string `json:"tipe_soal"`
	TanpaGambar bool   `json:"tanpa_gambar"`
	Model       string `json:"model"`
	ImageModel  string `json:"image_model"`
}

func jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, statusCode int, msg string) {
	jsonResponse(w, statusCode, map[string]string{"error": msg})
}

func GenerateWorksheet(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GenerateWorksheet] Incoming request")
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GenerateWorksheet] Failed to decode body: %v", err)
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[GenerateWorksheet] Topic: %s, Grade: %d, Qty: %d, Type: %s, NoImage: %t, Model: %s, ImageModel: %s", 
		req.Topik, req.Kelas, req.JumlahSoal, req.TipeSoal, req.TanpaGambar, req.Model, req.ImageModel)

	if req.Topik == "" {
		errorResponse(w, http.StatusBadRequest, "Topik tidak boleh kosong!")
		return
	}
	if req.Kelas < 1 || req.Kelas > 6 {
		errorResponse(w, http.StatusBadRequest, "Kelas tidak valid!")
		return
	}

	if req.TipeSoal == "" {
		req.TipeSoal = "pilihan_ganda"
	}

	cfg := services.GenerateConfig{
		Topik:       req.Topik,
		Kelas:       req.Kelas,
		JumlahSoal:  req.JumlahSoal,
		TipeSoal:    req.TipeSoal,
		TanpaGambar: req.TanpaGambar,
		Model:       req.Model,
		ImageModel:  req.ImageModel,
	}

	log.Printf("[GenerateWorksheet] Calling services.GenerateSoal")
	soalList, err := services.GenerateSoal(cfg)
	if err != nil {
		log.Printf("[GenerateWorksheet] GenerateSoal failed: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Gagal generate soal: "+err.Error())
		return
	}
	log.Printf("[GenerateWorksheet] GenerateSoal returned %d questions successfully", len(soalList))

	worksheet := models.Worksheet{
		JudulMateri:  req.Topik,
		TingkatKelas: req.Kelas,
		DataSoal:     soalList,
	}

	dataSoalJSON, err := json.Marshal(worksheet.DataSoal)
	if err != nil {
		log.Printf("[GenerateWorksheet] JSON marshal of questions failed: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Gagal marshal data soal")
		return
	}

	log.Printf("[GenerateWorksheet] Saving worksheet to database (size %d bytes)", len(dataSoalJSON))
	result, err := config.DB.Exec(
		"INSERT INTO worksheets (judul_materi, tingkat_kelas, data_soal) VALUES (?, ?, ?)",
		worksheet.JudulMateri, worksheet.TingkatKelas, dataSoalJSON,
	)
	if err != nil {
		log.Printf("[GenerateWorksheet] Database insertion failed: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Gagal simpan worksheet: "+err.Error())
		return
	}

	lastInsertID, _ := result.LastInsertId()
	worksheet.ID = uint(lastInsertID)
	log.Printf("[GenerateWorksheet] Worksheet saved successfully with ID: %d", worksheet.ID)

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":   "Worksheet berhasil dibuat!",
		"worksheet": worksheet,
	})
}

func GetAllHistory(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, judul_materi, tingkat_kelas, data_soal, created_at FROM worksheets ORDER BY created_at DESC")
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Gagal ambil riwayat: "+err.Error())
		return
	}
	defer rows.Close()

	worksheets := make([]models.Worksheet, 0)
	for rows.Next() {
		var ws models.Worksheet
		var dataSoalRaw []byte
		if err := rows.Scan(&ws.ID, &ws.JudulMateri, &ws.TingkatKelas, &dataSoalRaw, &ws.CreatedAt); err != nil {
			errorResponse(w, http.StatusInternalServerError, "Gagal scan row: "+err.Error())
			return
		}
		if err := json.Unmarshal(dataSoalRaw, &ws.DataSoal); err != nil {
			errorResponse(w, http.StatusInternalServerError, "Gagal unmarshal data soal: "+err.Error())
			return
		}
		worksheets = append(worksheets, ws)
	}

	if err := rows.Err(); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Gagal iterate riwayat: "+err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, worksheets)
}

func GetWorksheetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var ws models.Worksheet
	var dataSoalRaw []byte
	err := config.DB.QueryRow("SELECT id, judul_materi, tingkat_kelas, data_soal, created_at FROM worksheets WHERE id = ?", id).
		Scan(&ws.ID, &ws.JudulMateri, &ws.TingkatKelas, &dataSoalRaw, &ws.CreatedAt)
	if err == sql.ErrNoRows {
		errorResponse(w, http.StatusNotFound, "Worksheet tidak ditemukan")
		return
	} else if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Gagal ambil worksheet: "+err.Error())
		return
	}

	if err := json.Unmarshal(dataSoalRaw, &ws.DataSoal); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Gagal unmarshal data soal: "+err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, ws)
}

func DeleteWorksheet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := config.DB.Exec("DELETE FROM worksheets WHERE id = ?", id)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Gagal hapus worksheet: "+err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "Worksheet dihapus"})
}

func AddSoalToWorksheet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 1. Ambil worksheet
	var ws models.Worksheet
	var dataSoalRaw []byte
	err := config.DB.QueryRow("SELECT id, judul_materi, tingkat_kelas, data_soal, created_at FROM worksheets WHERE id = ?", id).
		Scan(&ws.ID, &ws.JudulMateri, &ws.TingkatKelas, &dataSoalRaw, &ws.CreatedAt)
	if err == sql.ErrNoRows {
		errorResponse(w, http.StatusNotFound, "Worksheet tidak ditemukan")
		return
	} else if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Gagal ambil worksheet: "+err.Error())
		return
	}

	if err := json.Unmarshal(dataSoalRaw, &ws.DataSoal); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Gagal unmarshal data soal: "+err.Error())
		return
	}

	if req.TipeSoal == "" {
		req.TipeSoal = "pilihan_ganda"
	}

	cfg := services.GenerateConfig{
		Topik:       req.Topik,
		Kelas:       req.Kelas,
		JumlahSoal:  req.JumlahSoal,
		TipeSoal:    req.TipeSoal,
		TanpaGambar: req.TanpaGambar,
		Model:       req.Model,
		ImageModel:  req.ImageModel,
	}

	newSoal, err := services.GenerateSoal(cfg)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Gagal generate soal tambahan: "+err.Error())
		return
	}

	ws.DataSoal = append(ws.DataSoal, newSoal...)

	newDataSoalRaw, err := json.Marshal(ws.DataSoal)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Gagal marshal data soal tambahan")
		return
	}

	_, err = config.DB.Exec("UPDATE worksheets SET data_soal = ? WHERE id = ?", newDataSoalRaw, id)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Gagal menyimpan soal tambahan: "+err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":   "Soal tambahan berhasil ditambahkan!",
		"worksheet": ws,
	})
}

func RegenerateImage(w http.ResponseWriter, r *http.Request) {
	type RegenerateReq struct {
		ImagePrompt string `json:"image_prompt"`
		ImageModel  string `json:"image_model"`
	}
	var req RegenerateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.ImagePrompt == "" {
		errorResponse(w, http.StatusBadRequest, "image_prompt is required")
		return
	}
	url := services.GenerateImageURL(req.ImagePrompt, req.ImageModel)
	jsonResponse(w, http.StatusOK, map[string]string{"image_url": url})
}

func ProxyImage(w http.ResponseWriter, r *http.Request) {
	prompt := r.URL.Query().Get("prompt")
	if prompt == "" {
		http.Error(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	imageModel := r.URL.Query().Get("image_model")
	if imageModel == "" {
		imageModel = r.URL.Query().Get("model")
	}

	imageData, contentType, err := services.GenerateImage(prompt, imageModel)
	if err != nil {
		http.Error(w, "Gagal ambil gambar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if contentType == "" {
		contentType = "image/jpeg"
	}

	// Set Cache-Control header (1 year, immutable)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(imageData)
}
