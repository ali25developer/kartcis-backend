# Admin Ticket Type Management APIs

Base URL: `/api/v1/admin`

**Authentication:** All endpoints require admin role

**Headers:**
```
Authorization: Bearer {admin_token}
```

---

## Create Ticket Type

**Endpoint:** `POST /admin/events/{event_id}/ticket-types`

**Path Parameters:**
- `event_id` - Event ID

**Request Body:**
```json
{
  "name": "Early Bird - Full Marathon",
  "description": "Harga spesial untuk pendaftar awal",
  "price": 350000.00,
  "original_price": 500000.00,
  "quota": 500,
  "sale_start_date": "2026-01-01T00:00:00Z",
  "sale_end_date": "2026-02-01T23:59:59Z"
}
```

**Response Success (201):**
```json
{
  "success": true,
  "message": "Ticket type berhasil dibuat",
  "data": {
    "id": 1,
    "event_id": 1,
    "name": "Early Bird - Full Marathon",
    "description": "Harga spesial untuk pendaftar awal",
    "price": 350000.00,
    "original_price": 500000.00,
    "quota": 500,
    "available": 500,
    "status": "available",
    "sale_start_date": "2026-01-01T00:00:00Z",
    "sale_end_date": "2026-02-01T23:59:59Z",
    "created_at": "2026-01-21T10:30:00Z"
  }
}
```

**Response Error (422):**
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "price": ["Price harus lebih dari 0"],
    "quota": ["Quota minimal 1"]
  }
}
```

---

## Get Ticket Type Detail

**Endpoint:** `GET /admin/ticket-types/{id}`

**Path Parameters:**
- `id` - Ticket type ID

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "event_id": 1,
    "name": "Early Bird - Full Marathon",
    "description": "Harga spesial untuk pendaftar awal",
    "price": 350000.00,
    "original_price": 500000.00,
    "quota": 500,
    "available": 245,
    "sold": 255,
    "status": "available",
    "sale_start_date": "2026-01-01T00:00:00Z",
    "sale_end_date": "2026-02-01T23:59:59Z",
    "created_at": "2026-01-21T10:30:00Z",
    "updated_at": "2026-01-21T15:30:00Z",
    "event": {
      "id": 1,
      "title": "Jakarta Marathon 2026",
      "slug": "jakarta-marathon-2026"
    },
    "revenue": 89250000.00
  }
}
```

---

## Update Ticket Type

**Endpoint:** `PUT /admin/ticket-types/{id}`

**Path Parameters:**
- `id` - Ticket type ID

**Request Body:**
```json
{
  "name": "Early Bird - Full Marathon (Updated)",
  "description": "Harga spesial untuk pendaftar awal",
  "price": 375000.00,
  "original_price": 500000.00,
  "quota": 600,
  "sale_start_date": "2026-01-01T00:00:00Z",
  "sale_end_date": "2026-02-15T23:59:59Z"
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "Ticket type berhasil diupdate",
  "data": {
    "id": 1,
    "name": "Early Bird - Full Marathon (Updated)",
    "price": 375000.00,
    "quota": 600,
    "updated_at": "2026-01-21T16:00:00Z"
  }
}
```

**Note:** 
- `available` akan di-adjust otomatis jika quota diubah
- Tidak bisa menurunkan quota di bawah jumlah yang sudah terjual

---

## Delete Ticket Type

**Endpoint:** `DELETE /admin/ticket-types/{id}`

**Path Parameters:**
- `id` - Ticket type ID

**Response Success (200):**
```json
{
  "success": true,
  "message": "Ticket type berhasil dihapus"
}
```

**Response Error (400):**
```json
{
  "success": false,
  "message": "Tidak dapat menghapus ticket type yang sudah terjual"
}
```

---

## Change Ticket Type Status

**Endpoint:** `PATCH /admin/ticket-types/{id}/status`

**Path Parameters:**
- `id` - Ticket type ID

**Request Body:**
```json
{
  "status": "unavailable"
}
```

**Status Options:**
- `available` - Aktif dan bisa dibeli
- `sold_out` - Kuota habis (auto-set oleh sistem)
- `unavailable` - Dinonaktifkan manual

**Response Success (200):**
```json
{
  "success": true,
  "message": "Status ticket type berhasil diupdate",
  "data": {
    "id": 1,
    "name": "Early Bird - Full Marathon",
    "status": "unavailable",
    "updated_at": "2026-01-21T16:15:00Z"
  }
}
```

---

## Get Ticket Sales Report

**Endpoint:** `GET /admin/ticket-types/{id}/sales`

**Path Parameters:**
- `id` - Ticket type ID

**Query Parameters:**
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "ticket_type_id": 1,
    "name": "Early Bird - Full Marathon",
    "total_sold": 255,
    "total_revenue": 89250000.00,
    "average_daily_sales": 12.75,
    "daily_sales": [
      {
        "date": "2026-01-21",
        "quantity": 12,
        "revenue": 4200000.00
      },
      {
        "date": "2026-01-20",
        "quantity": 8,
        "revenue": 2800000.00
      }
    ],
    "quota_info": {
      "total_quota": 500,
      "sold": 255,
      "available": 245,
      "percentage_sold": 51.00
    }
  }
}
```
