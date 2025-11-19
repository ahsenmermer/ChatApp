export default function MessageList({ messages = [] }) {
  if (!Array.isArray(messages)) return null;

  return (
    <div className="messages-container">
      {messages.map((m, i) => {
        if (!m) return null;

        const text = m.text || m.message || m.response;
        if (!text) return null; // boş mesajları render etme

        const from = m.from || (m.user_id ? "user" : "AI");

        return (
          <div
            key={m.id || i}
            className={`message-item ${
              from === "user" ? "from-user" : "from-ai"
            }`}
          >
            <div className="message-bubble">{text}</div>
          </div>
        );
      })}
    </div>
  );
}
