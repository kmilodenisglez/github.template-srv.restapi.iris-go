package dto

type DroneState uint

const (
	IDLE DroneState = iota
	LOADING
	LOADED
	DELIVERING
	DELIVERED
	RETURNING
)

type DroneModel uint

const (
	Lightweight DroneModel = iota
	Middleweight
	Cruiserweight
	Heavyweight
)

func (droneState DroneState) String() string {
	names := []string{"IDLE", "LOADING", "LOADED", "DELIVERING", "DELIVERED", "RETURNING"}
	if droneState < IDLE || droneState > RETURNING {
		return "unknown"
	}
	return names[droneState]
}
func (droneModel DroneModel) String() string {
	names := []string{"Lightweight", "Middleweight", "Cruiserweight", "Heavyweight"}
	if droneModel < Lightweight || droneModel > Heavyweight {
		return "unknown"
	}
	return names[droneModel]
}

type ConfigDB struct {
	IsPopulated bool `json:"isPopulated"`
}

// RequestDrone model
// @Description drone model without the weightLimit, it is used for endpoint request
// @Description the weight limit is calculated from the drone's model (Lightweight, Middleweight, Cruiserweight, Heavyweight)
type RequestDrone struct {
	SerialNumber    string     `json:"serialNumber" valid:"maxstringlength(100)"`
	Model           DroneModel `json:"model" valid:"drone_enum_validation~unknown drone model"`
	BatteryCapacity float64    `json:"batteryCapacity" valid:"range(0|100)"`
	State           DroneState `json:"state" valid:"drone_enum_validation~unknown drone state"`
}

// Drone model
// @Description Drone item information
type Drone struct {
	SerialNumber    string     `json:"serialNumber" valid:"maxstringlength(100)"`
	Model           DroneModel `json:"model" valid:"drone_enum_validation~unknown drone model"`
	WeightLimit     float64    `json:"weightLimit" valid:"required~the weight limit is between 1 and 500 gr,range(1|500)~the weight limit is between 1 and 500 gr"`
	BatteryCapacity float64    `json:"batteryCapacity" valid:"range(0|100)"`
	State           DroneState `json:"state" valid:"drone_enum_validation~unknown drone state"`
}

// Medication model
// @Description Medication item information
type Medication struct {
	Name   string  `json:"name" valid:"medication_name_validation~invalid name (allowed only letters - numbers - ‘-‘ - ‘_’)"`
	Weight float64 `json:"weight"`
	Code   string  `json:"code" valid:"medication_code_validation~invalid code (allowed only upper case letters - underscore and numbers)"` // we assume that the code is unique
	Image  string  `json:"image" valid:"base64"`
}

const (
	RegexpMedicationName  = "^[a-zA-Z0-9_-]*$" // allowed only letters, numbers, ‘-‘, ‘_’
	RegexpMedicationCode  = "^[A-Z0-9_]*$"     // allowed only upper case letters, underscore and numbers
	MaxSerialNumberLength = "100"              // serial number (100 characters max)
	WeightLimitDrone      = 500                // weight limit (500gr max)
)

type DroneBatteryLevel struct {
	SerialNumber    string  `json:"serialNumber"`
	BatteryCapacity float64 `json:"batteryCapacity"`
}

type LogEvent struct {
	Created             string              `json:"created"`
	UUID                string              `json:"uuid"`
	DronesBatteryLevels []DroneBatteryLevel `json:"dronesBatteryLevels"`
}

type StatusMsg struct {
	OK bool `json:"ok"`
}