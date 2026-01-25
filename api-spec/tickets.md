# Ticket APIs

Base URL: `/api/v1/tickets`

---

## Get My Tickets (Logged In User)

**Endpoint:** `GET /tickets/my-tickets`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
- `status` (optional) - 'upcoming' or 'past'

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "upcoming": [
      {
        "id": 1,
        "ticket_code": "TIX-1-1737456789-X7Y9",
        "qr_code": "https://api.kartcis.id/qr/TIX-1-1737456789-X7Y9",
        "attendee_name": "Budi Santoso",
        "attendee_email": "budi@gmail.com",
        "attendee_phone": "08123456789",
        "status": "active",
        "check_in_at": null,
        "created_at": "2026-01-21T14:25:30Z",
        "event": {
          "id": 1,
          "title": "Jakarta Marathon 2026",
          "slug": "jakarta-marathon-2026",
          "event_date": "2026-03-15",
          "event_time": "06:00:00",
          "venue": "Bundaran HI - Monas",
          "city": "Jakarta",
          "image": "https://storage.kartcis.id/events/jakarta-marathon.jpg"
        },
        "ticket_type": {
          "id": 1,
          "name": "Early Bird - Full Marathon",
          "price": 350000.00
        },
        "order": {
          "order_number": "ORD-1737456789123-ABCD",
          "total_amount": 700000.00
        }
      }
    ],
    "past": [
      {
        "id": 5,
        "ticket_code": "TIX-5-1735456789-A1B2",
        "qr_code": "https://api.kartcis.id/qr/TIX-5-1735456789-A1B2",
        "attendee_name": "Budi Santoso",
        "attendee_email": "budi@gmail.com",
        "attendee_phone": "08123456789",
        "status": "used",
        "check_in_at": "2025-12-15T05:30:00Z",
        "created_at": "2025-12-01T10:00:00Z",
        "event": {
          "id": 5,
          "title": "Bali Marathon 2025",
          "slug": "bali-marathon-2025",
          "event_date": "2025-12-15",
          "event_time": "06:00:00",
          "venue": "Pantai Sanur",
          "city": "Denpasar",
          "image": "https://storage.kartcis.id/events/bali-marathon.jpg"
        },
        "ticket_type": {
          "id": 10,
          "name": "Regular - Full Marathon",
          "price": 400000.00
        },
        "order": {
          "order_number": "ORD-1735456789111-XYZA",
          "total_amount": 400000.00
        }
      }
    ]
  }
}
```

---

## Get Ticket by Code (Guest Access)

**Endpoint:** `GET /tickets/{ticket_code}`

**Path Parameters:**
- `ticket_code` - Ticket code (e.g., TIX-1-1737456789-X7Y9)

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "ticket_code": "TIX-1-1737456789-X7Y9",
    "qr_code": "https://api.kartcis.id/qr/TIX-1-1737456789-X7Y9",
    "attendee_name": "Budi Santoso",
    "attendee_email": "budi@gmail.com",
    "attendee_phone": "08123456789",
    "status": "active",
    "check_in_at": null,
    "created_at": "2026-01-21T14:25:30Z",
    "event": {
      "id": 1,
      "title": "Jakarta Marathon 2026",
      "slug": "jakarta-marathon-2026",
      "event_date": "2026-03-15",
      "event_time": "06:00:00",
      "venue": "Bundaran HI - Monas",
      "address": "Jl. MH Thamrin, Jakarta Pusat",
      "city": "Jakarta",
      "image": "https://storage.kartcis.id/events/jakarta-marathon.jpg",
      "organizer_info": {
        "name": "Jakarta Sports Association",
        "phone": "021-12345678",
        "email": "info@jakartamarathon.com"
      }
    },
    "ticket_type": {
      "id": 1,
      "name": "Early Bird - Full Marathon",
      "price": 350000.00
    }
  }
}
```

**Response Error (404):**
```json
{
  "success": false,
  "message": "Tiket tidak ditemukan"
}
```

---

## Download Ticket PDF

**Endpoint:** `GET /tickets/{ticket_code}/download`

**Path Parameters:**
- `ticket_code` - Ticket code

**Response:**
- Content-Type: `application/pdf`
- Content-Disposition: `attachment; filename="ticket-TIX-1-1737456789-X7Y9.pdf"`

**Binary PDF file containing:**
- Event details
- Ticket type & attendee info
- QR code (large, scannable)
- Order number
- Terms & conditions

---

## Check-in Ticket (Scan QR)

**Endpoint:** `POST /tickets/check-in`

**Headers:**
```
Authorization: Bearer {admin_token}
```
*Requires admin or organizer role*

**Request Body:**
```json
{
  "ticket_code": "TIX-1-1737456789-X7Y9"
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "Check-in berhasil",
  "data": {
    "ticket_code": "TIX-1-1737456789-X7Y9",
    "attendee_name": "Budi Santoso",
    "event_title": "Jakarta Marathon 2026",
    "ticket_type": "Early Bird - Full Marathon",
    "check_in_at": "2026-03-15T05:30:00Z"
  }
}
```

**Response Error (400):**
```json
{
  "success": false,
  "message": "Tiket sudah pernah digunakan",
  "data": {
    "check_in_at": "2026-03-15T05:30:00Z"
  }
}
```

**Response Error (404):**
```json
{
  "success": false,
  "message": "Tiket tidak ditemukan"
}
```

---

## Verify Ticket (Check Status)

**Endpoint:** `GET /tickets/{ticket_code}/verify`

**Path Parameters:**
- `ticket_code` - Ticket code

**Response Success (200):**
```json
{
  "success": true,
  "data": {
    "ticket_code": "TIX-1-1737456789-X7Y9",
    "status": "active",
    "is_valid": true,
    "attendee_name": "Budi Santoso",
    "event_title": "Jakarta Marathon 2026",
    "event_date": "2026-03-15",
    "ticket_type": "Early Bird - Full Marathon",
    "check_in_at": null
  }
}
```

**For used ticket:**
```json
{
  "success": true,
  "data": {
    "ticket_code": "TIX-1-1737456789-X7Y9",
    "status": "used",
    "is_valid": false,
    "attendee_name": "Budi Santoso",
    "event_title": "Jakarta Marathon 2026",
    "event_date": "2026-03-15",
    "ticket_type": "Early Bird - Full Marathon",
    "check_in_at": "2026-03-15T05:30:00Z"
  }
}
```
