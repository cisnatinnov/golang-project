-- This is the SQL script that will be used to initialize the database schema.
-- We will evaluate you based on how well you design your database.
-- 1. How you design the tables.
-- 2. How you choose the data types and keys.
-- 3. How you name the fields.
-- In this assignment we will use PostgreSQL as the database.

-- This is the estate table. Remove this table and replace with your own tables. 
CREATE TABLE estate (
	id UUID PRIMARY KEY,
	length INT (100) NOT NULL,
	width INT (100) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tree (
	id UUID PRIMARY KEY,
	estate_id UUID NOT NULL,
	x INT (100) NOT NULL,
	y INT (100) NOT NULL,
	height INT (100) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (estate_id) REFERENCES estate(id)
);