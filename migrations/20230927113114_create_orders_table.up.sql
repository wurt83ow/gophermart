CREATE TYPE statuses AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(50) PRIMARY KEY, 
    number VARCHAR(50),     
    date date,  
    status statuses  
    user_id VARCHAR(50) NOT NULL,        
	FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE  
    );
CREATE UNIQUE INDEX IF NOT EXISTS uniq_order ON orders (number);
