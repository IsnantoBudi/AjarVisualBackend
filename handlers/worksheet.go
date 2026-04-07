package handlers

import (
"net/http"

"ajarvisual-backend/config"
"ajarvisual-backend/models"
"ajarvisual-backend/services"

"github.com/gin-gonic/gin"
)

type GenerateRequest struct {
Topik       string `json:"topik" binding:"required"`
Kelas       int    `json:"kelas" binding:"required,min=1,max=6"`
JumlahSoal  int    `json:"jumlah_soal" binding:"required,min=1,max=10"`
TipeSoal    string `json:"tipe_soal"`
TanpaGambar bool   `json:"tanpa_gambar"`
}

func GenerateWorksheet(c *gin.Context) {
var req GenerateRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
}

soalList, err := services.GenerateSoal(cfg)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate soal: " + err.Error()})
return
}

worksheet := models.Worksheet{
JudulMateri:  req.Topik,
TingkatKelas: req.Kelas,
DataSoal:     soalList,
}

if err := config.DB.Create(&worksheet).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal simpan worksheet"})
return
}

c.JSON(http.StatusOK, gin.H{
"message":   "Worksheet berhasil dibuat!",
"worksheet": worksheet,
})
}

func GetAllHistory(c *gin.Context) {
var worksheets []models.Worksheet
if err := config.DB.Order("created_at desc").Find(&worksheets).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil riwayat"})
return
}
c.JSON(http.StatusOK, worksheets)
}

func GetWorksheetByID(c *gin.Context) {
id := c.Param("id")
var worksheet models.Worksheet
if err := config.DB.First(&worksheet, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Worksheet tidak ditemukan"})
return
}
c.JSON(http.StatusOK, worksheet)
}

func DeleteWorksheet(c *gin.Context) {
id := c.Param("id")
if err := config.DB.Delete(&models.Worksheet{}, id).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal hapus worksheet"})
return
}
c.JSON(http.StatusOK, gin.H{"message": "Worksheet dihapus"})
}

func AddSoalToWorksheet(c *gin.Context) {
	id := c.Param("id")
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var worksheet models.Worksheet
	if err := config.DB.First(&worksheet, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Worksheet tidak ditemukan"})
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
	}

	newSoal, err := services.GenerateSoal(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate soal tambahan: " + err.Error()})
		return
	}

	worksheet.DataSoal = append(worksheet.DataSoal, newSoal...)

	if err := config.DB.Save(&worksheet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan soal tambahan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Soal tambahan berhasil ditambahkan!",
		"worksheet": worksheet,
	})
}

func RegenerateImage(c *gin.Context) {
	type RegenerateReq struct {
		ImagePrompt string `json:"image_prompt" binding:"required"`
	}
	var req RegenerateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	url := services.GenerateImageURL(req.ImagePrompt)
	c.JSON(http.StatusOK, gin.H{"image_url": url})
}

func ProxyImage(c *gin.Context) {
	prompt := c.Query("prompt")
	if prompt == "" {
		c.String(http.StatusBadRequest, "Prompt is required")
		return
	}

	imageData, contentType, err := services.GenerateImage(prompt)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal ambil gambar: "+err.Error())
		return
	}

	if contentType == "" {
		contentType = "image/jpeg"
	}

	c.Data(http.StatusOK, contentType, imageData)
}

