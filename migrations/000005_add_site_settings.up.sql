CREATE TABLE IF NOT EXISTS site_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed default values
INSERT INTO site_settings (key, value) VALUES 
('contact_email', 'support@kartcis.id'),
('contact_phone', '(+62) 8312-7246-830'),
('contact_address', 'Jl. Kaliurang Km 14, Yogyakarta'),
('facebook_url', 'https://facebook.com/'),
('twitter_url', 'https://twitter.com/'),
('instagram_url', 'https://instagram.com/')
ON CONFLICT (key) DO NOTHING;
