# Admin Category Management APIs

Base URL: `/api/v1/admin/categories`

**Authentication:** All endpoints require admin role

**Headers:**
```
Authorization: Bearer {admin_token}
```

---

## Get All Categories (Admin)

**Endpoint:** `GET /admin/categories`

**Query Parameters:**
- `include_inactive` (boolean) - Include inactive categories, default: false

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
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

## Create Category

**Endpoint:** `POST /admin/categories`

**Request Body:**
```json
{
  "name": "Workshop & Seminar",
  "slug": "workshop-seminar",
  "description": "Event workshop dan seminar profesional",
  "icon": "GraduationCap",
  "image": "https://storage.kartcis.id/categories/workshop.jpg",
  "display_order": 3
}
```

**Response Success (201):**
```json
{
  "success": true,
  "message": "Category berhasil dibuat",
  "data": {
    "id": 7,
    "name": "Workshop & Seminar",
    "slug": "workshop-seminar",
    "description": "Event workshop dan seminar profesional",
    "icon": "GraduationCap",
    "image": "https://storage.kartcis.id/categories/workshop.jpg",
    "is_active": true,
    "display_order": 3,
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
    "slug": ["Slug sudah digunakan"],
    "name": ["Name wajib diisi"]
  }
}
```

---

## Get Category Detail (Admin)

**Endpoint:** `GET /admin/categories/{id}`

**Path Parameters:**
- `id` - Category ID

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
    "total_revenue": 125000000.00,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "events": [
      {
        "id": 1,
        "title": "Jakarta Marathon 2026",
        "status": "published",
        "tickets_sold": 245
      }
    ]
  }
}
```

---

## Update Category

**Endpoint:** `PUT /admin/categories/{id}`

**Path Parameters:**
- `id` - Category ID

**Request Body:**
```json
{
  "name": "Marathon & Lari (Updated)",
  "slug": "marathon-lari",
  "description": "Updated description",
  "icon": "Footprints",
  "image": "https://storage.kartcis.id/categories/marathon-new.jpg",
  "display_order": 1
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "Category berhasil diupdate",
  "data": {
    "id": 1,
    "name": "Marathon & Lari (Updated)",
    "slug": "marathon-lari",
    "updated_at": "2026-01-21T15:30:00Z"
  }
}
```

---

## Delete Category

**Endpoint:** `DELETE /admin/categories/{id}`

**Path Parameters:**
- `id` - Category ID

**Response Success (200):**
```json
{
  "success": true,
  "message": "Category berhasil dihapus"
}
```

**Response Error (400):**
```json
{
  "success": false,
  "message": "Tidak dapat menghapus category yang memiliki event. Deactivate category sebagai gantinya."
}
```

---

## Activate/Deactivate Category

**Endpoint:** `PATCH /admin/categories/{id}/status`

**Path Parameters:**
- `id` - Category ID

**Request Body:**
```json
{
  "is_active": false
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "Category status berhasil diupdate",
  "data": {
    "id": 1,
    "name": "Marathon & Lari",
    "is_active": false,
    "updated_at": "2026-01-21T16:00:00Z"
  }
}
```

---

## Reorder Categories

**Endpoint:** `PUT /admin/categories/reorder`

**Request Body:**
```json
{
  "categories": [
    {"id": 1, "display_order": 1},
    {"id": 2, "display_order": 2},
    {"id": 3, "display_order": 3}
  ]
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "Category order berhasil diupdate"
}
```
