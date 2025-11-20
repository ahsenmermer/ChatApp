import React, { useEffect, useState } from "react";
import { v4 as uuidv4 } from "uuid";
import Sidebar from "../components/Sidebar.jsx";
import ChatWindow from "../components/ChatWindow.jsx";
import "../styles/ChatApp.css";
import api from "../api/api.js";
import { useAuth } from "../contexts/AuthContext.jsx";

export default function ChatApp() {
  const { user } = useAuth();
  const [conversations, setConversations] = useState([]);
  const [activeConversationId, setActiveConversationId] = useState(null);

  // Yeni sohbet oluştur
  const handleNewConversation = (title) => {
    const newId = uuidv4();
    const newConv = { id: newId, title, messages: [] };
    setConversations((prev) => [...prev, newConv]);
    setActiveConversationId(newId);
  };

  // Backend’den geçmişi çek
  useEffect(() => {
    if (!user?.id) return;

    const fetchHistory = async () => {
      try {
        const res = await api.get(`/api/chat/history/${user.id}`);
        const messages = res.data?.messages || res.data || [];

        if (messages.length === 0) {
          const newConv = { id: "default", title: "Sohbet 1", messages: [] };
          setConversations([newConv]);
          setActiveConversationId("default");
          return;
        }

        const convMap = {};
        messages.forEach((m) => {
          const convId = m.conversation_id || "default";
          if (!convMap[convId]) convMap[convId] = { id: convId, messages: [] };

          convMap[convId].messages.push({
            id: m.id,
            from: "user",
            text: m.user_message,
            timestamp: m.timestamp,
          });

          if (m.ai_response) {
            convMap[convId].messages.push({
              id: m.id + "-ai",
              from: "AI",
              text: m.ai_response,
              timestamp: m.timestamp,
            });
          }
        });

        // ✅ Her conversation içindeki mesajları timestamp'e göre sırala
        Object.values(convMap).forEach((conv) => {
          conv.messages.sort(
            (a, b) => new Date(a.timestamp) - new Date(b.timestamp)
          );
        });

        let convArr = Object.values(convMap).sort((a, b) => {
          const tA = new Date(a.messages[0]?.timestamp || 0);
          const tB = new Date(b.messages[0]?.timestamp || 0);
          return tA - tB;
        });

        convArr = convArr.map((conv, index) => ({
          ...conv,
          title: `Sohbet ${index + 1}`,
        }));

        setConversations(convArr);
        if (!activeConversationId) setActiveConversationId(convArr[0].id);
      } catch (err) {
        console.error("❌ Sohbet geçmişi alınamadı:", err);
      }
    };

    fetchHistory();
  }, [user]);

  // Mesaj geldiğinde conversation güncelle
  const handleUpdateConversation = (convId, updateFn) => {
    setConversations((prev) => {
      const newConversations = prev.map((c) =>
        c.id === convId ? { ...c, messages: updateFn(c.messages) } : c
      );
      return newConversations;
    });
  };

  const activeConversation =
    conversations.find((c) => c.id === activeConversationId) || null;

  return (
    <div className="chat-app">
      <Sidebar
        conversations={conversations}
        activeConversation={activeConversation}
        onSelectConversation={(conv) => setActiveConversationId(conv.id)}
        onNewConversation={handleNewConversation}
      />

      {activeConversation && (
        <ChatWindow
          key={activeConversation.id} // Bu önemli: React component yeniden mount edecek
          conversation={activeConversation}
          onUpdateConversation={handleUpdateConversation}
          user={user}
        />
      )}
    </div>
  );
}
