#!/usr/bin/env bash
# ============================================================================
#  loadgen.sh - PEMBANGKIT BEBAN (LOAD GENERATOR) untuk Linux / macOS
# ----------------------------------------------------------------------------
#  Script ini "menembakkan" banyak request ke aplikasi agar grafik di Datadog
#  bergerak. Pakai ini saat merekam screencast.
#
#  CARA PAKAI:
#    chmod +x loadgen.sh         # (sekali saja) beri izin eksekusi
#    ./loadgen.sh                # jalankan skenario default (campuran)
#    ./loadgen.sh normal         # hanya request normal (RPS naik, latency rendah)
#    ./loadgen.sh slow           # request lambat (latency p95/p99 naik)
#    ./loadgen.sh error          # request gagal (error rate naik)
#    ./loadgen.sh checkout       # memicu distributed tracing
#
#  Hentikan dengan menekan Ctrl + C.
# ============================================================================

# Alamat aplikasi. Ganti bila perlu.
BASE_URL="${BASE_URL:-http://localhost:8080}"

# Skenario yang dipilih (argumen pertama), default "mix".
SCENARIO="${1:-mix}"

echo "Load generator -> $BASE_URL  (skenario: $SCENARIO)"
echo "Tekan Ctrl+C untuk berhenti."
echo

# Fungsi menembak satu request dan menampilkan kode status singkat.
hit() {
  # -s diam, -o buang body, -w tampilkan kode status
  code=$(curl -s -o /dev/null -w "%{http_code}" "$1")
  echo "  $code  <-  $1"
}

# Loop tak terbatas sampai Ctrl+C.
while true; do
  case "$SCENARIO" in
    normal)
      hit "$BASE_URL/work"
      ;;
    slow)
      hit "$BASE_URL/work?slow=1"
      ;;
    error)
      hit "$BASE_URL/work?fail=1"
      ;;
    checkout)
      hit "$BASE_URL/checkout"
      ;;
    mix|*)
      # Campuran realistis: kebanyakan normal, sebagian lambat & error.
      hit "$BASE_URL/work"
      hit "$BASE_URL/work"
      hit "$BASE_URL/work?slow=1"
      hit "$BASE_URL/work?fail=1"
      hit "$BASE_URL/checkout"
      ;;
  esac
  # Jeda singkat agar tidak membanjiri terlalu cepat. Perkecil untuk RPS lebih tinggi.
  sleep 0.2
done
