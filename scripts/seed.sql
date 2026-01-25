-- Seed Data

-- Clear existing data (optional, careful in prod)
TRUNCATE TABLE ticket_types, tickets, orders, events, categories, users CASCADE;
ALTER SEQUENCE users_id_seq RESTART WITH 1;
ALTER SEQUENCE categories_id_seq RESTART WITH 1;
ALTER SEQUENCE events_id_seq RESTART WITH 1;
ALTER SEQUENCE ticket_types_id_seq RESTART WITH 1;
-- Or use TRUNCATE ... RESTART IDENTITY but it might not work depending on PG version/setup sometimes. Explicit sequence restart is safer for ensuring ID=1.

-- Users
INSERT INTO users (name, email, password, role, status, phone) VALUES
('Admin User', 'admin@kartcis.id', '$2a$10$Ph2yUSw50SiaWH8K6Lykceckr8sXwBVBKCyb1J9TnkqgvrBKypzRe', 'admin', 'active', '081234567890'),
('John Doe', 'john@example.com', '$2a$10$Ph2yUSw50SiaWH8K6Lykceckr8sXwBVBKCyb1J9TnkqgvrBKypzRe', 'user', 'active', '081234567891'),
('Jane Smith', 'jane@example.com', '$2a$10$Ph2yUSw50SiaWH8K6Lykceckr8sXwBVBKCyb1J9TnkqgvrBKypzRe', 'user', 'active', '081234567892'),
('Budi Organizer', 'budi@organizer.com', '$2a$10$Ph2yUSw50SiaWH8K6Lykceckr8sXwBVBKCyb1J9TnkqgvrBKypzRe', 'organizer', 'active', '081234567893');

-- REALISTIC DATA SEED

-- 1. Realistic Categories (Exactly 6)
INSERT INTO categories (name, slug, description, icon, is_active, display_order, image) VALUES
('Music & Concerts', 'music-concerts', 'Live music performances, festivals, and gigs.', 'Music', true, 1, 'http://localhost:8000/api/v1/uploads/dummy.jpg'),
('Sports & Fitness', 'sports-fitness', 'Marathons, football matches, and yoga sessions.', 'Activity', true, 2, 'http://localhost:8000/api/v1/uploads/dummy.jpg'),
('Technology & Coding', 'tech-coding', 'Developer conferences, workshops, and hackathons.', 'Cpu', true, 3, 'http://localhost:8000/api/v1/uploads/dummy.jpg'),
('Art & Museum', 'art-museum', 'Exhibitions, galleries, and creative workshops.', 'PenTool', true, 4, 'http://localhost:8000/api/v1/uploads/dummy.jpg'),
('Food & Culinary', 'food-culinary', 'Cooking classes, food festivals, and fine dining.', 'Coffee', true, 5, 'http://localhost:8000/api/v1/uploads/dummy.jpg'),
('Business & Startup', 'business-startup', 'Networking events, seminars, and pitch nights.', 'Briefcase', true, 6, 'http://localhost:8000/api/v1/uploads/dummy.jpg');

-- 2. Realistic Events (20+ items)
INSERT INTO events (title, slug, description, detailed_description, event_date, event_time, venue, city, organizer, image, quota, is_featured, status, category_id, min_price, max_price, custom_fields)
SELECT 
    CASE 
        WHEN i = 1 THEN 'Java Jazz Festival 2026'
        WHEN i = 2 THEN 'Jakarta Marathon'
        WHEN i = 3 THEN 'Go Conference Indonesia'
        WHEN i = 4 THEN 'Van Gogh Art Exhibition'
        WHEN i = 5 THEN 'Street Food Festival'
        WHEN i = 6 THEN 'Startup Pitch Night'
        WHEN i = 7 THEN 'Rock in Solo'
        WHEN i = 8 THEN 'Yoga in the Park'
        WHEN i = 9 THEN 'AI & Future Workshop'
        WHEN i = 10 THEN 'Local Pottery Class'
        WHEN i = 11 THEN 'Wine & Cheese Tasting'
        WHEN i = 12 THEN 'E-Commerce Summit'
        WHEN i = 13 THEN 'Indie Movie Night'
        WHEN i = 14 THEN 'Badminton Open'
        WHEN i = 15 THEN 'Cyber Security Seminar'
        WHEN i = 16 THEN 'Photography Walk'
        WHEN i = 17 THEN 'Baking Masterclass'
        WHEN i = 18 THEN 'Investasi Saham 101'
        WHEN i = 19 THEN 'Dewa 19 Reunion'
        WHEN i = 20 THEN 'Crossfit Games'
        ELSE 'Generic Event ' || i
    END,
    CASE 
        WHEN i = 1 THEN 'java-jazz-2026'
        WHEN i = 2 THEN 'jakarta-marathon-2026'
        WHEN i = 3 THEN 'go-conf-id-2026'
        WHEN i = 4 THEN 'van-gogh-exhibition'
        WHEN i = 5 THEN 'street-food-fest'
        WHEN i = 6 THEN 'startup-pitch'
        WHEN i = 7 THEN 'rock-in-solo'
        WHEN i = 8 THEN 'yoga-park'
        WHEN i = 9 THEN 'ai-future-ws'
        WHEN i = 10 THEN 'pottery-class'
        WHEN i = 11 THEN 'wine-cheese'
        WHEN i = 12 THEN 'ecommerce-summit'
        WHEN i = 13 THEN 'indie-movie'
        WHEN i = 14 THEN 'badminton-open'
        WHEN i = 15 THEN 'cyber-security'
        WHEN i = 16 THEN 'photo-walk'
        WHEN i = 17 THEN 'baking-masterclass'
        WHEN i = 18 THEN 'saham-101'
        WHEN i = 19 THEN 'dewa19-reunion'
        WHEN i = 20 THEN 'crossfit-games'
        ELSE 'event-' || i
    END,
    'Interesting and engaging event for everyone.',
    'This event features top speakers and incredible experiences tailored for enthusiasts.',
    CURRENT_DATE + (i || ' days')::interval,
    '19:00:00',
    CASE WHEN i % 2 = 0 THEN 'Istora Senayan' ELSE 'ICE BSD' END,
    CASE WHEN i % 3 = 0 THEN 'Jakarta' WHEN i % 3 = 1 THEN 'Tangerang' ELSE 'Surabaya' END,
    'Kartcis Organizer',
    'http://localhost:8000/api/v1/uploads/dummy.jpg',
    0,
    i <= 5,
    'published',
    (i % 6) + 1,
    100000 * (i % 5 + 1),
    250000 * (i % 5 + 1),
    CASE 
        WHEN i = 2 THEN '[{"name":"Shirt Size","type":"select","options":["S","M","L","XL","XXL"],"required":true},{"name":"Emergency Contact Name","type":"text","required":true},{"name":"Emergency Contact Phone","type":"text","required":true},{"name":"Blood Type","type":"select","options":["A","B","AB","O"],"required":false}]'
        WHEN i = 17 THEN '[{"name":"Dietary Restrictions","type":"text","required":false}]'
        ELSE NULL
    END
FROM generate_series(1, 25) s(i);

-- Update quota to some non-zero
UPDATE events SET quota = 500;

-- 3. Ticket Types
INSERT INTO ticket_types (event_id, name, description, price, original_price, quota, available)
SELECT 
    e.id, 
    CASE WHEN s.j = 1 THEN 'Regular' ELSE 'VIP' END,
    'Access to ' || (CASE WHEN s.j = 1 THEN 'general area' ELSE 'private lounge' END),
    e.min_price * s.j,
    CASE 
        WHEN e.id % 2 = 0 THEN (e.min_price * s.j) + 50000 -- Discounted for even events
        ELSE 0 -- No discount for odd events
    END,
    e.quota / 2,
    e.quota / 2
FROM events e, generate_series(1, 2) s(j);

-- 4. Realistic Orders (20+ items with varied users and data)
INSERT INTO orders (user_id, order_number, customer_name, customer_email, customer_phone, total_amount, status, payment_method, created_at)
SELECT 
    CASE WHEN i % 4 = 0 THEN 1 WHEN i % 4 = 1 THEN 2 WHEN i % 4 = 2 THEN 3 ELSE 4 END,
    'KRT-' || (100000 + i),
    CASE 
        WHEN i % 4 = 0 THEN 'Admin User'
        WHEN i % 4 = 1 THEN 'John Doe'
        WHEN i % 4 = 2 THEN 'Jane Smith'
        ELSE 'Budi Organizer'
    END,
    CASE 
        WHEN i % 4 = 0 THEN 'admin@kartcis.id'
        WHEN i % 4 = 1 THEN 'john@example.com'
        WHEN i % 4 = 2 THEN 'jane@example.com'
        ELSE 'budi@organizer.com'
    END,
    CASE 
        WHEN i % 4 = 0 THEN '081234567890'
        WHEN i % 4 = 1 THEN '081234567891'
        WHEN i % 4 = 2 THEN '081234567892'
        ELSE '081234567893'
    END,
    250000 + (i * 1000),
    CASE WHEN i % 3 = 0 THEN 'paid' WHEN i % 3 = 1 THEN 'cancelled' ELSE 'pending' END,
    CASE 
        WHEN i % 5 = 0 THEN 'BCA VA'
        WHEN i % 5 = 1 THEN 'Mandiri VA'
        WHEN i % 5 = 2 THEN 'BNI VA'
        WHEN i % 5 = 3 THEN 'GoPay'
        ELSE 'OVO'
    END,
    NOW() - (i || ' hours')::interval
FROM generate_series(1, 22) s(i);

-- Add specific Jakarta Marathon orders to ensure custom field responses
INSERT INTO orders (user_id, order_number, customer_name, customer_email, customer_phone, total_amount, status, payment_method, created_at)
VALUES
    (1, 'KRT-100023', 'Admin User', 'admin@kartcis.id', '081234567890', 300000, 'paid', 'BCA VA', NOW() - INTERVAL '1 day'),
    (2, 'KRT-100024', 'John Doe', 'john@example.com', '081234567891', 300000, 'pending', 'Mandiri VA', NOW() - INTERVAL '2 hours'),
    (3, 'KRT-100025', 'Jane Smith', 'jane@example.com', '081234567892', 600000, 'paid', 'BNI VA', NOW() - INTERVAL '5 hours');

-- 5. Tickets linked to orders with custom field responses
INSERT INTO tickets (order_id, event_id, ticket_type_id, ticket_code, attendee_name, attendee_email, attendee_phone, status, custom_field_responses)
SELECT 
    o.id,
    -- Vary events: cycle through different events (1-25)
    ((o.id - 1) % 25) + 1,
    -- Ticket type based on event
    (((o.id - 1) % 25) * 2) + (CASE WHEN o.id % 2 = 0 THEN 1 ELSE 2 END),
    'TKT-' || o.order_number || '-1',
    o.customer_name,
    o.customer_email,
    o.customer_phone,
    CASE WHEN o.status = 'cancelled' THEN 'cancelled' ELSE 'active' END,
    -- Add custom field responses for specific events
    CASE 
        -- Jakarta Marathon (event_id = 2)
        WHEN ((o.id - 1) % 25) + 1 = 2 THEN 
            CASE 
                WHEN o.id % 4 = 0 THEN '{"Shirt Size":"XL","Emergency Contact Name":"Budi Santoso","Emergency Contact Phone":"081234567890","Blood Type":"A"}'
                WHEN o.id % 4 = 1 THEN '{"Shirt Size":"L","Emergency Contact Name":"Dewi Lestari","Emergency Contact Phone":"081234567891","Blood Type":"B"}'
                WHEN o.id % 4 = 2 THEN '{"Shirt Size":"M","Emergency Contact Name":"Agus Wijaya","Emergency Contact Phone":"081234567892","Blood Type":"O"}'
                ELSE '{"Shirt Size":"XXL","Emergency Contact Name":"Rina Kusuma","Emergency Contact Phone":"081234567893","Blood Type":"AB"}'
            END
        -- Baking Masterclass (event_id = 17)
        WHEN ((o.id - 1) % 25) + 1 = 17 THEN 
            CASE 
                WHEN o.id % 3 = 0 THEN '{"Dietary Restrictions":"Vegetarian"}'
                WHEN o.id % 3 = 1 THEN '{"Dietary Restrictions":"No peanuts"}'
                ELSE '{"Dietary Restrictions":"None"}'
            END
        ELSE NULL
    END
FROM orders o;

-- Add tickets for Jakarta Marathon specific orders
INSERT INTO tickets (order_id, event_id, ticket_type_id, ticket_code, attendee_name, attendee_email, attendee_phone, status, custom_field_responses)
VALUES
    ((SELECT id FROM orders WHERE order_number = 'KRT-100023'), 2, 2, 'TKT-MARATHON-001', 'Admin User', 'admin@kartcis.id', '081234567890', 'active', '{"Shirt Size":"L","Emergency Contact Name":"Siti Rahayu","Emergency Contact Phone":"081298765432","Blood Type":"O"}'),
    ((SELECT id FROM orders WHERE order_number = 'KRT-100024'), 2, 2, 'TKT-MARATHON-002', 'John Doe', 'john@example.com', '081234567891', 'active', '{"Shirt Size":"XL","Emergency Contact Name":"Mary Doe","Emergency Contact Phone":"081298765433","Blood Type":"A"}'),
    ((SELECT id FROM orders WHERE order_number = 'KRT-100025'), 2, 2, 'TKT-MARATHON-003', 'Jane Smith', 'jane@example.com', '081234567892', 'active', '{"Shirt Size":"M","Emergency Contact Name":"Robert Smith","Emergency Contact Phone":"081298765434","Blood Type":"B"}'),
    ((SELECT id FROM orders WHERE order_number = 'KRT-100025'), 2, 27, 'TKT-MARATHON-004', 'Jane Smith', 'jane@example.com', '081234567892', 'active', '{"Shirt Size":"M","Emergency Contact Name":"Robert Smith","Emergency Contact Phone":"081298765434","Blood Type":"B"}');

-- Set 0% Admin Fee for specific events (e.g. Workshops/Education) that should be free from fees
UPDATE events SET fee_percentage = 0.0 WHERE id IN (9, 17, 21);
