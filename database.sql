-- This is the SQL script that will be used to initialize the database schema.
-- We will evaluate you based on how well you design your database.
-- 1. How you design the tables.
-- 2. How you choose the data types and keys.
-- 3. How you name the fields.
-- In this assignment we will use PostgreSQL as the database.

-- This is the estate table. Remove this table and replace with your own tables.
CREATE TABLE estates (
	id UUID PRIMARY KEY,
	length INT NOT NULL,
	width INT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE trees (
	id UUID PRIMARY KEY,
	estate_id UUID NOT NULL,
	x INT NOT NULL,
	y INT NOT NULL,
	height INT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (estate_id) REFERENCES estates(id)
);

CREATE TABLE users (
	id UUID PRIMARY KEY,
	username VARCHAR(255) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE person (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL UNIQUE,
	first_name VARCHAR(255),
	last_name VARCHAR(255),
	date_of_birth DATE,
	bio TEXT,
	avatar_url VARCHAR(500),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE person_email (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL,
	email VARCHAR(255) NOT NULL,
	is_primary BOOLEAN DEFAULT false,
	verified BOOLEAN DEFAULT false,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	UNIQUE(user_id, email)
);

CREATE TABLE person_phone (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL,
	phone VARCHAR(20) NOT NULL,
	type VARCHAR(50), -- 'mobile', 'home', 'work'
	is_primary BOOLEAN DEFAULT false,
	verified BOOLEAN DEFAULT false,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	UNIQUE(user_id, phone)
);

CREATE TABLE person_social_media (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL,
	platform VARCHAR(100) NOT NULL, -- 'twitter', 'linkedin', 'github', etc.
	username VARCHAR(255),
	profile_url VARCHAR(500),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	UNIQUE(user_id, platform)
);
