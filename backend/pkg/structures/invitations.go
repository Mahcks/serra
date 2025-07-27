package structures

// Invitation represents an invitation to join Serra
type Invitation struct {
	ID              int64     `json:"id"`
	Email           string    `json:"email"`
	Username        string    `json:"username"`
	Token           string    `json:"token,omitempty"` // Only include in creation response
	InvitedBy       string    `json:"invited_by"`
	InviterUsername string    `json:"inviter_username,omitempty"`
	Permissions     []string  `json:"permissions"`
	CreateMediaUser bool      `json:"create_media_user"`
	Status          string    `json:"status"` // pending, accepted, expired, cancelled
	ExpiresAt       string    `json:"expires_at"`
	AcceptedAt      string    `json:"accepted_at,omitempty"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
}

// CreateInvitationRequest represents a request to create an invitation
type CreateInvitationRequest struct {
	Email           string   `json:"email" validate:"required,email"`
	Username        string   `json:"username" validate:"required,min=3,max=50"`
	Permissions     []string `json:"permissions"`
	CreateMediaUser bool     `json:"create_media_user"`
	ExpiresInDays   int      `json:"expires_in_days"` // Defaults to 7 days
}

// AcceptInvitationRequest represents a request to accept an invitation
type AcceptInvitationRequest struct {
	Token           string `json:"token" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

// InvitationStats represents invitation statistics
type InvitationStats struct {
	PendingCount   int64 `json:"pending_count"`
	AcceptedCount  int64 `json:"accepted_count"`
	ExpiredCount   int64 `json:"expired_count"`
	CancelledCount int64 `json:"cancelled_count"`
	TotalCount     int64 `json:"total_count"`
}

// EmailSettings represents email configuration matching your settings
type EmailSettings struct {
	Enabled            bool   `json:"enabled"`
	RequireUserEmail   bool   `json:"require_user_email"`
	SenderName         string `json:"sender_name"`
	SenderAddress      string `json:"sender_address"`
	RequestAlert       bool   `json:"request_alert"`
	SMTPHost           string `json:"smtp_host"`
	SMTPPort           int    `json:"smtp_port"`
	EncryptionMethod   string `json:"encryption_method"` // "starttls", "implicit_tls", "none"
	UseSTARTTLS        bool   `json:"use_starttls"`
	AllowSelfSigned    bool   `json:"allow_self_signed"`
	SMTPUsername       string `json:"smtp_username"`
	SMTPPassword       string `json:"smtp_password"`
	PGPPrivateKey      string `json:"pgp_private_key,omitempty"`
	PGPPassword        string `json:"pgp_password,omitempty"`
}

// InvitationEmailData represents data for invitation email template
type InvitationEmailData struct {
	Username        string `json:"username"`
	InviterName     string `json:"inviter_name"`
	AppName         string `json:"app_name"`
	AppURL          string `json:"app_url"`
	AcceptURL       string `json:"accept_url"`
	MediaServerName string `json:"media_server_name"`
	MediaServerURL  string `json:"media_server_url"`
	ExpiresAt       string `json:"expires_at"`
}