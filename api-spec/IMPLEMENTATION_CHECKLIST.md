# Backend Implementation Checklist

Complete checklist untuk implement backend KARTCIS.ID dari zero ke production-ready.

## 🎯 Overview

Total endpoints: **70+**  
Estimated time: **2-4 weeks** (1 developer, full-time)  
Tech stack: **Go (Gin) / Node.js (Express) / PHP (Laravel)**

---

## 📋 Phase 1: Setup & Infrastructure (2-3 days)

### Database Setup
- [ ] Install MySQL 8.0+ atau PostgreSQL
- [ ] Create database `kartcis_db`
- [ ] Setup connection pooling
- [ ] Configure timezone to UTC

### Database Tables (10 tables)
Refer to: `database-schema.md`

- [ ] **users** - User accounts dengan role (user/admin/organizer)
- [ ] **social_accounts** - OAuth connections (Google, Facebook, Apple)
- [ ] **categories** - Event categories
- [ ] **events** - Event listings dengan rich data
- [ ] **ticket_types** - Ticket types per event
- [ ] **orders** - Customer orders (support guest checkout)
- [ ] **order_items** - Order line items
- [ ] **tickets** - Individual tickets dengan QR code
- [ ] **event_analytics** - Event statistics
- [ ] **activity_logs** - Admin activity logging

### Database Indexes
- [ ] `users(email)` - UNIQUE index
- [ ] `events(slug)` - UNIQUE index
- [ ] `events(category_id, status, event_date)` - Composite index
- [ ] `orders(order_number)` - UNIQUE index
- [ ] `orders(user_id, status)` - Composite index
- [ ] `tickets(ticket_code)` - UNIQUE index
- [ ] `tickets(order_id)` - Index

### Project Structure
```
backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   ├── models/          # Database models (GORM/Sequelize)
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # Auth, CORS, etc
│   ├── services/        # Business logic
│   ├── repositories/    # Database operations
│   └── utils/           # Helper functions
├── migrations/          # Database migrations
├── .env                 # Environment variables
└── go.mod / package.json
```

### Environment Variables
```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=kartcis_db

# JWT
JWT_SECRET=your-super-secret-key-change-this-in-production
JWT_EXPIRY=86400  # 24 hours in seconds

# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# Payment Gateway (Midtrans/Xendit)
PAYMENT_GATEWAY_SERVER_KEY=your-server-key
PAYMENT_GATEWAY_CLIENT_KEY=your-client-key
PAYMENT_GATEWAY_ENV=sandbox  # production

# Email
EMAIL_PROVIDER=sendgrid  # or aws-ses, mailgun
EMAIL_API_KEY=your-api-key
EMAIL_FROM=noreply@kartcis.id
EMAIL_FROM_NAME=KARTCIS.ID

# File Storage
STORAGE_PROVIDER=s3  # or gcs, local
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_BUCKET_NAME=kartcis-storage
AWS_REGION=ap-southeast-1

# Server
PORT=8080
FRONTEND_URL=http://localhost:5173
ALLOWED_ORIGINS=http://localhost:5173,https://kartcis.id

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=60  # seconds
```

### Middleware Setup
- [ ] **CORS** - Allow frontend origin
- [ ] **Rate Limiting** - Prevent abuse
- [ ] **Request Logger** - Log all requests
- [ ] **Error Handler** - Centralized error handling
- [ ] **Auth Middleware** - JWT verification
- [ ] **Admin Middleware** - Role check (admin only)

---

## 🔐 Phase 2: Authentication (3-4 days)

Refer to: `auth.md`

### Core Auth Endpoints
- [ ] `POST /auth/register` - User registration
  - Validate email format
  - Hash password (bcrypt/argon2)
  - Generate JWT token
  - Return user + token
  
- [ ] `POST /auth/login` - User login
  - Validate credentials
  - Generate JWT token
  - Update last_login_at
  - Support remember_me (extend token expiry)
  
- [ ] `POST /auth/logout` - Logout
  - Invalidate token (optional: token blacklist)
  
- [ ] `GET /auth/me` - Get current user
  - Verify JWT token
  - Return user data

### OAuth Google (Optional - Phase 3)
- [ ] `GET /auth/google` - Initiate Google OAuth
- [ ] `GET /auth/google/callback` - Google callback handler
- [ ] `POST /auth/google/one-tap` - Google One Tap
- [ ] `GET /auth/social` - Get connected accounts
- [ ] `DELETE /auth/social/{provider}` - Unlink account
- [ ] `POST /auth/set-password` - Set password for OAuth users

### Auth Testing
- [ ] Test register dengan valid data
- [ ] Test register dengan duplicate email → 422 error
- [ ] Test login dengan correct credentials
- [ ] Test login dengan wrong password → 401 error
- [ ] Test protected endpoint tanpa token → 401 error
- [ ] Test protected endpoint dengan invalid token → 401 error
- [ ] Test token expiry

---

## 🎫 Phase 3: Public APIs (4-5 days)

### Events API
Refer to: `events.md`

- [ ] `GET /events` - Get all events
  - Support pagination (page, limit)
  - Support filters (category, city, status, featured)
  - Support search (title, venue, description)
  - Support sorting (date_asc, date_desc, popular)
  - Support price range (min_price, max_price)
  - Support date range (date_from, date_to)
  - Include category & ticket_types
  
- [ ] `GET /events/{slug}` - Get event detail
  - Include all rich data (facilities, terms, agenda, faqs, organizer_info)
  - Include ticket_types dengan availability
  - Increment view_count
  
- [ ] `GET /events/featured` - Featured events
- [ ] `GET /events/popular` - Popular events (by view_count)
- [ ] `GET /events/upcoming` - Upcoming events (event_date >= today)
- [ ] `GET /cities` - Get cities dengan event count

### Categories API
Refer to: `categories.md`

- [ ] `GET /categories` - Get all categories
  - Include event_count
  - Order by display_order
  
- [ ] `GET /categories/{slug}` - Get category detail

### Orders API
Refer to: `orders.md`

- [ ] `POST /orders` - Create order (Checkout)
  - **Auth optional** (guest checkout support)
  - Validate ticket availability
  - Calculate total amount
  - Generate order_number: `ORD-{timestamp}-{random}`
  - Generate VA number (integration dengan payment gateway)
  - Set expires_at = now + 24 hours
  - Create order_items
  - **Don't create tickets yet** (wait for payment)
  - Return order with payment_details
  
- [ ] `GET /orders/{order_number}` - Get order detail
  - Include order_items dengan event & ticket_type
  - Include payment_details
  
- [ ] `POST /orders/payment-callback` - Payment webhook
  - Verify webhook signature
  - Update order status to 'paid'
  - Set paid_at timestamp
  - **Generate tickets** (create ticket records)
  - **Send email** dengan ticket PDF
  - Reduce ticket availability
  
- [ ] `POST /orders/{order_number}/simulate-payment` - Dev only
  - Mark order as paid
  - Generate tickets
  - Send email

### Tickets API
Refer to: `tickets.md`

- [ ] `GET /tickets/my-tickets` - Get user's tickets
  - **Auth required**
  - Group by upcoming & past (compare event_date dengan today)
  - Include event, ticket_type, order info
  
- [ ] `GET /tickets/{ticket_code}` - Get ticket detail
  - Public access (untuk guest)
  - Include event & organizer info
  
- [ ] `GET /tickets/{ticket_code}/download` - Download PDF
  - Generate PDF dengan QR code
  - Return as binary file
  
- [ ] `POST /tickets/check-in` - Check-in ticket
  - **Admin/Organizer only**
  - Validate ticket belum used
  - Update status to 'used'
  - Set check_in_at timestamp
  
- [ ] `GET /tickets/{ticket_code}/verify` - Verify ticket
  - Check ticket status
  - Return is_valid boolean

### Testing Public APIs
- [ ] Browse events dengan berbagai filter
- [ ] Search events
- [ ] View event detail
- [ ] Guest checkout (tanpa login)
- [ ] Logged in checkout
- [ ] View order status
- [ ] Payment simulation
- [ ] View my tickets (upcoming & past)
- [ ] Download ticket PDF

---

## 👑 Phase 4: Admin APIs (5-7 days)

### Admin Dashboard
Refer to: `admin-dashboard.md`

- [ ] `GET /admin/stats` - Dashboard summary
  - Total events (by status)
  - Total transactions (by status)
  - Revenue (today, week, month, total)
  - Users (total, new today/week/month)
  - Tickets sold (today, week, month, total)
  
- [ ] `GET /admin/dashboard/revenue` - Revenue overview
  - Chart data by period
  - Top events by revenue
  
- [ ] `GET /admin/dashboard/transactions-overview` - Transaction stats
- [ ] `GET /admin/dashboard/events-overview` - Event analytics
- [ ] `GET /admin/dashboard/users-overview` - User analytics

### Admin Events Management
Refer to: `admin-events.md`

- [ ] `GET /admin/events` - List all events (admin view)
  - Include tickets_sold & revenue
  - Support search & filters
  
- [ ] `POST /admin/events` - Create event
  - Validate all fields
  - Generate slug from title
  - Set status = 'draft' default
  - Upload image to storage (S3/GCS)
  
- [ ] `GET /admin/events/{id}` - Get event detail (admin)
  - Include statistics (views, tickets sold, revenue)
  
- [ ] `PUT /admin/events/{id}` - Update event
  - Validate fields
  - Update slug jika title berubah
  
- [ ] `DELETE /admin/events/{id}` - Delete event
  - Check jika sudah ada transaksi → reject
  - Soft delete atau hard delete
  
- [ ] `PATCH /admin/events/{id}/status` - Change status
  - draft → published
  - published → completed/cancelled
  - If cancelled, set cancel_reason
  
- [ ] `GET /admin/events/{id}/analytics` - Event analytics
  - Daily stats (views, tickets sold, revenue)
  - Ticket type breakdown
  - Traffic sources (optional)

### Admin Ticket Types
Refer to: `admin-tickets.md`

- [ ] `GET /admin/ticket-types?event_id={id}` - List ticket types
- [ ] `POST /admin/ticket-types` - Create ticket type
- [ ] `GET /admin/ticket-types/{id}` - Get detail
- [ ] `PUT /admin/ticket-types/{id}` - Update
- [ ] `DELETE /admin/ticket-types/{id}` - Delete (if no orders)
- [ ] `PATCH /admin/ticket-types/{id}/status` - Change status

### Admin Transactions
Refer to: `admin-transactions.md`

- [ ] `GET /admin/transactions` - List transactions
  - Support search (order_number, customer_name, email)
  - Support filters (status, date range)
  - Include summary stats
  - Pagination
  
- [ ] `GET /admin/transactions/{id}` - Transaction detail
  - Include order_items, tickets, user
  
- [ ] `POST /admin/transactions/{id}/resend-email` - Resend email
  - Only for paid transactions
  - Send email dengan ticket PDF
  
- [ ] `POST /admin/transactions/{id}/cancel` - Cancel transaction
  - Only pending transactions
  - Return ticket quota
  
- [ ] `POST /admin/transactions/{id}/mark-paid` - Manual payment
  - Mark as paid
  - Generate tickets
  - Send email
  
- [ ] `GET /admin/transactions/export` - Export
  - Format: CSV/Excel/PDF
  - Filter by status & date range
  
- [ ] `GET /admin/transactions/{id}/timeline` - Transaction history
- [ ] `GET /admin/transactions/revenue-summary` - Revenue stats

### Admin Categories
Refer to: `admin-categories.md`

- [ ] `GET /admin/categories` - List (admin view)
- [ ] `POST /admin/categories` - Create
- [ ] `GET /admin/categories/{id}` - Detail
- [ ] `PUT /admin/categories/{id}` - Update
- [ ] `DELETE /admin/categories/{id}` - Delete (if no events)
- [ ] `PATCH /admin/categories/{id}/reorder` - Reorder

### Admin Users
Refer to: `admin-users.md`

- [ ] `GET /admin/users` - List users
- [ ] `GET /admin/users/{id}` - User detail
- [ ] `POST /admin/users` - Create admin/organizer
- [ ] `PUT /admin/users/{id}` - Update user
- [ ] `DELETE /admin/users/{id}` - Delete user
- [ ] `PATCH /admin/users/{id}/role` - Change role
- [ ] `PATCH /admin/users/{id}/status` - Activate/Deactivate
- [ ] `GET /admin/users/{id}/activity` - Activity log
- [ ] `GET /admin/users/{id}/transactions` - User's transactions

### Admin Reports
Refer to: `admin-reports.md`

- [ ] `GET /admin/reports/sales` - Sales report
- [ ] `GET /admin/reports/events` - Events performance
- [ ] `GET /admin/reports/users` - User acquisition
- [ ] `GET /admin/reports/revenue` - Revenue by period
- [ ] `GET /admin/reports/top-events` - Top events
- [ ] `GET /admin/reports/export` - Export report

### Testing Admin Features
- [ ] Admin login (role = admin)
- [ ] Regular user cannot access admin endpoints → 403
- [ ] Dashboard stats accurate
- [ ] Create/update/delete events
- [ ] Manage ticket types
- [ ] View & filter transactions
- [ ] Resend email functionality
- [ ] Export transactions
- [ ] User management
- [ ] Reports generation

---

## 💳 Phase 5: Payment Integration (3-4 days)

### Payment Gateway Options
Choose one: **Midtrans** (recommended) or **Xendit**

### Midtrans Integration
- [ ] Install SDK
- [ ] Configure server key & client key
- [ ] Create VA (Virtual Account)
  - BCA: prefix 80777
  - Mandiri: prefix 8901
  - BNI: prefix 8808
  - BRI: prefix 26215
  
- [ ] Create e-wallet charge
  - OVO
  - GoPay
  - ShopeePay
  - DANA
  
- [ ] Create QRIS
- [ ] Setup webhook handler
  - Verify signature
  - Handle notification
  - Update order status
  - Generate tickets
  - Send email

### Payment Flow
1. User checkout → Backend create order
2. Backend call payment gateway API
3. Payment gateway return VA number / payment link
4. Backend save payment_details
5. User bayar via bank/e-wallet
6. Payment gateway send webhook notification
7. Backend verify & update order status
8. Backend generate tickets & send email

### Testing Payment
- [ ] Create order dengan berbagai payment methods
- [ ] VA number generated correctly
- [ ] Webhook received & processed
- [ ] Order status updated
- [ ] Tickets generated
- [ ] Email sent

---

## 📧 Phase 6: Email System (2-3 days)

### Email Provider Setup
Choose one: **SendGrid** / **AWS SES** / **Mailgun**

### Email Templates
- [ ] **Order Confirmation** (Pending payment)
  - Order number
  - Event details
  - Total amount
  - Payment instructions (VA/e-wallet)
  - Expiry time (24 jam countdown)
  - Payment link (optional)
  
- [ ] **Payment Success** (Ticket delivery)
  - Thank you message
  - Event details
  - Ticket attachments (PDF)
  - Download ticket link
  - QR code preview
  - Event reminder
  
- [ ] **Payment Reminder** (6 hours before expiry)
  - Reminder to complete payment
  - Payment instructions
  - Time remaining
  
- [ ] **Event Reminder**
  - 3 days before event
  - 1 day before event
  - Event details
  - Venue & time
  - What to bring
  
- [ ] **Welcome Email** (After register)
  - Welcome message
  - Account info
  - Browse events link

### Email Queue (Optional - Production)
- [ ] Setup queue system (Redis/RabbitMQ)
- [ ] Process emails asynchronously
- [ ] Retry failed emails

### Testing Emails
- [ ] Order confirmation sent after checkout
- [ ] Payment success email dengan PDF attachment
- [ ] Email readable di mobile & desktop
- [ ] Links working correctly
- [ ] PDF downloadable

---

## 📄 Phase 7: PDF Generation (2 days)

### PDF Library
Choose one: **wkhtmltopdf** / **PDFKit** (Node.js) / **FPDF** (PHP)

### Ticket PDF Content
- [ ] Header (KARTCIS.ID logo)
- [ ] Ticket code (large, bold)
- [ ] QR code (scannable, high quality)
- [ ] Event details
  - Title
  - Date & time
  - Venue & address
- [ ] Ticket type & price
- [ ] Attendee info (name, email, phone)
- [ ] Order number
- [ ] Terms & conditions
- [ ] Footer (contact info)

### PDF Features
- [ ] Generate PDF dari HTML template
- [ ] Generate QR code (ticket_code)
- [ ] Save to storage (S3/GCS)
- [ ] Return download URL
- [ ] Cache PDF (regenerate only if needed)

### Testing PDF
- [ ] PDF generated correctly
- [ ] QR code scannable
- [ ] All info displayed
- [ ] File size reasonable (<500KB per ticket)
- [ ] PDF viewable di mobile & desktop

---

## 🔍 Phase 8: Search & Optimization (2 days)

### Search Implementation
- [ ] Full-text search on events
  - Title, description, venue, city, organizer
- [ ] Case-insensitive search
- [ ] Search highlighting (optional)
- [ ] Search suggestions (optional)

### Database Optimization
- [ ] Add indexes (see database-schema.md)
- [ ] Optimize N+1 queries (use eager loading)
- [ ] Query profiling
- [ ] Slow query log

### Caching (Optional - Production)
- [ ] Setup Redis
- [ ] Cache event list (5 min TTL)
- [ ] Cache categories (30 min TTL)
- [ ] Cache dashboard stats (1 min TTL)
- [ ] Invalidate cache on update

### Performance
- [ ] API response time < 200ms
- [ ] Database query time < 50ms
- [ ] Image optimization (compress, resize)
- [ ] API response compression (gzip)

---

## 🧪 Phase 9: Testing & QA (3-4 days)

### Unit Tests
- [ ] Auth service (register, login, JWT)
- [ ] Event service (CRUD, filters)
- [ ] Order service (checkout, payment)
- [ ] Ticket service (generation, check-in)
- [ ] Payment service (webhook, verification)

### Integration Tests
- [ ] Complete checkout flow
- [ ] Payment webhook flow
- [ ] Email sending
- [ ] PDF generation
- [ ] Admin CRUD operations

### API Tests (Postman/Thunder Client)
- [ ] Test all public endpoints
- [ ] Test all admin endpoints
- [ ] Test error scenarios
- [ ] Test edge cases
- [ ] Test rate limiting

### Load Testing (Optional)
- [ ] Test concurrent requests
- [ ] Test database connection pool
- [ ] Test under high load
- [ ] Identify bottlenecks

### Security Testing
- [ ] SQL injection prevention
- [ ] XSS prevention
- [ ] CSRF protection
- [ ] JWT security
- [ ] Rate limiting
- [ ] Input validation
- [ ] Password hashing

---

## 🚀 Phase 10: Deployment (2-3 days)

### Deployment Options
Choose one:
- **VPS** (DigitalOcean, Linode, Vultr)
- **Cloud** (AWS EC2, Google Cloud, Azure)
- **PaaS** (Heroku, Railway, Render)
- **Container** (Docker + Kubernetes)

### Pre-Deployment Checklist
- [ ] Environment variables configured
- [ ] Database migrations ready
- [ ] SSL certificate (HTTPS)
- [ ] Domain configured
- [ ] Email service configured
- [ ] Payment gateway (production mode)
- [ ] File storage (S3/GCS)
- [ ] Monitoring setup (Sentry/New Relic)
- [ ] Logging setup

### Deployment Steps
1. [ ] Setup production server
2. [ ] Install dependencies
3. [ ] Configure environment variables
4. [ ] Run database migrations
5. [ ] Build application
6. [ ] Setup reverse proxy (Nginx/Caddy)
7. [ ] Configure SSL (Let's Encrypt)
8. [ ] Setup process manager (PM2/systemd)
9. [ ] Test production API
10. [ ] Connect frontend to production API

### Post-Deployment
- [ ] Monitor API logs
- [ ] Monitor error logs
- [ ] Monitor performance
- [ ] Setup backup (database, files)
- [ ] Setup CI/CD (optional)

---

## 📊 Progress Tracking

### Overall Progress

**Phase 1:** Setup & Infrastructure ⬜⬜⬜⬜⬜ 0%  
**Phase 2:** Authentication ⬜⬜⬜⬜⬜ 0%  
**Phase 3:** Public APIs ⬜⬜⬜⬜⬜ 0%  
**Phase 4:** Admin APIs ⬜⬜⬜⬜⬜ 0%  
**Phase 5:** Payment Integration ⬜⬜⬜⬜⬜ 0%  
**Phase 6:** Email System ⬜⬜⬜⬜⬜ 0%  
**Phase 7:** PDF Generation ⬜⬜⬜⬜⬜ 0%  
**Phase 8:** Search & Optimization ⬜⬜⬜⬜⬜ 0%  
**Phase 9:** Testing & QA ⬜⬜⬜⬜⬜ 0%  
**Phase 10:** Deployment ⬜⬜⬜⬜⬜ 0%  

**Total Progress:** 0%

---

## 🎓 Learning Resources

### Go (Gin Framework)
- Official Docs: https://gin-gonic.com/docs/
- GORM (ORM): https://gorm.io/docs/
- JWT: github.com/golang-jwt/jwt

### Node.js (Express)
- Express Docs: https://expressjs.com/
- Sequelize (ORM): https://sequelize.org/
- JWT: jsonwebtoken

### PHP (Laravel)
- Laravel Docs: https://laravel.com/docs
- Eloquent ORM: Built-in
- JWT: tymon/jwt-auth

### Payment Gateway
- Midtrans: https://docs.midtrans.com/
- Xendit: https://developers.xendit.co/

### Email Services
- SendGrid: https://docs.sendgrid.com/
- AWS SES: https://docs.aws.amazon.com/ses/

---

## 📞 Support & Questions

Jika stuck atau butuh clarification:
1. ✅ Check API spec files di `/api-spec/`
2. ✅ Check ENDPOINTS_CHEATSHEET.md
3. ✅ Check FRONTEND_INTEGRATION_GUIDE.md
4. ✅ Check TypeScript types di `/src/app/types/index.ts`
5. ✅ Test dengan frontend yang sudah ready

---

**Happy Coding! 🚀**

**Last Updated:** January 21, 2026
