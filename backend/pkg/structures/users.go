package structures

// User types
const (
	UserTypeMediaServer = "media_server"
	UserTypeLocal       = "local"
)

// LocalUserRegistrationRequest represents a request to create a local user
type LocalUserRegistrationRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LocalUserLoginRequest represents a local user login request  
type LocalUserLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// ChangePasswordRequest represents a password change request for local users
type ChangePasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type GetAllUsersResponse struct {
	Total int64                 `json:"total"`
	Users []UserWithPermissions `json:"users"`
}