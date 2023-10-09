CREATE TABLE IF NOT EXISTS users (
	user_id VARCHAR(50) PRIMARY KEY, 
    email TEXT,
	hash bytea,
	name VARCHAR(150) 	 
	);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_users ON users (email);
