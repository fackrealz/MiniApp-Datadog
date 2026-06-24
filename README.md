# Mini App & Instrumentasi Datadog — Panduan Lengkap untuk Pemula

Aplikasi web kecil ini dibuat untuk **Tugas L1 Monitoring — Komponen 2 (Mini Application & Instrumentasi)**, dan dirancang agar langsung **terkorelasi dengan Komponen 1 (Eksplorasi & Pemahaman Datadog)**.

Satu aplikasi ini bisa mendemonstrasikan **keempat objektif** sekaligus:

| Objektif (Komponen 1) | Dibuktikan oleh | Cara memicu |
|---|---|---|
| **RPS** (Requests Per Second) | endpoint `/work` | tembak banyak request (pakai `loadgen`) |
| **Latency** (p50/p95/p99) | endpoint `/work?slow=1` | request lambat 0.8–2 detik |
| **Error Rate** | endpoint `/work?fail=1` | request balas HTTP 500 |
| **Distributed Tracing** | endpoint `/checkout` | satu request memunculkan beberapa span bertingkat |

> Kamu **tidak perlu paham Go** untuk menjalankannya. Semua berjalan di dalam Docker. Tapi setiap baris kode di `main.go` sudah diberi komentar Bahasa Indonesia agar kamu bisa menjelaskannya saat presentasi.

---

## Daftar Isi
1. Apa yang kamu butuhkan
2. Penjelasan tiap file
3. Langkah menjalankan (paling penting)
4. Membuktikan tiap objektif di Datadog
5. Menghubungkan ke dashboard & alert (Komponen 1)
6. Pertanyaan yang mungkin ditanya penilai
7. Troubleshooting

---

## 1. Apa yang kamu butuhkan

- **Docker Desktop** terpasang dan berjalan (Windows/Mac) atau Docker Engine (Linux). Ini satu-satunya hal yang wajib diinstall.
- **Akun Datadog** (free trial 14 hari sudah cukup).
- **API Key** dari Datadog: buka *Organization Settings → API Keys → New Key*, lalu salin.
- Mengetahui **site/region** akunmu (lihat alamat browser saat login, mis. `app.datadoghq.com` = site `datadoghq.com`, `ap1.datadoghq.com` = site `ap1.datadoghq.com`).

---

## 2. Penjelasan tiap file

| File | Fungsi |
|---|---|
| `main.go` | Kode aplikasi Go. Inti instrumentasi ada di sini (berkomentar lengkap). |
| `go.mod` | Daftar dependensi (library Datadog). |
| `Dockerfile` | Resep membungkus aplikasi menjadi container. |
| `docker-compose.yml` | Menjalankan Agent + aplikasi sekaligus. |
| `.env.example` | Contoh tempat menaruh API Key (salin jadi `.env`). |
| `loadgen.sh` / `loadgen.ps1` | Pembangkit beban untuk menggerakkan grafik (Linux/Mac & Windows). |
| `.gitignore` | Mencegah file rahasia (`.env`) ikut ter-push ke Git. |

---

## 3. Langkah menjalankan (ikuti berurutan)

**Langkah 1 — Siapkan file rahasia.** Di dalam folder proyek:

```bash
cp .env.example .env
```

Buka `.env`, isi `DD_API_KEY` dengan API Key kamu, dan sesuaikan `DD_SITE`.

**Langkah 2 — Nyalakan semuanya dengan satu perintah:**

```bash
docker compose up --build
```

Perintah ini akan: mengompilasi aplikasi Go, menyalakan Datadog Agent, lalu menjalankan aplikasi. Tunggu sampai muncul log `listening on :8080`.

**Langkah 3 — Cek aplikasi hidup.** Buka browser ke `http://localhost:8080`. Kamu akan melihat daftar endpoint.

**Langkah 4 — Bangkitkan beban.** Buka terminal **baru** (biarkan yang tadi tetap jalan):

```bash
# Linux / macOS:
chmod +x loadgen.sh
./loadgen.sh

# Windows (PowerShell):
.\loadgen.ps1
```

**Langkah 5 — Lihat data di Datadog.** Buka Datadog → **APM → Services**. Dalam 1–2 menit service bernama `mini-app` akan muncul lengkap dengan grafik.

Untuk berhenti: tekan `Ctrl+C` di terminal, lalu `docker compose down`.

---

## 4. Membuktikan tiap objektif di Datadog

Jalankan skenario spesifik agar tiap sinyal terlihat jelas saat merekam video:

```bash
./loadgen.sh normal     # RPS naik, latency rendah, 0% error
./loadgen.sh slow       # latency p95/p99 melonjak
./loadgen.sh error      # error rate naik (HTTP 500)
./loadgen.sh checkout   # distributed tracing (span bertingkat)
```

**RPS** — di APM → Services → `mini-app`, lihat grafik *Requests* / *Throughput*. Saat loadgen jalan, garis naik.

**Latency** — di grafik *Latency*, tampilkan p50, p95, p99. Jalankan skenario `slow` dan tunjukkan p99 melonjak jauh di atas p50.

**Error Rate** — di grafik *Errors*. Jalankan skenario `error`; persentase error naik. Inilah pemicu alert kamu nanti.

**Distributed Tracing** — APM → Traces → klik salah satu trace `/checkout`. Kamu akan melihat *flame graph* dengan span: `checkout.request` → `validate.order` → `db.save_order` (paling lebar = bottleneck) → `payment.charge`. Tunjuk span database sebagai bottleneck saat presentasi.

---

## 5. Menghubungkan ke dashboard & alert (Komponen 1)

Aplikasi ini adalah "alat peraga" untuk dashboard dan alert yang kamu buat di Komponen 1:

- **Dashboard:** buat widget Timeseries memakai metrik `trace.http.request.hits` (RPS) dan `trace.http.request.duration` (latency p50/p95/p99) dengan filter `service:mini-app`. Tambah Query Value untuk error rate.
- **Alert (Monitor):** buat monitor yang menyala bila error rate `service:mini-app` melewati ambang (mis. > 5% selama 5 menit). Saat kamu jalankan `./loadgen.sh error`, monitor akan berubah merah dan mengirim notifikasi ke Telegram/Teams.

Dengan begitu, video screencast kamu bisa menunjukkan alur penuh: aplikasi jalan → grafik bergerak → alert menyala → notifikasi masuk.

---

## 6. Pertanyaan yang mungkin ditanya penilai (siapkan jawaban)

- **"Bagaimana aplikasi mengirim data ke Datadog?"** Lewat library `dd-trace-go`. `httptrace.NewServeMux` otomatis membungkus tiap request HTTP menjadi span dan mengirimkannya ke Datadog Agent di port 8126; Agent meneruskannya ke Datadog cloud.
- **"Apa beda trace dan span?"** Trace = seluruh perjalanan satu request. Span = satu langkah di dalamnya. Lihat `/checkout` yang punya beberapa span.
- **"Kenapa pakai p99, bukan rata-rata?"** Rata-rata menyembunyikan request paling lambat. p99 menangkap 1% terburuk yang justru paling perlu dideteksi.
- **"Di mana bottleneck pada /checkout?"** Pada span `db.save_order` yang sengaja paling lambat — terlihat paling lebar di flame graph.

---

## 7. Troubleshooting

- **Service tidak muncul di APM** → pastikan `DD_API_KEY` & `DD_SITE` di `.env` benar; tunggu 1–2 menit; pastikan loadgen sudah dijalankan (tanpa request, tak ada data).
- **`docker compose` tidak dikenal** → pakai Docker versi baru, atau coba `docker-compose` (pakai tanda hubung).
- **Port 8080 sudah dipakai** → ubah pemetaan port di `docker-compose.yml`, mis. `"9090:8080"`, lalu akses `http://localhost:9090`.
- **Error rate tidak naik** → pastikan menjalankan skenario `error`, dan lihat grafik dengan rentang waktu *Past 15 Minutes*.

---

### Catatan kejujuran akademik
Kode ini boleh dikembangkan dengan bantuan AI coding assistant (sesuai aturan tugas), **tetapi pastikan kamu memahami tiap baris** — komentar di `main.go` dibuat untuk itu. Sebutkan nama anggota kelompok bila dikerjakan bersama saat pengumpulan.
