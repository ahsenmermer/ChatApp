-- ---------------------------
-- Subscription Service Tables
-- ---------------------------

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS subscription_quotas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subscription_id UUID REFERENCES subscription_plans(id),
    quota INT NOT NULL  -- artık mesaj sayısı bazlı
);

CREATE TABLE IF NOT EXISTS user_subscription (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    subscription_id UUID REFERENCES subscription_plans(id),
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL
);

ALTER TABLE user_subscription
ADD COLUMN IF NOT EXISTS remaining_quota INT DEFAULT 0;

-- Free plan ekle (1000 mesaj hakkı)
INSERT INTO subscription_plans (id, name, description)
VALUES (uuid_generate_v4(), 'Free', 'Free plan with 1000 messages quota')
ON CONFLICT (name) DO NOTHING;

INSERT INTO subscription_quotas (id, subscription_id, quota)
SELECT uuid_generate_v4(), id, 1000 FROM subscription_plans WHERE name='Free'
ON CONFLICT DO NOTHING;
