package helper

import (
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/service_errors"
	validation "github.com/minisource/go-common/validations"
)

type BaseHttpResponse struct {
	Result           any                           `json:"result"`
	Success          bool                          `json:"success"`
	ResultCode       ResultCode                    `json:"resultCode"`
	Message          string                        `json:"message,omitempty"`
	ValidationErrors *[]validation.ValidationError `json:"validationErrors,omitempty"`
	Error            any                           `json:"error,omitempty"`
}

func GenerateBaseResponse(result any, success bool, resultCode ResultCode) *BaseHttpResponse {
	return &BaseHttpResponse{
		Result:     result,
		Success:    success,
		ResultCode: resultCode,
	}
}

func GenerateBaseResponseWithError(result any, success bool, resultCode ResultCode, err error) *BaseHttpResponse {
	return &BaseHttpResponse{
		Result:     result,
		Success:    success,
		ResultCode: resultCode,
		Error:      err.Error(),
	}
}

func GenerateBaseResponseWithAnyError(result any, success bool, resultCode ResultCode, err any) *BaseHttpResponse {
	return &BaseHttpResponse{
		Result:     result,
		Success:    success,
		ResultCode: resultCode,
		Error:      err,
	}
}

func GenerateBaseResponseWithValidationError(result any, success bool, resultCode ResultCode, err error) *BaseHttpResponse {
	return &BaseHttpResponse{
		Result:           result,
		Success:          success,
		ResultCode:       resultCode,
		ValidationErrors: validation.GetValidationErrors(err),
	}
}

// GenerateBaseResponseWithMessage generates response with i18n message
func GenerateBaseResponseWithMessage(ctx interface{}, result any, success bool, resultCode ResultCode, messageKey string) *BaseHttpResponse {
	return &BaseHttpResponse{
		Result:     result,
		Success:    success,
		ResultCode: resultCode,
		Message:    i18n.T(ctx, messageKey),
	}
}

// GenerateBaseResponseWithServiceError generates response with service error (environment-aware)
func GenerateBaseResponseWithServiceError(ctx interface{}, result any, success bool, resultCode ResultCode, err *service_errors.ServiceError, isDevelopment bool) *BaseHttpResponse {
	response := &BaseHttpResponse{
		Result:     result,
		Success:    success,
		ResultCode: resultCode,
	}

	if err != nil {
		// Translate error message if it's a known error code
		if err.Code != "" {
			response.Message = i18n.T(ctx, "errors."+err.Code)
		} else {
			response.Message = err.EndUserMessage
		}

		// Add detailed error info in development mode
		if isDevelopment {
			response.Error = err.GetDetails(true)
		}
	}

	return response
}

// GenerateI18nResponse generates response with translated message
func GenerateI18nResponse(ctx interface{}, result any, success bool, resultCode ResultCode, messageKey string, params ...map[string]interface{}) *BaseHttpResponse {
	return &BaseHttpResponse{
		Result:     result,
		Success:    success,
		ResultCode: resultCode,
		Message:    i18n.T(ctx, messageKey, params...),
	}
}
