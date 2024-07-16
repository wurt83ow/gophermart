CREATE TABLE IF NOT EXISTS savings_account ( 
	user_id VARCHAR(50) NOT NULL,	
    processed_at timestamp without time zone NOT NULL, 
	id_order_in VARCHAR(50),		
    id_order_out VARCHAR(50),
    accrual numeric,	     
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE 
	);
