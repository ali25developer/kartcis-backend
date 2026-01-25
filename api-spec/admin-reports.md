# Admin Reports & Analytics APIs

Base URL: `/api/v1/admin/reports`

**Authentication:** All endpoints require admin role

**Headers:**
```
Authorization: Bearer {admin_token}
```

---

## Revenue Report

**Endpoint:** `GET /admin/reports/revenue`

**Query Parameters:**
- `period` (string) - daily/weekly/monthly/yearly
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD
- `group_by` (string) - event/category/payment_method

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_revenue": 125000000.00,
      "total_transactions": 234,
      "total_tickets_sold": 890,
      "average_transaction": 534188.03,
      "growth_percentage": 23.45
    },
    "period_data": [
      {
        "period": "2026-01-21",
        "revenue": 5000000.00,
        "transactions": 12,
        "tickets_sold": 45
      },
      {
        "period": "2026-01-20",
        "revenue": 7500000.00,
        "transactions": 18,
        "tickets_sold": 67
      }
    ],
    "by_event": [
      {
        "event_id": 1,
        "event_title": "Jakarta Marathon 2026",
        "category": "Marathon & Lari",
        "revenue": 35000000.00,
        "transactions": 120,
        "tickets_sold": 245,
        "percentage": 28.00
      }
    ],
    "by_category": [
      {
        "category_id": 1,
        "category_name": "Marathon & Lari",
        "revenue": 52000000.00,
        "transactions": 145,
        "percentage": 41.60
      }
    ],
    "by_payment_method": [
      {
        "method": "BCA Virtual Account",
        "revenue": 52000000.00,
        "transactions": 89,
        "percentage": 41.60
      }
    ]
  }
}
```

---

## Sales Report

**Endpoint:** `GET /admin/reports/sales`

**Query Parameters:**
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD
- `event_id` (int) - Filter by event
- `category_id` (int) - Filter by category

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_tickets_sold": 890,
      "total_revenue": 125000000.00,
      "total_orders": 234,
      "average_ticket_price": 140449.44,
      "conversion_rate": 12.34
    },
    "by_event": [
      {
        "event_id": 1,
        "event_title": "Jakarta Marathon 2026",
        "event_date": "2026-03-15",
        "tickets_sold": 245,
        "revenue": 35000000.00,
        "quota": 10000,
        "quota_used_percentage": 2.45
      }
    ],
    "by_ticket_type": [
      {
        "event_id": 1,
        "event_title": "Jakarta Marathon 2026",
        "ticket_type_id": 1,
        "ticket_type_name": "Early Bird - Full Marathon",
        "price": 350000.00,
        "sold": 150,
        "revenue": 52500000.00,
        "quota": 500,
        "quota_used_percentage": 30.00
      }
    ],
    "daily_sales": [
      {
        "date": "2026-01-21",
        "tickets_sold": 45,
        "revenue": 5000000.00,
        "orders": 12
      }
    ]
  }
}
```

---

## User Report

**Endpoint:** `GET /admin/reports/users`

**Query Parameters:**
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_users": 1234,
      "new_users_this_period": 234,
      "active_users": 1180,
      "inactive_users": 54,
      "users_with_orders": 456,
      "users_without_orders": 778,
      "conversion_rate": 36.95
    },
    "registration_trend": [
      {
        "date": "2026-01-21",
        "new_users": 12,
        "cumulative": 1234
      },
      {
        "date": "2026-01-20",
        "new_users": 18,
        "cumulative": 1222
      }
    ],
    "by_role": {
      "user": 1217,
      "admin": 5,
      "organizer": 12
    },
    "top_customers": [
      {
        "user_id": 15,
        "name": "Budi Santoso",
        "email": "budi@gmail.com",
        "total_orders": 12,
        "total_spent": 8500000.00,
        "total_tickets": 34
      }
    ],
    "user_retention": {
      "returning_customers": 123,
      "one_time_customers": 333,
      "retention_rate": 26.97
    }
  }
}
```

---

## Event Performance Report

**Endpoint:** `GET /admin/reports/event-performance`

**Query Parameters:**
- `event_id` (int) - Specific event ID
- `category_id` (int) - Filter by category
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "events": [
      {
        "event_id": 1,
        "event_title": "Jakarta Marathon 2026",
        "category": "Marathon & Lari",
        "event_date": "2026-03-15",
        "status": "published",
        "metrics": {
          "total_views": 5432,
          "unique_visitors": 3214,
          "tickets_sold": 245,
          "revenue": 35000000.00,
          "conversion_rate": 7.62,
          "average_ticket_price": 142857.14
        },
        "quota_info": {
          "total_quota": 10000,
          "sold": 245,
          "available": 9755,
          "percentage_sold": 2.45
        },
        "ticket_type_performance": [
          {
            "ticket_type": "Early Bird - Full Marathon",
            "price": 350000.00,
            "sold": 150,
            "revenue": 52500000.00,
            "sell_through_rate": 30.00
          }
        ],
        "traffic_sources": [
          {
            "source": "direct",
            "views": 2345,
            "percentage": 43.15
          },
          {
            "source": "social",
            "views": 1876,
            "percentage": 34.52
          }
        ]
      }
    ],
    "comparison": {
      "best_performing": {
        "event_id": 1,
        "event_title": "Jakarta Marathon 2026",
        "metric": "revenue",
        "value": 35000000.00
      },
      "worst_performing": {
        "event_id": 15,
        "event_title": "Small Workshop ABC",
        "metric": "revenue",
        "value": 500000.00
      }
    }
  }
}
```

---

## Payment Method Report

**Endpoint:** `GET /admin/reports/payment-methods`

**Query Parameters:**
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_transactions": 234,
      "total_revenue": 125000000.00
    },
    "by_method": [
      {
        "payment_method": "BCA Virtual Account",
        "transactions": 89,
        "revenue": 52000000.00,
        "percentage": 41.60,
        "average_transaction": 584269.66,
        "success_rate": 94.50
      },
      {
        "payment_method": "OVO",
        "transactions": 56,
        "revenue": 31000000.00,
        "percentage": 24.80,
        "average_transaction": 553571.43,
        "success_rate": 96.20
      },
      {
        "payment_method": "GoPay",
        "transactions": 45,
        "revenue": 24000000.00,
        "percentage": 19.20,
        "average_transaction": 533333.33,
        "success_rate": 95.80
      }
    ],
    "trend": [
      {
        "date": "2026-01-21",
        "methods": {
          "BCA Virtual Account": 5000000.00,
          "OVO": 2500000.00,
          "GoPay": 1800000.00
        }
      }
    ]
  }
}
```

---

## Category Performance Report

**Endpoint:** `GET /admin/reports/categories`

**Query Parameters:**
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_categories": 6,
      "total_events": 45,
      "total_revenue": 125000000.00
    },
    "by_category": [
      {
        "category_id": 1,
        "category_name": "Marathon & Lari",
        "total_events": 23,
        "published_events": 18,
        "total_views": 12345,
        "tickets_sold": 456,
        "revenue": 52000000.00,
        "percentage": 41.60,
        "average_event_revenue": 2260869.57,
        "growth_percentage": 15.30
      },
      {
        "category_id": 2,
        "category_name": "Musik & Konser",
        "total_events": 15,
        "published_events": 12,
        "total_views": 8765,
        "tickets_sold": 289,
        "revenue": 38000000.00,
        "percentage": 30.40,
        "average_event_revenue": 2533333.33,
        "growth_percentage": 22.50
      }
    ]
  }
}
```

---

## Export Report

**Endpoint:** `GET /admin/reports/export`

**Query Parameters:**
- `report_type` (string) - revenue/sales/users/event-performance
- `format` (string) - csv/excel/pdf
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD
- Additional filters based on report_type

**Response:**
- Content-Type: Based on format
- Content-Disposition: `attachment; filename="report-revenue-2026-01-21.xlsx"`

**Excel/CSV Structure:**
- Summary sheet
- Detailed data sheet
- Charts & visualizations (Excel only)
- Filters applied info

---

## Custom Report Builder

**Endpoint:** `POST /admin/reports/custom`

**Request Body:**
```json
{
  "name": "Custom Monthly Revenue by Category",
  "metrics": ["revenue", "tickets_sold", "transactions"],
  "dimensions": ["category", "event"],
  "filters": {
    "date_from": "2026-01-01",
    "date_to": "2026-01-31",
    "category_id": [1, 2, 3],
    "status": "paid"
  },
  "group_by": "category",
  "sort_by": "revenue",
  "sort_order": "desc"
}
```

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "report_name": "Custom Monthly Revenue by Category",
    "generated_at": "2026-01-21T16:00:00Z",
    "period": {
      "from": "2026-01-01",
      "to": "2026-01-31"
    },
    "results": [
      {
        "category": "Marathon & Lari",
        "revenue": 52000000.00,
        "tickets_sold": 456,
        "transactions": 145
      }
    ]
  }
}
```
