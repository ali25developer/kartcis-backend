# Category APIs (Public)

Base URL: `/api/v1/categories`

---

## Get All Categories

**Endpoint:** `GET /categories`

**Response Success (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Marathon & Lari",
      "slug": "marathon-lari",
      "description": "Event lari marathon, half marathon, fun run",
      "icon": "Footprints",
      "image": "https://storage.kartcis.id/categories/marathon.jpg",
      "is_active": true,
      "display_order": 1,
      "event_count": 23,
      "created_at": "2026-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "name": "Musik & Konser",
      "slug": "musik-konser",
      "description": "Konser musik, festival musik",
      "icon": "Music",
      "image": "https://storage.kartcis.id/categories/musik.jpg",
      "is_active": true,
      "display_order": 2,
      "event_count": 15,
      "created_at": "2026-01-01T00:00:00Z"
    },
    {
      "id": 3,
      "name": "Workshop & Seminar",
      "slug": "workshop-seminar",
      "description": "Workshop, seminar, pelatihan",
      "icon": "GraduationCap",
      "image": "https://storage.kartcis.id/categories/workshop.jpg",
      "is_active": true,
      "display_order": 3,
      "event_count": 12,
      "created_at": "2026-01-01T00:00:00Z"
    },
    {
      "id": 4,
      "name": "Olahraga",
      "slug": "olahraga",
      "description": "Event olahraga umum",
      "icon": "Dumbbell",
      "image": "https://storage.kartcis.id/categories/olahraga.jpg",
      "is_active": true,
      "display_order": 4,
      "event_count": 8,
      "created_at": "2026-01-01T00:00:00Z"
    },
    {
      "id": 5,
      "name": "Kuliner",
      "slug": "kuliner",
      "description": "Festival kuliner, food expo",
      "icon": "UtensilsCrossed",
      "image": "https://storage.kartcis.id/categories/kuliner.jpg",
      "is_active": true,
      "display_order": 5,
      "event_count": 6,
      "created_at": "2026-01-01T00:00:00Z"
    },
    {
      "id": 6,
      "name": "Charity & Sosial",
      "slug": "charity-sosial",
      "description": "Event charity, sosial",
      "icon": "Heart",
      "image": "https://storage.kartcis.id/categories/charity.jpg",
      "is_active": true,
      "display_order": 6,
      "event_count": 4,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

## Get Category Detail

**Endpoint:** `GET /categories/{slug}`

**Path Parameters:**
- `slug` - Category slug

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Marathon & Lari",
    "slug": "marathon-lari",
    "description": "Event lari marathon, half marathon, fun run",
    "icon": "Footprints",
    "image": "https://storage.kartcis.id/categories/marathon.jpg",
    "is_active": true,
    "display_order": 1,
    "event_count": 23,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  }
}
```

**Response Error (404):**
```json
{
  "success": false,
  "message": "Category tidak ditemukan"
}
```
