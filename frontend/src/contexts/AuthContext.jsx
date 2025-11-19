// src/contexts/AuthContext.jsx
import React, { createContext, useContext, useState } from "react";
import api from "../api/api.js";

const AuthContext = createContext();
export const useAuth = () => useContext(AuthContext);

export function AuthProvider({ children }) {
  // Basit ve güvenli: başlangıç null. (undefined ile yükleniyor durumunu karmaşaya sokmayalım)
  const [user, setUser] = useState(() => {
    try {
      const stored = localStorage.getItem("chatapp_user");
      return stored ? JSON.parse(stored) : null;
    } catch {
      return null;
    }
  });

  const login = async (email, password) => {
    try {
      const res = await api.post("/api/auth/login", { email, password });

      // ✅ backend token yerine direkt user dönüyor, onu yakalayalım
      const userData = res.data.user || res.data;
      const token = res.data.token || "dummy-token"; // token yoksa geçici bir değer

      // LocalStorage’a kaydet
      localStorage.setItem("chatapp_token", token);
      localStorage.setItem("chatapp_user", JSON.stringify(userData));

      setUser(userData);
      return userData;
    } catch (error) {
      console.error("❌ Login hatası:", error);
      throw error;
    }
  };

  const register = async (username, email, password) => {
    const res = await api.post("/api/auth/register", {
      username,
      email,
      password,
    });
    return res.data;
  };

  const logout = () => {
    localStorage.removeItem("chatapp_token");
    localStorage.removeItem("chatapp_user");
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}
