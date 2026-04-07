package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"ajarvisual-backend/models"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GenerateConfig struct {
	Topik       string `json:"topik"`
	Kelas       int    `json:"kelas"`
	JumlahSoal  int    `json:"jumlah_soal"`
	TipeSoal    string `json:"tipe_soal"`
	TanpaGambar bool   `json:"tanpa_gambar"`
}

// geminiMatchingPair is a temporary struct for parsing the Gemini raw output
type geminiMatchingPair struct {
	Kiri         string `json:"kiri"`
	Kanan        string `json:"kanan"`
	KiriIsImage  bool   `json:"kiri_is_image"`
	KananIsImage bool   `json:"kanan_is_image"`
	KiriPrompt   string `json:"kiri_prompt,omitempty"`
	KananPrompt  string `json:"kanan_prompt,omitempty"`
}

// geminiSoal is temporary for raw Gemini JSON parsing
type geminiSoal struct {
	Pertanyaan   string               `json:"pertanyaan"`
	JawabanBenar string               `json:"jawaban_benar,omitempty"`
	Opsi         []string             `json:"opsi,omitempty"`
	PasanganItem []geminiMatchingPair `json:"pasangan_item,omitempty"`
	TipeSoal     string               `json:"tipe_soal"`
	TanpaGambar  bool                 `json:"tanpa_gambar"`
	ImagePrompt  string               `json:"image_prompt,omitempty"`
}

func GenerateSoal(cfg GenerateConfig) ([]models.Soal, error) {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Println("Gemini client error:", err)
		return nil, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")
	model.SetTemperature(0.8)

	var formatJson string
	if cfg.TipeSoal == "mencocokkan" {
		if cfg.TanpaGambar {
			// Text-only matching
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
			// Illustration matching: left side = text/word, right side = image illustration
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
	} else if cfg.TanpaGambar {
		instruksiImage = `- MENGABAIKAN image_prompt (wajib isi string kosong "")`
	} else {
		instruksiImage = `- image_prompt harus deskriptif dan cocok untuk kartun anak-anak`
	}

	jumlahSoalPrompt := fmt.Sprintf("Buat %d soal berjenis \"%s\"", cfg.JumlahSoal, cfg.TipeSoal)
	if cfg.TipeSoal == "mencocokkan" {
		jumlahSoalPrompt = fmt.Sprintf("Buat 1 soal berjenis \"%s\" yang memuat tepat %d pasangan item di dalamnya", cfg.TipeSoal, cfg.JumlahSoal)
	}

	prompt := fmt.Sprintf(`Kamu adalah guru ahli untuk anak SD kelas %d di Indonesia.
%s tentang topik: "%s".

PENTING: Balas HANYA dengan JSON array valid, tanpa markdown, tanpa penjelasan tambahan.

Format setiap soal seperti ini:
%s

Pastikan:
- KUNCI JAWABAN (jawaban_benar / pasangan) HARUS 100%% AKURAT, VALID, DAN SESUAI FAKTA NYATA. Dilarang memberikan jawaban yang salah!
- Pertanyaan sesuai level kelas %d SD
- Bila pilihan_ganda atau benar_salah, jawaban_benar HARUS identik persis dengan yang ada di dalam array opsi
- Bila isian_singkat, buat array opsi MENJADI KOSONG []
- Bila mencocokkan, pasangan kiri dan kanan harus tepat dan bersesuaian.
%s
- Hanya output JSON, tidak ada teks lain`, cfg.Kelas, jumlahSoalPrompt, cfg.Topik, formatJson, cfg.Kelas, instruksiImage)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	rawText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Clean the response - remove markdown code blocks if present
	rawText = strings.TrimSpace(rawText)
	rawText = strings.TrimPrefix(rawText, "```json")
	rawText = strings.TrimPrefix(rawText, "```")
	rawText = strings.TrimSuffix(rawText, "```")
	rawText = strings.TrimSpace(rawText)

	// Parse into our geminiSoal intermediate structs
	var rawList []geminiSoal
	err = json.Unmarshal([]byte(rawText), &rawList)
	if err != nil {
		log.Println("JSON parse error:", err)
		log.Println("Raw response:", rawText)
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	// Convert geminiSoal -> models.Soal and generate images
	soalList := make([]models.Soal, 0, len(rawList))
	for _, raw := range rawList {
		soal := models.Soal{
			Pertanyaan:   raw.Pertanyaan,
			JawabanBenar: raw.JawabanBenar,
			Opsi:         raw.Opsi,
			TipeSoal:     raw.TipeSoal,
			TanpaGambar:  cfg.TanpaGambar,
			ImagePrompt:  raw.ImagePrompt,
		}

		if soal.TipeSoal == "" {
			soal.TipeSoal = cfg.TipeSoal
		}

		// Handle non-matching image
		if !cfg.TanpaGambar && cfg.TipeSoal != "mencocokkan" && raw.ImagePrompt != "" {
			soal.ImageURL = GenerateImageURL(raw.ImagePrompt)
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
						pair.KiriURL = GenerateImageURL(p.KiriPrompt)
					}
					if p.KananIsImage && p.KananPrompt != "" {
						pair.KananURL = GenerateImageURL(p.KananPrompt)
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
