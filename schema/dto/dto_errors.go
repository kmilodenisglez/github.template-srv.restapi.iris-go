package dto

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
)

// Problem api documentation
type Problem struct {
	Status uint   `example:"503"`
	Title  string `example:"err_code"`
	Detail string `example:"Some error details"`
}

type validationError struct {
	ActualTag string `json:"tag"`
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	Param     string `json:"param"`
}

// NewProblem construct a new api error struct and return a pointer to it
//
// - s [uint] ~ HTTP status tu respond
//
// - t [string] ~ Title of the error
//
// - d [string] ~ Description or detail of the error
func NewProblem(s uint, t string, d string) *Problem {
	return &Problem{Status: s, Title: t, Detail: d}
}

// HandleError the error, below you will find the right way to do that...
func HandleError(ctx iris.Context, err error, code int) {
	if errs, ok := err.(validator.ValidationErrors); ok {
		// Wrap the errors with JSON format, the underline library returns the errors as interface.
		validationErrors := wrapValidationErrors(errs)

		// Fire an application/json+problem response and stop the handlers chain.
		ctx.StopWithProblem(code, iris.NewProblem().
			Title("Validation error").
			Detail("One or more fields failed to be validated").
			Type(ctx.RouteName()).
			Key("errors", validationErrors))
		return
	}

	// It's probably an internal JSON error, let's dont give more info here.
	ctx.StopWithStatus(iris.StatusInternalServerError)
	return
}

func wrapValidationErrors(errs validator.ValidationErrors) []validationError {
	validationErrors := make([]validationError, 0, len(errs))
	for _, validationErr := range errs {
		validationErrors = append(validationErrors, validationError{
			ActualTag: validationErr.ActualTag(),
			Namespace: validationErr.Namespace(),
			Kind:      validationErr.Kind().String(),
			Type:      validationErr.Type().String(),
			Value:     fmt.Sprintf("%v", validationErr.Value()),
			Param:     validationErr.Param(),
		})
	}

	return validationErrors
}
