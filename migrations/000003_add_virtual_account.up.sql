-- Add virtual account number to orders
ALTER TABLE orders ADD COLUMN virtual_account_number VARCHAR(50);
