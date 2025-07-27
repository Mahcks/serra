package apiErrors

import (
	"fmt"
	"strings"

	"github.com/mahcks/serra/utils"
	"github.com/valyala/fasthttp"
)

type APIError interface {
	Error() string
	Message() string
	Code() int
	SetDetail(str string, a ...any) APIError
	SetFields(d Fields) APIError
	GetFields() Fields
	ExpectedHTTPStatus() int
	WithHTTPStatus(s int) APIError
}

type apiErrorFunc func() APIError

var (
	// Generic client errors
	ErrUnauthorized            apiErrorFunc = DefineError(10401, "Authorization Required", fasthttp.StatusUnauthorized)
	ErrTokenExpired            apiErrorFunc = DefineError(10402, "Token Expired", fasthttp.StatusUnauthorized)
	ErrInvalidToken            apiErrorFunc = DefineError(10403, "Invalid Token", fasthttp.StatusUnauthorized)
	ErrInsufficientPermissions apiErrorFunc = DefineError(10404, "Insufficient Permissions", fasthttp.StatusForbidden)
	ErrBadRequest              apiErrorFunc = DefineError(10405, "Bad Request", fasthttp.StatusBadRequest)
	ErrForbidden               apiErrorFunc = DefineError(10406, "Forbidden", fasthttp.StatusForbidden)
	ErrConflict                apiErrorFunc = DefineError(10407, "Conflict", fasthttp.StatusConflict)      // Used for conflicts like duplicate entries.
	ErrBadGateway              apiErrorFunc = DefineError(10408, "Bad Gateway", fasthttp.StatusBadGateway) // Used for upstream service errors.
	ErrTooManyRequests         apiErrorFunc = DefineError(10409, "Too Many Requests", fasthttp.StatusTooManyRequests)

	// Client type errors
	ErrValidationRejected         apiErrorFunc = DefineError(10410, "Validation Rejected", fasthttp.StatusBadRequest)
	ErrMissingEnvironmentVariable apiErrorFunc = DefineError(10411, "Missing Required Environment Variable", fasthttp.StatusBadRequest)

	// Server errors
	ErrInternalServerError apiErrorFunc = DefineError(10500, "Internal Server Error", fasthttp.StatusInternalServerError)
	ErrNotFound            apiErrorFunc = DefineError(10501, "Not Found", fasthttp.StatusNotFound)
	ErrInvalidSignature    apiErrorFunc = DefineError(10502, "Invalid Signature", fasthttp.StatusForbidden)

	// Request processing errors
	ErrNoRadarrInstances    apiErrorFunc = DefineError(10600, "No Radarr instances are configured. Please contact your administrator to set up movie automation.", fasthttp.StatusInternalServerError)
	ErrNoSonarrInstances    apiErrorFunc = DefineError(10601, "No Sonarr instances are configured. Please contact your administrator to set up TV show automation.", fasthttp.StatusInternalServerError)
	ErrInvalidQualityProfile apiErrorFunc = DefineError(10602, "The configured quality profile is invalid. Please contact your administrator to fix the automation setup.", fasthttp.StatusInternalServerError)
	ErrRadarrConnection     apiErrorFunc = DefineError(10603, "Unable to connect to Radarr. The movie automation service may be down.", fasthttp.StatusBadGateway)
	ErrSonarrConnection     apiErrorFunc = DefineError(10604, "Unable to connect to Sonarr. The TV show automation service may be down.", fasthttp.StatusBadGateway)

	// Request validation errors
	ErrDuplicateRequest     apiErrorFunc = DefineError(10610, "You have already requested this content. Check your existing requests.", fasthttp.StatusConflict)
	ErrInvalidMediaType     apiErrorFunc = DefineError(10611, "Invalid content type. Only movies and TV shows are supported.", fasthttp.StatusBadRequest)
	ErrMissingTMDBID        apiErrorFunc = DefineError(10612, "This content is missing required information. Please try a different title.", fasthttp.StatusBadRequest)
	ErrInvalidSeasons       apiErrorFunc = DefineError(10613, "The selected seasons are invalid. Please choose valid season numbers.", fasthttp.StatusBadRequest)
	ErrRequestNotApproved   apiErrorFunc = DefineError(10614, "This request has not been approved yet and cannot be processed.", fasthttp.StatusBadRequest)
	ErrSeasonParsingFailed  apiErrorFunc = DefineError(10615, "Unable to process the selected seasons. Please try requesting again.", fasthttp.StatusBadRequest)

	// Permission errors
	ErrNoRequestPermission  apiErrorFunc = DefineError(10620, "You don't have permission to request this type of content. Contact your administrator for access.", fasthttp.StatusForbidden)
	ErrNoApprovalPermission apiErrorFunc = DefineError(10621, "You don't have permission to approve requests. Contact your administrator for access.", fasthttp.StatusForbidden)
	ErrNoManagePermission   apiErrorFunc = DefineError(10622, "You don't have permission to manage requests. Contact your administrator for access.", fasthttp.StatusForbidden)
	ErrNo4KPermission       apiErrorFunc = DefineError(10623, "You don't have permission to request 4K content. Contact your administrator for access.", fasthttp.StatusForbidden)

	// Processing errors
	ErrRadarrAddFailed      apiErrorFunc = DefineError(10630, "Failed to add movie to download queue. The automation service may be experiencing issues.", fasthttp.StatusBadGateway)
	ErrSonarrAddFailed      apiErrorFunc = DefineError(10631, "Failed to add TV show to download queue. The automation service may be experiencing issues.", fasthttp.StatusBadGateway)
	ErrProcessingTimeout    apiErrorFunc = DefineError(10632, "Request processing timed out. Please try again later or contact support.", fasthttp.StatusInternalServerError)
)

type apiError struct {
	message            string
	code               int
	fields             Fields
	expectedHTTPStatus int
}

type Fields map[string]interface{}

func (e *apiError) Error() string {
	return fmt.Sprintf("[%d] %s", e.code, strings.ToLower(e.message))
}

func (e *apiError) Message() string {
	return e.message
}

func (e *apiError) Code() int {
	return e.code
}

func (e *apiError) SetDetail(str string, a ...any) APIError {
	e.message = e.message + ": " + utils.Ternary(len(a) > 0, fmt.Sprintf(str, a...), str)
	return e
}

func (e *apiError) SetFields(d Fields) APIError {
	e.fields = d
	return e
}

func (e *apiError) GetFields() Fields {
	return e.fields
}

func (e *apiError) ExpectedHTTPStatus() int {
	return e.expectedHTTPStatus
}

func (e *apiError) WithHTTPStatus(s int) APIError {
	e.expectedHTTPStatus = s
	return e
}

func DefineError(code int, message string, httpStatus int) func() APIError {
	return func() APIError {
		return &apiError{
			message:            message,
			code:               code,
			fields:             Fields{},
			expectedHTTPStatus: httpStatus,
		}
	}
}
