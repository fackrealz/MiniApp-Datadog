# ============================================================================
#  Dockerfile - membungkus aplikasi Go menjadi container siap jalan
# ----------------------------------------------------------------------------
#  Memakai "multi-stage build":
#    Tahap 1 (builder): mengompilasi kode Go menjadi 1 file binary.
#    Tahap 2 (runtime): menyalin binary itu ke image kecil agar ringan.
#  Pemula tidak perlu menginstall Go di laptop; semua terjadi di dalam Docker.
# ============================================================================

# ---------- Tahap 1: Build ----------
FROM golang:1.23-alpine AS builder

# Direktori kerja di dalam container.
WORKDIR /app

# Salin file modul lebih dulu agar cache Docker efektif (dependensi jarang berubah).
COPY go.mod ./

# go mod tidy mengunduh & mengunci semua dependensi (dd-trace-go dll).
# Membutuhkan koneksi internet saat build.
RUN go mod tidy

# Salin sisa kode sumber.
COPY . .

# Kompilasi menjadi binary statis bernama "server".
# CGO_ENABLED=0 membuat binary tidak bergantung pada library sistem -> portable.
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

# ---------- Tahap 2: Runtime ----------
FROM alpine:3.20

WORKDIR /app

# Salin HANYA binary hasil kompilasi dari tahap builder.
COPY --from=builder /server /server

# Aplikasi mendengarkan di port 8080.
EXPOSE 8080

# Perintah yang dijalankan saat container start.
ENTRYPOINT ["/server"]
