CREATE TABLE IF NOT EXISTS users (
	id VARCHAR(50) PRIMARY KEY, 
    email TEXT,
	hash bytea,
	name TEXT 	 
	);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_user ON users (email);
