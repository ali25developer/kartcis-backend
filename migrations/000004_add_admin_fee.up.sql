-- Add fee_percentage to events (default 5.0)
ALTER TABLE events ADD COLUMN fee_percentage DECIMAL(5, 2) DEFAULT 5.0;

-- Add admin_fee to orders (to record fee at transaction time)
ALTER TABLE orders ADD COLUMN admin_fee DECIMAL(15, 2) DEFAULT 0;
