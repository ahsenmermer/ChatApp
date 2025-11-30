export default function MessageList({ messages = [] }) {
  if (!Array.isArray(messages)) return null;

  return (
    <div className="messages-container">
      {messages.map((m, i) => {
        if (!m) return null;

        const text = m.text || m.message || m.response;
        if (!text) return null;

        const from = m.from || (m.user_id ? "user" : "AI");

        return (
          <div
            key={m.id || i}
            className={`message-item ${
              from === "user" ? "from-user" : "from-ai"
            }`}
          >
            {/* ✅ Dosya mesajı için özel render */}
            {m.isFile ? (
              <div className="file-message-bubble">
                <div className="file-message-icon">
                  {m.fileType === "pdf" ? (
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                      <path
                        d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                        stroke="currentColor"
                        strokeWidth="2"
                        fill="currentColor"
                        fillOpacity="0.1"
                      />
                      <path
                        d="M14 2v6h6"
                        stroke="currentColor"
                        strokeWidth="2"
                      />
                    </svg>
                  ) : (
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
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
                      />
                    </svg>
                  )}
                </div>
                <div className="file-message-info">
                  <span className="file-message-name">{m.fileName}</span>
                  <span className="file-message-type">
                    {m.fileType === "pdf" ? "PDF" : "Görsel"}
                  </span>
                </div>
              </div>
            ) : (
              <div className="message-bubble">{text}</div>
            )}
          </div>
        );
      })}
    </div>
  );
}
