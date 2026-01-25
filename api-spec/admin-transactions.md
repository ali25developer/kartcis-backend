# Admin Transaction Management APIs

Base URL: `/api/v1/admin/transactions`

**Authentication:** All endpoints require admin role

**Headers:**
```
Authorization: Bearer {admin_token}
```

---

## Get All Transactions

**Endpoint:** `GET /admin/transactions`

**Query Parameters:**
- `page` (int) - Default: 1
- `limit` (int) - Default: 20
- `search` (string) - Search by order_number, customer_name, customer_email, event_title
- `status` (string) - Filter: all/pending/paid/cancelled/expired
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD
- `sort` (string) - newest/oldest/amount_asc/amount_desc

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "transactions": [
      {
        "id": 1,
        "order_number": "ORD-1737456789123-ABCD",
        "customer_name": "Budi Santoso",
        "customer_email": "budi@gmail.com",
        "customer_phone": "08123456789",
        "event_title": "Jakarta Marathon 2026",
        "event_date": "2026-03-15",
        "ticket_type": "Early Bird - Full Marathon",
        "quantity": 2,
        "total_amount": 700000.00,
        "status": "paid",
        "payment_method": "BCA Virtual Account",
        "created_at": "2026-01-21T10:30:00Z",
        "expires_at": null,
        "paid_at": "2026-01-21T14:25:30Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 12,
      "total_items": 234,
      "per_page": 20
    },
    "summary": {
      "total": 234,
      "pending": 12,
      "paid": 210,
      "expired": 8,
      "cancelled": 4,
      "total_revenue": 125000000.00,
      "pending_revenue": 8400000.00
    }
  }
}
```

---

## Get Transaction Detail

**Endpoint:** `GET /admin/transactions/{id}`

**Path Parameters:**
- `id` - Transaction ID or Order Number

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "order_number": "ORD-1737456789123-ABCD",
    "user_id": 1,
    "customer_name": "Budi Santoso",
    "customer_email": "budi@gmail.com",
    "customer_phone": "08123456789",
    "total_amount": 700000.00,
    "status": "paid",
    "payment_method": "BCA Virtual Account",
    "payment_details": {
      "bank": "BCA",
      "va_number": "1234567890123456",
      "account_name": "KARTCIS.ID"
    },
    "expires_at": "2026-01-22T10:30:00Z",
    "paid_at": "2026-01-21T14:25:30Z",
    "created_at": "2026-01-21T10:30:00Z",
    "updated_at": "2026-01-21T14:25:30Z",
    "order_items": [
      {
        "id": 1,
        "event_id": 1,
        "ticket_type_id": 1,
        "quantity": 2,
        "unit_price": 350000.00,
        "subtotal": 700000.00,
        "event": {
          "id": 1,
          "title": "Jakarta Marathon 2026",
          "slug": "jakarta-marathon-2026",
          "event_date": "2026-03-15",
          "venue": "Bundaran HI - Monas",
          "city": "Jakarta",
          "image": "https://storage.kartcis.id/events/jakarta-marathon.jpg"
        },
        "ticket_type": {
          "id": 1,
          "name": "Early Bird - Full Marathon",
          "price": 350000.00
        }
      }
    ],
    "tickets": [
      {
        "ticket_code": "TIX-1-1737456789-X7Y9",
        "attendee_name": "Budi Santoso",
        "status": "active",
        "check_in_at": null
      },
      {
        "ticket_code": "TIX-1-1737456789-Z8W7",
        "attendee_name": "Budi Santoso",
        "status": "active",
        "check_in_at": null
      }
    ],
    "user": {
      "id": 1,
      "name": "Budi Santoso",
      "email": "budi@gmail.com"
    }
  }
}
```

---

## Resend Ticket Email

**Endpoint:** `POST /admin/transactions/{id}/resend-email`

**Path Parameters:**
- `id` - Transaction ID or Order Number

**Response Success (200):**
```json
{
  "success": true,
  "message": "Email tiket berhasil dikirim ulang ke budi@gmail.com"
}
```

**Response Error (400):**
```json
{
  "success": false,
  "message": "Hanya transaksi dengan status paid yang bisa resend email"
}
```

---

## Cancel Transaction

**Endpoint:** `POST /admin/transactions/{id}/cancel`

**Path Parameters:**
- `id` - Transaction ID or Order Number

**Request Body:**
```json
{
  "reason": "Customer request / Duplicate order / Fraud suspicion"
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "Transaction berhasil dibatalkan",
  "data": {
    "order_number": "ORD-1737456789123-ABCD",
    "status": "cancelled",
    "cancelled_at": "2026-01-21T16:30:00Z"
  }
}
```

**Note:** 
- Hanya bisa cancel transaction dengan status `pending`
- Ticket quota akan dikembalikan

---

## Mark as Paid (Manual)

**Endpoint:** `POST /admin/transactions/{id}/mark-paid`

**Path Parameters:**
- `id` - Transaction ID or Order Number

**Request Body:**
```json
{
  "payment_proof": "https://storage.kartcis.id/proofs/payment-123.jpg",
  "notes": "Verified via bank transfer confirmation"
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "Transaction berhasil ditandai sebagai paid",
  "data": {
    "order_number": "ORD-1737456789123-ABCD",
    "status": "paid",
    "paid_at": "2026-01-21T16:45:00Z",
    "tickets_generated": 2
  }
}
```

**Note:**
- Hanya untuk payment method yang membutuhkan verifikasi manual
- Akan generate tickets otomatis
- Akan kirim email ke customer

---

## Export Transactions

**Endpoint:** `GET /admin/transactions/export`

**Query Parameters:**
- `format` (string) - csv/excel/pdf
- `status` (string) - Filter by status
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD

**Response:**
- Content-Type: `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` (Excel)
- Content-Type: `text/csv` (CSV)
- Content-Type: `application/pdf` (PDF)
- Content-Disposition: `attachment; filename="transactions-export-2026-01-21.xlsx"`

**Excel/CSV Columns:**
- Order Number
- Customer Name
- Customer Email
- Customer Phone
- Event Title
- Event Date
- Ticket Type
- Quantity
- Total Amount
- Status
- Payment Method
- Created At
- Paid At

---

## Get Transaction Timeline

**Endpoint:** `GET /admin/transactions/{id}/timeline`

**Path Parameters:**
- `id` - Transaction ID or Order Number

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "order_number": "ORD-1737456789123-ABCD",
    "timeline": [
      {
        "status": "created",
        "description": "Order dibuat",
        "timestamp": "2026-01-21T10:30:00Z",
        "metadata": {
          "payment_method": "BCA Virtual Account",
          "total_amount": 700000.00
        }
      },
      {
        "status": "payment_pending",
        "description": "Menunggu pembayaran",
        "timestamp": "2026-01-21T10:30:05Z",
        "metadata": {
          "expires_at": "2026-01-22T10:30:00Z"
        }
      },
      {
        "status": "paid",
        "description": "Pembayaran diterima",
        "timestamp": "2026-01-21T14:25:30Z",
        "metadata": {
          "transaction_id": "PAY-987654321"
        }
      },
      {
        "status": "tickets_generated",
        "description": "Tiket berhasil digenerate",
        "timestamp": "2026-01-21T14:25:35Z",
        "metadata": {
          "ticket_count": 2
        }
      },
      {
        "status": "email_sent",
        "description": "Email tiket dikirim ke budi@gmail.com",
        "timestamp": "2026-01-21T14:25:40Z"
      }
    ]
  }
}
```

---

## Get Revenue Summary

**Endpoint:** `GET /admin/transactions/revenue-summary`

**Query Parameters:**
- `period` (string) - daily/weekly/monthly/yearly
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "total_revenue": 125000000.00,
    "total_transactions": 234,
    "average_transaction_value": 534188.03,
    "by_status": {
      "paid": {
        "count": 210,
        "revenue": 125000000.00
      },
      "pending": {
        "count": 12,
        "revenue": 8400000.00
      }
    },
    "by_payment_method": [
      {
        "method": "BCA Virtual Account",
        "count": 89,
        "revenue": 52000000.00,
        "percentage": 41.60
      },
      {
        "method": "OVO",
        "count": 56,
        "revenue": 31000000.00,
        "percentage": 24.80
      }
    ],
    "by_event": [
      {
        "event_id": 1,
        "event_title": "Jakarta Marathon 2026",
        "count": 120,
        "revenue": 35000000.00,
        "percentage": 28.00
      }
    ]
  }
}
```
