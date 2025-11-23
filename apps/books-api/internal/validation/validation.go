package validation

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors,omitempty"`
}

func BindAndValidateJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			resp := formatValidationErrors(verrs)
			c.AbortWithStatusJSON(http.StatusBadRequest, resp)
			return false
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST_BODY",
			Message: "invalid request body",
			Errors: []FieldError{
				{
					Field:   "",
					Rule:    "syntax",
					Message: err.Error(),
				},
			},
		})
		return false
	}

	return true
}

func formatValidationErrors(verrs validator.ValidationErrors) ErrorResponse {
	fields := make([]FieldError, 0, len(verrs))

	for _, fe := range verrs {
		jsonField := toJSONFieldName(fe.Field())
		fields = append(fields, FieldError{
			Field:   jsonField,
			Rule:    fe.Tag(),
			Message: buildMessage(jsonField, fe),
		})
	}

	return ErrorResponse{
		Code:    "VALIDATION_ERROR",
		Message: "validation failed",
		Errors:  fields,
	}
}

func toJSONFieldName(field string) string {
	if field == "" {
		return field
	}
	return strings.ToLower(field[:1]) + field[1:]
}

func buildMessage(field string, fe validator.FieldError) string {
	if fe.Tag() == "required" {
		return field + " is required"
	}
	return field + " is invalid (" + fe.Tag() + ")"
}
