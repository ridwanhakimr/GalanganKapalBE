# 🚢 Shipyard Management System - Backend API

Repositori ini berisi layanan *Backend* (API) untuk **Shipyard Management System** (Sistem Manajemen Inventaris Galangan Kapal). Dibangun untuk menangani logika bisnis yang kompleks, transaksi atomik, dan keamanan akses berlapis.

## 🛠️ Teknologi (Tech Stack)

* **Bahasa**: Golang (Go)
* **Web Framework**: [Fiber v2](https://gofiber.io/) (Cepat & Ringan)
* **ORM**: [GORM](https://gorm.io/)
* **Database**: PostgreSQL (Di-host via Supabase)
* **Keamanan**: JWT (JSON Web Tokens) & Custom RBAC Middleware

## ✨ Fitur Utama (Core Features)

1. **Role-Based Access Control (RBAC)**: *Middleware* ketat yang memastikan hanya Admin yang bisa mengubah Master Data, Supervisor yang bisa menyetujui, dan Staff yang bisa mengajukan permintaan.
2. **Transaksi Atomik**: Memastikan pemotongan stok di Database (`Quantity`) hanya terjadi secara valid dan mutlak apabila Supervisor menekan *Approve*.
3. **Audit Log & Stock Ledger**: Mencatat setiap aktivitas secara tidak terhapuskan (*Immutable*). Mulai dari modifikasi data (Audit Log) hingga pencatatan detik keluar masuknya arus fisik barang (Buku Besar).
4. **Master Data CRUD**: API lengkap untuk Manajemen Gudang (*Warehouses*) dan Kategori (*Categories*).

## 🚀 Panduan Instalasi (Setup Guide)

Ikuti langkah-langkah berikut untuk menjalankan server API di mesin lokal Anda.

### 1. Kloning Repositori
```bash
git clone https://github.com/USERNAME_ANDA/NAMA_REPO_BACKEND_ANDA.git
cd NAMA_REPO_BACKEND_ANDA
```

### 2. Konfigurasi Database (Environment Variables)
Mintalah file `.env` kepada pemilik repositori ini (atau buat file bernama `.env` di folder utama backend). Isinya harus memuat kredensial rahasia seperti ini:
```env
APP_PORT=3000
DATABASE_URL="postgres://postgres.xxxxx:password@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres"
JWT_SECRET="rahasia_super_kuat"
```

### 3. Mengunduh Dependensi
```bash
go mod tidy
```

### 4. Menjalankan Server
```bash
go run ./cmd/api/main.go
```
*Tanda Berhasil: Terminal akan memunculkan tulisan "Successfully connected to the database!" dan "Server berjalan di port 3000". Skema database juga akan melakukan Auto-Migrate secara otomatis.*
