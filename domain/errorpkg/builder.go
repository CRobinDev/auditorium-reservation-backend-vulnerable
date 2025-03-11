package errorpkg

import "github.com/google/uuid"

type ErrorResponse struct {
	HttpStatusCode int        `json:"-"`
	Message        string     `json:"message"`
	Detail         any        `json:"detail,omitempty"`
	ErrorCode      string     `json:"error_code,omitempty"`
	TraceID        *uuid.UUID `json:"trace_id,omitempty"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

func NewError(httpStatusCode int) *ErrorResponse {
	return &ErrorResponse{
		HttpStatusCode: httpStatusCode,
	}
}

func (e *ErrorResponse) WithMessage(message string) *ErrorResponse {
	e.Message = message
	return e
}

func (e *ErrorResponse) WithDetail(payload any) *ErrorResponse {
	e.Detail = payload
	return e
}

func (e *ErrorResponse) WithErrorCode(errorCode string) *ErrorResponse {
	e.ErrorCode = errorCode
	return e
}

func (e *ErrorResponse) WithTraceID(traceID uuid.UUID) *ErrorResponse {
	e.TraceID = &traceID
	return e
}
