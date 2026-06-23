// go.mod mendefinisikan module Go beserta dependensinya.
//
// PENTING UNTUK PEMULA:
// Versi dependensi di bawah sengaja dibiarkan agar diselesaikan otomatis.
// Setelah meletakkan file ini, JALANKAN perintah berikut sekali saja:
//
//     go mod tidy
//
// Perintah itu akan mengunduh dd-trace-go versi terbaru yang cocok dan
// mengisi file go.sum secara otomatis. Jika kamu menjalankan lewat Docker
// (disarankan, lihat README), langkah ini sudah ditangani di dalam Dockerfile.

module mini-app

go 1.23

require (
	github.com/DataDog/dd-trace-go/contrib/net/http/v2 v2.3.0
	github.com/DataDog/dd-trace-go/v2 v2.3.0
)
