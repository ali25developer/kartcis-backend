-- Create referral_codes table
CREATE TABLE IF NOT EXISTS referral_codes (
    id SERIAL PRIMARY KEY,
    code VARCHAR(255) UNIQUE NOT NULL,
    partner_name VARCHAR(255) NOT NULL,
    user_id INTEGER,
    event_id INTEGER,
    discount_type VARCHAR(50) DEFAULT 'none',
    discount_value DECIMAL(16,2) DEFAULT 0,
    max_uses INTEGER DEFAULT 0,
    used_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create bank_transactions table
CREATE TABLE IF NOT EXISTS bank_transactions (
    id SERIAL PRIMARY KEY,
    order_id INTEGER,
    reference_id VARCHAR(255) UNIQUE NOT NULL,
    amount DECIMAL(16,2) NOT NULL,
    sender VARCHAR(255),
    bank_name VARCHAR(255),
    transaction_date TIMESTAMP WITH TIME ZONE,
    raw_data TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add referral_code column to orders if not exists
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='referral_code') THEN
        ALTER TABLE orders ADD COLUMN referral_code VARCHAR(255);
    END IF;
END $$;
