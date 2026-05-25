-- Run this in Supabase SQL Editor
-- mini-render database schema

CREATE TABLE IF NOT EXISTS users (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(100) NOT NULL,
  email       VARCHAR(255) UNIQUE NOT NULL,
  username    VARCHAR(50)  UNIQUE NOT NULL,
  password    VARCHAR(255) NOT NULL,
  created_at  TIMESTAMP DEFAULT NOW(),
  plan        VARCHAR(20)  DEFAULT 'free'
);

CREATE TABLE IF NOT EXISTS deployments (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username     VARCHAR(50) NOT NULL REFERENCES users(username),
  app_name     VARCHAR(100) NOT NULL,
  app_url      TEXT NOT NULL,
  port         INTEGER NOT NULL,
  status       VARCHAR(20) DEFAULT 'sleeping',
  runtime      VARCHAR(20) DEFAULT 'python',
  deployed_at  TIMESTAMP DEFAULT NOW(),
  last_active  TIMESTAMP DEFAULT NOW(),
  UNIQUE(username, app_name)
);
