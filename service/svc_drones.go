package service

import (
	"fmt"
	"restapi.app/lib"

	"github.com/kataras/iris/v12"
	"github.com/tidwall/buntdb"
	"restapi.app/repo/db"
	"restapi.app/schema"
	"restapi.app/schema/dto"
)

// region ======== SETUP =================================================================

// ISvcDrones Drones request service interface
type ISvcDrones interface {
	IsPopulateDBSvc() bool
	PopulateDBSvc() *dto.Problem

	// user functions

	GetUserSvc(id string, filter bool) (*dto.User, *dto.Problem)
	GetUsersSvc() (*[]dto.User, *dto.Problem)

	// drone functions

	GetADroneSvc(serialNumber string) (*dto.Drone, *dto.Problem)
	GetDronesSvc(filters ...string) (*[]dto.Drone, *dto.Problem)
	RegisterDroneSvc(drone *dto.Drone) *dto.Problem
	ExistDroneSvc(serialNumber string) (bool, *dto.Problem)

	// medication functions

	GetMedicationsSvc() (*[]dto.Medication, *dto.Problem)
	CheckingLoadedMedicationsItemsSvc(serialNumberDrone string) (*[]string, *dto.Problem)
	LoadMedicationItemsADroneSvc(serialNumberDrone string, medicationItemIDs []interface{}) *dto.Problem
}

type svcDronesReqs struct {
	reposDrones *db.RepoDrones
}

// endregion =============================================================================

// NewSvcDronesReqs instantiate the Drones request services
func NewSvcDronesReqs(reposDrones *db.RepoDrones) ISvcDrones {
	return &svcDronesReqs{reposDrones}
}

// region ======== METHODS ======================================================

func (s *svcDronesReqs) IsPopulateDBSvc() bool {
	return (*s.reposDrones).IsPopulated()
}

func (s *svcDronesReqs) PopulateDBSvc() *dto.Problem {
	err := (*s.reposDrones).PopulateDB()

	switch {
	case err == buntdb.ErrNotFound:
		return lib.NewProblem(iris.StatusPreconditionFailed, schema.ErrBuntdbItemNotFound, err.Error())
	case err != nil:
		if err.Error() == schema.ErrBuntdbPopulated {
			return lib.NewProblem(iris.StatusInternalServerError, schema.ErrBuntdbPopulated, "the database has already been populated")
		}
		return lib.NewProblem(iris.StatusExpectationFailed, schema.ErrBuntdb, err.Error())
	}
	return nil
}

func (s *svcDronesReqs) GetUserSvc(id string, filter bool) (*dto.User, *dto.Problem) {
	res, err := (*s.reposDrones).GetUser(id, filter)
	if err != nil {
		return nil, lib.NewProblem(iris.StatusExpectationFailed, schema.ErrBuntdb, err.Error())
	}
	return res, nil
}

func (s *svcDronesReqs) GetUsersSvc() (*[]dto.User, *dto.Problem) {
	res, err := (*s.reposDrones).GetUsers()
	if err != nil {
		return nil, lib.NewProblem(iris.StatusExpectationFailed, schema.ErrBuntdb, err.Error())
	}
	return res, nil
}

// GetADroneSvc get a specific drone
func (s *svcDronesReqs) GetADroneSvc(serialNumber string) (*dto.Drone, *dto.Problem) {
	res, err := (*s.reposDrones).GetDrone(serialNumber)
	// Getting non-existent values will cause an ErrNotFound error.
	if err == buntdb.ErrNotFound {
		return nil, lib.NewProblem(iris.StatusPreconditionFailed, schema.ErrBuntdbItemNotFound, err.Error())
	} else if err != nil {
		return nil, lib.NewProblem(iris.StatusExpectationFailed, schema.ErrBuntdb, err.Error())
	}

	return res, nil
}

func (s *svcDronesReqs) GetDronesSvc(filters ...string) (*[]dto.Drone, *dto.Problem) {
	var filter = ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	res, err := (*s.reposDrones).GetDrones(filter)
	if err != nil {
		return nil, lib.NewProblem(iris.StatusExpectationFailed, schema.ErrBuntdb, err.Error())
	}

	return res, nil
}

func (s *svcDronesReqs) RegisterDroneSvc(drone *dto.Drone) *dto.Problem {
	err := (*s.reposDrones).RegisterDrone(drone)
	if err != nil {
		return lib.NewProblem(iris.StatusExpectationFailed, schema.ErrBuntdb, err.Error())
	}
	return nil
}

func (s *svcDronesReqs) ExistDroneSvc(serialNumber string) (bool, *dto.Problem) {
	err := (*s.reposDrones).ExistDrone(serialNumber)
	// Getting non-existent values will cause an ErrNotFound error.
	if err == buntdb.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, lib.NewProblem(iris.StatusExpectationFailed, schema.ErrBuntdb, err.Error())
	}
	return true, nil
}

func (s *svcDronesReqs) GetMedicationsSvc() (*[]dto.Medication, *dto.Problem) {
	res, err := (*s.reposDrones).GetMedications()
	if err != nil {
		return nil, lib.NewProblem(iris.StatusExpectationFailed, schema.ErrBuntdb, err.Error())
	}
	return res, nil
}

func (s *svcDronesReqs) CheckingLoadedMedicationsItemsSvc(serialNumberDrone string) (*[]string, *dto.Problem) {
	// check that the drone exists in the database
	err := (*s.reposDrones).ExistDrone(serialNumberDrone)
	// Getting non-existent values will cause an ErrNotFound error.
	if err == buntdb.ErrNotFound {
		return nil, lib.NewProblem(iris.StatusPreconditionFailed, schema.ErrBuntdbItemNotFound, fmt.Sprintf("the drone with serial number %s does not exist", serialNumberDrone))
	} else if err != nil {
		return nil, lib.NewProblem(iris.StatusExpectationFailed, schema.ErrBuntdb, err.Error())
	}

	// if the drone exists, then we check if it has medication items associated with it
	res, err := (*s.reposDrones).CheckingLoadedMedicationsItems(serialNumberDrone)

	// Getting non-existent values will cause an ErrNotFound error.
	// if it throws the ErrNotFound error, it is that the drone is not loading medication items
	if err == buntdb.ErrNotFound {
		return &[]string{}, nil
	} else if err != nil {
		return nil, lib.NewProblem(iris.StatusExpectationFailed, schema.ErrBuntdb, err.Error())
	}
	return res, nil
}

func (s *svcDronesReqs) LoadMedicationItemsADroneSvc(serialNumberDrone string, medicationItemIDs []interface{}) *dto.Problem {
	// get drone if exist
	drone, errP := s.GetADroneSvc(serialNumberDrone)
	if errP != nil {
		return errP
	}

	// prevent the drone from being in LOADING state if the battery level is **below 25%**
	if drone.BatteryCapacity < 25.0 {
		return lib.NewProblem(iris.StatusPreconditionFailed, schema.ErrDroneVeryLowBatteryKey, schema.ErrDroneVeryLowBattery.Error())
	} else if drone.State != dto.IDLE {
		return lib.NewProblem(iris.StatusPreconditionFailed, schema.ErrDroneBusyKey, schema.ErrDroneBusy.Error())
	}

	err := (*s.reposDrones).LoadMedicationItemsADrone(drone, medicationItemIDs)
	if err == buntdb.ErrNotFound {
		return lib.NewProblem(iris.StatusPreconditionFailed, schema.ErrDroneMaximumLoadWeightExceededKey, err.Error())
	} else if err != nil {
		return lib.NewProblem(iris.StatusExpectationFailed, schema.ErrBuntdb, err.Error())
	}
	return nil
}
