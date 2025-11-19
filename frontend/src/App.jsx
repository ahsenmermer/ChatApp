import React from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import Login from "./pages/Login";
import Register from "./pages/Register";
import ChatApp from "./pages/ChatApp";
import { useAuth } from "./contexts/AuthContext";

function PrivateRoute({ children }) {
  const { user } = useAuth();
  // Basit: user varsa geçir, yoksa login'e yönlendir.
  return user ? children : <Navigate to="/login" replace />;
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route
        path="/app"
        element={
          <PrivateRoute>
            <ChatApp />
          </PrivateRoute>
        }
      />
      <Route path="/" element={<Navigate to="/app" replace />} />
    </Routes>
  );
}
