package main

// ============================================================================
//  MINI APPLICATION & INSTRUMENTASI DATADOG
//  Tugas L1 Monitoring - Komponen 2
// ----------------------------------------------------------------------------
//  Aplikasi web kecil ini sengaja dirancang untuk MEMBUKTIKAN ke-4 objektif
//  dari Komponen 1 (Eksplorasi & Pemahaman Datadog):
//
//    1. RPS (Requests Per Second) -> endpoint /work bisa dibanjiri request
//    2. Latency (p50/p95/p99)     -> endpoint /work?slow=1 memperlambat respon
//    3. Error Rate                -> endpoint /work?fail=1 memunculkan HTTP 500
//    4. Distributed Tracing       -> /checkout memanggil beberapa span bertingkat
//                                    (validasi -> simpan DB -> panggil service bayar)
//
//  Semua request otomatis menjadi TRACE di Datadog karena kita memakai
//  httptrace.NewServeMux() dari library resmi dd-trace-go v2.
// ============================================================================

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	// Library resmi Datadog untuk Go (v2).
	// tracer  = inti tracing (start/stop, membuat span manual)
	// httptrace = membungkus net/http agar tiap request otomatis jadi span
	httptrace "github.com/DataDog/dd-trace-go/contrib/net/http/v2"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
)

// getenv mengambil environment variable, atau memakai nilai default bila kosong.
// Dipakai agar konfigurasi (nama service, env) bisa diatur dari luar tanpa
// mengubah kode.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	// ------------------------------------------------------------------
	// 1) START TRACER
	// ------------------------------------------------------------------
	// tracer.Start() menyalakan koneksi ke Datadog Agent (default localhost:8126).
	// Di Docker, alamat Agent diatur lewat env DD_AGENT_HOST (lihat docker-compose).
	// WithService  -> nama service yang muncul di APM Datadog.
	// WithEnv      -> tag lingkungan (dev/staging/prod) untuk memfilter data.
	// WithServiceVersion -> versi, berguna untuk membandingkan rilis.
	serviceName := getenv("DD_SERVICE", "mini-app")
	env := getenv("DD_ENV", "dev")

	// Catatan: untuk memastikan SEMUA trace terkirim saat demo, set environment
	// variable DD_TRACE_SAMPLE_RATE=1 (sudah diatur di docker-compose.yml).
	// Cara ini lebih sederhana & aman bagi pemula dibanding mengonfigurasi
	// sampling lewat kode.
	tracer.Start(
		tracer.WithService(serviceName),
		tracer.WithEnv(env),
		tracer.WithServiceVersion("1.0.0"),
	)
	// defer memastikan tracer dimatikan rapi saat aplikasi berhenti,
	// sehingga trace terakhir sempat terkirim ke Datadog.
	defer tracer.Stop()

	// ------------------------------------------------------------------
	// 2) ROUTER YANG SUDAH DI-INSTRUMENTASI
	// ------------------------------------------------------------------
	// httptrace.NewServeMux() menggantikan http.NewServeMux() biasa.
	// Bedanya: setiap request yang masuk OTOMATIS dibungkus menjadi sebuah
	// span (root span) sehingga muncul sebagai trace di Datadog APM,
	// lengkap dengan metrik hits (untuk RPS), duration (untuk latency),
	// dan status (untuk error rate). Inilah jembatan ke Komponen 1.
	mux := httptrace.NewServeMux(httptrace.WithService(serviceName))

	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/work", handleWork)         // demonstrasi RPS, Latency, Error
	mux.HandleFunc("/checkout", handleCheckout) // demonstrasi Distributed Tracing

	port := getenv("PORT", "8080")
	addr := ":" + port

	log.Printf("[mini-app] service=%s env=%s listening on %s", serviceName, env, addr)
	log.Printf("[mini-app] Datadog Agent host=%s", getenv("DD_AGENT_HOST", "localhost"))

	// http.ListenAndServe menjalankan server. Memakai mux yang sudah ditrace
	// di atas, jadi seluruh endpoint otomatis terpantau.
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server gagal berjalan: %v", err)
	}
}

// ============================================================================
//  HANDLER: HALAMAN UTAMA
// ============================================================================
// Menampilkan petunjuk singkat agar pengguna tahu endpoint apa saja yang ada.
func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "Mini App Datadog - Simulator Monitoring")
	fmt.Fprintln(w, "----------------------------------------")
	fmt.Fprintln(w, "GET /work             -> request normal (cepat, sukses)")
	fmt.Fprintln(w, "GET /work?slow=1      -> request lambat (menaikkan latency)")
	fmt.Fprintln(w, "GET /work?fail=1      -> request gagal (HTTP 500, menaikkan error rate)")
	fmt.Fprintln(w, "GET /checkout         -> request multi-span (distributed tracing)")
	fmt.Fprintln(w, "GET /health           -> health check")
}

// ============================================================================
//  HANDLER: HEALTH CHECK
// ============================================================================
// Endpoint ringan untuk mengecek aplikasi hidup. Berguna juga sebagai
// pembanding latency rendah di dashboard.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ============================================================================
//  HANDLER: /work  (INTI DEMONSTRASI GOLDEN SIGNALS)
// ============================================================================
// Endpoint ini bisa berperilaku tiga macam tergantung query parameter:
//
//   /work          -> cepat & sukses    (latency rendah, status 200)
//   /work?slow=1   -> lambat            (latency tinggi, status 200)
//   /work?fail=1   -> gagal             (status 500 -> error rate naik)
//
// Dengan membanjiri endpoint ini (lihat loadgen.sh), RPS pun ikut naik.
func handleWork(w http.ResponseWriter, r *http.Request) {
	// Ambil span aktif dari context request (dibuat otomatis oleh httptrace).
	// Kita pakai untuk menambahkan tag/keterangan agar trace lebih informatif.
	span, _ := tracer.SpanFromContext(r.Context())

	// --- Skenario GAGAL: menaikkan Error Rate ---
	if r.URL.Query().Get("fail") == "1" {
		// Simulasikan sedikit kerja sebelum gagal.
		time.Sleep(time.Duration(20+rand.Intn(40)) * time.Millisecond)
		if span != nil {
			// Menandai span sebagai error -> Datadog menghitungnya di error rate.
			// SetTag("error", true) menyalakan flag error; ext.ErrorMsg mengisi
			// pesannya. httptrace juga otomatis menandai error karena status 500.
			span.SetTag("error", true)
			span.SetTag(ext.ErrorMsg, "simulasi kegagalan internal")
			span.SetTag("work.scenario", "fail")
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error (simulasi)",
		})
		return
	}

	// --- Skenario LAMBAT: menaikkan Latency (p95/p99) ---
	if r.URL.Query().Get("slow") == "1" {
		// Tidur 800ms s.d. 2000ms untuk meniru operasi berat.
		delay := time.Duration(800+rand.Intn(1200)) * time.Millisecond
		time.Sleep(delay)
		if span != nil {
			span.SetTag("work.scenario", "slow")
			span.SetTag("work.delay_ms", delay.Milliseconds())
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":   "ok-but-slow",
			"delay_ms": delay.Milliseconds(),
		})
		return
	}

	// --- Skenario NORMAL: cepat & sukses ---
	// Latency kecil (10-60ms) agar p50 tetap rendah.
	time.Sleep(time.Duration(10+rand.Intn(50)) * time.Millisecond)
	if span != nil {
		span.SetTag("work.scenario", "normal")
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ============================================================================
//  HANDLER: /checkout  (DEMONSTRASI DISTRIBUTED TRACING)
// ============================================================================
// Satu request /checkout sengaja dipecah menjadi BEBERAPA SPAN bertingkat,
// meniru perjalanan request antar-service:
//
//   checkout.request                (root span, dibuat httptrace)
//     |-- validate.order            (child span)
//     |-- db.save_order             (child span - paling lambat = bottleneck)
//     |-- payment.charge            (child span)
//
// Di flame graph Datadog, span db.save_order akan tampak PALING LEBAR,
// sehingga kamu bisa menunjuk "inilah bottleneck-nya" saat presentasi.
func handleCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Langkah 1: validasi pesanan (cepat)
	validateOrder(ctx)

	// Langkah 2: simpan ke database (sengaja dibuat paling lambat)
	saveOrder(ctx)

	// Langkah 3: proses pembayaran (kadang gagal untuk variasi error)
	if err := chargePayment(ctx); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": "pembayaran gagal (simulasi)",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "checkout sukses",
		"orderId": fmt.Sprintf("ORD-%d", rand.Intn(100000)),
	})
}

// validateOrder membuat child span "validate.order".
// tracer.StartSpanFromContext mengaitkan span baru ini sebagai anak dari
// span yang sedang aktif di context (yaitu checkout.request).
func validateOrder(ctx context.Context) {
	span, _ := tracer.StartSpanFromContext(ctx, "validate.order",
		tracer.ResourceName("validate pesanan"))
	defer span.Finish() // span ditutup saat fungsi selesai -> durasinya tercatat

	time.Sleep(time.Duration(15+rand.Intn(25)) * time.Millisecond)
	span.SetTag("order.valid", true)
}

// saveOrder membuat child span "db.save_order".
// Sengaja paling lambat (300-900ms) agar menjadi bottleneck yang jelas
// terlihat di flame graph.
func saveOrder(ctx context.Context) {
	span, _ := tracer.StartSpanFromContext(ctx, "db.save_order",
		tracer.ResourceName("INSERT INTO orders"),
		tracer.SpanType(ext.SpanTypeSQL)) // ditandai sebagai operasi database
	defer span.Finish()

	delay := time.Duration(300+rand.Intn(600)) * time.Millisecond
	time.Sleep(delay)
	span.SetTag("db.rows_affected", 1)
	span.SetTag("db.delay_ms", delay.Milliseconds())
}

// chargePayment membuat child span "payment.charge".
// 10% kemungkinan gagal untuk menambah variasi data error pada trace.
func chargePayment(ctx context.Context) error {
	span, _ := tracer.StartSpanFromContext(ctx, "payment.charge",
		tracer.ResourceName("charge kartu"))

	time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)

	if rand.Float64() < 0.10 { // 10% gagal
		err := fmt.Errorf("gateway pembayaran menolak transaksi")
		// Finish dengan WithError adalah cara resmi v2 menandai span error:
		// otomatis mengisi error message, type, dan stack trace.
		span.Finish(tracer.WithError(err))
		return err
	}
	span.SetTag("payment.status", "approved")
	span.Finish()
	return nil
}

// ============================================================================
//  UTILITAS
// ============================================================================
// writeJSON menulis respon dalam format JSON beserta status code.
// Dibuat terpisah agar tidak menulis kode yang sama berulang kali.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
