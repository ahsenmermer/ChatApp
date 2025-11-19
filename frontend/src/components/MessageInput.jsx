import React, { useState } from "react";
import api from "../api/api.js";
import "../styles/MessageInput.css";

export default function MessageInput({ onSend, onFileUpload }) {
  const [text, setText] = useState("");
  const [sending, setSending] = useState(false);
  const [showMenu, setShowMenu] = useState(false);

  const submit = async (e) => {
    e.preventDefault();
    if (!text.trim()) return;
    setSending(true);

    try {
      await onSend(text.trim());
      setText("");
    } finally {
      setSending(false);
    }
  };

  const handleFileSelect = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    setShowMenu(false);

    await onFileUpload(file);
  };

  return (
    <div className="input-container">
      {/* + Button */}
      <div className="left-buttons">
        <button
          type="button"
          className="plus-btn"
          onClick={() => setShowMenu(!showMenu)}
        >
          +
        </button>

        {showMenu && (
          <div className="upload-menu">
            <label>
              📄 PDF Yükle
              <input
                type="file"
                accept="application/pdf"
                onChange={handleFileSelect}
                hidden
              />
            </label>

            <label>
              🖼️ Fotoğraf Yükle
              <input
                type="file"
                accept="image/*"
                onChange={handleFileSelect}
                hidden
              />
            </label>
          </div>
        )}
      </div>

      {/* Text Input & Send */}
      <form onSubmit={submit} className="message-form">
        <input
          type="text"
          placeholder="Mesaj yazın..."
          value={text}
          onChange={(e) => setText(e.target.value)}
        />
        <button type="submit">{sending ? "..." : "Gönder"}</button>
      </form>
    </div>
  );
}
