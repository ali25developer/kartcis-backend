# Admin Event Management APIs

Base URL: `/api/v1/admin/events`

**Authentication:** All endpoints require admin role

**Headers:**
```
Authorization: Bearer {admin_token}
```

---

## Get All Events (Admin)

**Endpoint:** `GET /admin/events`

**Query Parameters:**
- `page` (int) - Default: 1
- `limit` (int) - Default: 20
- `search` (string) - Search by title, venue
- `status` (string) - draft/published/completed/cancelled/sold-out
- `category_id` (int) - Filter by category
- `sort` (string) - date_asc/date_desc/newest/oldest

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "events": [
      {
        "id": 1,
        "title": "Jakarta Marathon 2026",
        "slug": "jakarta-marathon-2026",
        "event_date": "2026-03-15",
        "city": "Jakarta",
        "status": "published",
        "is_featured": true,
        "view_count": 5432,
        "quota": 10000,
        "tickets_sold": 245,
        "revenue": 35000000.00,
        "created_at": "2026-01-01T10:00:00Z",
        "category": {
          "id": 1,
          "name": "Marathon & Lari"
        },
        "organizer": {
          "id": 1,
          "name": "Admin User"
        }
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 3,
      "total_items": 45,
      "per_page": 20
    }
  }
}
```

---

## Create Event

**Endpoint:** `POST /admin/events`

**Request Body:**
```json
{
  "category_id": 1,
  "title": "Jakarta Marathon 2026",
  "description": "Marathon internasional terbesar di Jakarta",
  "detailed_description": "<p>Full HTML description...</p>",
  "facilities": [
    "Medali finisher untuk semua kategori",
    "Sertifikat digital",
    "Running jersey official"
  ],
  "terms": [
    "Peserta harus berusia minimal 17 tahun",
    "Tiket tidak dapat dikembalikan"
  ],
  "agenda": [
    {"time": "04:00", "activity": "Registration"},
    {"time": "06:00", "activity": "Race Start"}
  ],
  "organizer_info": {
    "name": "Jakarta Sports Association",
    "description": "Penyelenggara event olahraga profesional",
    "phone": "021-12345678",
    "email": "info@jakartamarathon.com",
    "website": "https://jakartamarathon.com",
    "instagram": "@jakartamarathon"
  },
  "faqs": [
    {
      "question": "Kapan pengambilan race pack?",
      "answer": "2 hari sebelum event di Jakarta Convention Center"
    }
  ],
  "event_date": "2026-03-15",
  "event_time": "06:00",
  "venue": "Bundaran HI - Monas",
  "address": "Jl. MH Thamrin, Jakarta Pusat",
  "city": "Jakarta",
  "province": "DKI Jakarta",
  "organizer": "Jakarta Sports Association",
  "quota": 10000,
  "image": "https://storage.kartcis.id/events/jakarta-marathon.jpg",
  "is_featured": true,
  "status": "draft"
}
```

**Response Success (201):**
```json
{
  "success": true,
  "message": "Event berhasil dibuat",
  "data": {
    "id": 1,
    "title": "Jakarta Marathon 2026",
    "slug": "jakarta-marathon-2026",
    "category_id": 1,
    "description": "Marathon internasional terbesar di Jakarta",
    "event_date": "2026-03-15",
    "event_time": "06:00:00",
    "venue": "Bundaran HI - Monas",
    "city": "Jakarta",
    "status": "draft",
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
    "title": ["Title sudah digunakan"],
    "event_date": ["Event date harus tanggal di masa depan"]
  }
}
```

---

## Get Event Detail (Admin)

**Endpoint:** `GET /admin/events/{id}`

**Path Parameters:**
- `id` - Event ID

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "category_id": 1,
    "organizer_id": 1,
    "title": "Jakarta Marathon 2026",
    "slug": "jakarta-marathon-2026",
    "description": "Marathon internasional...",
    "detailed_description": "<p>Full HTML...</p>",
    "facilities": ["Medali", "Sertifikat"],
    "terms": ["Minimal 17 tahun"],
    "agenda": [{"time": "06:00", "activity": "Start"}],
    "organizer_info": {...},
    "faqs": [...],
    "event_date": "2026-03-15",
    "event_time": "06:00:00",
    "venue": "Bundaran HI - Monas",
    "address": "Jl. MH Thamrin, Jakarta Pusat",
    "city": "Jakarta",
    "province": "DKI Jakarta",
    "organizer": "Jakarta Sports Association",
    "quota": 10000,
    "image": "https://storage.kartcis.id/events/jakarta-marathon.jpg",
    "is_featured": true,
    "view_count": 5432,
    "status": "published",
    "created_at": "2026-01-01T10:00:00Z",
    "updated_at": "2026-01-21T14:30:00Z",
    "category": {
      "id": 1,
      "name": "Marathon & Lari",
      "slug": "marathon-lari"
    },
    "ticket_types": [
      {
        "id": 1,
        "name": "Early Bird - Full Marathon",
        "price": 350000.00,
        "quota": 500,
        "available": 245,
        "status": "available"
      }
    ],
    "statistics": {
      "total_views": 5432,
      "total_tickets_sold": 245,
      "total_revenue": 35000000.00,
      "tickets_available": 9755
    }
  }
}
```

---

## Update Event

**Endpoint:** `PUT /admin/events/{id}`

**Path Parameters:**
- `id` - Event ID

**Request Body:** Same as Create Event

**Response Success (200):**
```json
{
  "success": true,
  "message": "Event berhasil diupdate",
  "data": {
    "id": 1,
    "title": "Jakarta Marathon 2026",
    "updated_at": "2026-01-21T15:30:00Z"
  }
}
```

---

## Delete Event

**Endpoint:** `DELETE /admin/events/{id}`

**Path Parameters:**
- `id` - Event ID

**Response Success (200):**
```json
{
  "success": true,
  "message": "Event berhasil dihapus"
}
```

**Response Error (400):**
```json
{
  "success": false,
  "message": "Tidak dapat menghapus event yang sudah memiliki transaksi. Cancel event sebagai gantinya."
}
```

---

## Change Event Status

**Endpoint:** `PATCH /admin/events/{id}/status`

**Path Parameters:**
- `id` - Event ID

**Request Body:**
```json
{
  "status": "published",
  "cancel_reason": null
}
```

**Status Options:**
- `draft` - Belum dipublish
- `published` - Aktif dan bisa dibeli
- `completed` - Event sudah selesai
- `cancelled` - Event dibatalkan
- `sold-out` - Tiket habis

**For cancelled status:**
```json
{
  "status": "cancelled",
  "cancel_reason": "Cuaca buruk, event ditunda hingga pemberitahuan lebih lanjut"
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "Status event berhasil diupdate",
  "data": {
    "id": 1,
    "title": "Jakarta Marathon 2026",
    "status": "published",
    "updated_at": "2026-01-21T15:30:00Z"
  }
}
```

---

## Get Event Analytics

**Endpoint:** `GET /admin/events/{id}/analytics`

**Path Parameters:**
- `id` - Event ID

**Query Parameters:**
- `date_from` (date) - YYYY-MM-DD
- `date_to` (date) - YYYY-MM-DD

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "event_id": 1,
    "event_title": "Jakarta Marathon 2026",
    "total_views": 5432,
    "total_tickets_sold": 245,
    "total_revenue": 35000000.00,
    "average_ticket_price": 142857.14,
    "conversion_rate": 4.51,
    "daily_stats": [
      {
        "date": "2026-01-21",
        "views": 234,
        "tickets_sold": 12,
        "revenue": 4200000.00
      },
      {
        "date": "2026-01-20",
        "views": 189,
        "tickets_sold": 8,
        "revenue": 2800000.00
      }
    ],
    "ticket_type_breakdown": [
      {
        "ticket_type_id": 1,
        "name": "Early Bird - Full Marathon",
        "sold": 150,
        "revenue": 52500000.00,
        "percentage": 61.22
      },
      {
        "ticket_type_id": 2,
        "name": "Regular - Half Marathon",
        "sold": 95,
        "revenue": 23750000.00,
        "percentage": 38.78
      }
    ],
    "traffic_sources": [
      {
        "source": "direct",
        "views": 2345,
        "percentage": 43.15
      },
      {
        "source": "search",
        "views": 1876,
        "percentage": 34.52
      },
      {
        "source": "social",
        "views": 1211,
        "percentage": 22.33
      }
    ]
  }
}
```
