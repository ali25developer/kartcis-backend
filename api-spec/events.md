# Event APIs (Public)

Base URL: `/api/v1/events`

---

## Get All Events

**Endpoint:** `GET /events`

**Query Parameters:**
- `page` (int)
- `limit` (int)
- `search` (string)
- `category` (string)
- `featured` (boolean)

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
        "description": "Marathon internasional terbesar di Jakarta...",
        "detailed_description": "<p>Jakarta Marathon 2026...</p>",
        "event_date": "2026-03-15",
        "event_time": "06:00:00",
        "venue": "Bundaran HI - Monas",
        "city": "Jakarta",
        "organizer": "Jakarta Sports Association",
        "quota": 10000,
        "image": "https://storage.kartcis.id/events/jakarta-marathon.jpg",
        "is_featured": true,
        "status": "published",
        "category_id": 1,
        "min_price": 250000.00,
        "max_price": 500000.00,
        "created_at": "2026-01-01T10:00:00Z",
        "updated_at": "2026-01-01T10:00:00Z",
        "category": {
           "id": 1,
           "name": "Marathon",
           "slug": "marathon",
           "description": "Lari",
           "created_at": "...",
           "updated_at": "..."
        },
        "ticket_types": [
           {
             "id": 101,
             "event_id": 1,
             "name": "VIP",
             "description": "VIP Ticket",
             "price": 500000,
             "originalPrice": 600000,
             "quota": 100,
             "available": 50,
             "sold": 50,
             "status": "available",
             "created_at": "...",
             "updated_at": "..."
           }
        ]
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 5,
      "total_items": 58,
      "per_page": 12
    }
  }
}
```

---

## Get Event Detail

**Endpoint:** `GET /events/{id}` (or `{slug}`)

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "Jakarta Marathon 2026",
    "...": "Same fields as above"
  }
}
```
