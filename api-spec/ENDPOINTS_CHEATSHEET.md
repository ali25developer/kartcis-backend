# API Endpoints Cheatsheet

Quick reference untuk semua endpoints. Untuk detail lengkap, lihat file masing-masing feature.

## 📋 Legend

- 🔓 Public (No auth required)
- 🔐 Auth Required (User/Admin)
- 👑 Admin Only

---

## 🔐 Authentication (`/api/v1/auth`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/auth/register` | 🔓 | Register user baru |
| POST | `/auth/login` | 🔓 | Login dengan email & password |
| POST | `/auth/logout` | 🔐 | Logout dan invalidate token |
| GET | `/auth/me` | 🔐 | Get current user data |
| GET | `/auth/google` | 🔓 | Initiate Google OAuth |
| GET | `/auth/google/callback` | 🔓 | Google OAuth callback |
| POST | `/auth/google/one-tap` | 🔓 | Google One Tap login |
| GET | `/auth/social` | 🔐 | Get connected social accounts |
| DELETE | `/auth/social/{provider}` | 🔐 | Unlink social account |
| POST | `/auth/set-password` | 🔐 | Set password for OAuth users |

---

## 🎫 Events (`/api/v1/events`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/events` | 🔓 | Get all events dengan filter & pagination |
| GET | `/events/{slug}` | 🔓 | Get event detail by slug |
| GET | `/events/popular` | 🔓 | Get popular events |
| GET | `/events/featured` | 🔓 | Get featured events |
| GET | `/events/upcoming` | 🔓 | Get upcoming events |
| GET | `/cities` | 🔓 | Get cities dengan event count |

**Query Parameters (GET /events):**
- `page`, `limit` - Pagination
- `category`, `city`, `province` - Filter
- `search` - Search by title/venue/description
- `featured`, `status` - Filter
- `sort` - date_asc/date_desc/popular/newest
- `min_price`, `max_price` - Price range
- `date_from`, `date_to` - Date range

---

## 📂 Categories (`/api/v1/categories`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/categories` | 🔓 | Get all categories |
| GET | `/categories/{slug}` | 🔓 | Get category detail |

---

## 🛒 Orders (`/api/v1/orders`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/orders` | 🔓* | Create order (checkout) |
| GET | `/orders/{order_number}` | 🔓 | Get order detail |
| POST | `/orders/{order_number}/simulate-payment` | 🔓 | Simulate payment (dev only) |
| POST | `/orders/payment-callback` | 🔓** | Payment webhook (from gateway) |

*Auth optional - jika ada token, order linked ke user  
**Webhook endpoint, butuh signature verification

---

## 🎟️ Tickets (`/api/v1/tickets`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/tickets/my-tickets` | 🔐 | Get user's tickets (upcoming & past) |
| GET | `/tickets/{ticket_code}` | 🔓 | Get ticket detail (guest access) |
| GET | `/tickets/{ticket_code}/download` | 🔓 | Download ticket PDF |
| POST | `/tickets/check-in` | 👑 | Check-in ticket (scan QR) |
| GET | `/tickets/{ticket_code}/verify` | 🔓 | Verify ticket status |

---

## 👑 Admin - Dashboard (`/api/v1/admin`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/admin/stats` | 👑 | Get dashboard summary stats |
| GET | `/admin/dashboard/revenue` | 👑 | Revenue overview & charts |
| GET | `/admin/dashboard/transactions-overview` | 👑 | Transaction stats & recent |
| GET | `/admin/dashboard/events-overview` | 👑 | Event analytics overview |
| GET | `/admin/dashboard/users-overview` | 👑 | User analytics |

**Query Parameters (GET /admin/stats):**
- `event_id` - Filter by specific event
- `start_date`, `end_date` - Filter by date range (YYYY-MM-DD)

---

## 👑 Admin - Events (`/api/v1/admin/events`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/admin/events` | 👑 | Get all events (admin view) |
| POST | `/admin/events` | 👑 | Create new event |
| GET | `/admin/events/{id}` | 👑 | Get event detail (admin) |
| PUT | `/admin/events/{id}` | 👑 | Update event |
| DELETE | `/admin/events/{id}` | 👑 | Delete event |
| PATCH | `/admin/events/{id}/status` | 👑 | Change event status |
| GET | `/admin/events/{id}/analytics` | 👑 | Get event analytics |

---

## 👑 Admin - Ticket Types (`/api/v1/admin/ticket-types`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/admin/ticket-types?event_id={id}` | 👑 | Get ticket types for event |
| POST | `/admin/ticket-types` | 👑 | Create ticket type |
| GET | `/admin/ticket-types/{id}` | 👑 | Get ticket type detail |
| PUT | `/admin/ticket-types/{id}` | 👑 | Update ticket type |
| DELETE | `/admin/ticket-types/{id}` | 👑 | Delete ticket type |
| PATCH | `/admin/ticket-types/{id}/status` | 👑 | Change ticket type status |

---

## 👑 Admin - Categories (`/api/v1/admin/categories`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/admin/categories` | 👑 | Get all categories (admin) |
| POST | `/admin/categories` | 👑 | Create category |
| GET | `/admin/categories/{id}` | 👑 | Get category detail (admin) |
| PUT | `/admin/categories/{id}` | 👑 | Update category |
| DELETE | `/admin/categories/{id}` | 👑 | Delete category |
| PATCH | `/admin/categories/{id}/reorder` | 👑 | Reorder categories |

---

## 👑 Admin - Transactions (`/api/v1/admin/transactions`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/admin/transactions` | 👑 | Get all transactions dengan filter |
| GET | `/admin/transactions/{id}` | 👑 | Get transaction detail |
| POST | `/admin/transactions/{id}/resend-email` | 👑 | Resend ticket email |
| POST | `/admin/transactions/{id}/cancel` | 👑 | Cancel transaction |
| POST | `/admin/transactions/{id}/mark-paid` | 👑 | Mark as paid (manual) |
| GET | `/admin/transactions/export` | 👑 | Export transactions (CSV/Excel/PDF) |
| GET | `/admin/transactions/{id}/timeline` | 👑 | Get transaction timeline |
| GET | `/admin/transactions/revenue-summary` | 👑 | Get revenue summary |

**Query Parameters (GET /admin/transactions):**
- `page`, `limit` - Pagination
- `search` - Search by order_number/customer_email/customer_name
- `status` - Filter: all/pending/paid/cancelled/expired
- `event_id` - Filter by specific event
- `start_date`, `end_date` - Filter by date range (YYYY-MM-DD)
- `date_from`, `date_to` - Filter by date range (Alternate style)
- `sort` - newest/oldest/amount_asc/amount_desc

---

## 👑 Admin - Users (`/api/v1/admin/users`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/admin/users` | 👑 | Get all users dengan filter |
| GET | `/admin/users/{id}` | 👑 | Get user detail |
| POST | `/admin/users` | 👑 | Create user (admin/organizer) |
| PUT | `/admin/users/{id}` | 👑 | Update user |
| DELETE | `/admin/users/{id}` | 👑 | Delete user |
| PATCH | `/admin/users/{id}/role` | 👑 | Change user role |
| PATCH | `/admin/users/{id}/status` | 👑 | Activate/Deactivate user |
| GET | `/admin/users/{id}/activity` | 👑 | Get user activity log |
| GET | `/admin/users/{id}/transactions` | 👑 | Get user's transactions |

---

## 👑 Admin - Reports (`/api/v1/admin/reports`)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/admin/reports/sales` | 👑 | Sales report |
| GET | `/admin/reports/events` | 👑 | Events performance report |
| GET | `/admin/reports/users` | 👑 | User acquisition report |
| GET | `/admin/reports/revenue` | 👑 | Revenue report by period |
| GET | `/admin/reports/top-events` | 👑 | Top performing events |
| GET | `/admin/reports/export` | 👑 | Export report (CSV/Excel/PDF) |

---

## 📊 Summary Statistics

### Total Endpoints: **70+**

**Public APIs:** 15 endpoints
- Auth: 10
- Events: 6
- Categories: 2
- Orders: 4
- Tickets: 5

**Admin APIs:** 55+ endpoints
- Dashboard: 5
- Event Management: 7
- Ticket Types: 6
- Categories: 6
- Transactions: 9
- Users: 9
- Reports: 6

---

## 🔄 Common Response Patterns

### List with Pagination
```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "current_page": 1,
      "total_pages": 5,
      "total_items": 58,
      "per_page": 12
    }
  }
}
```

### Single Resource
```json
{
  "success": true,
  "data": { /* resource object */ }
}
```

### Create/Update Success
```json
{
  "success": true,
  "message": "Resource berhasil dibuat/diupdate",
  "data": { /* created/updated resource */ }
}
```

### Delete Success
```json
{
  "success": true,
  "message": "Resource berhasil dihapus"
}
```

### Validation Error
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "field_name": ["Error message 1", "Error message 2"]
  }
}
```

---

## 🎯 Priority Implementation Order

### Phase 1: MVP (Must Have)
1. ✅ Auth (login, register)
2. ✅ Events (list, detail)
3. ✅ Categories (list)
4. ✅ Orders (create)
5. ✅ Tickets (my tickets, detail)

### Phase 2: Admin & Payment
6. ✅ Admin Dashboard (stats)
7. ✅ Admin Events (CRUD)
8. ✅ Admin Transactions (list, detail)
9. ✅ Payment Gateway Integration
10. ✅ Email Notifications

### Phase 3: Advanced Features
11. ✅ OAuth Google
12. ✅ PDF Generation
13. ✅ Reports & Analytics
14. ✅ Check-in System
15. ✅ Admin User Management

---

## 📝 Notes

- Semua tanggal gunakan format ISO 8601: `YYYY-MM-DDTHH:mm:ssZ`
- Semua price gunakan DECIMAL(12,2) - format: `350000.00`
- Pagination default: `page=1`, `limit=12` (public) atau `limit=20` (admin)
- Search case-insensitive
- Soft delete untuk critical data (users, orders, tickets)
- Hard delete untuk non-critical (categories bisa hard delete jika tidak ada event)

---

**Version:** 1.0  
**Last Updated:** January 21, 2026
