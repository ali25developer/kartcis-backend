# Order & Checkout APIs

Base URL: `/api/v1/orders`

---

## Create Order (Checkout)

**Endpoint:** `POST /orders`

**Headers (Optional):**
```
Authorization: Bearer {token}
```
*Token optional - untuk logged in user. Jika tanpa token = guest checkout*

**Request Body:**
```json
{
  "items": [
    {
      "event_id": 1,
      "ticket_type_id": 1,
      "quantity": 2
    },
    {
      "event_id": 2,
      "ticket_type_id": 3,
      "quantity": 1
    }
  ],
  "payment_method": "BCA Virtual Account",
  "customer_info": {
    "name": "Budi Santoso",
    "email": "budi@gmail.com",
    "phone": "08123456789"
  }
}
```

**Payment Methods:**
- BCA Virtual Account
- Mandiri Virtual Account
- BNI Virtual Account
- BRI Virtual Account
- OVO
- GoPay
- ShopeePay
- DANA
- QRIS

**Response Success (201):**
```json
{
  "success": true,
  "message": "Order berhasil dibuat",
  "data": {
    "id": 1,
    "order_number": "ORD-1737456789123-ABCD",
    "user_id": 1,
    "customer_name": "Budi Santoso",
    "customer_email": "budi@gmail.com",
    "customer_phone": "08123456789",
    "total_amount": 700000.00,
    "status": "pending",
    "payment_method": "BCA Virtual Account",
    "payment_details": {
      "bank": "BCA",
      "va_number": "1234567890123456",
      "account_name": "KARTCIS.ID",
      "amount": 700000.00,
      "instructions": [
        "Login ke BCA Mobile/KlikBCA",
        "Pilih menu Transfer",
        "Input nomor Virtual Account: 1234567890123456",
        "Masukkan nominal: Rp 700.000",
        "Konfirmasi pembayaran"
      ]
    },
    "expires_at": "2026-01-22T10:30:00Z",
    "created_at": "2026-01-21T10:30:00Z",
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
    ]
  }
}
```

**Response Error (422):**
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "items.0.quantity": ["Tiket tidak mencukupi. Tersedia: 1, diminta: 2"]
  }
}
```

---

## Get Order Detail

**Endpoint:** `GET /orders/{order_number}`

**Path Parameters:**
- `order_number` - Order number (e.g., ORD-1737456789123-ABCD)

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
    "status": "pending",
    "payment_method": "BCA Virtual Account",
    "payment_details": {
      "bank": "BCA",
      "va_number": "1234567890123456",
      "account_name": "KARTCIS.ID",
      "amount": 700000.00,
      "instructions": [
        "Login ke BCA Mobile/KlikBCA",
        "Pilih menu Transfer",
        "Input nomor Virtual Account: 1234567890123456",
        "Masukkan nominal: Rp 700.000",
        "Konfirmasi pembayaran"
      ]
    },
    "expires_at": "2026-01-22T10:30:00Z",
    "paid_at": null,
    "created_at": "2026-01-21T10:30:00Z",
    "updated_at": "2026-01-21T10:30:00Z",
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
          "event_time": "06:00:00",
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
    ]
  }
}
```

**Response Error (404):**
```json
{
  "success": false,
  "message": "Order tidak ditemukan"
}
```

---

## Simulate Payment (Development Only)

**Endpoint:** `POST /orders/{order_number}/simulate-payment`

**Path Parameters:**
- `order_number` - Order number

*Note: Endpoint ini hanya untuk development/testing*

**Response Success (200):**
```json
{
  "success": true,
  "message": "Pembayaran berhasil disimulasikan",
  "data": {
    "order_number": "ORD-1737456789123-ABCD",
    "status": "paid",
    "paid_at": "2026-01-21T14:25:30Z",
    "tickets_generated": 2
  }
}
```

---

## Payment Status Webhook

**Endpoint:** `POST /orders/payment-callback`

**Headers:**
```
X-Callback-Token: {secret_token}
```

**Request Body (dari payment gateway):**
```json
{
  "order_number": "ORD-1737456789123-ABCD",
  "transaction_id": "PAY-987654321",
  "status": "paid",
  "paid_at": "2026-01-21T14:25:30Z",
  "payment_method": "BCA Virtual Account",
  "amount": 700000.00
}
```

**Response Success (200):**
```json
{
  "success": true,
  "message": "Payment confirmed"
}
```

*Note: Endpoint ini dipanggil oleh payment gateway, bukan frontend*
