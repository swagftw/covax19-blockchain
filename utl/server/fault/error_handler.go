package fault

import (
	"github.com/labstack/echo/v4"
)

func (he *HTTPError) Error() string {
	return he.Message
}

func New(code string, message string, statusCode int) error {
	err := &HTTPError{
		ErrorCode:  code,
		Message:    message,
		StatusCode: statusCode,
	}

	return err
}

type HTTPError struct {
	ErrorCode  string `json:"errorCode"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func ErrorHandler(err error, ctx echo.Context) {
	if err == nil {
		return
	}

	httpError := new(HTTPError)

	switch e := err.(type) {
	case *HTTPError:
		httpError = e
	case *echo.HTTPError:
		httpError.StatusCode = e.Code
		httpError.Message = e.Error()
		httpError.ErrorCode = "ECHO_ERROR"
	default:
		httpError.StatusCode = 500
		httpError.Message = err.Error()
		httpError.ErrorCode = "UNKNOWN_ERROR"
	}

	// Send response
	if !ctx.Response().Committed {
		_ = ctx.JSON(httpError.StatusCode, httpError)
	}
}
