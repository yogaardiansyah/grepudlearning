Berikut adalah dokumen lengkap dalam format Markdown (`.md`) yang sudah dirapikan, diperdetail, dan dilengkapi dengan penjelasan ala **ELI5 (Explain Like I'm 5)** agar mudah dipahami oleh siapa pun, dari junior hingga stakeholder.

---

# 📚 Distributed Transaction Patterns in Microservices

Dalam arsitektur microservices, transaksi lintas service tidak bisa menggunakan **ACID transaction** biasa seperti di monolith. Karena setiap service memiliki database sendiri, kita membutuhkan pola khusus untuk menjaga konsistensi data (Eventual Consistency).

Dokumen ini menjelaskan pola distribusi transaksi, langkah praktis migrasi dari sistem saat ini, serta penjelasan sederhana untuk setiap konsep.

---

## 1️⃣ Two-Phase Commit (2PC)

### 📌 Konsep
Pola klasik yang menggunakan **Transaction Coordinator** untuk memastikan semua service commit atau rollback secara bersamaan.

*   **Prepare Phase:** Coordinator bertanya ke semua service, "Siap commit?"
*   **Commit Phase:** Jika semua jawab "Siap", data ditulis permanen. Jika satu saja gagal, semua rollback.

### 🛠️ Evolusi dari Sistem Sekarang
*   **Status:** Tidak direkomendasikan untuk sistem modern.
*   **Jika ingin diterapkan:** Perlu menginstal software middleware coordinator (seperti Atomikos atau JTA) dan database harus mendukung XA Transactions.

### ⚠️ Kelemahan
*   **Blocking:** Jika coordinator mati, semua service terkunci (lock).
*   **Slow:** Menunggu semua service merespons membuat performa drop.

> ### 👶 ELI5 (Penjelasan Anak 5 Tahun)
> Bayangkan seorang guru (Coordinator) dan murid-muridnya. Guru bertanya, "Semua sudah pegang pensil?" Guru baru akan bilang "Mulai menulis!" kalau **semua** murid menjawab "Sudah!". Jika ada satu saja murid yang pensilnya patah, guru menyuruh **semua** murid meletakkan pensilnya kembali. Ini adil, tapi kalau ada satu murid yang lambat, semua orang jadi menunggu lama.

---

## 2️⃣ Saga Pattern

Saga adalah solusi modern menggunakan **Compensating Action** (tindakan pembatalan) alih-alih mengunci database.

### 2.1 Orchestrated Saga
Ada satu service pusat yang bertindak sebagai otak/konduktor.

#### 🛠️ Langkah Migrasi:
1.  **Buat Orchestrator Service:** Service baru (misal: `order-saga-manager`) untuk mengontrol alur.
2.  **Ubah Flow:** Frontend tidak lagi memanggil banyak service, cukup panggil Orchestrator.
3.  **State Machine:** Buat logika status: `PENDING` -> `PAID` -> `SUCCESS`.
4.  **Endpoint Kompensasi:** Tambahkan API `/cancel-order` atau `/refund-payment` di service terkait.

> ### 👶 ELI5
> Seperti seorang **Wedding Organizer**. Dia menelepon katering, dekorasi, dan fotografer. Jika katering bilang "Habis", si WO yang bertugas menelepon tukang dekorasi untuk batal pasang tenda.

---

### 2.2 Choreography Saga
Tidak ada bos. Setiap service saling memberi tahu lewat "pengumuman" (Event).

#### 🛠️ Langkah Migrasi:
1.  **Pasang Message Broker:** Install Kafka, RabbitMQ, atau Google Pub/Sub.
2.  **Publish Event:** Saat Order dibuat, `OrderService` melempar pesan "Order_Dibuat".
3.  **Listen Event:** `PaymentService` mendengar pesan itu, memproses bayar, lalu melempar pesan "Bayar_Sukses".
4.  **Auto-Update:** `OrderService` mendengar "Bayar_Sukses" dan mengubah status jadi "Selesai".

> ### 👶 ELI5
> Seperti **Grup WhatsApp**. Tidak ada ketua. Jika satu orang share "Saya sudah beli kopi", orang lain langsung otomatis "Saya siapkan gelas". Semuanya bergerak sendiri-sendiri karena melihat pesan di grup.

---

## 3️⃣ Event-Driven Architecture (EDA)

### 📌 Konsep
Seluruh komunikasi antar service berbasis **Event** (peristiwa), bukan perintah langsung (HTTP Call).

#### 🛠️ Langkah Migrasi:
1.  **Matikan HTTP Call:** Hapus kode `axios.post('http://payment-service/...')` di dalam service.
2.  **Ganti dengan Publisher:** Gunakan library library seperti `kafkajs` atau `amqplib`.
3.  **Event Schema:** Tentukan format JSON pesan agar semua service paham isinya.

> ### 👶 ELI5
> Seperti **Papan Pengumuman** di sekolah. Guru tidak mendatangi murid satu-persatu untuk bilang "Besok Libur". Guru cukup tempel di papan pengumuman. Siapa pun yang berkepentingan tinggal melihat papan itu kapan saja mereka sempat.

---

## 4️⃣ CQRS (Command Query Responsibility Segregation)

### 📌 Konsep
Memisahkan jalur **Menulis** (Command) dan jalur **Membaca** (Query) ke database yang berbeda.

#### 🛠️ Langkah Migrasi:
1.  **Pisahkan Codebase:** Buat folder `internal/commands` (untuk Create/Update/Delete) dan `internal/queries` (untuk List/Get).
2.  **Gunakan Read DB:** Gunakan database cepat seperti **Redis** atau **ElasticSearch** khusus untuk Read.
3.  **Sinkronisasi:** Saat ada Command di Postgres, kirim event ke Kafka untuk mengupdate data di Redis.

> ### 👶 ELI5
> Seperti **Restoran Cepat Saji**. Ada **Kasir 1** khusus buat pesan makanan (Menulis/Command), dan ada **Layar Monitor** besar untuk melihat apakah makanan sudah siap (Membaca/Query). Kamu tidak tanya ke kasir "Pesanan saya nomor berapa?", kamu cukup lihat layar monitor.

---

## 5️⃣ Event Sourcing

### 📌 Konsep
Kita tidak menyimpan status terakhir, tapi menyimpan **semua riwayat kejadian**.

#### 🛠️ Langkah Migrasi:
1.  **Ubah Tabel Database:** Jangan gunakan `UPDATE orders SET status = 'DONE'`.
2.  **Log-Based:** Tambahkan baris baru setiap ada perubahan:
    *   `Order_Created`
    *   `Payment_Received`
    *   `Order_Shipped`
3.  **Replay Logic:** Buat fungsi untuk menghitung status terakhir berdasarkan seluruh riwayat tersebut.

> ### 👶 ELI5
> Seperti **Buku Tabungan Bank**. Di bukumu tidak cuma tertulis "Saldo: 1 Juta". Tapi ada daftar: "Masuk 500rb", "Keluar 100rb", "Masuk 600rb". Jika ingin tahu saldo akhir, kamu hitung semua riwayatnya dari awal sampai akhir.

---

## 📊 Perbandingan Pola

| Pattern | Kompleksitas | Skalabilitas | Penggunaan Modern |
| :--- | :--- | :--- | :--- |
| **2PC** | Rendah | Rendah | Sangat Jarang |
| **Orchestrated Saga** | Sedang | Tinggi | Sangat Umum |
| **Choreography Saga** | Tinggi | Sangat Tinggi | Standar Industri |
| **Event-Driven** | Tinggi | Sangat Tinggi | Cloud-Native |
| **CQRS** | Sedang | Tinggi | Sistem High-Traffic |
| **Event Sourcing** | Sangat Tinggi | Sangat Tinggi | Fintech / Audit Trail |

---

## 🚀 Roadmap Evolusi (Step-by-Step)

Jika sistem Anda saat ini masih menggunakan Synchronous HTTP, mulailah secara bertahap:

1.  **Tahap 1 (Stability):** Implementasikan **Orchestrated Saga**. Ini yang paling mudah karena kontrol tetap terpusat, tapi mulai memperkenalkan pembatalan transaksi (compensating action).
2.  **Tahap 2 (Decoupling):** Mulai perkenalkan Message Broker (Kafka/RabbitMQ) dan pindah ke **Choreography Saga** agar service tidak saling bergantung secara langsung.
3.  **Tahap 3 (Performance):** Jika aplikasi mulai lambat saat mengambil data, terapkan **CQRS**. Pisahkan database baca dan tulis.
4.  **Tahap 4 (Advanced):** Terapkan **Event Sourcing** hanya jika Anda butuh fitur audit trail yang sangat ketat (seperti sistem keuangan).

---

## 🎯 Kesimpulan
Tidak ada pola yang "paling benar", yang ada adalah pola yang "paling cocok".
*   Butuh konsistensi & alur jelas? **Orchestrated Saga**.
*   Butuh sistem yang bisa tumbuh sangat besar? **Choreography / Event-Driven**.
*   Butuh dashboard super cepat? **CQRS**.

---
*Dokumen ini dibuat untuk panduan teknis pengembangan arsitektur microservices.*