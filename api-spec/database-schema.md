# Database Schema

## Tables Overview

1. `users` - User accounts
2. `social_accounts` - OAuth connections (Google, etc)
3. `categories` - Event categories
4. `events` - Event listings
5. `ticket_types` - Ticket types per event
6. `orders` - Customer orders
7. `order_items` - Order line items
8. `tickets` - Individual tickets
9. `event_analytics` - Event statistics
10. `activity_logs` - Admin activity logs

---

## 1. users

```sql
CREATE TABLE users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255), -- NULL untuk OAuth users
  phone VARCHAR(20),
  role ENUM('user', 'admin', 'organizer') DEFAULT 'user',
  avatar VARCHAR(500),
  is_active BOOLEAN DEFAULT TRUE,
  email_verified_at TIMESTAMP NULL,
  last_login_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**Roles:**
- `user` - Regular user
- `admin` - Full system access
- `organizer` - Can create/manage own events

---

## 2. social_accounts

```sql
CREATE TABLE social_accounts (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  provider VARCHAR(50) NOT NULL, -- 'google', 'facebook', 'apple'
  provider_id VARCHAR(255) NOT NULL,
  provider_token TEXT,
  provider_refresh_token TEXT,
  provider_data JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE KEY (provider, provider_id)
);
```

---

## 3. categories

```sql
CREATE TABLE categories (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(100) UNIQUE NOT NULL,
  description TEXT,
  icon VARCHAR(50),
  image VARCHAR(500),
  is_active BOOLEAN DEFAULT TRUE,
  display_order INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**Default Categories:**
- Marathon & Lari
- Musik & Konser
- Workshop & Seminar
- Olahraga
- Kuliner
- Charity & Sosial

---

## 4. events

```sql
CREATE TABLE events (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  category_id BIGINT NOT NULL,
  organizer_id BIGINT,
  title VARCHAR(255) NOT NULL,
  slug VARCHAR(255) UNIQUE NOT NULL,
  description TEXT NOT NULL,
  detailed_description TEXT,
  facilities JSON, -- ["Medali", "Sertifikat", ...]
  terms JSON, -- ["Minimal 17 tahun", ...]
  agenda JSON, -- [{"time":"06:00","activity":"Start"}]
  organizer_info JSON, -- {"name":"","phone":"","email":""}
  faqs JSON, -- [{"question":"","answer":""}]
  event_date DATE NOT NULL,
  event_time TIME,
  venue VARCHAR(255) NOT NULL,
  address TEXT,
  city VARCHAR(100) NOT NULL,
  province VARCHAR(100),
  organizer VARCHAR(255) NOT NULL,
  quota INT NOT NULL DEFAULT 0,
  image VARCHAR(500),
  is_featured BOOLEAN DEFAULT FALSE,
  view_count INT DEFAULT 0,
  status ENUM('draft', 'published', 'completed', 'cancelled', 'sold-out') DEFAULT 'draft',
  cancel_reason TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (category_id) REFERENCES categories(id),
  FOREIGN KEY (organizer_id) REFERENCES users(id)
);
```

---

## 5. ticket_types

```sql
CREATE TABLE ticket_types (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  event_id BIGINT NOT NULL,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  price DECIMAL(12,2) NOT NULL,
  original_price DECIMAL(12,2), -- Untuk show discount
  quota INT NOT NULL,
  available INT NOT NULL,
  status ENUM('available', 'sold_out', 'unavailable') DEFAULT 'available',
  sale_start_date TIMESTAMP,
  sale_end_date TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);
```

---

## 6. orders

```sql
CREATE TABLE orders (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NULL, -- NULL untuk guest checkout
  order_number VARCHAR(50) UNIQUE NOT NULL,
  customer_name VARCHAR(255) NOT NULL,
  customer_email VARCHAR(255) NOT NULL,
  customer_phone VARCHAR(20) NOT NULL,
  total_amount DECIMAL(12,2) NOT NULL,
  status ENUM('pending', 'paid', 'cancelled', 'expired') DEFAULT 'pending',
  payment_method VARCHAR(50) NOT NULL,
  payment_details JSON,
  expires_at TIMESTAMP NOT NULL, -- 24 jam dari created_at
  paid_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);
```

---

## 7. order_items

```sql
CREATE TABLE order_items (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  order_id BIGINT NOT NULL,
  event_id BIGINT NOT NULL,
  ticket_type_id BIGINT NOT NULL,
  quantity INT NOT NULL,
  unit_price DECIMAL(12,2) NOT NULL,
  subtotal DECIMAL(12,2) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
  FOREIGN KEY (event_id) REFERENCES events(id),
  FOREIGN KEY (ticket_type_id) REFERENCES ticket_types(id)
);
```

---

## 8. tickets

```sql
CREATE TABLE tickets (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  order_id BIGINT NOT NULL,
  event_id BIGINT NOT NULL,
  ticket_type_id BIGINT NOT NULL,
  ticket_code VARCHAR(50) UNIQUE NOT NULL,
  qr_code TEXT,
  attendee_name VARCHAR(255) NOT NULL,
  attendee_email VARCHAR(255) NOT NULL,
  attendee_phone VARCHAR(20) NOT NULL,
  status ENUM('active', 'used', 'cancelled') DEFAULT 'active',
  check_in_at TIMESTAMP NULL,
  checked_in_by BIGINT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
  FOREIGN KEY (event_id) REFERENCES events(id),
  FOREIGN KEY (ticket_type_id) REFERENCES ticket_types(id),
  FOREIGN KEY (checked_in_by) REFERENCES users(id)
);
```

---

## 9. event_analytics

```sql
CREATE TABLE event_analytics (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  event_id BIGINT NOT NULL,
  date DATE NOT NULL,
  views INT DEFAULT 0,
  tickets_sold INT DEFAULT 0,
  revenue DECIMAL(12,2) DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
  UNIQUE KEY (event_id, date)
);
```

---

## 10. activity_logs

```sql
CREATE TABLE activity_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT,
  action VARCHAR(100) NOT NULL,
  entity_type VARCHAR(50),
  entity_id BIGINT,
  description TEXT,
  ip_address VARCHAR(45),
  user_agent TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);
```

**Actions:** `create_event`, `update_event`, `delete_event`, `create_category`, `update_user_role`, etc.
