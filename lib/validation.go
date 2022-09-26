package lib

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/schema/dto"
	reg "regexp"
)

// ValidateString validate a string given a regular expression
func ValidateString(data string, regexp string) bool {
	return reg.MustCompile(regexp).MatchString(data)
}

// ValidateStringCollection validate a string collection given a regular expression
func ValidateStringCollection(data []interface{}, regexp string) bool {
	var fn govalidator.ConditionIterator = func(value interface{}, index int) bool {
		fmt.Println(value.(string))
		return reg.MustCompile(regexp).MatchString(value.(string))
	}
	return govalidator.ValidateArray(data, fn)
}

// InitValidator Activate behavior to require all fields and adding new validators
func InitValidator() {
	govalidator.SetFieldsRequiredByDefault(false)

	// Add your own struct validation tags
	// validates medication name
	govalidator.TagMap["medication_name_validation"] = func(str string) bool {
		return reg.MustCompile(dto.RegexpMedicationName).MatchString(str)
	}

	// validates medication code
	govalidator.TagMap["medication_code_validation"] = func(str string) bool {
		return reg.MustCompile(dto.RegexpMedicationCode).MatchString(str)
	}

	// validates that an enum is within the interval
	govalidator.TagMap["drone_enum_validation"] = func(str string) bool {
		return str != "unknown"
	}
}

func ValidateSerialNumberDrone(serialNumber string) bool {
	return govalidator.MaxStringLength(serialNumber, dto.MaxSerialNumberLength)
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