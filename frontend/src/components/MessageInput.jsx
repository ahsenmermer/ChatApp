import React, { useState, useRef } from "react";
import "../styles/MessageInput.css";

export default function MessageInput({
  onSend,
  onFileUpload,
  currentFileId,
  uploadedFile,
  onRemoveFile,
}) {
  const [text, setText] = useState("");
  const [sending, setSending] = useState(false);
  const [showMenu, setShowMenu] = useState(false);
  const [uploadingFile, setUploadingFile] = useState(null);
  const fileInputRef = useRef(null);

  const submit = async (e) => {
    e.preventDefault();
    if (!text.trim() || sending) return;
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
    setUploadingFile(file.name);

    try {
      await onFileUpload(file);
    } finally {
      setUploadingFile(null);
    }

    // Input'u temizle
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  return (
    <div className="message-input-wrapper">
      {/* 📎 Dosya Önizleme Badge */}
      {uploadedFile && (
        <div className="file-preview-badge">
          <div className="file-info">
            <span className="file-icon">
              {/* ✅ PDF veya fotoğraf ikonu */}
              {uploadedFile.name.toLowerCase().endsWith(".pdf") ? "📄" : "🖼️"}
            </span>
            <div className="file-details">
              <span className="file-name">{uploadedFile.name}</span>
              <span className="file-status">
                {uploadedFile.status === "processing" && "⏳ İşleniyor..."}
                {uploadedFile.status === "ready" &&
                  `✅ ${uploadedFile.chunks} bölüm - Soru sorabilirsiniz!`}
                {uploadedFile.status === "failed" && "❌ İşlenemedi"}
              </span>
            </div>
          </div>
          <button
            type="button"
            className="remove-file-btn"
            onClick={onRemoveFile}
            title="Dosyayı kaldır"
          >
            ✕
          </button>
        </div>
      )}

      <div className="input-container">
        {/* Attachment Button */}
        <div className="attachment-wrapper">
          <button
            type="button"
            className="attachment-btn"
            onClick={() => setShowMenu(!showMenu)}
            disabled={uploadingFile || uploadedFile}
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
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                  <path
                    d="M7 18h10M9.5 12.5l2.5-2.5 2.5 2.5M12 10v7"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                  />
                  <rect
                    x="4"
                    y="4"
                    width="16"
                    height="16"
                    rx="2"
                    stroke="currentColor"
                    strokeWidth="2"
                  />
                </svg>
                <span>PDF Yükle</span>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="application/pdf"
                  onChange={handleFileSelect}
                  hidden
                />
              </label>

              <label className="upload-option">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                  <rect
                    x="3"
                    y="3"
                    width="18"
                    height="18"
                    rx="2"
                    stroke="currentColor"
                    strokeWidth="2"
                  />
                  <circle cx="8.5" cy="8.5" r="1.5" fill="currentColor" />
                  <path
                    d="M21 15l-5-5L5 21"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                  />
                </svg>
                <span>Fotoğraf Yükle</span>
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

        {/* Message Form */}
        <form onSubmit={submit} className="message-form">
          <input
            type="text"
            placeholder={
              uploadingFile
                ? `📤 ${uploadingFile} yükleniyor...`
                : uploadedFile?.status === "ready"
                ? `"${uploadedFile.name}" hakkında soru sorun...`
                : uploadedFile?.status === "processing"
                ? "Dosya işleniyor, lütfen bekleyin..."
                : "Mesaj yazın..."
            }
            value={text}
            onChange={(e) => setText(e.target.value)}
            disabled={
              sending || uploadingFile || uploadedFile?.status === "processing"
            }
            className="message-input"
          />
          <button
            type="submit"
            disabled={
              !text.trim() ||
              sending ||
              uploadingFile ||
              uploadedFile?.status === "processing"
            }
            className="send-btn"
          >
            {sending ? (
              <span className="loading-spinner">⏳</span>
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
        </form>
      </div>
    </div>
  );
}
