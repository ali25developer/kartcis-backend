package models

import (
	"time"

	"gorm.io/gorm"
)

type SiteSetting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"uniqueIndex" json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID              uint            `gorm:"primaryKey" json:"id"`
	Name            string          `json:"name"`
	Email           string          `gorm:"unique" json:"email"`
	Password        string          `json:"-"` // Can be empty for OAuth users
	Phone           string          `json:"phone"`
	Role            string          `json:"role" gorm:"default:user"`     // admin, organizer, user
	CustomFee       *float64        `json:"custom_fee"`                   // Specific fee for this organizer
	Status          string          `json:"status" gorm:"default:active"` // active, inactive, banned
	Avatar          string          `json:"avatar"`
	EmailVerifiedAt *time.Time      `json:"email_verified_at"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	SocialAccounts  []SocialAccount `json:"social_accounts" gorm:"foreignKey:UserID"`
}

type SocialAccount struct {
	ID                   uint   `gorm:"primaryKey" json:"id"`
	UserID               uint   `json:"user_id"`
	Provider             string `json:"provider"`
	ProviderID           string `json:"provider_id"`
	ProviderToken        string `json:"-"`
	ProviderRefreshToken string `json:"-"`
	// ProviderData    postgres.Jsonb `json:"provider_data" gorm:"type:jsonb"` // Complex in Gorm without proper strict, skipping struct binding for simplified demo or using string
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Category struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Slug         string    `gorm:"uniqueIndex" json:"slug"`
	Description  string    `json:"description"`
	Icon         string    `json:"icon"`
	Image        string    `json:"image"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Event struct {
	ID                  uint         `gorm:"primaryKey" json:"id"`
	Title               string       `json:"title"`
	Slug                string       `json:"slug"`
	Description         string       `json:"description"`
	DetailedDescription string       `json:"detailed_description"`
	EventDate           time.Time    `json:"event_date"`
	EventTime           string       `json:"event_time"`
	Venue               string       `json:"venue"`
	City                string       `json:"city"`
	Organizer           string       `json:"organizer"` // Display name
	OrganizerID         uint         `json:"organizer_id"`
	OrganizerUser       User         `json:"organizer_user" gorm:"foreignKey:OrganizerID"`
	Image               string       `json:"image"`
	Quota               int          `json:"quota"`
	IsFeatured          bool         `json:"is_featured"`
	Status              string       `json:"status"` // draft, published, completed, cancelled, sold_out
	CategoryID          uint         `json:"category_id"`
	Category            Category     `json:"category" gorm:"foreignKey:CategoryID"`
	MinPrice            float64      `json:"min_price"`
	MaxPrice            float64      `json:"max_price"`
	TicketTypes         []TicketType `json:"ticket_types" gorm:"foreignKey:EventID"`
	CustomFields        string       `json:"custom_fields"` // JSON string for form definition
	FeePercentage       float64      `json:"fee_percentage"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

type TicketType struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	EventID       uint      `json:"event_id"`
	Event         Event     `json:"event" gorm:"foreignKey:EventID"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	OriginalPrice float64   `json:"original_price"`
	Quota         int       `json:"quota"`
	Available     int       `json:"available"`
	Sold          int       `json:"sold" gorm:"-"` // Virtual field: Quota - Available
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Hooks to calculate Sold field
func (tt *TicketType) AfterFind(tx *gorm.DB) (err error) {
	tt.Sold = tt.Quota - tt.Available
	return
}

func (tt *TicketType) AfterSave(tx *gorm.DB) (err error) {
	tt.Sold = tt.Quota - tt.Available
	return
}

type Order struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	UserID               *uint      `json:"user_id"`
	OrderNumber          string     `json:"order_number"`
	CustomerName         string     `json:"customer_name"`
	CustomerEmail        string     `json:"customer_email"`
	CustomerPhone        string     `json:"customer_phone"`
	TotalAmount          float64    `json:"total_amount"`
	AdminFee             float64    `json:"admin_fee"`
	UniqueCode           int        `json:"unique_code"`
	Status               string     `json:"status"`
	PaymentMethod        string     `json:"payment_method"`
	VirtualAccountNumber string     `json:"virtual_account_number"`
	PaymentURL           string     `json:"payment_url"`  // URL for E-Wallet redirect / QRIS
	PaymentData          string     `json:"payment_data"` // JSON string for raw gateway response
	PaymentInstructions  string     `json:"payment_instructions"`
	PaidAt               *time.Time `json:"paid_at"`
	ExpiresAt            *time.Time `json:"expires_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	Tickets              []Ticket   `json:"tickets" gorm:"foreignKey:OrderID"`
}

type Ticket struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	OrderID              *uint      `json:"order_id"`
	Order                Order      `json:"order" gorm:"foreignKey:OrderID"`
	EventID              uint       `json:"event_id"`
	Event                Event      `json:"event" gorm:"foreignKey:EventID"`
	TicketTypeID         uint       `json:"ticket_type_id"`
	TicketType           TicketType `json:"ticket_type" gorm:"foreignKey:TicketTypeID"`
	TicketCode           string     `json:"ticket_code"`
	AttendeeName         string     `json:"attendee_name"`
	AttendeeEmail        string     `json:"attendee_email"`
	AttendeePhone        string     `json:"attendee_phone"`
	Status               string     `json:"status"` // active, used
	CheckInAt            *time.Time `json:"check_in_at"`
	CustomFieldResponses string     `json:"custom_field_responses"` // JSON string with responses
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type ActivityLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Action    string    `json:"action"` // login, update_profile, purchase, etc.
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Details   string    `json:"details"` // JSON string or text
	CreatedAt time.Time `json:"created_at"`
}

type OrderStatusHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `json:"order_id"`
	Order     Order     `json:"-" gorm:"foreignKey:OrderID"`
	Status    string    `json:"status"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

type RequestLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    *uint     `json:"user_id"` // Nullable
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	Latency   int64     `json:"latency"` // in milliseconds:w
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type PasswordReset struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `json:"email" gorm:"index"`
	Token     string    `json:"token" gorm:"index"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
