package permissions

// Owner permission (Super admin - all permissions)
const (
	Owner = "owner" // Full system access, cannot be revoked
)

// Administrative permissions (Owner/Admin only)
const (
	AdminUsers    = "admin.users"    // Manage user accounts and permissions
	AdminServices = "admin.services" // Configure Radarr, Sonarr, download clients
	AdminSystem   = "admin.system"   // System settings, storage, webhooks
)

// Request permissions (Most users get these)
const (
	RequestMovies = "request.movies" // Submit movie requests
	RequestSeries = "request.series" // Submit TV series requests
)

// 4K Request permissions (Special permission due to storage/bandwidth costs)
const (
	Request4KMovies = "request.4k_movies" // Submit 4K movie requests
	Request4KSeries = "request.4k_series" // Submit 4K TV series requests
)

// Auto-approval permissions (Automatically approve requests without manual review)
const (
	RequestAutoApproveMovies = "request.auto_approve_movies"    // Automatically approve movie requests
	RequestAutoApproveSeries = "request.auto_approve_series"    // Automatically approve TV series requests
	RequestAutoApprove4KMovies = "request.auto_approve_4k_movies" // Automatically approve 4K movie requests
	RequestAutoApprove4KSeries = "request.auto_approve_4k_series" // Automatically approve 4K TV series requests
)

// Request management permissions (Moderators)
const (
	RequestsView    = "requests.view"    // View all user requests
	RequestsApprove = "requests.approve" // Approve/deny requests
	RequestsManage  = "requests.manage"  // Edit/delete any requests
)

// Default permissions everyone gets (no permission check needed)
// - View main dashboard
// - View calendar/upcoming releases
// - View their own requests
// - View recently added media
// - View active downloads and progress
// - View all dashboard widgets

// All available permissions for individual assignment
var AllPermissions = []string{
	// Owner (super admin)
	Owner,

	// Administrative
	AdminUsers,
	AdminServices,
	AdminSystem,

	// Request permissions
	RequestMovies,
	RequestSeries,
	Request4KMovies,
	Request4KSeries,

	// Auto-approval permissions
	RequestAutoApproveMovies,
	RequestAutoApproveSeries,
	RequestAutoApprove4KMovies,
	RequestAutoApprove4KSeries,

	// Request management
	RequestsView,
	RequestsApprove,
	RequestsManage,
}

// GetPermissionDescription returns a human-readable description for a permission
func GetPermissionDescription(permission string) string {
	descriptions := map[string]string{
		// Owner (super admin)
		Owner: "Full system access - cannot be revoked",

		// Administrative (Owner/Admin only)
		AdminUsers:    "Manage user accounts and permissions",
		AdminServices: "Configure Radarr, Sonarr, and download clients",
		AdminSystem:   "Manage system settings, storage, and webhooks",

		// Request permissions
		RequestMovies: "Submit movie requests",
		RequestSeries: "Submit TV series requests",

		// 4K Request permissions
		Request4KMovies: "Submit 4K movie requests",
		Request4KSeries: "Submit 4K TV series requests",

		// Auto-approval permissions
		RequestAutoApproveMovies: "Automatically approve movie requests",
		RequestAutoApproveSeries: "Automatically approve TV series requests",
		RequestAutoApprove4KMovies: "Automatically approve 4K movie requests",
		RequestAutoApprove4KSeries: "Automatically approve 4K TV series requests",

		// Request management
		RequestsView:    "View all user requests",
		RequestsApprove: "Approve or deny pending requests",
		RequestsManage:  "Edit or delete any user requests",
	}

	if desc, exists := descriptions[permission]; exists {
		return desc
	}
	return "Unknown permission"
}

// GetAllPermissions returns all available permissions for selection
func GetAllPermissions() []string {
	return AllPermissions
}

// PermissionInfo holds permission details for UI display
type PermissionInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Dangerous   bool   `json:"dangerous"` // Marks admin-level permissions
}

// GetPermissionsByCategory returns permissions grouped by category for UI display
func GetPermissionsByCategory() map[string][]string {
	return map[string][]string{
		"Owner": {
			Owner,
		},
		"Administrative": {
			AdminUsers,
			AdminServices,
			AdminSystem,
		},
		"Request Content": {
			RequestMovies,
			RequestSeries,
			Request4KMovies,
			Request4KSeries,
		},
		"Auto-Approve Requests": {
			RequestAutoApproveMovies,
			RequestAutoApproveSeries,
			RequestAutoApprove4KMovies,
			RequestAutoApprove4KSeries,
		},
		"Manage Requests": {
			RequestsView,
			RequestsApprove,
			RequestsManage,
		},
	}
}

// GetPermissionInfo returns detailed information about a permission
func GetPermissionInfo(permission string) PermissionInfo {
	info := PermissionInfo{
		ID:          permission,
		Description: GetPermissionDescription(permission),
	}

	switch permission {
	// Owner permission - extremely dangerous
	case Owner:
		info.Name = "Owner"
		info.Category = "Owner"
		info.Dangerous = true
	// Administrative permissions - dangerous
	case AdminUsers:
		info.Name = "Manage Users"
		info.Category = "Administrative"
		info.Dangerous = true
	case AdminServices:
		info.Name = "Manage Services"
		info.Category = "Administrative"
		info.Dangerous = true
	case AdminSystem:
		info.Name = "System Settings"
		info.Category = "Administrative"
		info.Dangerous = true

	// Request permissions - safe
	case RequestMovies:
		info.Name = "Request Movies"
		info.Category = "Request Content"
	case RequestSeries:
		info.Name = "Request Series"
		info.Category = "Request Content"
	case Request4KMovies:
		info.Name = "Request 4K Movies"
		info.Category = "Request Content"
	case Request4KSeries:
		info.Name = "Request 4K Series"
		info.Category = "Request Content"

	// Auto-approval permissions
	case RequestAutoApproveMovies:
		info.Name = "Auto-Approve Movies"
		info.Category = "Auto-Approve Requests"
	case RequestAutoApproveSeries:
		info.Name = "Auto-Approve Series"
		info.Category = "Auto-Approve Requests"
	case RequestAutoApprove4KMovies:
		info.Name = "Auto-Approve 4K Movies"
		info.Category = "Auto-Approve Requests"
	case RequestAutoApprove4KSeries:
		info.Name = "Auto-Approve 4K Series"
		info.Category = "Auto-Approve Requests"

	// Request management permissions - moderate risk
	case RequestsView:
		info.Name = "View All Requests"
		info.Category = "Manage Requests"
	case RequestsApprove:
		info.Name = "Approve Requests"
		info.Category = "Manage Requests"
	case RequestsManage:
		info.Name = "Manage Requests"
		info.Category = "Manage Requests"
	}

	return info
}

// GetAllPermissionInfo returns detailed info for all permissions
func GetAllPermissionInfo() []PermissionInfo {
	var permissions []PermissionInfo
	for _, perm := range AllPermissions {
		permissions = append(permissions, GetPermissionInfo(perm))
	}
	return permissions
}

// IsOwnerPermission checks if a permission is owner level
func IsOwnerPermission(permission string) bool {
	return permission == Owner
}

// IsAdminPermission checks if a permission is administrative level
func IsAdminPermission(permission string) bool {
	return permission == AdminUsers || permission == AdminServices || permission == AdminSystem
}

// Is4KPermission checks if a permission is for 4K content
func Is4KPermission(permission string) bool {
	return permission == Request4KMovies || permission == Request4KSeries
}

// IsValidPermission checks if a permission string is valid
func IsValidPermission(permission string) bool {
	_, exists := map[string]bool{
		// Owner permission
		Owner: true,
		// Admin permissions
		AdminUsers: true, AdminServices: true, AdminSystem: true,
		// Request permissions
		RequestMovies: true, RequestSeries: true,
		// 4K Request permissions
		Request4KMovies: true, Request4KSeries: true,
		// Auto-approval permissions
		RequestAutoApproveMovies: true, RequestAutoApproveSeries: true,
		RequestAutoApprove4KMovies: true, RequestAutoApprove4KSeries: true,
		// Request management
		RequestsView: true, RequestsApprove: true, RequestsManage: true,
	}[permission]
	return exists
}
