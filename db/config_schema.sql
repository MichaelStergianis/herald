-- Settings related
CREATE SCHEMA IF NOT EXISTS config;

CREATE TABLE IF NOT EXISTS config.preferences (
       id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS config.users (
      id SERIAL PRIMARY KEY,

      preference_id INTEGER REFERENCES config.preferences(id),

      user_name VARCHAR UNIQUE NOT NULL,
      email VARCHAR NOT NULL,
      password VARCHAR NOT NULL
);
