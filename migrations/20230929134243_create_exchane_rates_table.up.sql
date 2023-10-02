CREATE TYPE currencies AS ENUM ('rur', 'usd', 'euro');
CREATE TABLE IF NOT EXISTS exchane_rates (
	at_date date, 
    currency currencies,
	rate numeric
);
 