CREATE TYPE statuses AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE IF NOT EXISTS orders (
    order_id VARCHAR(50) PRIMARY KEY, 
    number VARCHAR(50),     
    date timestamp with time zone NOT NULL DEFAULT (current_timestamp AT TIME ZONE 'UTC'),       
    status statuses,  
    user_id VARCHAR(50) NOT NULL,        
	FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE  
    );
CREATE UNIQUE INDEX IF NOT EXISTS uniq_order ON orders (number);
