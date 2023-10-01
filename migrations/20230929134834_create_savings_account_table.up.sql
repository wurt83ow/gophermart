CREATE TABLE IF NOT EXISTS savings_account ( 
	user_id VARCHAR(50) NOT NULL,	
    processed_at date, 
	id_order_in VARCHAR(50),		
    id_order_out VARCHAR(50),
    count_points integer,	     
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE 
	);
