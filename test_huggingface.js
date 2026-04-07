const fs = require('fs');
const path = require('path');

// Membaca file .env secara manual agar tidak perlu install library dotenv
function loadEnv() {
  try {
    const envPath = path.join(__dirname, '.env');
    const envFile = fs.readFileSync(envPath, 'utf8');
    envFile.split('\n').forEach(line => {
      const match = line.match(/^([^=]+)=(.*)$/);
      if (match) {
        process.env[match[1].trim()] = match[2].trim();
      }
    });
  } catch (error) {
    console.log("⚠️ File .env tidak ditemukan di folder backend.");
  }
}

loadEnv();

const HF_TOKEN = process.env.HF_TOKEN;

if (!HF_TOKEN) {
  console.error("❌ Token HF_TOKEN tidak ditemukan di dalam .env!");
  process.exit(1);
}

// Menggunakan model tercepat dari Hugging Face
const MODEL_ID = 'black-forest-labs/FLUX.1-schnell';
const API_URL = `https://router.huggingface.co/hf-inference/models/${MODEL_ID}`;

async function generateImage(prompt) {
  console.log(`🚀 Meminta image untuk prompt: "${prompt}"...`);
  
  try {
    const response = await fetch(API_URL, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${HF_TOKEN}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ inputs: prompt }), 
    });

    if (!response.ok) {
      const errorText = await response.text();
      
      try {
        const errorJson = JSON.parse(errorText);
        // Menangani error cold start model
        if (response.status === 503 && errorJson.estimated_time) {
          const waitTime = errorJson.estimated_time;
          console.log(`[Info] ⏳ Model sedang dimuat. Harap tunggu sekitar ${waitTime.toFixed(1)} detik...`);
          await new Promise(resolve => setTimeout(resolve, waitTime * 1000));
          return generateImage(prompt); // retry
        }
      } catch (e) {}

      throw new Error(`API Error HTTP ${response.status}: ${errorText}`);
    }

    const arrayBuffer = await response.arrayBuffer();
    return Buffer.from(arrayBuffer);

  } catch (error) {
    console.error("❌ Gagal saat memanggil API:", error.message);
    throw error;
  }
}

async function main() {
  const prompt = "A cute futuristic robot cat reading a book in a glowing cyberpunk library, highly detailed, 4k";
  const outputPath = path.join(__dirname, "hasil_generate.jpg");

  try {
    const imageBuffer = await generateImage(prompt);
    fs.writeFileSync(outputPath, imageBuffer);
    console.log(`✅ Sukses! Gambar berhasil disimpan sebagai: ${outputPath}`);
  } catch (error) {
    console.log("❌ Proses terhenti karena error.");
  }
}

main();
