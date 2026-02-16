import axios from "axios";

// Konfigurasi instance Axios dengan base URL dari backend Go Anda.
// Pastikan URL ini sesuai dengan alamat backend Anda berjalan.
const api = axios.create({
  baseURL: "http://localhost:8080/auth", // Sesuaikan jika port atau host berbeda
  headers: {
    "Content-Type": "application/json",
  },
});

export default api; 
