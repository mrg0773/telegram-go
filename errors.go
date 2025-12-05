package telegram

import "fmt"

// APIError represents Telegram API error
type APIError struct {
	Code        int
	Description string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("telegram api error: code=%d, description=%s", e.Code, e.Description)
}

// IsBlockedError checks if error is "bot was blocked by the user" (403)
func IsBlockedError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == 403
	}
	return false
}

// IsRateLimitError checks if error is rate limit (429)
func IsRateLimitError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == 429
	}
	return false
}

// IsNotFoundError checks if error is not found (404)
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == 404
	}
	return false
}

// IsBadRequestError checks if error is bad request (400)
func IsBadRequestError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == 400
	}
	return false
}

// IsUnauthorizedError checks if error is unauthorized (401)
func IsUnauthorizedError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == 401
	}
	return false
}

// IsForbiddenError checks if error is forbidden (403)
func IsForbiddenError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == 403
	}
	return false
}

// GetErrorCode returns error code if it's APIError, otherwise -1
func GetErrorCode(err error) int {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code
	}
	return -1
}
