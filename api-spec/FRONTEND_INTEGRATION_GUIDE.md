# Frontend Integration Guide

Guide lengkap untuk menghubungkan backend dengan frontend React yang sudah ada.

## 📍 Frontend Code Structure

```
src/
├── app/
│   ├── components/         # UI Components
│   ├── pages/             # Page components
│   ├── contexts/          # React Context (Auth, Events)
│   ├── services/          # API service layer
│   │   ├── api.ts        # Mock API (GANTI INI dengan real API)
│   │   └── adminApi.ts   # Mock Admin API (GANTI INI)
│   ├── types/            # TypeScript interfaces
│   │   └── index.ts      # Match dengan API response
│   └── utils/            # Helper functions
```

## 🔄 API Service Files yang Perlu Diganti

### 1. `/src/app/services/api.ts`

**Current:** Mock API dengan localStorage  
**Action:** Ganti dengan real API calls

```typescript
// BEFORE (Mock)
const api = {
  events: {
    getAll: async () => {
      await delay(300);
      return { success: true, data: mockEvents };
    }
  }
}

// AFTER (Real API)
const API_BASE_URL = process.env.VITE_API_URL || 'http://localhost:8080/api/v1';

const api = {
  events: {
    getAll: async () => {
      const response = await fetch(`${API_BASE_URL}/events`);
      return await response.json();
    }
  }
}
```

### 2. `/src/app/services/adminApi.ts`

**Current:** Mock Admin API dengan localStorage  
**Action:** Ganti dengan real API calls + auth token

```typescript
// BEFORE (Mock)
export const adminApi = {
  getTransactions: async (params) => {
    const transactions = JSON.parse(localStorage.getItem('admin_transactions'));
    return { success: true, data: { transactions, stats, pagination } };
  }
}

// AFTER (Real API)
const API_BASE_URL = process.env.VITE_API_URL || 'http://localhost:8080/api/v1';

const getAuthToken = () => localStorage.getItem('auth_token');

export const adminApi = {
  getTransactions: async (params) => {
    const queryString = new URLSearchParams(params).toString();
    const response = await fetch(`${API_BASE_URL}/admin/transactions?${queryString}`, {
      headers: {
        'Authorization': `Bearer ${getAuthToken()}`,
        'Content-Type': 'application/json',
      }
    });
    return await response.json();
  }
}
```

## 📋 TypeScript Interfaces (/src/app/types/index.ts)

Frontend sudah punya complete TypeScript interfaces yang **MATCH 100%** dengan API response. Backend harus return exact format ini:

### Core Interfaces

```typescript
export interface Event {
  id: number;
  title: string;
  slug: string;
  description: string;
  detailed_description?: string;
  facilities?: string[];
  terms?: string[];
  agenda?: AgendaItem[];
  organizer_info?: OrganizerInfo;
  faqs?: FAQ[];
  event_date: string;        // YYYY-MM-DD
  event_time: string | null; // HH:mm:ss
  venue: string;
  city: string;
  organizer: string;
  quota: number;
  image: string | null;
  is_featured: boolean;
  status: 'draft' | 'published' | 'completed' | 'cancelled' | 'sold-out';
  created_at: string;
  updated_at: string;
  category_id: number;
  category?: Category;
  ticket_types?: TicketType[];
}

export interface Category {
  id: number;
  name: string;
  slug: string;
  description: string | null;
  created_at: string;
  updated_at: string;
}

export interface TicketType {
  id: number;
  event_id: number;
  name: string;
  description: string | null;
  price: number;
  originalPrice?: number;
  quota: number;
  available: number;
  status: 'available' | 'sold_out' | 'unavailable';
  created_at: string;
  updated_at: string;
}

export interface Order {
  id: number;
  user_id: number | null;
  order_number: string;
  total_amount: number;
  status: 'pending' | 'paid' | 'cancelled' | 'expired';
  payment_method: string;
  payment_details: any;
  expires_at: string;
  paid_at: string | null;
  created_at: string;
  updated_at: string;
  order_items?: OrderItem[];
}

export interface User {
  id: number;
  name: string;
  email: string;
  phone: string | null;
  role?: 'user' | 'admin';
  created_at: string;
  updated_at: string;
}
```

**👉 PENTING:** Semua field name harus **exact match**. Frontend tidak akan work jika field name berbeda.

## 🔐 Authentication Context (/src/app/contexts/AuthContext.tsx)

Frontend menggunakan React Context untuk manage auth state.

### Expected Auth Flow

```typescript
// 1. LOGIN
const login = async (email: string, password: string, rememberMe: boolean) => {
  const response = await fetch(`${API_BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, remember_me: rememberMe })
  });
  
  const data = await response.json();
  
  if (data.success) {
    // Save to localStorage
    localStorage.setItem('auth_token', data.data.token);
    localStorage.setItem('user_data', JSON.stringify(data.data.user));
    localStorage.setItem('token_expiry', data.data.expires_in);
  }
}

// 2. GET CURRENT USER
const checkAuth = async () => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch(`${API_BASE_URL}/auth/me`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  const data = await response.json();
  return data.data; // User object
}

// 3. LOGOUT
const logout = async () => {
  const token = localStorage.getItem('auth_token');
  
  await fetch(`${API_BASE_URL}/auth/logout`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  localStorage.removeItem('auth_token');
  localStorage.removeItem('user_data');
  localStorage.removeItem('token_expiry');
}
```

### Token Storage

Frontend menyimpan di **localStorage**:
- `auth_token` - JWT token
- `user_data` - User object (JSON string)
- `token_expiry` - Timestamp expiry

## 📱 Frontend Page → API Endpoint Mapping

### HomePage (`/src/app/pages/HomePage.tsx`)

**Uses:**
- `GET /events` - List semua events
- `GET /events/featured` - Featured events carousel
- `GET /categories` - Category filter

**Implementation:**
```typescript
useEffect(() => {
  const fetchData = async () => {
    // Get all events
    const eventsRes = await fetch(`${API_BASE_URL}/events?page=1&limit=12`);
    const eventsData = await eventsRes.json();
    setEvents(eventsData.data.events);
    
    // Get featured
    const featuredRes = await fetch(`${API_BASE_URL}/events/featured?limit=5`);
    const featuredData = await featuredRes.json();
    setFeatured(featuredData.data);
    
    // Get categories
    const categoriesRes = await fetch(`${API_BASE_URL}/categories`);
    const categoriesData = await categoriesRes.json();
    setCategories(categoriesData.data);
  };
  
  fetchData();
}, []);
```

---

### EventDetailPage (`/src/app/pages/EventDetailPage.tsx`)

**Uses:**
- `GET /events/{slug}` - Event detail lengkap

**Implementation:**
```typescript
useEffect(() => {
  const fetchEvent = async () => {
    const response = await fetch(`${API_BASE_URL}/events/${slug}`);
    const data = await response.json();
    
    if (data.success) {
      setEvent(data.data);
    }
  };
  
  fetchEvent();
}, [slug]);
```

**Expected Response Fields:**
- Basic info (title, description, date, venue)
- `detailed_description` - Rich HTML content
- `facilities` - Array of strings
- `terms` - Array of strings
- `agenda` - Array of {time, activity}
- `organizer_info` - Object with contact info
- `faqs` - Array of {question, answer}
- `ticket_types` - Array of ticket options

---

### CheckoutPage (`/src/app/pages/CheckoutPage.tsx`)

**Uses:**
- `POST /orders` - Create order / checkout

**Implementation:**
```typescript
const handleCheckout = async () => {
  const token = localStorage.getItem('auth_token');
  
  const orderData = {
    items: cartItems.map(item => ({
      event_id: item.event_id,
      ticket_type_id: item.ticket_type_id,
      quantity: item.quantity
    })),
    payment_method: selectedPaymentMethod,
    customer_info: {
      name: customerName,
      email: customerEmail,
      phone: customerPhone
    }
  };
  
  const response = await fetch(`${API_BASE_URL}/orders`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token && { 'Authorization': `Bearer ${token}` })
    },
    body: JSON.stringify(orderData)
  });
  
  const data = await response.json();
  
  if (data.success) {
    // Redirect ke payment page dengan order number
    navigate(`/payment/${data.data.order_number}`);
  }
}
```

---

### PaymentPage (`/src/app/pages/PaymentPage.tsx`)

**Uses:**
- `GET /orders/{order_number}` - Get order detail & payment info

**Implementation:**
```typescript
useEffect(() => {
  const fetchOrder = async () => {
    const response = await fetch(`${API_BASE_URL}/orders/${orderNumber}`);
    const data = await response.json();
    
    if (data.success) {
      setOrder(data.data);
      setPaymentDetails(data.data.payment_details);
      setExpiresAt(data.data.expires_at);
    }
  };
  
  fetchOrder();
  
  // Poll every 10 seconds untuk check payment status
  const interval = setInterval(fetchOrder, 10000);
  return () => clearInterval(interval);
}, [orderNumber]);
```

**Expected `payment_details` format:**
```json
{
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
}
```

---

### MyTicketsPage (`/src/app/pages/MyTicketsPage.tsx`)

**Uses:**
- `GET /tickets/my-tickets` - Get user's tickets

**Implementation:**
```typescript
useEffect(() => {
  const fetchTickets = async () => {
    const token = localStorage.getItem('auth_token');
    
    const response = await fetch(`${API_BASE_URL}/tickets/my-tickets`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    const data = await response.json();
    
    if (data.success) {
      setUpcomingTickets(data.data.upcoming);
      setPastTickets(data.data.past);
    }
  };
  
  fetchTickets();
}, []);
```

---

### AdminDashboard (`/src/app/pages/AdminDashboard.tsx`)

**Uses:**
- `GET /admin/stats` - Dashboard statistics
- `GET /admin/transactions` - Transaction list dengan filter

**Implementation:**
```typescript
useEffect(() => {
  const fetchDashboard = async () => {
    const token = localStorage.getItem('auth_token');
    
    // Get stats
    const statsRes = await fetch(`${API_BASE_URL}/admin/stats`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    const statsData = await statsRes.json();
    setStats(statsData.data);
    
    // Get transactions
    const params = new URLSearchParams({
      page: currentPage.toString(),
      limit: '20',
      status: filterStatus,
      search: searchQuery
    });
    
    const transRes = await fetch(`${API_BASE_URL}/admin/transactions?${params}`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    const transData = await transRes.json();
    setTransactions(transData.data.transactions);
    setPagination(transData.data.pagination);
  };
  
  fetchDashboard();
}, [currentPage, filterStatus, searchQuery]);
```

---

## 🔄 API Request/Response Examples

### Example 1: Browse Events with Filters

**Frontend Code:**
```typescript
const filters = {
  category: 'marathon-lari',
  city: 'Jakarta',
  search: 'marathon',
  page: 1,
  limit: 12
};

const params = new URLSearchParams(filters);
const response = await fetch(`${API_BASE_URL}/events?${params}`);
```

**Expected Backend Response:**
```json
{
  "success": true,
  "data": {
    "events": [
      {
        "id": 1,
        "title": "Jakarta Marathon 2026",
        "slug": "jakarta-marathon-2026",
        "description": "...",
        "event_date": "2026-03-15",
        "event_time": "06:00:00",
        "venue": "Bundaran HI - Monas",
        "city": "Jakarta",
        "image": "https://...",
        "is_featured": true,
        "status": "published",
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
            "available": 245,
            "quota": 500,
            "status": "available"
          }
        ],
        "min_price": 250000.00,
        "max_price": 500000.00
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 5,
      "total_items": 58,
      "per_page": 12,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

---

### Example 2: Login

**Frontend Code:**
```typescript
const loginData = {
  email: 'budi@gmail.com',
  password: 'password123',
  remember_me: false
};

const response = await fetch(`${API_BASE_URL}/auth/login`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(loginData)
});
```

**Expected Backend Response:**
```json
{
  "success": true,
  "message": "Login berhasil",
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

### Example 3: Checkout

**Frontend Code:**
```typescript
const checkoutData = {
  items: [
    {
      event_id: 1,
      ticket_type_id: 1,
      quantity: 2
    }
  ],
  payment_method: "BCA Virtual Account",
  customer_info: {
    name: "Budi Santoso",
    email: "budi@gmail.com",
    phone: "08123456789"
  }
};

const response = await fetch(`${API_BASE_URL}/orders`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}` // Optional
  },
  body: JSON.stringify(checkoutData)
});
```

**Expected Backend Response:**
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
          "event_date": "2026-03-15",
          "venue": "Bundaran HI - Monas",
          "image": "https://..."
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

---

## ⚠️ Common Pitfalls

### 1. Field Name Mismatch
❌ Backend return: `event_title`  
✅ Frontend expect: `title`

**Solution:** Match exact field names di TypeScript interfaces

### 2. Date Format
❌ Backend return: `"15-03-2026"` atau `"2026/03/15"`  
✅ Frontend expect: `"2026-03-15"` (ISO 8601)

### 3. Nested Objects
❌ Backend return flat data  
✅ Frontend expect nested: `event.category`, `order.order_items`

### 4. Price Format
❌ Backend return: `350000` (integer)  
✅ Frontend expect: `350000.00` (decimal with 2 places)

### 5. Array vs Object
❌ Backend return: `"facilities": "Medal, Certificate"`  
✅ Frontend expect: `"facilities": ["Medal", "Certificate"]`

---

## 🧪 Testing Checklist

### Phase 1: Authentication
- [ ] Login dengan email & password
- [ ] Register user baru
- [ ] Get current user data
- [ ] Logout
- [ ] Token expiry handling

### Phase 2: Public Features
- [ ] Browse events (list, filter, search)
- [ ] View event detail
- [ ] View categories
- [ ] Featured events carousel

### Phase 3: Checkout Flow
- [ ] Add items to cart (frontend localStorage)
- [ ] Checkout (create order)
- [ ] View payment instructions
- [ ] Check order status

### Phase 4: User Features
- [ ] View my tickets (upcoming & past)
- [ ] Download ticket PDF
- [ ] Ticket detail

### Phase 5: Admin Features
- [ ] Admin login (role check)
- [ ] Dashboard stats
- [ ] View transactions list
- [ ] Filter & search transactions
- [ ] Resend ticket email
- [ ] Event management (CRUD)

---

## 🔧 Environment Variables

Create `.env` file di frontend:

```bash
# Development
VITE_API_URL=http://localhost:8080/api/v1

# Production
VITE_API_URL=https://api.kartcis.id/api/v1
```

Usage di code:
```typescript
const API_BASE_URL = import.meta.env.VITE_API_URL;
```

---

## 📞 Need Help?

Jika ada yang tidak jelas:
1. Check TypeScript interfaces di `/src/app/types/index.ts`
2. Lihat mock implementation di `/src/app/services/api.ts`
3. Check API spec files untuk detail endpoint
4. Test dengan Postman/Thunder Client

---

**Last Updated:** January 21, 2026
