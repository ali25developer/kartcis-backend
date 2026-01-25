# KARTCIS.ID - API Specification

## 📋 Overview

Complete API specification untuk sistem ticketing event KARTCIS.ID. Backend tinggal implement sesuai spec ini, frontend sudah siap integrasi.

**Base URL Production:** `https://api.kartcis.id/api/v1`  
**Base URL Development:** `http://localhost:8080/api/v1`

## 🎯 Quick Start untuk Backend Developer

1. **Baca [database-schema.md](./database-schema.md)** - Setup 10 tables
2. **Pilih feature** - Implement endpoint per file di bawah
3. **Follow response format** - Semua response pakai wrapper standar
4. **Test** - Frontend sudah ready, tinggal ganti base URL

## 📁 API Documentation Structure

### Database & Models
- **[database-schema.md](./database-schema.md)** - 10 tables lengkap dengan relasi, indexes, dan constraints

### Public APIs (No Auth Required - Kecuali yang disebutkan)
- **[auth.md](./auth.md)** - Register, Login, OAuth Google, Set Password, Social Account Management
- **[events.md](./events.md)** - Browse events, Search, Filter, Event detail, Popular/Featured events
- **[categories.md](./categories.md)** - List categories dengan event count
- **[orders.md](./orders.md)** - Checkout (with/without login), Payment methods, Order tracking, Payment webhook
- **[tickets.md](./tickets.md)** - My tickets (auth required), Download PDF, Check-in, Verify ticket

### Admin APIs (Requires Admin Role)
- **[admin-dashboard.md](./admin-dashboard.md)** - Dashboard stats, Revenue overview, Charts, Analytics
- **[admin-events.md](./admin-events.md)** - Event CRUD, Change status, Analytics per event
- **[admin-tickets.md](./admin-tickets.md)** - Ticket type management (CRUD)
- **[admin-categories.md](./admin-categories.md)** - Category management (CRUD)
- **[admin-users.md](./admin-users.md)** - User management, Role management
- **[admin-transactions.md](./admin-transactions.md)** - Transaction list, Detail, Resend email, Export, Revenue summary
- **[admin-reports.md](./admin-reports.md)** - Reports & analytics (Sales, Events, Users)

## 🔐 Authentication

### JWT Bearer Token
Sebagian besar endpoint membutuhkan JWT token di header:
```
Authorization: Bearer {token}
```

### Token Flow
1. **Login/Register** → Dapat token dengan expiry 2-24 jam
2. **Save token** di localStorage/cookie (frontend)
3. **Include token** di semua request yang butuh auth
4. **Check expiry** dan refresh jika perlu

### Role-Based Access
- **user** - Regular user (my tickets, checkout)
- **admin** - Full access (dashboard, event management, transactions)
- **organizer** - Manage own events (optional, bisa di-skip dulu)

## 📊 Standard Response Format

### Success Response (HTTP 200/201)
```json
{
  "success": true,
  "message": "Success message (optional)",
  "data": { 
    // Response data
  }
}
```

### Error Response (HTTP 400/401/403/404/422/500)
```json
{
  "success": false,
  "message": "Error message",
  "error": "Detailed error (optional)",
  "errors": {
    "field_name": ["Validation error 1", "Validation error 2"]
  }
}
```

### Pagination Format
```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "current_page": 1,
      "total_pages": 5,
      "total_items": 58,
      "per_page": 12,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

## 🔢 HTTP Status Codes

- `200` - OK (Success)
- `201` - Created (Resource successfully created)
- `400` - Bad Request (Invalid input)
- `401` - Unauthorized (Authentication required or failed)
- `403` - Forbidden (Insufficient permissions)
- `404` - Not Found (Resource not found)
- `422` - Unprocessable Entity (Validation error)
- `500` - Internal Server Error

## 🚀 Integration Steps

### Step 1: Setup Database
```bash
# Run migrations untuk create 10 tables
# Lihat database-schema.md untuk struktur lengkap
```

### Step 2: Implement Core Features (Priority Order)

#### Priority 1 (Must Have)
1. **Auth** (`auth.md`) - Login, Register, JWT middleware
2. **Events** (`events.md`) - GET all events, GET detail
3. **Categories** (`categories.md`) - GET categories
4. **Orders** (`orders.md`) - POST create order
5. **Tickets** (`tickets.md`) - GET my tickets

#### Priority 2 (Important)
6. **Admin Dashboard** (`admin-dashboard.md`) - Stats
7. **Admin Events** (`admin-events.md`) - Event CRUD
8. **Admin Transactions** (`admin-transactions.md`) - List & detail
9. **Payment Gateway** - Integrate VA/e-wallet (Midtrans/Xendit)

#### Priority 3 (Nice to Have)
10. **OAuth Google** (`auth.md`) - Google login
11. **Email Notifications** - Ticket delivery, payment reminder
12. **PDF Generation** - Ticket PDF with QR code
13. **Reports** (`admin-reports.md`) - Analytics

### Step 3: Test dengan Frontend
```bash
# Frontend sudah ready, tinggal update base URL
# File: src/app/services/api.ts
# Ganti: const API_BASE_URL = 'http://localhost:8080/api/v1'
```

## 💡 Important Notes

### Guest Checkout
- User bisa checkout **tanpa login** (user_id = NULL)
- Email tiket dikirim ke customer_email
- Tiket bisa diakses via link di email

### Logged In Checkout
- User login → order.user_id terisi
- Tiket tersimpan di "My Tickets"
- Email tetap dikirim sebagai backup

### Ticket Generation
- Generate tickets **setelah payment confirmed**
- Ticket code format: `TIX-{event_id}-{timestamp}-{random}`
- QR code harus unique dan scannable

### Payment Integration
- Support VA (BCA, Mandiri, BNI, BRI)
- Support e-wallet (OVO, GoPay, DANA, ShopeePay)
- Support QRIS
- 24 jam expiry untuk pending orders
- Webhook untuk payment notification

### Data Validation
- Event date harus di masa depan
- Ticket quantity harus <= available
- Email format valid
- Phone number format Indonesia (08xxx)

## 🛠️ Tech Stack Recommendations

### Backend Framework
- **Go (Gin)** - Recommended (fast, simple)
- **Node.js (Express)** - Alternative
- **PHP (Laravel)** - Alternative

### Database
- **PostgreSQL** - Recommended (robust, good for transactions)
- **MySQL 8.0+** - Alternative (yang dipakai di schema)

### Payment Gateway
- **Midtrans** - Complete (VA, e-wallet, QRIS)
- **Xendit** - Alternative

### Email Service
- **SendGrid** - Reliable
- **AWS SES** - Cheap & scalable
- **Mailgun** - Alternative

### File Storage
- **AWS S3** - Recommended
- **Google Cloud Storage** - Alternative
- **Local Storage** - Development only

### PDF Generation
- **wkhtmltopdf** - HTML to PDF
- **PDFKit** (Node.js) - Programmatic PDF
- **FPDF** (PHP) - Simple & fast

## 📧 Email Templates Needed

1. **Order Confirmation** - Pending payment dengan VA/payment details
2. **Payment Success** - Tiket attachment (PDF) + download link
3. **Payment Reminder** - 6 jam sebelum expiry
4. **Event Reminder** - 3 hari & 1 hari sebelum event
5. **Welcome Email** - Setelah register

## 🔒 Security Checklist

- [ ] JWT secret key secure (env variable)
- [ ] Password hashing (bcrypt/argon2)
- [ ] SQL injection prevention (prepared statements)
- [ ] XSS prevention (sanitize input)
- [ ] CSRF protection
- [ ] Rate limiting (login, checkout, API calls)
- [ ] CORS configuration
- [ ] HTTPS only (production)
- [ ] Payment webhook signature verification

## 📈 Performance Optimization

- [ ] Database indexes (see database-schema.md)
- [ ] Query optimization (N+1 problem)
- [ ] Caching (Redis) - Event list, Categories
- [ ] Image optimization (compress, CDN)
- [ ] API response compression (gzip)
- [ ] Connection pooling
- [ ] Pagination semua list endpoints

## 🧪 Testing Priority

1. **Auth flow** - Register → Login → Get user
2. **Event browsing** - List → Filter → Detail
3. **Checkout flow** - Add to cart → Checkout → Payment
4. **Ticket access** - My tickets → Download PDF
5. **Admin operations** - Create event → Manage transactions

## 📞 Support

Jika ada yang kurang jelas atau butuh tambahan endpoint:
1. Check file-file `.md` di folder ini untuk detail lengkap
2. Lihat response examples di setiap endpoint
3. Match dengan TypeScript types di `/src/app/types/index.ts` (frontend)

---

**Version:** 1.0  
**Last Updated:** January 21, 2026  
**Frontend Compatible:** ✅ Ready to integrate