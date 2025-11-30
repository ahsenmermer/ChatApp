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
  const [uploadedFile, setUploadedFile] = useState(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [conversation?.messages]);

  const sendMessage = async (text) => {
    if (!user?.id || !conversation?.id) return;

    // ✅ 1. Eğer dosya varsa, önce dosyayı mesaj olarak ekle
    if (uploadedFile && uploadedFile.status === "ready") {
      const fileMsg = {
        id: Date.now() - 1,
        from: "user",
        text: `📎 ${uploadedFile.name}`,
        fileType: uploadedFile.name.toLowerCase().endsWith(".pdf")
          ? "pdf"
          : "image",
        fileName: uploadedFile.name,
        isFile: true,
        timestamp: new Date().toISOString(),
      };
      onUpdateConversation(conversation.id, (prev) => [...prev, fileMsg]);

      // ✅ Badge'i input'tan kaldır (ama currentFileId'yi TUTUYORUZ!)
      setUploadedFile(null);
    }

    // ✅ 2. Kullanıcı mesajını ekle
    const userMsg = {
      id: Date.now(),
      from: "user",
      text,
      timestamp: new Date().toISOString(),
    };

    onUpdateConversation(conversation.id, (prev) => [...prev, userMsg]);

    try {
      const payload = {
        user_id: user.id,
        message: text,
        conversation_id: conversation.id,
      };

      // ✅ 3. currentFileId varsa ekle
      if (currentFileId) {
        payload.file_id = currentFileId;
      }

      const res = await api.post("/api/chat", payload);

      // ✅ 4. AI cevabını ekle
      const aiMsg = {
        id: Date.now() + 1,
        from: "AI",
        text: res.data.response || "",
        timestamp: new Date().toISOString(),
      };

      onUpdateConversation(conversation.id, (prev) => [...prev, aiMsg]);
    } catch (err) {
      console.error("❌ Mesaj gönderilemedi:", err);
      const errorMsg = {
        id: Date.now() + 2,
        from: "AI",
        text: "❌ Mesaj gönderilirken bir hata oluştu.",
        timestamp: new Date().toISOString(),
      };
      onUpdateConversation(conversation.id, (prev) => [...prev, errorMsg]);
    }
  };

  // ✅ YENİ: İlk init mesajı gönder (user_id ve conversation_id kaydetmek için)
  const sendInitialMessage = async (fileId, fileName) => {
    try {
      await api.post("/api/chat", {
        user_id: user.id,
        message: `_file_init_${fileName}`,
        conversation_id: conversation.id,
        file_id: fileId,
      });

      console.log("✅ Initial file message sent to register user info");
    } catch (err) {
      console.error("⚠️ Failed to send initial file message:", err);
    }
  };

  const onFileUpload = async (file) => {
    if (!conversation?.id || !user?.id) return;

    setUploadedFile({
      name: file.name,
      status: "processing",
      chunks: 0,
      fileId: null,
    });

    const form = new FormData();
    form.append("file", file);

    try {
      const uploadRes = await api.post("/api/upload", form, {
        headers: { "Content-Type": "multipart/form-data" },
        timeout: 120000,
      });

      const fileId = uploadRes.data.file_id;

      if (!fileId) {
        throw new Error("File ID not returned from server");
      }

      const maxAttempts = 60;
      let attempts = 0;

      const checkStatus = async () => {
        while (attempts < maxAttempts) {
          try {
            const statusRes = await api.get(`/api/file/status/${fileId}`);
            const status = statusRes.data.status;

            console.log(`📊 Polling attempt ${attempts + 1}: status=${status}`);

            // ✅ SADECE "completed" veya "ready" olduğunda devam et
            if (status === "completed" || status === "ready") {
              setCurrentFileId(fileId);
              setUploadedFile({
                name: file.name,
                status: "ready",
                chunks: statusRes.data.total_chunks || 0,
                fileId: fileId,
              });

              // ✅ YENİ: İlk mesajı otomatik gönder
              await sendInitialMessage(fileId, file.name);

              return;
            } else if (status === "failed") {
              throw new Error("File processing failed");
            }

            attempts++;
            await new Promise((resolve) => setTimeout(resolve, 1000));
          } catch (err) {
            if (err.response?.status === 404) {
              attempts++;
              await new Promise((resolve) => setTimeout(resolve, 1000));
            } else {
              throw err;
            }
          }
        }

        throw new Error("File processing timeout after 60 seconds");
      };

      await checkStatus();
    } catch (e) {
      console.error("❌ Dosya işleme hatası:", e);

      setUploadedFile({
        name: file.name,
        status: "failed",
        chunks: 0,
        fileId: null,
      });

      const errorMsg = {
        id: Date.now() + 3,
        from: "AI",
        text: `❌ Dosya işlenirken hata: ${e.message}`,
        timestamp: new Date().toISOString(),
      };
      onUpdateConversation(conversation.id, (prev) => [...prev, errorMsg]);
    }
  };

  const handleRemoveFile = () => {
    setUploadedFile(null);
    setCurrentFileId(null);
  };

  if (!conversation) {
    return (
      <div className="chat-window empty-state">
        <div className="empty-content">
          <h3>💬 Sohbet Seçin</h3>
          <p>Sol taraftan bir sohbet seçin veya yeni sohbet başlatın</p>
        </div>
      </div>
    );
  }

  return (
    <div className="chat-window">
      <div className="chat-header">
        <h3>{conversation?.title || "Yeni Sohbet"}</h3>
        {uploadedFile?.status === "ready" && (
          <span className="rag-badge">🔍 RAG Aktif</span>
        )}
      </div>

      <div className="chat-messages">
        <MessageList messages={conversation?.messages || []} />
        <div ref={messagesEndRef} />
      </div>

      <div className="chat-input-area">
        <MessageInput
          onSend={sendMessage}
          onFileUpload={onFileUpload}
          currentFileId={currentFileId}
          uploadedFile={uploadedFile}
          onRemoveFile={handleRemoveFile}
        />
      </div>
    </div>
  );
}
