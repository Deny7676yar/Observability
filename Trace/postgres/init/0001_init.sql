CREATE USER gopher
WITH PASSWORD 'P@ssw0rd';

CREATE DATABASE app
    WITH OWNER gopher
    TEMPLATE = 'template0'
    ENCODING = 'utf-8'
    LC_COLLATE = 'C.UTF-8'
    LC_CTYPE = 'C.UTF-8';

\c app
SET ROLE gopher;

CREATE TABLE IF NOT EXISTS users (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(150) NOT NULL
);

CREATE TABLE IF NOT EXISTS articles (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	title varchar(150) NOT NULL,
	user_id uuid REFERENCES users
);