#  AjarVisual Backend API

> REST API untuk platform generator soal otomatis berbasis AI  dibangun dengan Go, Gin, dan TiDB Cloud.

## Tech Stack

| Layer | Teknologi | Keterangan |
|-------|-----------|------------|
| Language | **Go 1.22+** | Performa tinggi, concurrency native |
| Framework | **Gin** | HTTP router paling populer di ekosistem Go |
| ORM | **GORM** | Abstraksi database dengan auto-migrate |
| Database | **TiDB Cloud** | Distributed MySQL-compatible cloud DB |
| AI - Teks | **Gemini 1.5 Flash** | Google GenAI untuk generate soal |
| AI - Gambar | **Pollinations.ai** | Free image generation, zero API key |
| Config | **godotenv** | Load .env file |

---

## Struktur Direktori

```
backend/
 main.go                  # Entry point, Gin server, routes
 .env                     # Environment variables (jangan di-commit!)
 go.mod
 go.sum
 config/
    db.go                # TiDB Cloud connection + auto-migrate
 models/
    worksheet.go         # Struct Worksheet & Soal, custom JSON scanner
 services/
    gemini.go            # Integrasi Gemini 1.5 Flash API
    pollinations.go      # Generator URL gambar Pollinations.ai
 handlers/
     worksheet.go         # Handler untuk semua endpoint API
```

---

## Setup & Instalasi

### Prerequisites
- **Go 1.22+**  [download](https://go.dev/dl/)
- Akun **TiDB Cloud**  [tidbcloud.com](https://tidbcloud.com)
- **Gemini API Key**  [Google AI Studio](https://aistudio.google.com)

### 1. Clone & masuk ke folder backend

```bash
cd backend
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Buat file `.env`

```env
GEMINI_API_KEY=your_gemini_api_key_here

TIDB_HOST=gateway01.ap-southeast-1.prod.aws.tidbcloud.com
TIDB_PORT=4000
TIDB_USER=your_tidb_username
TIDB_PASSWORD=your_tidb_password
TIDB_DATABASE=test

PORT=8080
FRONTEND_URL=http://localhost:3000
```

### 4. Jalankan server

```bash
go run main.go
```

Server berjalan di **http://localhost:8080**

---

## API Endpoints

### Base URL: `http://localhost:8080/api`

#### `GET /health`
Health check  pastikan server berjalan.

**Response:**
```json
{
  "status": "ok",
  "message": "AjarVisual API is running"
}
```

---

#### `POST /generate`
Generate worksheet soal baru menggunakan Gemini AI.

**Request Body:**
```json
{
  "topik": "Penjumlahan Buah",
  "kelas": 3,
  "jumlah_soal": 5
}
```

| Field | Type | Validasi | Keterangan |
|-------|------|----------|------------|
| `topik` | string | required | Topik materi yang ingin dibuat soal |
| `kelas` | int | 16 | Tingkat kelas SD |
| `jumlah_soal` | int | 510 | Jumlah soal yang diinginkan |

**Response:**
```json
{
  "message": "Worksheet berhasil dibuat!",
  "worksheet": {
    "id": 1,
    "judul_materi": "Penjumlahan Buah",
    "tingkat_kelas": 3,
    "data_soal": [
      {
        "pertanyaan": "Berapa jumlah apel di bawah ini?",
        "jawaban_benar": "5",
        "opsi": ["3", "4", "5", "6"],
        "image_prompt": "5 red apples on a wooden table, cartoon style for kids",
        "image_url": "https://image.pollinations.ai/prompt/..."
      }
    ],
    "created_at": "2026-04-06T07:23:48Z"
  }
}
```

---

#### `GET /history`
Ambil semua worksheet yang tersimpan, diurutkan dari terbaru.

**Response:** `[]Worksheet`

---

#### `GET /history/:id`
Ambil satu worksheet berdasarkan ID.

---

#### `DELETE /history/:id`
Hapus worksheet berdasarkan ID.

---

#### `POST /regenerate-image`
Generate ulang URL gambar untuk soal tertentu.

**Request Body:**
```json
{
  "image_prompt": "5 red apples on wooden table, cartoon style"
}
```

**Response:**
```json
{
  "image_url": "https://image.pollinations.ai/prompt/..."
}
```

---

## Database Schema

```sql
CREATE TABLE worksheets (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    judul_materi VARCHAR(255) NOT NULL,
    tingkat_kelas INT DEFAULT 1,
    data_soal   JSON NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Schema di-generate otomatis saat pertama kali server jalan (`AutoMigrate`).

---

## Cara Kerja Integrasi AI

### Gemini 1.5 Flash
1. Menerima input `topik`, `kelas`, `jumlah_soal` dari request
2. Membangun prompt yang memaksa Gemini untuk output **JSON murni** (tanpa markdown)
3. Melakukan parsing & validasi response JSON
4. Jika parsing gagal, error dikembalikan ke client

### Pollinations.ai
- **Zero API Key**  cukup format URL dengan prompt yang di-encode
- Setiap soal mendapat gambar unik berdasarkan `image_prompt`
- Seed digenerate dari hash prompt untuk konsistensi

```go
// Contoh URL yang dihasilkan
https://image.pollinations.ai/prompt/5%20red%20apples%2C%20cartoon%20style?width=512&height=512&nologo=true&seed=12345
```

---

## Tips Development

```bash
# Build binary untuk production
go build -o ajarvisual-api main.go

# Cek compile error
go vet ./...

# Format kode
gofmt -w .

# Lihat semua dependency
go list -m all
```

---

## Troubleshooting

| Masalah | Solusi |
|---------|--------|
| `Failed to connect to TiDB` | Cek kredensial di `.env`, pastikan IP tidak diblokir firewall |
| `failed to parse Gemini response` | Gemini kadang menambahkan markdown  sudah di-handle dengan `strings.TrimPrefix` |
| Port conflict | Ganti `PORT=8080` di `.env` |
| DNS resolution failed | Coba `nslookup gateway01.ap-southeast-1.prod.aws.tidbcloud.com` |

---

## Kontributor

**Isnanto Budi**  Backend Engineer

> *"Generated by AjarVisual AI by Isnanto Budi"*
