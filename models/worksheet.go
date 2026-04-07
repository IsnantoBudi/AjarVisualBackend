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

type Soal struct {
	Pertanyaan   string         `json:"pertanyaan"`
	JawabanBenar string         `json:"jawaban_benar,omitempty"`
	Opsi         []string       `json:"opsi,omitempty"`
	Pasangan     map[string]string `json:"pasangan,omitempty"` // legacy
	PasanganItem []MatchingPair `json:"pasangan_item,omitempty"` // new structured matching
	TipeSoal     string         `json:"tipe_soal"`
	TanpaGambar  bool           `json:"tanpa_gambar"`
	ImagePrompt  string         `json:"image_prompt,omitempty"`
	ImageURL     string         `json:"image_url,omitempty"`
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
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	JudulMateri  string    `json:"judul_materi" gorm:"type:varchar(255);not null"`
	TingkatKelas int       `json:"tingkat_kelas" gorm:"default:1"`
	DataSoal     SoalList  `json:"data_soal" gorm:"type:json;not null"`
	CreatedAt    time.Time `json:"created_at"`
}
