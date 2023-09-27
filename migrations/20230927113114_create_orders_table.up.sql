CREATE TABLE IF NOT EXISTS orders (
	id VARCHAR(50) PRIMARY KEY, 
	number TEXT,	 
	user_id VARCHAR(50) NOT NULL,	
	is_deleted BOOLEAN NOT NULL   
	);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_order ON orders (code);