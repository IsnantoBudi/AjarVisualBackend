package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// MatchingPair represents a single pair in a matching question
type MatchingPair struct {
	Kiri         string `json:"kiri"`
	Kanan        string `json:"kanan"`
	KiriIsImage  bool   `json:"kiri_is_image,omitempty"`
	KananIsImage bool   `json:"kanan_is_image,omitempty"`
	KiriURL      string `json:"kiri_url,omitempty"`
	KananURL     string `json:"kanan_url,omitempty"`
	KiriPrompt   string `json:"kiri_prompt,omitempty"`
	KananPrompt  string `json:"kanan_prompt,omitempty"`
}

// MathBlock represents a column or group of math drill items
type MathBlock struct {
	JudulBlok string   `json:"judul_blok,omitempty"`
	Items     []string `json:"items"`
}

type Soal struct {
	Pertanyaan        string            `json:"pertanyaan"`
	JawabanBenar      string            `json:"jawaban_benar,omitempty"`
	Opsi              []string          `json:"opsi,omitempty"`
	Pasangan          map[string]string `json:"pasangan,omitempty"` // legacy
	PasanganItem      []MatchingPair    `json:"pasangan_item,omitempty"` // new structured matching
	TipeSoal          string            `json:"tipe_soal"`
	TanpaGambar       bool              `json:"tanpa_gambar"`
	ImagePrompt       string            `json:"image_prompt,omitempty"`
	ImageURL          string            `json:"image_url,omitempty"`
	// Field LKPD baru
	SukuKataAwal      string            `json:"suku_kata_awal,omitempty"`      // e.g. "Sa", "Ku"
	PilihanSukuKata   []string          `json:"pilihan_suku_kata,omitempty"`   // e.g. ["pa", "pi"]
	HurufDepan        string            `json:"huruf_depan,omitempty"`        // e.g. "A", "B"
	SisaKata          string            `json:"sisa_kata,omitempty"`          // e.g. "pi", "ola"
	OpsiKata          []string          `json:"opsi_kata,omitempty"`          // e.g. ["Sapu", "Saku", "Suka"]
	HurufAcak         string            `json:"huruf_acak,omitempty"`         // e.g. "L P E A"
	JumlahHuruf       int               `json:"jumlah_huruf,omitempty"`       // e.g. 4
	MathBlocks        []MathBlock       `json:"math_blocks,omitempty"`        // e.g. 6 blocks for math drill
}

type SoalList []Soal

func (s SoalList) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SoalList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan SoalList")
	}
	return json.Unmarshal(bytes, s)
}

type Worksheet struct {
	ID           uint       `json:"id"`
	JudulMateri  string     `json:"judul_materi"`
	TingkatKelas int        `json:"tingkat_kelas"`
	DataSoal     SoalList   `json:"data_soal"`
	CreatedAt    *time.Time `json:"created_at"`
}
