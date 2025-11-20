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
  const [selectedFile, setSelectedFile] = useState(null);
  const [text, setText] = useState("");

  // Sohbet değiştiğinde input ve dosya alanlarını temizle
  useEffect(() => {
    setSelectedFile(null);
    setText("");
  }, [conversation.id]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [conversation?.messages]);

  const sendMessage = async (messageText) => {
    if (!user?.id || !conversation?.id) return;

    const userMsg = { id: Date.now(), from: "user", text: messageText };
    onUpdateConversation(conversation.id, (prev) => [...prev, userMsg]);

    try {
      const res = await api.post("/api/chat/", {
        user_id: user.id,
        message: messageText,
        conversation_id: conversation.id,
      });

      const aiMsg = {
        id: Date.now() + 1,
        from: "AI",
        text: res.data.response || "",
      };
      onUpdateConversation(conversation.id, (prev) => [...prev, aiMsg]);
    } catch (err) {
      console.error("❌ Mesaj gönderilemedi:", err);
      const errorMsg = {
        id: Date.now() + 2,
        from: "AI",
        text: "❌ Mesaj gönderilirken bir hata oluştu.",
      };
      onUpdateConversation(conversation.id, (prev) => [...prev, errorMsg]);
    }
  };

  const onFileUpload = async (file, additionalMessage = "") => {
    if (!conversation?.id || !user?.id) return;

    const userText = additionalMessage
      ? `📎 ${file.name}\n\n${additionalMessage}`
      : `📎 ${file.name}`;
    const userMsg = { id: Date.now(), from: "user", text: userText };
    const loadingId = Date.now() + 1;
    const loadingMsg = {
      id: loadingId,
      from: "AI",
      text: "📄 Dosya işleniyor...",
    };

    onUpdateConversation(conversation.id, (prev) => [
      ...prev,
      userMsg,
      loadingMsg,
    ]);

    const form = new FormData();
    form.append("file", file);

    try {
      const res = await api.post("/api/upload", form, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      const ocrText = res.data.text || "OCR metni alınamadı.";

      const ocrMsg = {
        id: Date.now() + 2,
        from: "AI",
        text: `📄 **Dosya İşlendi**\n\n${ocrText.substring(0, 500)}${
          ocrText.length > 500 ? "..." : ""
        }`,
      };

      onUpdateConversation(conversation.id, (prev) => [
        ...prev.filter((msg) => msg.id !== loadingId),
        ocrMsg,
      ]);

      await api.post("/api/chat/", {
        user_id: user.id,
        message: `[DOSYA: ${file.name}] ${
          additionalMessage || "Dosya yüklendi"
        }`,
        conversation_id: conversation.id,
      });
    } catch (e) {
      console.error("❌ Dosya işleme hatası:", e);
      const errorMsg = {
        id: Date.now() + 3,
        from: "AI",
        text: "❌ Dosya işlenirken bir hata oluştu.",
      };
      onUpdateConversation(conversation.id, (prev) => [
        ...prev.filter((msg) => msg.id !== loadingId),
        errorMsg,
      ]);
    }
  };

  if (!conversation) return <div className="chat-window">Sohbet seçiniz.</div>;

  return (
    <div className="chat-window">
      <div className="chat-header">
        <h3>{conversation?.title || "Yeni Sohbet"}</h3>
      </div>

      <div className="chat-messages">
        <MessageList messages={conversation?.messages || []} />
        <div ref={messagesEndRef} />
      </div>

      <MessageInput
        onSend={sendMessage}
        onFileUpload={onFileUpload}
        selectedFile={selectedFile}
        setSelectedFile={setSelectedFile}
        text={text}
        setText={setText}
      />
    </div>
  );
}
