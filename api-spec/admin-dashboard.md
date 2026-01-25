# Admin Dashboard APIs

Base URL: `/api/v1/admin`

**Authentication:** All endpoints require admin role

**Headers:**
```
Authorization: Bearer {admin_token}
```

---

## Get Dashboard Stats

**Endpoint:** `GET /admin/stats`

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "total_events": 45,
    "published_events": 32,
    "draft_events": 8,
    "completed_events": 5,
    "total_transactions": 234,
    "pending_transactions": 12,
    "completed_transactions": 210,
    "expired_transactions": 8,
    "cancelled_transactions": 4,
    "total_revenue": 125000000.00,
    "today_revenue": 5000000.00,
    "this_week_revenue": 28000000.00,
    "this_month_revenue": 125000000.00,
    "total_users": 1234,
    "new_users_today": 12,
    "new_users_this_week": 56,
    "new_users_this_month": 234,
    "total_tickets_sold": 890,
    "tickets_sold_today": 23,
    "tickets_sold_this_week": 145,
    "tickets_sold_this_month": 890
  }
}
```

---

## Get Revenue Overview

**Endpoint:** `GET /admin/dashboard/revenue`

**Query Parameters:**
- `period` (string) - 'daily', 'weekly', 'monthly', 'yearly'
- `start_date` (date) - YYYY-MM-DD
- `end_date` (date) - YYYY-MM-DD

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "total_revenue": 125000000.00,
    "total_transactions": 234,
    "average_transaction": 534188.03,
    "chart_data": [
      {
        "date": "2026-01-15",
        "revenue": 5000000.00,
        "transactions": 12
      },
      {
        "date": "2026-01-16",
        "revenue": 7500000.00,
        "transactions": 18
      },
      {
        "date": "2026-01-17",
        "revenue": 3200000.00,
        "transactions": 8
      }
    ],
    "top_events": [
      {
        "event_id": 1,
        "event_title": "Jakarta Marathon 2026",
        "revenue": 35000000.00,
        "tickets_sold": 120
      },
      {
        "event_id": 2,
        "event_title": "Java Jazz Festival 2026",
        "revenue": 28000000.00,
        "tickets_sold": 95
      }
    ]
  }
}
```

---

## Get Transaction Overview

**Endpoint:** `GET /admin/dashboard/transactions-overview`

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "total": 234,
    "pending": 12,
    "completed": 210,
    "expired": 8,
    "cancelled": 4,
    "pending_percentage": 5.13,
    "completed_percentage": 89.74,
    "expired_percentage": 3.42,
    "cancelled_percentage": 1.71,
    "recent_transactions": [
      {
        "id": 1,
        "order_number": "ORD-1737456789123-ABCD",
        "customer_name": "Budi Santoso",
        "total_amount": 700000.00,
        "status": "completed",
        "created_at": "2026-01-21T10:30:00Z"
      }
    ]
  }
}
```

---

## Get Event Analytics Overview

**Endpoint:** `GET /admin/dashboard/events-overview`

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "total_events": 45,
    "published": 32,
    "draft": 8,
    "completed": 5,
    "most_viewed": [
      {
        "id": 1,
        "title": "Jakarta Marathon 2026",
        "view_count": 5432,
        "tickets_sold": 245
      },
      {
        "id": 2,
        "title": "Java Jazz Festival 2026",
        "view_count": 4321,
        "tickets_sold": 189
      }
    ],
    "best_selling": [
      {
        "id": 1,
        "title": "Jakarta Marathon 2026",
        "tickets_sold": 245,
        "revenue": 35000000.00
      },
      {
        "id": 2,
        "title": "Java Jazz Festival 2026",
        "tickets_sold": 189,
        "revenue": 28000000.00
      }
    ]
  }
}
```

---

## Get User Analytics

**Endpoint:** `GET /admin/dashboard/users-overview`

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "total_users": 1234,
    "active_users": 1180,
    "inactive_users": 54,
    "total_admins": 5,
    "total_organizers": 12,
    "new_users_this_month": 234,
    "growth_percentage": 23.45,
    "user_growth_chart": [
      {
        "date": "2026-01-15",
        "new_users": 12
      },
      {
        "date": "2026-01-16",
        "new_users": 18
      }
    ]
  }
}
```
