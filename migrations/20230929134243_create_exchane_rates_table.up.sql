CREATE TYPE currencies AS ENUM ('rur', 'usd', 'euro');
CREATE TABLE IF NOT EXISTS exchane_rates (
	at_date timestamp without time zone NOT NULL, 
    currency currencies,
	rate numeric
);
 