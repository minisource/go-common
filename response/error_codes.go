package response

// Common error codes for use across all microservices
const (
	// Authentication errors
	ErrCodeUnauthorized       = "AUTH_UNAUTHORIZED"
	ErrCodeTokenExpired       = "AUTH_TOKEN_EXPIRED"
	ErrCodeTokenInvalid       = "AUTH_TOKEN_INVALID"
	ErrCodeSessionExpired     = "AUTH_SESSION_EXPIRED"
	ErrCodeInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	ErrCodeAccountLocked      = "AUTH_ACCOUNT_LOCKED"
	ErrCodeAccountDisabled    = "AUTH_ACCOUNT_DISABLED"
	ErrCodeMFARequired        = "AUTH_MFA_REQUIRED"
	ErrCodeMFAInvalid         = "AUTH_MFA_INVALID"

	// Authorization errors
	ErrCodeForbidden          = "AUTHZ_FORBIDDEN"
	ErrCodeInsufficientScope  = "AUTHZ_INSUFFICIENT_SCOPE"
	ErrCodeResourceNotAllowed = "AUTHZ_RESOURCE_NOT_ALLOWED"
	ErrCodeTenantAccessDenied = "AUTHZ_TENANT_ACCESS_DENIED"

	// Validation errors
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeRequiredField    = "VALIDATION_REQUIRED"
	ErrCodeInvalidFormat    = "VALIDATION_INVALID_FORMAT"
	ErrCodeInvalidEmail     = "VALIDATION_INVALID_EMAIL"
	ErrCodeInvalidPhone     = "VALIDATION_INVALID_PHONE"
	ErrCodeMinLength        = "VALIDATION_MIN_LENGTH"
	ErrCodeMaxLength        = "VALIDATION_MAX_LENGTH"
	ErrCodeMinValue         = "VALIDATION_MIN_VALUE"
	ErrCodeMaxValue         = "VALIDATION_MAX_VALUE"
	ErrCodeInvalidEnum      = "VALIDATION_INVALID_ENUM"
	ErrCodeInvalidUUID      = "VALIDATION_INVALID_UUID"
	ErrCodeInvalidDate      = "VALIDATION_INVALID_DATE"
	ErrCodePasswordWeak     = "VALIDATION_PASSWORD_WEAK"
	ErrCodePasswordMismatch = "VALIDATION_PASSWORD_MISMATCH"

	// Resource errors
	ErrCodeNotFound      = "RESOURCE_NOT_FOUND"
	ErrCodeAlreadyExists = "RESOURCE_ALREADY_EXISTS"
	ErrCodeConflict      = "RESOURCE_CONFLICT"
	ErrCodeGone          = "RESOURCE_GONE"
	ErrCodeLocked        = "RESOURCE_LOCKED"

	// Rate limiting errors
	ErrCodeRateLimited     = "RATE_LIMIT_EXCEEDED"
	ErrCodeQuotaExceeded   = "QUOTA_EXCEEDED"
	ErrCodeTooManyRequests = "TOO_MANY_REQUESTS"

	// Input errors
	ErrCodeBadRequest       = "BAD_REQUEST"
	ErrCodeInvalidJSON      = "INVALID_JSON"
	ErrCodeInvalidQuery     = "INVALID_QUERY_PARAMS"
	ErrCodeMissingHeader    = "MISSING_REQUIRED_HEADER"
	ErrCodeUnsupportedMedia = "UNSUPPORTED_MEDIA_TYPE"
	ErrCodePayloadTooLarge  = "PAYLOAD_TOO_LARGE"

	// Server errors
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeDatabaseError      = "DATABASE_ERROR"
	ErrCodeCacheError         = "CACHE_ERROR"
	ErrCodeExternalService    = "EXTERNAL_SERVICE_ERROR"
	ErrCodeTimeout            = "OPERATION_TIMEOUT"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeNotImplemented     = "NOT_IMPLEMENTED"

	// Business logic errors
	ErrCodeOperationFailed   = "OPERATION_FAILED"
	ErrCodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	ErrCodeLimitExceeded     = "LIMIT_EXCEEDED"
	ErrCodeExpired           = "EXPIRED"
	ErrCodeAlreadyProcessed  = "ALREADY_PROCESSED"
	ErrCodeInvalidState      = "INVALID_STATE"
	ErrCodeDependencyFailed  = "DEPENDENCY_FAILED"

	// File/Storage errors
	ErrCodeFileNotFound    = "FILE_NOT_FOUND"
	ErrCodeFileTooLarge    = "FILE_TOO_LARGE"
	ErrCodeInvalidFileType = "INVALID_FILE_TYPE"
	ErrCodeUploadFailed    = "UPLOAD_FAILED"
	ErrCodeStorageError    = "STORAGE_ERROR"

	// Tenant errors
	ErrCodeTenantNotFound      = "TENANT_NOT_FOUND"
	ErrCodeTenantSuspended     = "TENANT_SUSPENDED"
	ErrCodeTenantLimitExceeded = "TENANT_LIMIT_EXCEEDED"

	// OTP errors
	ErrCodeOTPExpired     = "OTP_EXPIRED"
	ErrCodeOTPInvalid     = "OTP_INVALID"
	ErrCodeOTPMaxAttempts = "OTP_MAX_ATTEMPTS"
)

// ErrorCodeToStatus maps error codes to HTTP status codes
var ErrorCodeToStatus = map[string]int{
	// 400 Bad Request
	ErrCodeBadRequest:       400,
	ErrCodeInvalidJSON:      400,
	ErrCodeInvalidQuery:     400,
	ErrCodeValidationFailed: 400,

	// 401 Unauthorized
	ErrCodeUnauthorized:       401,
	ErrCodeTokenExpired:       401,
	ErrCodeTokenInvalid:       401,
	ErrCodeInvalidCredentials: 401,
	ErrCodeSessionExpired:     401,

	// 403 Forbidden
	ErrCodeForbidden:          403,
	ErrCodeInsufficientScope:  403,
	ErrCodeResourceNotAllowed: 403,
	ErrCodeTenantAccessDenied: 403,
	ErrCodeAccountLocked:      403,
	ErrCodeAccountDisabled:    403,
	ErrCodeTenantSuspended:    403,

	// 404 Not Found
	ErrCodeNotFound:       404,
	ErrCodeTenantNotFound: 404,
	ErrCodeFileNotFound:   404,

	// 409 Conflict
	ErrCodeConflict:         409,
	ErrCodeAlreadyExists:    409,
	ErrCodeAlreadyProcessed: 409,

	// 410 Gone
	ErrCodeGone:    410,
	ErrCodeExpired: 410,

	// 413 Payload Too Large
	ErrCodePayloadTooLarge: 413,
	ErrCodeFileTooLarge:    413,

	// 415 Unsupported Media Type
	ErrCodeUnsupportedMedia: 415,
	ErrCodeInvalidFileType:  415,

	// 422 Unprocessable Entity
	ErrCodeRequiredField:    422,
	ErrCodeInvalidFormat:    422,
	ErrCodeInvalidEmail:     422,
	ErrCodePasswordWeak:     422,
	ErrCodePasswordMismatch: 422,

	// 423 Locked
	ErrCodeLocked: 423,

	// 429 Too Many Requests
	ErrCodeRateLimited:     429,
	ErrCodeQuotaExceeded:   429,
	ErrCodeTooManyRequests: 429,
	ErrCodeOTPMaxAttempts:  429,

	// 500 Internal Server Error
	ErrCodeInternalError:   500,
	ErrCodeDatabaseError:   500,
	ErrCodeCacheError:      500,
	ErrCodeOperationFailed: 500,
	ErrCodeStorageError:    500,
	ErrCodeUploadFailed:    500,

	// 501 Not Implemented
	ErrCodeNotImplemented: 501,

	// 502 Bad Gateway
	ErrCodeExternalService:  502,
	ErrCodeDependencyFailed: 502,

	// 503 Service Unavailable
	ErrCodeServiceUnavailable: 503,

	// 504 Gateway Timeout
	ErrCodeTimeout: 504,
}

// GetStatusForCode returns the HTTP status code for an error code
func GetStatusForCode(code string) int {
	if status, ok := ErrorCodeToStatus[code]; ok {
		return status
	}
	return 500 // Default to internal error
}
