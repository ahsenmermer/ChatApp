// src/api/api.js
import axios from "axios";

// 🌐 API Gateway adresini .env'den al
// Eğer tanımlı değilse 127.0.0.1 kullan (Docker uyumlu)
const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://127.0.0.1:8085";

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 10000, // 10 saniye zaman aşımı (istek takılmaz)
});

// 🔑 Token interceptor — kullanıcı giriş yaptıysa header’a ekle
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("chatapp_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// ⚠️ Hata yakalama — örneğin token süresi bitmişse otomatik yönlendirme
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (!error.response) {
      console.error("🌐 Ağ hatası veya sunucuya ulaşılamadı:", error.message);
    } else if (error.response.status === 401) {
      console.warn("🔒 Yetkisiz! Giriş sayfasına yönlendiriliyor...");
      localStorage.removeItem("chatapp_token");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);

export default api;
