package lib

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"reflect"
	reg "regexp"
	"restapi.app/schema/dto"
	"strings"
)

// ValidateString validate a string given a regular expression
func ValidateString(data string, regexp string) bool {
	return reg.MustCompile(regexp).MatchString(data)
}

// ValidateStringCollection validate a string collection given a regular expression
func ValidateStringCollection(data []interface{}, regexp string) bool {
	var fn govalidator.ConditionIterator = func(value interface{}, index int) bool {
		return reg.MustCompile(regexp).MatchString(value.(string))
	}
	return govalidator.ValidateArray(data, fn)
}

// ValidateStringCollectionUsingValidator10 validation into arrays type
// e.g.
// tag = "required,max=10,min=1,dive,max=12"
//
// 			max=10 		 -> Max array len
// 			dive, max=12 -> Max length of every array element
func ValidateStringCollectionUsingValidator10(validate *validator.Validate, data any, tag string) bool {
	// variable must be a slice
	if reflect.TypeOf(data).Kind() != reflect.Slice {
		return false
	}

	errs := validate.Var(data, tag)
	if errs != nil {
		return false
	}
	if reflect.TypeOf(data).Elem().Kind() != reflect.String {
		return false
	}
	return true
}

// InitValidator Activate behavior to require all fields and adding new validators
func InitValidator(validate *validator.Validate) error {
	// Add your own struct validation tags
	// validates medication name
	err := validate.RegisterValidation("medication_name_validation", func(fl validator.FieldLevel) bool {
		fmt.Println("medication_name_validation: ", fl.Field().String(), reg.MustCompile(dto.RegexpMedicationCode).MatchString(fl.Field().String()))
		return reg.MustCompile(dto.RegexpMedicationName).MatchString(fl.Field().String())
	})
	if err != nil {
		return err
	}

	// validates medication code
	err = validate.RegisterValidation("medication_code_validation", func(fl validator.FieldLevel) bool {
		fmt.Println("medication_code_validation: ", fl.Field().String(), reg.MustCompile(dto.RegexpMedicationCode).MatchString(fl.Field().String()))
		return reg.MustCompile(dto.RegexpMedicationCode).MatchString(fl.Field().String())
	})
	if err != nil {
		return err
	}

	// validates that an enum is within the interval
	err = validate.RegisterValidation("drone_enum_validation", func(fl validator.FieldLevel) bool {
		value := fl.Field().Interface().(dto.DroneModel)
		return value.String() != "unknown"
	})
	if err != nil {
		return err
	}

	// validates that an enum is within the interval
	err = validate.RegisterValidation("drone_state_validation", func(fl validator.FieldLevel) bool {
		value := fl.Field().Interface().(dto.DroneState)
		return value.String() != "unknown"
	})
	if err != nil {
		return err
	}

	return nil
}

func ValidateSerialNumberDrone(validate *validator.Validate, serialNumber string) bool {
	return validate.Var(serialNumber, fmt.Sprintf("required,max=%s", dto.MaxSerialNumberLength)) == nil
}

func CalculateDroneWeightLimit(model dto.DroneModel) float64 {
	switch model {
	case dto.Lightweight:
		return dto.WeightLimitDrone / 4
	case dto.Middleweight:
		return dto.WeightLimitDrone / 3
	case dto.Cruiserweight:
		return dto.WeightLimitDrone / 2
	}

	return dto.WeightLimitDrone
}

// NotBlank is the validation function for validating if the current field
// has a value or length greater than zero, or is not a space only string.
// example: v.RegisterValidation("notblank", NotBlank)
func NotBlank(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return len(strings.TrimSpace(field.String())) > 0
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array:
		return field.Len() > 0
	case reflect.Ptr, reflect.Interface, reflect.Func:
		return !field.IsNil()
	default:
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}

func addTranslation(validate *validator.Validate, trans *ut.Translator, tag string, errMessage string) {
	registerFn := func(ut ut.Translator) error {
		return ut.Add(tag, errMessage, false)
	}

	transFn := func(ut ut.Translator, fe validator.FieldError) string {
		param := fe.Param()
		tag := fe.Tag()

		t, err := ut.T(tag, fe.Field(), param)
		if err != nil {
			return fe.(error).Error()
		}
		return t
	}

	_ = validate.RegisterTranslation(tag, *trans, registerFn, transFn)
}

func translateError(err error, trans ut.Translator) (errs []error) {
	if err == nil {
		return nil
	}
	validatorErrs := err.(validator.ValidationErrors)
	for _, e := range validatorErrs {
		translatedErr := fmt.Errorf(e.Translate(trans))
		errs = append(errs, translatedErr)
	}
	return errs
}
