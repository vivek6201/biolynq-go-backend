package utils

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

type GenericResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp string      `json:"timestamp"`
}

type ErrorResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Error     interface{} `json:"error,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp string      `json:"timestamp"`
}

func SendSuccess(c fiber.Ctx, status int, message string, data interface{}) error {
	reqID := getRequestID(c)
	return c.Status(status).JSON(GenericResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: reqID,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func SendError(c fiber.Ctx, status int, message string, errDetails interface{}) error {
	reqID := getRequestID(c)

	var errVal interface{}
	if err, ok := errDetails.(error); ok {
		errVal = fiber.Map{"details": err.Error()}
	} else if errStr, ok := errDetails.(string); ok {
		errVal = fiber.Map{"details": errStr}
	} else {
		errVal = errDetails
	}

	return c.Status(status).JSON(ErrorResponse{
		Success:   false,
		Message:   message,
		Error:     errVal,
		RequestID: reqID,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func getRequestID(c fiber.Ctx) string {
	return requestid.FromContext(c)
}
