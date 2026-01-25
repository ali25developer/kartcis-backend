# Authentication APIs

Base URL: `/api/v1/auth`

---

## Register

**Endpoint:** `POST /auth/register`

**Request Body:**
```json
{
  "name": "Budi Santoso",
  "email": "budi@gmail.com",
  "phone": "08123456789",
  "password": "password123",
  "password_confirmation": "password123"
}
```

**Response Success (201):**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "name": "Budi Santoso",
      "email": "budi@gmail.com",
      "phone": "08123456789",
      "role": "user",
      "created_at": "2026-01-21T10:30:00Z",
      "updated_at": "2026-01-21T10:30:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 7200
  }
}
```

---

## Login

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "email": "budi@gmail.com",
  "password": "password123"
}
```

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "name": "Budi Santoso",
      "email": "budi@gmail.com",
      "phone": "08123456789",
      "role": "user",
      "created_at": "2026-01-21T10:30:00Z",
      "updated_at": "2026-01-21T10:30:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 7200
  }
}
```

---

## Get Current User

**Endpoint:** `GET /auth/me`

**Headers:**
```
Authorization: Bearer {token}
```

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
    "created_at": "2026-01-21T10:30:00Z",
    "updated_at": "2026-01-21T10:30:00Z"
  }
}
```

---

## Logout

**Endpoint:** `POST /auth/logout`

**Headers:**
```
Authorization: Bearer {token}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "Logout successful"
}
```
