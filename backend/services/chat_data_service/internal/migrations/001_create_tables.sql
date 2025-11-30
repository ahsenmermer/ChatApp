CREATE TABLE IF NOT EXISTS chat_messages (
    user_id String,
    user_message String,
    ai_response String,
    conversation_id String DEFAULT '',
    timestamp DateTime
)
ENGINE = MergeTree()
ORDER BY (user_id, timestamp);