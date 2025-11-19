import React from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext.jsx";
import "../styles/Sidebar.css";

export default function Sidebar({
  conversations,
  activeConversation,
  onSelectConversation,
  onNewConversation,
}) {
  const { logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const handleNewChat = () => {
    const title = `Sohbet ${conversations.length + 1}`;
    onNewConversation(title);
  };

  return (
    <div className="sidebar">
      <div className="sidebar-header">
        <h2>ChatApp</h2>
        <button className="logout-btn" onClick={handleLogout}>
          Çıkış
        </button>
      </div>

      <button className="new-chat-btn" onClick={handleNewChat}>
        + Yeni Sohbet
      </button>

      <div className="conversation-list">
        {conversations.length === 0 ? (
          <p className="empty">Henüz sohbet yok</p>
        ) : (
          conversations.map((conv) => (
            <div
              key={conv.id}
              className={`conversation-item ${
                conv.id === activeConversation?.id ? "active" : ""
              }`}
              onClick={() => onSelectConversation(conv)}
            >
              {conv.title}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
