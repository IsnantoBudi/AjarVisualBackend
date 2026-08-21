# 🚀 AjarVisual Backend API

> REST API berkinerja tinggi untuk platform generator soal otomatis & Lembar Kerja Peserta Didik (LKPD) berbasis AI interaktif. Dibangun dengan Go 1.22+, mendukung eksekusi native maupun **Serverless WebAssembly di Cloudflare Workers**.

---

## 🛠️ Tech Stack & Integrasi

| Komponen | Teknologi | Keterangan |
| :--- | :--- | :--- |
| **Language & Runtime** | **Go 1.22+** / **WebAssembly (Wasm)** | Native Go untuk server/VPS atau Wasm untuk Cloudflare Workers |
| **HTTP Framework** | **Gin** / Custom Go Handler | Routing REST API berlatensi rendah |
| **Database** | **Cloudflare D1** & **TiDB Cloud / MySQL** | Penyimpanan riwayat worksheet dan butir soal |
| **LLM - Teks & Soal** | **Ollama Cloud** (`gemma4`, `minimax-m2.5`) | Generasi butir soal, anagram kata, dan drill hitung |
| **AI - Ilustrasi Gambar** | **Hugging Face FLUX.1-schnell** & **Pollinations.ai** | Generasi gambar edukatif kartun berlatar putih bersih |
| **Deployment Engine** | **Cloudflare Workers (Wrangler)** / **Docker** | Serverless Edge Computing global |

---

## 📂 Struktur Direktori

```text
backend/
├── config/             # Manajemen konfigurasi environment & database
├── handlers/           # Controller handler endpoint REST API
├── models/             # Data struct Worksheet, Soal, LKPD, dan MathBlock
├── services/           # Integrasi Ollama Cloud, Hugging Face, & Pollinations
├── build.js            # Script kompilasi Go ke WebAssembly (main.wasm)
├── index.js            # Entry point Cloudflare Workers wrapper
├── main.go             # Entry point native Go HTTP server
├── wrangler.toml       # Konfigurasi deployment Cloudflare Workers & D1
└── .env                # Environment variables (JANGAN di-commit)
```

---

## ⚙️ Konfigurasi Environment (`.env`)

Buat berkas `.env` di dalam folder `backend/`:

```env
# ── Database (TiDB Cloud / MySQL / Cloudflare D1) ──
TIDB_HOST=gateway01.ap-southeast-1.prod.aws.tidbcloud.com
TIDB_PORT=4000
TIDB_USER=your_tidb_user
TIDB_PASSWORD=your_tidb_password
TIDB_DATABASE=test

# ── Server & CORS ──
PORT=8080
FRONTEND_URL=http://localhost:3000,https://ajar-visual.vercel.app,https://isnantobudi.online
BACKEND_URL=https://ajarvisual.isnantobudi.online

# ── AI API Keys ──
OLLAMA_CLOUD_API=your_ollama_cloud_api_key
OLLAMA_MODEL=gemma4
HF_TOKEN=your_huggingface_token
```

---

## 💻 Panduan Menjalankan Secara Lokal (Local Development)

### 1. Prasyarat
- **Go 1.22+** ([Unduh Go](https://go.dev/dl/))
- **Node.js 18+** & npm (untuk build Wasm jika diperlukan)

### 2. Jalankan Server Go Lokal
```bash
cd backend
go mod tidy
go run .
```
Server lokal akan aktif di `http://localhost:8080`.

---

## 🚀 Panduan Deployment

Backend AjarVisual mendukung beberapa opsi deployment sesuai kebutuhan infrastruktur:

### Opsi 1: Cloudflare Workers (Rekomendasi / Production Saat Ini)
Deploy backend sebagai serverless edge WebAssembly menggunakan Cloudflare Workers dan database Cloudflare D1:

1. **Login ke Cloudflare via CLI**:
   ```bash
   npx wrangler login
   ```
2. **Deploy ke Cloudflare Workers**:
   ```bash
   npx wrangler deploy
   ```
   > Script `build.js` akan otomatis mengompilasi kode Go menjadi `main.wasm` dan mengunggahnya ke Cloudflare Workers.

3. **Set Environment Secrets di Cloudflare**:
   ```bash
   npx wrangler secret put OLLAMA_CLOUD_API
   npx wrangler secret put HF_TOKEN
   ```

4. **Custom Domain**:
   - Di dashboard Cloudflare Workers, hubungkan custom domain: `https://ajarvisual.isnantobudi.online`

---

### Opsi 2: VPS / Linux Server (Docker / Systemd)
Jika ingin mendeploy backend sebagai service mandiri di VPS (Ubuntu, Debian, dll.):

1. **Build Binary Native Linux**:
   ```bash
   GOOS=linux GOARCH=amd64 go build -o ajarvisual-api .
   ```
2. **Jalankan via Systemd atau PM2**:
   ```bash
   ./ajarvisual-api
   ```
3. **Konfigurasi Reverse Proxy Nginx**:
   ```nginx
   server {
       server_name ajarvisual.isnantobudi.online;
       location / {
           proxy_pass http://127.0.0.1:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
   }
   ```

---

## 📑 Daftar Endpoint REST API

| Method | Endpoint | Keterangan |
| :--- | :--- | :--- |
| `GET` | `/api/health` | Health check & verifikasi status server |
| `POST` | `/api/generate` | Generate lembar kerja LKPD baru dengan AI |
| `POST` | `/api/worksheets/:id/add-soal` | Menambahkan butir soal ke worksheet yang sudah ada |
| `GET` | `/api/history` | Mengambil seluruh riwayat lembar kerja |
| `GET` | `/api/history/:id` | Mengambil 1 lembar kerja spesifik berdasarkan ID |
| `DELETE` | `/api/history/:id` | Menghapus lembar kerja dari database |
| `POST` | `/api/regenerate-image` | Membuat ulang URL gambar untuk soal tertentu |
| `GET` | `/api/image-proxy` | Proxy image generator (Pollinations / Hugging Face FLUX) |

---

## 📝 Format 9 Tipe Lembar Kerja (LKPD) yang Didukung

1. **Lengkapi Suku Kata** (`lengkapi_suku_kata`) - *Wajib Bergambar*
2. **Tulis Huruf Depan** (`huruf_depan`) - *Wajib Bergambar*
3. **Lingkari Kata Sesuai Gambar** (`lingkari_kata`) - *Wajib Bergambar*
4. **Menyusun Kata / Anagram** (`susun_kata`) - *Wajib Bergambar*
5. **Drill Matematika (30 Baris Hitung Grid A4)** (`drill_matematika`) - *Tanpa Gambar*
6. **Pilihan Ganda (A/B/C/D)** (`pilihan_ganda`) - *Gambar Opsional*
7. **Mencocokkan Garis** (`mencocokkan`) - *Gambar Opsional*
8. **Benar / Salah** (`benar_salah`) - *Gambar Opsional*
9. **Isian Singkat** (`isian_singkat`) - *Gambar Opsional*

---

## 👨‍💻 Kontributor
- **Isnanto Budi** - Full-Stack Developer
