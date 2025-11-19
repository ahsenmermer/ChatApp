import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext.jsx";
import "../styles/Auth.css";

export default function Login() {
  const { login, user } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState(null);
  const [loading, setLoading] = useState(false);

  // Eğer kullanıcı zaten varsa otomatik yönlendir
  useEffect(() => {
    if (user) {
      navigate("/app", { replace: true });
    }
  }, [user, navigate]);

  const submit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setErr(null);
    try {
      // login fonksiyonu setUser çağırıyor
      await login(email, password);
      // ekstra güvenlik: login başarılıysa navigate et
      navigate("/app", { replace: true });
    } catch (error) {
      setErr(error.response?.data?.error || error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-container">
      <form className="auth-form" onSubmit={submit}>
        <h2>Giriş Yap</h2>
        {err && <div className="error">{err}</div>}
        <input
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          type="email"
          required
        />
        <input
          placeholder="Şifre"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          type="password"
          required
        />
        <button type="submit">{loading ? "Bekleyin..." : "Giriş Yap"}</button>
        <a href="/register">Hesap oluştur</a>
      </form>
    </div>
  );
}
