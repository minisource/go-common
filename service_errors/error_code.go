package service_errors

const (
	// Token
	UnExpectedError = "unexpected_error"
	ClaimsNotFound  = "claims_not_found"
	TokenRequired   = "token_required"
	TokenExpired    = "token_expired"
	TokenInvalid    = "token_invalid"

	// OTP
	OptExists   = "otp_exists"
	OtpUsed     = "otp_used"
	OtpNotValid = "otp_not_valid"

	// User
	UserDisabled     = "user_disabled"
	EmailExists      = "email_exists"
	UsernameExists   = "username_exists"
	PermissionDenied = "permission_denied"

	// DB
	RecordNotFound = "record_not_found"

	// Validation
	ValidationError = "validation_error"

	// HTTP
	BadRequest         = "bad_request"
	Unauthorized       = "unauthorized"
	Forbidden          = "forbidden"
	NotFound           = "not_found"
	Conflict           = "conflict"
	InternalError      = "internal_error"
	ServiceUnavailable = "service_unavailable"
	TooManyRequests    = "too_many_requests"
)
