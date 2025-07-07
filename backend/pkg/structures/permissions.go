package structures

// AssignPermissionRequest represents a request to assign a permission to a user
type AssignPermissionRequest struct {
	Permission string `json:"permission" validate:"required"`
}

// BulkUpdatePermissionsRequest represents a request to update all permissions for a user
type BulkUpdatePermissionsRequest struct {
	Permissions []string `json:"permissions" validate:"required"`
}

// UserPermissionResponse represents a user's permission information
type UserPermissionResponse struct {
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Permissions []PermissionInfo       `json:"permissions"`
}

// PermissionInfo represents detailed permission information for API responses
type PermissionInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Dangerous   bool   `json:"dangerous"`
}

// PermissionsListResponse represents the response for listing all permissions
type PermissionsListResponse struct {
	Permissions []PermissionInfo            `json:"permissions"`
	Categories  map[string][]string         `json:"categories"`
}

// UserPermissionsUpdateLog represents an audit log entry for permission changes
type UserPermissionsUpdateLog struct {
	UserID       string   `json:"user_id"`
	UpdatedBy    string   `json:"updated_by"`
	Added        []string `json:"added"`
	Removed      []string `json:"removed"`
	Timestamp    int64    `json:"timestamp"`
}

// UserWithPermissions represents a user with their assigned permissions
type UserWithPermissions struct {
	ID          string           `json:"id"`
	Username    string           `json:"username"`
	Email       string           `json:"email"`
	AvatarUrl   string           `json:"avatar_url,omitempty"`
	UserType    string           `json:"user_type"`
	CreatedAt   string           `json:"created_at,omitempty"`
	Permissions []PermissionInfo `json:"permissions"`
}