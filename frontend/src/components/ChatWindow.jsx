import React, { useRef, useEffect, useState } from "react";
import api from "../api/api.js";
import MessageList from "./MessageList.jsx";
import MessageInput from "./MessageInput.jsx";
import "../styles/ChatWindow.css";

export default function ChatWindow({
  conversation,
  onUpdateConversation,
  user,
}) {
  const messagesEndRef = useRef(null);
  const [currentFileId, setCurrentFileId] = useState(null);
  const [fileStatus, setFileStatus] = useState(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [conversation?.messages]);

  // Dosya durumunu kontrol et
  const checkFileStatus = async (fileId) => {
    try {
      const res = await api.get(`/api/file/status/${fileId}`);
      return res.data;
    } catch (err) {
      console.error("❌ Dosya durumu kontrol edilemedi:", err);
      return null;
    }
  };

  // Polling ile dosya hazır olana kadar bekle
  const waitForFile = async (fileId) => {
    const maxAttempts = 30; // 30 saniye (30 * 1000ms)
    let attempts = 0;

    return new Promise((resolve, reject) => {
      const interval = setInterval(async () => {
        attempts++;
        const status = await checkFileStatus(fileId);

        if (status?.status === "ready") {
          clearInterval(interval);
          resolve(status);
        } else if (status?.status === "failed") {
          clearInterval(interval);
          reject(new Error("Dosya işleme başarısız oldu"));
        } else if (attempts >= maxAttempts) {
          clearInterval(interval);
          reject(new Error("Dosya işleme zaman aşımına uğradı"));
        }

        // Durum mesajını güncelle
        setFileStatus(
          `Dosya işleniyor... (${status?.total_chunks || 0} chunk)`
        );
      }, 1000);
    });
  };

  // Normal mesaj gönder
  const sendMessage = async (messageText, fileId = null) => {
    if (!user?.id || !conversation?.id) return;

    const userMsg = {
      id: Date.now(),
      from: "user",
      text: messageText,
    };
    onUpdateConversation(conversation.id, (prev) => [...prev, userMsg]);

    try {
      const payload = {
        user_id: user.id,
        message: messageText,
        conversation_id: conversation.id,
      };

      // Eğer file_id varsa ekle
      if (fileId) {
        payload.file_id = fileId;
      }

      const res = await api.post("/api/chat", payload);

      const aiMsg = {
        id: Date.now() + 1,
        from: "AI",
        text: res.data.response || "Yanıt alınamadı.",
      };
      onUpdateConversation(conversation.id, (prev) => [...prev, aiMsg]);
    } catch (err) {
      console.error("❌ Mesaj gönderilemedi:", err);
      const errorMsg = {
        id: Date.now() + 2,
        from: "AI",
        text:
          err.response?.data?.error ||
          "❌ Mesaj gönderilirken bir hata oluştu.",
      };
      onUpdateConversation(conversation.id, (prev) => [...prev, errorMsg]);
    }
  };

  // Dosya yükleme
  const onFileUpload = async (file, additionalMessage = "") => {
    if (!conversation?.id || !user?.id) return;

    const userText = additionalMessage
      ? `📎 Dosya yükleniyor: ${file.name}\n\n${additionalMessage}`
      : `📎 Dosya yükleniyor: ${file.name}`;

    const userMsg = { id: Date.now(), from: "user", text: userText };
    const loadingId = Date.now() + 1;
    const loadingMsg = {
      id: loadingId,
      from: "AI",
      text: "📄 Dosya yükleniyor ve işleniyor...",
    };

    onUpdateConversation(conversation.id, (prev) => [
      ...prev,
      userMsg,
      loadingMsg,
    ]);

    const form = new FormData();
    form.append("file", file);

    try {
      // 1. Dosyayı yükle
      const uploadRes = await api.post("/api/upload", form, {
        headers: { "Content-Type": "multipart/form-data" },
      });

      const fileId = uploadRes.data.file_id;
      if (!fileId) {
        throw new Error("file_id alınamadı");
      }

      setCurrentFileId(fileId);

      // 2. Dosya işlenene kadar bekle
      onUpdateConversation(conversation.id, (prev) =>
        prev.map((msg) =>
          msg.id === loadingId
            ? {
                ...msg,
                text: "⏳ Dosya işleniyor, lütfen bekleyin...",
              }
            : msg
        )
      );

      await waitForFile(fileId);

      // 3. Dosya hazır mesajı
      const readyMsg = {
        id: Date.now() + 2,
        from: "AI",
        text: `✅ Dosya başarıyla işlendi!\n\n📄 **${file.name}**\n\nArtık bu dosya hakkında soru sorabilirsiniz.`,
      };

      onUpdateConversation(conversation.id, (prev) => [
        ...prev.filter((msg) => msg.id !== loadingId),
        readyMsg,
      ]);

      setFileStatus(null);

      // 4. Eğer kullanıcı mesaj yazmışsa, otomatik olarak gönder
      if (additionalMessage.trim()) {
        setTimeout(() => {
          sendMessage(additionalMessage, fileId);
        }, 500);
      }
    } catch (err) {
      console.error("❌ Dosya işleme hatası:", err);
      const errorMsg = {
        id: Date.now() + 3,
        from: "AI",
        text: `❌ Dosya işlenemedi: ${err.message}`,
      };
      onUpdateConversation(conversation.id, (prev) => [
        ...prev.filter((msg) => msg.id !== loadingId),
        errorMsg,
      ]);
      setFileStatus(null);
      setCurrentFileId(null);
    }
  };

  if (!conversation) return <div className="chat-window">Sohbet seçiniz.</div>;

  return (
    <div className="chat-window">
      <div className="chat-header">
        <h3>{conversation?.title || "Yeni Sohbet"}</h3>
        {fileStatus && <div className="file-status-badge">{fileStatus}</div>}
        {currentFileId && !fileStatus && (
          <div className="file-ready-badge">
            📄 Dosya hazır - Soru sorabilirsiniz
          </div>
        )}
      </div>

      <div className="chat-messages">
        <MessageList messages={conversation?.messages || []} />
        <div ref={messagesEndRef} />
      </div>

      <MessageInput
        onSend={(text) => sendMessage(text, currentFileId)}
        onFileUpload={onFileUpload}
      />
    </div>
  );
}
