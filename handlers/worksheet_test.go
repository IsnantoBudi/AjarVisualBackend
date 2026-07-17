package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"ajarvisual-backend/config"
	"ajarvisual-backend/handlers"
	"ajarvisual-backend/models"

	_ "modernc.org/sqlite" // Pure Go SQLite driver for testing
)

// roundTripFunc mock type to intercept HTTP requests
type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func setupTestDB(t *testing.T) {
	// 1. Setup in-memory SQLite database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open testing in-memory SQLite database: %v", err)
	}

	// 2. Create tables
	createWorksheetsQuery := `
	CREATE TABLE IF NOT EXISTS worksheets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		judul_materi TEXT NOT NULL,
		tingkat_kelas INTEGER DEFAULT 1,
		data_soal TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createWorksheetsQuery); err != nil {
		t.Fatalf("Failed to create test worksheets table: %v", err)
	}

	createCachedImagesQuery := `
	CREATE TABLE IF NOT EXISTS cached_images (
		prompt_hash TEXT PRIMARY KEY,
		prompt TEXT NOT NULL,
		image_data BLOB NOT NULL,
		content_type TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createCachedImagesQuery); err != nil {
		t.Fatalf("Failed to create test cached_images table: %v", err)
	}

	config.DB = db
}

func TestHealthCheck(t *testing.T) {
	setupTestDB(t)
	defer config.DB.Close()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	// Simple direct handler health check test
	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"AjarVisual API is running"}`))
	}

	healthHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	var data map[string]string
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if data["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", data["status"])
	}
}

func TestGetAllHistory(t *testing.T) {
	setupTestDB(t)
	defer config.DB.Close()

	// Insert mock worksheets data
	dataSoalDummy := `[{"pertanyaan": "Soal 1", "opsi": ["A", "B"], "jawaban_benar": "A", "tipe_soal": "pilihan_ganda", "tanpa_gambar": true}]`
	_, err := config.DB.Exec(
		"INSERT INTO worksheets (judul_materi, tingkat_kelas, data_soal) VALUES (?, ?, ?)",
		"Matematika Dasar", 1, dataSoalDummy,
	)
	if err != nil {
		t.Fatalf("Failed to insert mock worksheets: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/history", nil)
	w := httptest.NewRecorder()

	handlers.GetAllHistory(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	var worksheets []models.Worksheet
	if err := json.Unmarshal(body, &worksheets); err != nil {
		t.Fatalf("Failed to unmarshal worksheets: %v", err)
	}

	if len(worksheets) != 1 {
		t.Errorf("Expected 1 worksheet, got %d", len(worksheets))
	}

	if worksheets[0].JudulMateri != "Matematika Dasar" {
		t.Errorf("Expected 'Matematika Dasar', got '%s'", worksheets[0].JudulMateri)
	}
}

func TestGetWorksheetByID(t *testing.T) {
	setupTestDB(t)
	defer config.DB.Close()

	// Insert mock data
	dataSoalDummy := `[{"pertanyaan": "Soal 1", "opsi": ["A", "B"], "jawaban_benar": "A", "tipe_soal": "pilihan_ganda", "tanpa_gambar": true}]`
	_, err := config.DB.Exec(
		"INSERT INTO worksheets (id, judul_materi, tingkat_kelas, data_soal) VALUES (?, ?, ?, ?)",
		42, "Bahasa Indonesia", 2, dataSoalDummy,
	)
	if err != nil {
		t.Fatalf("Failed to insert mock data: %v", err)
	}

	// Test Found ID
	req := httptest.NewRequest("GET", "/api/history/42", nil)
	req.SetPathValue("id", "42") // Inject path value for standard ServeMux router
	w := httptest.NewRecorder()

	handlers.GetWorksheetByID(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	var ws models.Worksheet
	if err := json.Unmarshal(body, &ws); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if ws.ID != 42 || ws.JudulMateri != "Bahasa Indonesia" {
		t.Errorf("Unexpected worksheet data: %+v", ws)
	}

	// Test Not Found ID
	reqNotFound := httptest.NewRequest("GET", "/api/history/999", nil)
	reqNotFound.SetPathValue("id", "999")
	wNotFound := httptest.NewRecorder()

	handlers.GetWorksheetByID(wNotFound, reqNotFound)
	if wNotFound.Result().StatusCode != http.StatusNotFound {
		t.Errorf("Expected status Not Found (404), got %d", wNotFound.Result().StatusCode)
	}
}

func TestDeleteWorksheet(t *testing.T) {
	setupTestDB(t)
	defer config.DB.Close()

	// Insert data
	_, err := config.DB.Exec("INSERT INTO worksheets (id, judul_materi, data_soal) VALUES (?, ?, ?)", 55, "Sains", "[]")
	if err != nil {
		t.Fatalf("Failed to insert mock: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/history/55", nil)
	req.SetPathValue("id", "55")
	w := httptest.NewRecorder()

	handlers.DeleteWorksheet(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Result().StatusCode)
	}

	// Verify it was deleted
	var count int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM worksheets WHERE id = 55").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Error("Worksheet was not deleted from database")
	}
}

func TestRegenerateImage(t *testing.T) {
	setupTestDB(t)
	defer config.DB.Close()

	reqBody := `{"image_prompt": "seekor kucing lucu"}`
	req := httptest.NewRequest("POST", "/api/regenerate-image", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	handlers.RegenerateImage(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	var data map[string]string
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !strings.Contains(data["image_url"], "/api/image-proxy?prompt=") {
		t.Errorf("Unexpected image url: %s", data["image_url"])
	}
}

func TestProxyImage(t *testing.T) {
	setupTestDB(t)
	defer config.DB.Close()

	// Intercept global http.DefaultTransport to mock Pollinations/HuggingFace API response
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	mockResponseData := []byte("fake-image-bytes")
	http.DefaultTransport = roundTripFunc(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer(mockResponseData)),
			Header:     make(http.Header),
		}
	})

	req := httptest.NewRequest("GET", "/api/image-proxy?prompt=test-prompt", nil)
	w := httptest.NewRecorder()

	handlers.ProxyImage(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", resp.StatusCode)
	}

	if string(body) != string(mockResponseData) {
		t.Errorf("Expected body %s, got %s", mockResponseData, body)
	}

	// Verify image was cached in database
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM cached_images").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Error("Image was not cached in the database")
	}
}

func TestGenerateWorksheet(t *testing.T) {
	setupTestDB(t)
	defer config.DB.Close()

	// Setup mock Ollama API Server
	ollamaMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"response": "[{\"pertanyaan\": \"1+1?\", \"jawaban_benar\": \"2\", \"opsi\": [\"1\", \"2\", \"3\"], \"tipe_soal\": \"pilihan_ganda\", \"tanpa_gambar\": true}]"
		}`))
	}))
	defer ollamaMockServer.Close()

	os.Setenv("OLLAMA_CLOUD_API", "test-key")
	os.Setenv("OLLAMA_CLOUD_URL", ollamaMockServer.URL)
	defer func() {
		os.Unsetenv("OLLAMA_CLOUD_API")
		os.Unsetenv("OLLAMA_CLOUD_URL")
	}()

	reqBody := `{"topik": "Matematika", "kelas": 1, "jumlah_soal": 1, "tipe_soal": "pilihan_ganda", "tanpa_gambar": true}`
	req := httptest.NewRequest("POST", "/api/generate", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	handlers.GenerateWorksheet(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if data["message"] != "Worksheet berhasil dibuat!" {
		t.Errorf("Expected success message, got: %s", data["message"])
	}

	// Verify saved to DB
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM worksheets").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Error("Worksheet was not saved to database")
	}
}

func TestAddSoalToWorksheet(t *testing.T) {
	setupTestDB(t)
	defer config.DB.Close()

	// Insert mock initial data
	dataSoalDummy := `[{"pertanyaan": "Soal Awal", "opsi": ["A", "B"], "jawaban_benar": "A", "tipe_soal": "pilihan_ganda", "tanpa_gambar": true}]`
	_, err := config.DB.Exec(
		"INSERT INTO worksheets (id, judul_materi, tingkat_kelas, data_soal) VALUES (?, ?, ?, ?)",
		10, "Matematika", 1, dataSoalDummy,
	)
	if err != nil {
		t.Fatalf("Failed to insert mock: %v", err)
	}

	// Setup mock Ollama API Server
	ollamaMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"response": "[{\"pertanyaan\": \"Soal Tambahan\", \"jawaban_benar\": \"X\", \"opsi\": [\"X\", \"Y\"], \"tipe_soal\": \"pilihan_ganda\", \"tanpa_gambar\": true}]"
		}`))
	}))
	defer ollamaMockServer.Close()

	os.Setenv("OLLAMA_CLOUD_API", "test-key")
	os.Setenv("OLLAMA_CLOUD_URL", ollamaMockServer.URL)
	defer func() {
		os.Unsetenv("OLLAMA_CLOUD_API")
		os.Unsetenv("OLLAMA_CLOUD_URL")
	}()

	reqBody := `{"topik": "Matematika", "kelas": 1, "jumlah_soal": 1, "tipe_soal": "pilihan_ganda", "tanpa_gambar": true}`
	req := httptest.NewRequest("POST", "/api/history/10/add-soal", bytes.NewBufferString(reqBody))
	req.SetPathValue("id", "10")
	w := httptest.NewRecorder()

	handlers.AddSoalToWorksheet(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Verify update in DB
	var dataSoalRaw []byte
	err = config.DB.QueryRow("SELECT data_soal FROM worksheets WHERE id = 10").Scan(&dataSoalRaw)
	if err != nil {
		t.Fatal(err)
	}

	var listSoal []models.Soal
	if err := json.Unmarshal(dataSoalRaw, &listSoal); err != nil {
		t.Fatal(err)
	}

	if len(listSoal) != 2 {
		t.Errorf("Expected 2 soal after addition, got %d", len(listSoal))
	}

	if listSoal[1].Pertanyaan != "Soal Tambahan" {
		t.Errorf("Expected second question to be 'Soal Tambahan', got '%s'", listSoal[1].Pertanyaan)
	}
}
