import React, { useState, useRef } from "react";
import "../styles/MessageInput.css";

export default function MessageInput({ onSend, onFileUpload }) {
  const [text, setText] = useState("");
  const [sending, setSending] = useState(false);
  const [showMenu, setShowMenu] = useState(false);
  const [selectedFile, setSelectedFile] = useState(null);
  const fileInputRef = useRef(null);

  const submit = async (e) => {
    e.preventDefault();

    // Eğer dosya varsa, önce dosyayı yükle
    if (selectedFile) {
      setSending(true);
      try {
        await onFileUpload(selectedFile, text.trim());
        setText("");
        setSelectedFile(null);
      } finally {
        setSending(false);
      }
      return;
    }

    // Sadece metin varsa
    if (!text.trim()) return;
    setSending(true);

    try {
      await onSend(text.trim());
      setText("");
    } finally {
      setSending(false);
    }
  };

  const handleFileSelect = (e, type) => {
    const file = e.target.files[0];
    if (!file) return;

    setSelectedFile(file);
    setShowMenu(false);
  };

  const removeFile = () => {
    setSelectedFile(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  return (
    <div className="input-wrapper">
      {/* Dosya Önizleme */}
      {selectedFile && (
        <div className="file-preview">
          <div className="file-info">
            <span className="file-icon">
              {selectedFile.type.includes("pdf") ? "📄" : "🖼️"}
            </span>
            <span className="file-name">{selectedFile.name}</span>
            <span className="file-size">
              {(selectedFile.size / 1024).toFixed(1)} KB
            </span>
          </div>
          <button
            type="button"
            className="remove-file-btn"
            onClick={removeFile}
          >
            ✕
          </button>
        </div>
      )}

      {/* Input Alanı */}
      <form onSubmit={submit} className="message-input-form">
        <div className="input-container">
          {/* + Butonu ve Menü */}
          <div className="attachment-section">
            <button
              type="button"
              className="attach-btn"
              onClick={() => setShowMenu(!showMenu)}
              title="Dosya ekle"
            >
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <path
                  d="M12 5v14M5 12h14"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                />
              </svg>
            </button>

            {showMenu && (
              <div className="upload-dropdown">
                <label className="upload-option">
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="application/pdf"
                    onChange={(e) => handleFileSelect(e, "pdf")}
                    hidden
                  />
                  <span className="option-icon">📄</span>
                  <span className="option-text">PDF Yükle</span>
                </label>

                <label className="upload-option">
                  <input
                    type="file"
                    accept="image/*"
                    onChange={(e) => handleFileSelect(e, "image")}
                    hidden
                  />
                  <span className="option-icon">🖼️</span>
                  <span className="option-text">Fotoğraf Yükle</span>
                </label>
              </div>
            )}
          </div>

          {/* Text Input */}
          <input
            type="text"
            className="text-input"
            placeholder={
              selectedFile ? "Mesaj ekle (opsiyonel)..." : "Mesaj yazın..."
            }
            value={text}
            onChange={(e) => setText(e.target.value)}
            disabled={sending}
          />

          {/* Gönder Butonu */}
          <button
            type="submit"
            className="send-btn"
            disabled={sending || (!text.trim() && !selectedFile)}
          >
            {sending ? (
              <div className="spinner" />
            ) : (
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                <path
                  d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            )}
          </button>
        </div>
      </form>
    </div>
  );
}
