# Admin User Management APIs

Base URL: `/api/v1/admin/users`

**Authentication:** All endpoints require admin role

**Headers:**
```
Authorization: Bearer {admin_token}
```

---

## Get All Users

**Endpoint:** `GET /admin/users`

**Query Parameters:**
- `page` (int) - Default: 1
- `limit` (int) - Default: 20
- `search` (string) - Search by name, email
- `role` (string) - Filter by role: user/admin/organizer
- `is_active` (boolean) - Filter by active status
- `sort` (string) - newest/oldest/name_asc/name_desc

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": 1,
        "name": "Budi Santoso",
        "email": "budi@gmail.com",
        "phone": "08123456789",
        "role": "user",
        "avatar": "https://lh3.googleusercontent.com/a/...",
        "is_active": true,
        "email_verified_at": "2026-01-15T10:30:00Z",
        "last_login_at": "2026-01-21T08:00:00Z",
        "created_at": "2026-01-10T08:00:00Z",
        "total_orders": 5,
        "total_spent": 2500000.00
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 62,
      "total_items": 1234,
      "per_page": 20
    },
    "summary": {
      "total_users": 1234,
      "total_admins": 5,
      "total_organizers": 12,
      "active_users": 1180,
      "inactive_users": 54
    }
  }
}
```

---

## Get User Detail

**Endpoint:** `GET /admin/users/{id}`

**Path Parameters:**
- `id` - User ID

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Budi Santoso",
    "email": "budi@gmail.com",
    "phone": "08123456789",
    "role": "user",
    "avatar": "https://lh3.googleusercontent.com/a/...",
    "is_active": true,
    "email_verified_at": "2026-01-15T10:30:00Z",
    "last_login_at": "2026-01-21T08:00:00Z",
    "created_at": "2026-01-10T08:00:00Z",
    "updated_at": "2026-01-21T08:00:00Z",
    "social_accounts": [
      {
        "provider": "google",
        "provider_email": "budi@gmail.com",
        "connected_at": "2026-01-10T08:00:00Z"
      }
    ],
    "statistics": {
      "total_orders": 5,
      "total_spent": 2500000.00,
      "total_tickets": 12,
      "upcoming_events": 3,
      "past_events": 2
    },
    "recent_orders": [
      {
        "order_number": "ORD-1737456789123-ABCD",
        "total_amount": 700000.00,
        "status": "paid",
        "created_at": "2026-01-21T10:30:00Z"
      }
    ]
  }
}
```

---

## Create User

**Endpoint:** `POST /admin/users`

**Request Body:**
```json
{
  "name": "New User",
  "email": "newuser@gmail.com",
  "phone": "08123456789",
  "password": "password123",
  "role": "user"
}
```

**Response Success (201):**
```json
{
  "success": true,
  "message": "User berhasil dibuat",
  "data": {
    "id": 1235,
    "name": "New User",
    "email": "newuser@gmail.com",
    "phone": "08123456789",
    "role": "user",
    "is_active": true,
    "created_at": "2026-01-21T10:30:00Z"
  }
}
```

---

## Update User

**Endpoint:** `PUT /admin/users/{id}`

**Path Parameters:**
- `id` - User ID

**Request Body:**
```json
{
  "name": "Budi Santoso (Updated)",
  "email": "budi@gmail.com",
  "phone": "08123456790"
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "User berhasil diupdate",
  "data": {
    "id": 1,
    "name": "Budi Santoso (Updated)",
    "email": "budi@gmail.com",
    "phone": "08123456790",
    "updated_at": "2026-01-21T15:30:00Z"
  }
}
```

---

## Change User Role

**Endpoint:** `PATCH /admin/users/{id}/role`

**Path Parameters:**
- `id` - User ID

**Request Body:**
```json
{
  "role": "admin"
}
```

**Roles:**
- `user` - Regular user
- `admin` - Full system access
- `organizer` - Can create/manage own events

**Response Success (200):**
```json
{
  "success": true,
  "message": "User role berhasil diupdate",
  "data": {
    "id": 1,
    "name": "Budi Santoso",
    "role": "admin",
    "updated_at": "2026-01-21T16:00:00Z"
  }
}
```

---

## Activate/Deactivate User

**Endpoint:** `PATCH /admin/users/{id}/status`

**Path Parameters:**
- `id` - User ID

**Request Body:**
```json
{
  "is_active": false,
  "reason": "Spam or fraudulent activity"
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "User status berhasil diupdate",
  "data": {
    "id": 1,
    "name": "Budi Santoso",
    "is_active": false,
    "updated_at": "2026-01-21T16:15:00Z"
  }
}
```

---

## Delete User

**Endpoint:** `DELETE /admin/users/{id}`

**Path Parameters:**
- `id` - User ID

**Response Success (200):**
```json
{
  "success": true,
  "message": "User berhasil dihapus"
}
```

**Response Error (400):**
```json
{
  "success": false,
  "message": "Tidak dapat menghapus user dengan pending/paid orders. Deactivate user sebagai gantinya."
}
```

---

## Get User Activity Log

**Endpoint:** `GET /admin/users/{id}/activity`

**Path Parameters:**
- `id` - User ID

**Query Parameters:**
- `page` (int) - Default: 1
- `limit` (int) - Default: 20

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "activities": [
      {
        "id": 123,
        "action": "login",
        "description": "User logged in",
        "ip_address": "192.168.1.1",
        "user_agent": "Mozilla/5.0...",
        "created_at": "2026-01-21T08:00:00Z"
      },
      {
        "id": 122,
        "action": "create_order",
        "description": "Created order ORD-1737456789123-ABCD",
        "ip_address": "192.168.1.1",
        "created_at": "2026-01-21T10:30:00Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 5,
      "total_items": 98,
      "per_page": 20
    }
  }
}
```

---

## Reset User Password

**Endpoint:** `POST /admin/users/{id}/reset-password`

**Path Parameters:**
- `id` - User ID

**Request Body:**
```json
{
  "new_password": "newpassword123"
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "Password berhasil direset"
}
```
