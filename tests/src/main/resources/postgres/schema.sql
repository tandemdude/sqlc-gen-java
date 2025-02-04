-- Users table
CREATE TABLE users (
    user_id       UUID PRIMARY KEY,
    username      VARCHAR(50) UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tokens table
CREATE TABLE tokens (
    token_id SERIAL PRIMARY KEY,
    user_id  UUID NOT NULL,
    token    VARCHAR(255) UNIQUE NOT NULL,
    expiry   TIMESTAMP NOT NULL
);

-- Chats table
CREATE TABLE chats (
    chat_id    SERIAL PRIMARY KEY,
    chat_name  VARCHAR(100) NOT NULL,
    created_at_time TIME WITH TIME ZONE DEFAULT NOW()::time with time zone
);

-- Messages table with array column for attachments
CREATE TABLE messages (
    message_id SERIAL PRIMARY KEY,
    chat_id    INTEGER NOT NULL,
    user_id    UUID NOT NULL,
    content    TEXT NOT NULL,
    attachments TEXT[], -- Array column for storing attachment URLs
    sent_at_time    TIME DEFAULT NOW()::time without time zone,
    sent_at_date    DATE DEFAULT NOW()::date
);
