package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"ajarvisual-backend/config"
	"ajarvisual-backend/models"
)

type GenerateConfig struct {
	Topik       string `json:"topik"`
	Kelas       int    `json:"kelas"`
	JumlahSoal  int    `json:"jumlah_soal"`
	TipeSoal    string `json:"tipe_soal"`
	TanpaGambar bool   `json:"tanpa_gambar"`
	Model       string `json:"model"`
	ImageModel  string `json:"image_model"`
}

type ollamaMatchingPair struct {
	Kiri         string `json:"kiri"`
	Kanan        string `json:"kanan"`
	KiriIsImage  bool   `json:"kiri_is_image"`
	KananIsImage bool   `json:"kanan_is_image"`
	KiriPrompt   string `json:"kiri_prompt,omitempty"`
	KananPrompt  string `json:"kanan_prompt,omitempty"`
}

type ollamaSoal struct {
	Pertanyaan        string               `json:"pertanyaan"`
	JawabanBenar      string               `json:"jawaban_benar,omitempty"`
	Opsi              []string             `json:"opsi,omitempty"`
	PasanganItem      []ollamaMatchingPair `json:"pasangan_item,omitempty"`
	TipeSoal          string               `json:"tipe_soal"`
	TanpaGambar       bool                 `json:"tanpa_gambar"`
	ImagePrompt       string               `json:"image_prompt,omitempty"`
	SukuKataAwal      string               `json:"suku_kata_awal,omitempty"`
	PilihanSukuKata   []string             `json:"pilihan_suku_kata,omitempty"`
	HurufDepan        string               `json:"huruf_depan,omitempty"`
	SisaKata          string               `json:"sisa_kata,omitempty"`
	OpsiKata          []string             `json:"opsi_kata,omitempty"`
	HurufAcak         string               `json:"huruf_acak,omitempty"`
	JumlahHuruf       int                  `json:"jumlah_huruf,omitempty"`
	MathBlocks        []models.MathBlock   `json:"math_blocks,omitempty"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Format string `json:"format,omitempty"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func GenerateSoal(cfg GenerateConfig) ([]models.Soal, error) {
	apiKey := config.Getenv("OLLAMA_CLOUD_API")
	if apiKey == "" {
		return nil, fmt.Errorf("OLLAMA_CLOUD_API not set in environment")
	}

	modelName := cfg.Model
	if modelName == "" {
		modelName = config.Getenv("OLLAMA_MODEL")
		if modelName == "" {
			modelName = "gemma4" // default fallback
		}
	}

	// Prepare the prompt formatting
	var formatJson string
	if cfg.TipeSoal == "mencocokkan" {
		if cfg.TanpaGambar {
			formatJson = `[
  {
    "pertanyaan": "Instruksi soal mencocokkan terkait topik ini",
    "pasangan_item": [
      {"kiri": "teks item kiri 1", "kanan": "teks item kanan 1", "kiri_is_image": false, "kanan_is_image": false},
      {"kiri": "teks item kiri 2", "kanan": "teks item kanan 2", "kiri_is_image": false, "kanan_is_image": false},
      {"kiri": "teks item kiri 3", "kanan": "teks item kanan 3", "kiri_is_image": false, "kanan_is_image": false},
      {"kiri": "teks item kiri 4", "kanan": "teks item kanan 4", "kiri_is_image": false, "kanan_is_image": false},
      {"kiri": "teks item kiri 5", "kanan": "teks item kanan 5", "kiri_is_image": false, "kanan_is_image": false}
    ],
    "tipe_soal": "mencocokkan",
    "tanpa_gambar": true,
    "image_prompt": ""
  }
]`
		} else {
			formatJson = `[
  {
    "pertanyaan": "Cocokkan nama benda atau kata di sebelah kiri dengan gambarnya di sebelah kanan!",
    "pasangan_item": [
      {"kiri": "nama/kata 1", "kanan": "nama/kata 1 (sama dengan kiri)", "kiri_is_image": false, "kanan_is_image": true, "kiri_prompt": "", "kanan_prompt": "deskripsi ilustrasi kartun anak untuk nama/kata 1, simple clipart style"},
      {"kiri": "nama/kata 2", "kanan": "nama/kata 2 (sama dengan kiri)", "kiri_is_image": false, "kanan_is_image": true, "kiri_prompt": "", "kanan_prompt": "deskripsi ilustrasi kartun anak untuk nama/kata 2, simple clipart style"},
      {"kiri": "nama/kata 3", "kanan": "nama/kata 3 (sama dengan kiri)", "kiri_is_image": false, "kanan_is_image": true, "kiri_prompt": "", "kanan_prompt": "deskripsi ilustrasi kartun anak untuk nama/kata 3, simple clipart style"},
      {"kiri": "nama/kata 4", "kanan": "nama/kata 4 (sama dengan kiri)", "kiri_is_image": false, "kanan_is_image": true, "kiri_prompt": "", "kanan_prompt": "deskripsi ilustrasi kartun anak untuk nama/kata 4, simple clipart style"},
      {"kiri": "nama/kata 5", "kanan": "nama/kata 5 (sama dengan kiri)", "kiri_is_image": false, "kanan_is_image": true, "kiri_prompt": "", "kanan_prompt": "deskripsi ilustrasi kartun anak untuk nama/kata 5, simple clipart style"}
    ],
    "tipe_soal": "mencocokkan",
    "tanpa_gambar": false,
    "image_prompt": ""
  }
]`
		}
	} else if cfg.TipeSoal == "lengkapi_suku_kata" {
		formatJson = fmt.Sprintf(`[
  {
    "pertanyaan": "Lengkapi suku kata di bawah ini dengan tepat!",
    "suku_kata_awal": "Sa",
    "pilihan_suku_kata": ["pa", "pi"],
    "jawaban_benar": "pi",
    "tipe_soal": "lengkapi_suku_kata",
    "tanpa_gambar": %t,
    "image_prompt": "seekor sapi hitam putih kartun lucu untuk anak, simple white background"
  }
]`, cfg.TanpaGambar)
	} else if cfg.TipeSoal == "huruf_depan" {
		formatJson = fmt.Sprintf(`[
  {
    "pertanyaan": "Tulis huruf depan yang hilang pada kotak!",
    "huruf_depan": "B",
    "sisa_kata": "ola",
    "jawaban_benar": "B",
    "tipe_soal": "huruf_depan",
    "tanpa_gambar": %t,
    "image_prompt": "sebuah bola sepak kartun berwarna, simple white background"
  }
]`, cfg.TanpaGambar)
	} else if cfg.TipeSoal == "lingkari_kata" {
		formatJson = fmt.Sprintf(`[
  {
    "pertanyaan": "Lingkari kata sesuai gambar berikut!",
    "opsi_kata": ["Sapu", "Saku", "Suka"],
    "jawaban_benar": "Sapu",
    "tipe_soal": "lingkari_kata",
    "tanpa_gambar": %t,
    "image_prompt": "sebuah sapu ijuk kuning bersih kartun untuk anak, simple white background"
  }
]`, cfg.TanpaGambar)
	} else if cfg.TipeSoal == "susun_kata" {
		formatJson = fmt.Sprintf(`[
  {
    "pertanyaan": "Susunlah huruf-huruf ini menjadi kata yang tepat!",
    "huruf_acak": "L P E A",
    "jumlah_huruf": 4,
    "jawaban_benar": "APEL",
    "tipe_soal": "susun_kata",
    "tanpa_gambar": %t,
    "image_prompt": "buah apel merah segar kartun anak, simple white background"
  }
]`, cfg.TanpaGambar)
	} else if cfg.TipeSoal == "drill_matematika" {
		formatJson = `[
  {
    "pertanyaan": "Berhitung Penjumlahan / Pengurangan Cepat",
    "math_blocks": [
      {"judul_blok": "Kotak 1", "items": ["1 + 2 =", "2 + 4 =", "4 + 6 =", "7 + 8 =", "9 + 1 ="]},
      {"judul_blok": "Kotak 2", "items": ["3 + 3 =", "2 + 2 =", "4 + 5 =", "6 + 3 =", "7 + 9 ="]},
      {"judul_blok": "Kotak 3", "items": ["8 + 4 =", "6 + 2 =", "2 + 3 =", "7 + 1 =", "5 + 6 ="]},
      {"judul_blok": "Kotak 4", "items": ["5 + 1 =", "8 + 4 =", "4 + 3 =", "7 + 5 =", "6 + 1 ="]},
      {"judul_blok": "Kotak 5", "items": ["1 + 9 =", "8 + 2 =", "7 + 6 =", "6 + 6 =", "5 + 9 ="]},
      {"judul_blok": "Kotak 6", "items": ["2 + 9 =", "3 + 4 =", "7 + 4 =", "1 + 3 =", "4 + 1 ="]}
    ],
    "tipe_soal": "drill_matematika",
    "tanpa_gambar": true,
    "image_prompt": ""
  }
]`
	} else if cfg.TipeSoal == "benar_salah" {
		formatJson = fmt.Sprintf(`[
  {
    "pertanyaan": "teks pernyataan atau fakta terkait topik",
    "jawaban_benar": "Benar",
    "opsi": ["Benar", "Salah"],
    "tipe_soal": "benar_salah",
    "tanpa_gambar": %t,
    "image_prompt": "deskripsi gambar ilustrasi style kartun (jika tanpa visual set kosong)"
  }
]`, cfg.TanpaGambar)
	} else if cfg.TipeSoal == "isian_singkat" {
		formatJson = fmt.Sprintf(`[
  {
    "pertanyaan": "teks pertanyaan isian singkat",
    "jawaban_benar": "jawaban harus singkat 1 atau 2 kata yang valid",
    "opsi": [],
    "tipe_soal": "isian_singkat",
    "tanpa_gambar": %t,
    "image_prompt": "deskripsi gambar ilustrasi style kartun (jika tanpa visual set kosong)"
  }
]`, cfg.TanpaGambar)
	} else {
		formatJson = fmt.Sprintf(`[
  {
    "pertanyaan": "teks pertanyaan",
    "jawaban_benar": "jawaban yang benar (salah satu dari opsi)",
    "opsi": ["opsi A", "opsi B", "opsi C", "opsi D"],
    "tipe_soal": "pilihan_ganda",
    "tanpa_gambar": %t,
    "image_prompt": "deskripsi gambar ilustrasi style kartun (jika tanpa visual set kosong)"
  }
]`, cfg.TanpaGambar)
	}

	var instruksiImage string
	if cfg.TipeSoal == "mencocokkan" && !cfg.TanpaGambar {
		instruksiImage = fmt.Sprintf(`- kanan_prompt wajib deskriptif, cocok untuk ilustrasi kartun anak-anak (SD)
- kiri berisi kata atau teks, kanan_is_image selalu true untuk soal ilustrasi
- Pastikan pasangan_item berisi tepat %d item`, cfg.JumlahSoal)
	} else if cfg.TipeSoal == "drill_matematika" || cfg.TanpaGambar {
		instruksiImage = `- MENGABAIKAN image_prompt (wajib isi string kosong "")`
	} else {
		instruksiImage = `- image_prompt harus deskriptif dan cocok untuk kartun anak-anak, berlatar belakang putih bersih (white background clipart)`
	}

	jumlahSoalPrompt := fmt.Sprintf("Buat %d soal berjenis \"%s\"", cfg.JumlahSoal, cfg.TipeSoal)
	if cfg.TipeSoal == "mencocokkan" {
		jumlahSoalPrompt = fmt.Sprintf("Buat 1 soal berjenis \"%s\" yang memuat tepat %d pasangan item di dalamnya", cfg.TipeSoal, cfg.JumlahSoal)
	} else if cfg.TipeSoal == "drill_matematika" {
		jumlahSoalPrompt = "Buat 1 set drill matematika yang memuat 6 math_blocks (masing-masing 5 soal aritmatika)"
	}

	prompt := fmt.Sprintf(`Kamu adalah guru ahli untuk anak TK dan SD kelas %d di Indonesia.
%s tentang topik: "%s".

PENTING: Balas HANYA dengan JSON array valid. Format response harus berupa JSON array of objects.

Format setiap soal seperti ini:
%s

Pastikan:
- KUNCI JAWABAN (jawaban_benar / pasangan / kata) HARUS 100%% AKURAT, VALID, DAN SESUAI FAKTA NYATA. Dilarang memberikan kata/jawaban yang salah!
- Tingkat kesulitan sesuai level kelas %d SD / TK
- Bila lengkapi_suku_kata: kata terdiri dari 2 suku kata (CV-CV seperti Sa-pi, Ku-da, Ru-sa, Ba-ju). Pilihan suku kata berisi 2 opsi (1 benar, 1 pengecoh mirip).
- Bila huruf_depan: objek kata konkret (seperti Bola, Api, Ceri, Durian). huruf_depan huruf ke-1, sisa_kata huruf ke-2 sampai akhir.
- Bila lingkari_kata: sediakan 3 kata mirip (opsi_kata) yang satu adalah jawaban_benar sesuai gambar.
- Bila susun_kata: acak huruf kata secara jelas (huruf_acak diberi spasi seperti "L P E A", jawaban_benar huruf kapital "APEL", jumlah_huruf 4).
- Bila drill_matematika: sediakan tepat 6 blok, tiap blok berisi 5 soal aritmatika tingkat kelas %d.
- Bila pilihan_ganda atau benar_salah, jawaban_benar HARUS identik persis dengan yang ada di dalam array opsi
- Bila isian_singkat, buat array opsi MENJADI KOSONG []
- Bila mencocokkan, pasangan kiri dan kanan harus tepat dan bersesuaian.
%s
- Hanya output JSON array, tidak ada teks lain`, cfg.Kelas, jumlahSoalPrompt, cfg.Topik, formatJson, cfg.Kelas, cfg.Kelas, instruksiImage)

	log.Printf("[Ollama] Preparing request for model: %s, topic: %s", modelName, cfg.Topik)

	// Call Ollama Cloud API
	ollamaReq := OllamaRequest{
		Model:  modelName,
		Prompt: prompt,
		Format: "json",
		Stream: false,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://ollama.com/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	// Clean raw text
	rawText := strings.TrimSpace(ollamaResp.Response)
	rawText = strings.TrimPrefix(rawText, "```json")
	rawText = strings.TrimPrefix(rawText, "```")
	rawText = strings.TrimSuffix(rawText, "```")
	rawText = strings.TrimSpace(rawText)

	// Parse into our raw list
	var rawList []ollamaSoal
	err = json.Unmarshal([]byte(rawText), &rawList)
	if err != nil {
		log.Println("JSON parse error:", err)
		log.Println("Raw response:", rawText)
		return nil, fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	// Convert rawList -> models.Soal and generate images
	soalList := make([]models.Soal, 0, len(rawList))
	for _, raw := range rawList {
		soal := models.Soal{
			Pertanyaan:      raw.Pertanyaan,
			JawabanBenar:    raw.JawabanBenar,
			Opsi:            raw.Opsi,
			TipeSoal:        raw.TipeSoal,
			TanpaGambar:     cfg.TanpaGambar,
			ImagePrompt:     raw.ImagePrompt,
			SukuKataAwal:    raw.SukuKataAwal,
			PilihanSukuKata: raw.PilihanSukuKata,
			HurufDepan:      raw.HurufDepan,
			SisaKata:        raw.SisaKata,
			OpsiKata:        raw.OpsiKata,
			HurufAcak:       raw.HurufAcak,
			JumlahHuruf:     raw.JumlahHuruf,
			MathBlocks:      raw.MathBlocks,
		}

		if soal.TipeSoal == "" {
			soal.TipeSoal = cfg.TipeSoal
		}

		if cfg.TipeSoal == "drill_matematika" {
			soal.TanpaGambar = true
		}

		// Handle image generation for types with images
		if !cfg.TanpaGambar && cfg.TipeSoal != "mencocokkan" && cfg.TipeSoal != "drill_matematika" && raw.ImagePrompt != "" {
			soal.ImageURL = GenerateImageURL(raw.ImagePrompt, cfg.ImageModel)
		}

		// Handle matching items
		if cfg.TipeSoal == "mencocokkan" && len(raw.PasanganItem) > 0 {
			pairs := make([]models.MatchingPair, 0, len(raw.PasanganItem))
			for _, p := range raw.PasanganItem {
				pair := models.MatchingPair{
					Kiri:         p.Kiri,
					Kanan:        p.Kanan,
					KiriIsImage:  p.KiriIsImage,
					KananIsImage: p.KananIsImage,
					KiriPrompt:   p.KiriPrompt,
					KananPrompt:  p.KananPrompt,
				}
				// Generate image URLs for image-type items
				if !cfg.TanpaGambar {
					if p.KiriIsImage && p.KiriPrompt != "" {
						pair.KiriURL = GenerateImageURL(p.KiriPrompt, cfg.ImageModel)
					}
					if p.KananIsImage && p.KananPrompt != "" {
						pair.KananURL = GenerateImageURL(p.KananPrompt, cfg.ImageModel)
					}
				}
				pairs = append(pairs, pair)
			}
			soal.PasanganItem = pairs
		}

		soalList = append(soalList, soal)
	}

	return soalList, nil
}
