package db

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/brianvoe/gofakeit/v6"
	jsoniter "github.com/json-iterator/go"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/lib"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/schema"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/schema/dto"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/service/utils"
	"github.com/tidwall/buntdb"
	"log"
	"strconv"
	"strings"
)

// region ======== SETUP =================================================================

type RepoDrones interface {
	IsPopulated() bool
	PopulateDB() error

	GetUser(field string, filterOptional ...bool) (*dto.User, error)
	GetUsers() (*[]dto.User, error)

	GetDrone(serialNumber string) (*dto.Drone, error)
	GetDrones(filter string) (*[]dto.Drone, error)
	RegisterDrone(drone *dto.Drone) error
	CheckingLoadedMedicationsItems(serialNumber string) (*[]string, error)
	LoadMedicationItemsADrone(drone *dto.Drone, medicationItemIDs []interface{}) error
	ExistDrone(serialNumber string) error

	GetMedications() (*[]dto.Medication, error)
}

type repoDrones struct {
	DBUserLocation string
}

// endregion =============================================================================

func NewRepoDrones(svcConf *utils.SvcConfig) RepoDrones {
	return &repoDrones{DBUserLocation: svcConf.StoreDBPath}
}

// region ======== METHODS ===============================================================

func (r *repoDrones) IsPopulated() bool {
	db, err := r.loadDB()
	if err != nil {
		return false
	}
	defer db.Close()
	return isPopulated(db)
}


// PopulateDB Populate the database with the initial information only if "IsPopulated" is
// false or does not exist
//nolint:gocognit
func (r *repoDrones) PopulateDB() error {
	db, err := r.loadDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// If it is already populated, the execution of the function stops
	if isPopulated(db) {return errors.New(schema.ErrBuntdbPopulated)}

	var fakeUsersList = fakeUsers()
	var fakeDronesList = fakeDrones()
	var fakeMedicationsList = fakeMedications()

	log.Println("writing users in database")
	err = db.Update(func(tx *buntdb.Tx) error {
		for i := 0; i < len(fakeUsersList); i++ {
			res, err := jsoniter.MarshalToString(fakeUsersList[i])
			log.Printf("user #%d: %s", i, res)
			if err != nil {
				return err
			}
			_, _, err = tx.Set(strconv.Itoa(i), res, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	log.Println("successfully added users")

	log.Println("writing drones in database")
	err = db.Update(func(tx *buntdb.Tx) error {
		for i := 0; i < len(fakeDronesList); i++ {
			res, err := jsoniter.MarshalToString(fakeDronesList[i])
			if err != nil {
				return err
			}
			// add drone value with "serialnumber" key
			_, _, err = tx.Set("drone:"+fakeDronesList[i].SerialNumber, res, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	log.Println("successfully added drones")

	log.Println("writing medications in database")
	err = db.Update(func(tx *buntdb.Tx) error {
		for i := 0; i < len(fakeMedicationsList); i++ {
			res, err := jsoniter.MarshalToString(fakeMedicationsList[i])
			if err != nil {
				return err
			}
			// add drone value with "code" key
			_, _, err = tx.Set("med:"+fakeMedicationsList[i].Code, res, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	log.Println("successfully added medications")

	// set IsPopulated to true
	err = db.Update(func(tx *buntdb.Tx) error {
		res, err := jsoniter.MarshalToString(dto.ConfigDB{IsPopulated: true})
		if err != nil {
			return err
		}
		_, _, err = tx.Set("config", res, nil)
		return err
	})
	if err != nil {
		return err
	}
	log.Println("'IsPopulated' has been set to true")

	return nil
}

// GetUser get the user from the DB
func (r *repoDrones) GetUser(field string, filterOptional ...bool) (*dto.User, error) {
	filter := false
	if len(filterOptional) > 0 {
		filter = filterOptional[0]
	}
	user := dto.User{}

	// Open the data.db file. It will be created if it doesn't exist.
	db, err := buntdb.Open(r.DBUserLocation)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.CreateIndex("username", "*", buntdb.IndexString)
	if err != nil {
		return nil, err
	}
	err = db.View(func(tx *buntdb.Tx) error {
		if filter {
			err := tx.Ascend("username", func(key, value string) bool {
				if strings.Contains(value, field) {
					err := jsoniter.UnmarshalFromString(value, &user)
					if err != nil {
						return false
					}
					return false
				}

				return true
			})
			return err
		}
		// filter = false
		//value, err := tx.Get(field)
		//if err != nil {
		//	return err
		//}
		//err = jsoniter.UnmarshalFromString(value, &user)
		//if err != nil {
		//	return err
		//}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUsers return a list of dto.User
func (r *repoDrones) GetUsers() (*[]dto.User, error) {
	// Open the data.db file. It will be created if it doesn't exist.
	db, err := buntdb.Open(r.DBUserLocation)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	user := dto.User{}
	var list []dto.User

	err = db.CreateIndex("username", "*", buntdb.IndexString)
	if err != nil {
		return nil, err
	}
	err = db.View(func(tx *buntdb.Tx) error {
		tx.Ascend("username", func(key, value string) bool {
			err = jsoniter.UnmarshalFromString(value, &user)
			if err == nil {
				list = append(list, user)
			}
			return err == nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &list, nil
}

// region ======== Drones ======================================================

// GetDrone get a specific drone
func (r *repoDrones) GetDrone(serialNumber string) (*dto.Drone, error) {
	db, err := r.loadDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	drone := dto.Drone{}

	db.CreateIndex("drone_state", "drone:*", buntdb.IndexJSON("batteryCapacity"))
	err = db.View(func(tx *buntdb.Tx) error {
		value, err := tx.Get("drone:"+serialNumber)
		if err != nil{
			return err
		}
		err = jsoniter.UnmarshalFromString(value, &drone)
		if err != nil{
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &drone, nil
}

// GetDrones A read-only transaction, return drones in db
// allows filtering by a specific string field
func (r *repoDrones) GetDrones(filter string) (*[]dto.Drone, error) {
	db, err := r.loadDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	drone := dto.Drone{}
	dronesList := make([]dto.Drone, 0)
	// custom index: sort drones descending by battery capacity
	db.CreateIndex("drone_state", "drone:*", buntdb.IndexJSON("batteryCapacity"))
	err = db.View(func(tx *buntdb.Tx) error {
		if filter != "" {
			err := tx.Descend("drone_state", func(key, value string) bool {
				if strings.Contains(value, filter) {
					err = jsoniter.UnmarshalFromString(value, &drone)
					if err == nil {
						dronesList = append(dronesList, drone)
					}
					return err == nil
				}
				return true
			})
			return err
		}
		err := tx.Descend("drone_state", func(key, value string) bool {
			err = jsoniter.UnmarshalFromString(value, &drone)
			if err == nil {
				dronesList = append(dronesList, drone)
			}
			return err == nil
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return &dronesList, nil
}

func (r *repoDrones) RegisterDrone(drone *dto.Drone) error {
	db, err := r.loadDB()
	if err != nil {
		return err
	}
	defer db.Close()

	log.Printf("writing the drone '%s' in database", drone.SerialNumber)
	err = db.Update(func(tx *buntdb.Tx) error {
		res, err := jsoniter.MarshalToString(drone)
		if err != nil {
			return err
		}
		_, _, err = tx.Set("drone:"+drone.SerialNumber, res, nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	log.Println("successfully added drone")
	return  nil
}

// CheckingLoadedMedicationsItems checking loaded medication items for a given drone
func (r *repoDrones) CheckingLoadedMedicationsItems(serialNumber string) (*[]string, error) {
	db, err := r.loadDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// medications id slice loaded by the drone
	loadedMeds := make([]string, 0)

	db.CreateIndex("loaded_medications", "loaded_medications:*", buntdb.IndexString)
	err = db.View(func(tx *buntdb.Tx) error {
		value, err := tx.Get("loaded_medications:"+serialNumber)
		if err != nil{
			return err
		}
		err = jsoniter.UnmarshalFromString(value, &loadedMeds)
		if err != nil{
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &loadedMeds, nil
}

func (r *repoDrones) LoadMedicationItemsADrone(drone *dto.Drone, medicationItemIDs []interface{}) error {
	db, err := r.loadDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// begin: validating medication item IDs
	medication := dto.Medication{}
	medicationIdsRealMap := make(map[string]float64)

	db.CreateIndex("medication_id", "med:*", buntdb.IndexJSON("weight"))
	err = db.View(func(tx *buntdb.Tx) error {
		err := tx.Descend("medication_id", func(key, value string) bool {
			err = jsoniter.UnmarshalFromString(value, &medication)
			if err == nil {
				medicationIdsRealMap[medication.Code] = medication.Weight
			}
			return err == nil
		})
		return err
	})
	if err != nil {
		return err
	}

	// to guarantee non-repeated id
	medicationItemIDs = lib.Unique(medicationItemIDs)

	// compares the request IDs (medicationItemIDs) with the collection obtained from the database (medicationIdsRealMap)
	// also returns the total weight
	packedTotalWeight, allIDValid := thereAreAll(medicationIdsRealMap, medicationItemIDs)
	if !allIDValid {
		return fmt.Errorf("at least one of the medication items does not exist")
	}

	// prevent the drone from being loaded with more weight that it can carry
	if packedTotalWeight > drone.WeightLimit {
		return schema.ErrDroneMaximumLoadWeightExceeded
	}

	// end: validating medication item IDs

	log.Printf("loading a drone '%s' with medication items: %s", drone.SerialNumber, medicationItemIDs)
	err = db.Update(func(tx *buntdb.Tx) error {
		res, err := jsoniter.MarshalToString(medicationItemIDs)
		if err != nil {
			return err
		}
		_, _, err = tx.Set("loaded_medications:"+drone.SerialNumber, res, nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	log.Println("successfully loaded medication items")

	return  nil
}

func (r *repoDrones) ExistDrone(serialNumber string) error {
	db, err := r.loadDB()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get("drone:"+serialNumber)
		if err != nil {
			return err
		}
		return nil
	})
	// Getting non-existent values will cause an ErrNotFound error.
	if err != nil {
		return err
	}

	return nil
}

// endregion ======== Drones ======================================================

// region ======== Medications ======================================================

func (r *repoDrones) GetMedications() (*[]dto.Medication, error) {
	db, err := r.loadDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	medication := dto.Medication{}
	medicationsList := make([]dto.Medication, 0)
	// custom index: sort medications descending by weight
	db.CreateIndex("medication_state", "med:*", buntdb.IndexJSON("weight"))
	err = db.View(func(tx *buntdb.Tx) error {
		err := tx.Descend("medication_state", func(key, value string) bool {
			err = jsoniter.UnmarshalFromString(value, &medication)
			if err == nil {
				medicationsList = append(medicationsList, medication)
			}
			return err == nil
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return &medicationsList, nil
}


// endregion ======== Medications ======================================================

// region ======== PRIVATE AUX ===========================================================
func (r *repoDrones) loadDB() (*buntdb.DB, error) {
	log.Println("Load DB ", r.DBUserLocation)
	// Open the data.db file. It will be created if it doesn't exist.
	db, err := buntdb.Open(r.DBUserLocation)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return db, nil
}

func isPopulated(db *buntdb.DB) bool {
	log.Println("checking if StoreDB has already been populated")
	configDB := dto.ConfigDB{}
	db.CreateIndex("config", "config", buntdb.IndexString)
	err := db.View(func(tx *buntdb.Tx) error {
		value, err := tx.Get("config")
		if err != nil{
			return err
		}
		err = jsoniter.UnmarshalFromString(value, &configDB)
		if err != nil{
			log.Println("Unmarshal Error in IsPopulated func: ", err)
			return err
		}
		return nil
	})
	if err == buntdb.ErrNotFound {
		log.Println("Database not found")
		return false
	} else if err != nil {
		panic(err)
	}

	return configDB.IsPopulated
}

func fakeUsers() []dto.User {
	var users = []dto.User{{
		Passphrase: "0b14d501a594442a01c6859541bcb3e8164d183d32937b851835442f69d5c94e", // password1
		Username:   "richard.sargon@meinermail.com",
		Name:       "Richard Sargon",
	}, {
		Passphrase: "6cf615d5bcaac778352a8f1f3360d23f02f34ec182e259897fd6ce485d7870d4", // password2
		Username:   "tom.carter@meinermail.com",
		Name:       "Tom Carter",
	}}
	return users
}

func fakeDrones() []dto.Drone {
	uuid := "123e4567-e89b-12d3-a456-4266141740"
	var drones = []dto.Drone{{
		SerialNumber:    uuid+"01",
		Model:           dto.Cruiserweight,
		WeightLimit:     lib.CalculateDroneWeightLimit(dto.Cruiserweight),
		BatteryCapacity: 45,
		State:           dto.IDLE,
	}, {
		SerialNumber:    uuid+"02",
		Model:           dto.Middleweight,
		WeightLimit:     lib.CalculateDroneWeightLimit(dto.Middleweight),
		BatteryCapacity: 56.4,
		State:           dto.DELIVERED,
	}, {
		SerialNumber:    uuid+"03",
		Model:           dto.Heavyweight,
		WeightLimit:     lib.CalculateDroneWeightLimit(dto.Heavyweight),
		BatteryCapacity: 99.2,
		State:           dto.LOADING,
	}, {
		SerialNumber:    uuid+"04",
		Model:           dto.Middleweight,
		WeightLimit:     lib.CalculateDroneWeightLimit(dto.Middleweight),
		BatteryCapacity: 35.6,
		State:           dto.RETURNING,
	}, {
		SerialNumber:    uuid+"05",
		Model:           dto.Heavyweight,
		WeightLimit:     lib.CalculateDroneWeightLimit(dto.Heavyweight),
		BatteryCapacity: 52.9,
		State:           dto.DELIVERING,
	}, {
		SerialNumber:    uuid+"06",
		Model:           dto.Lightweight,
		WeightLimit:     lib.CalculateDroneWeightLimit(dto.Lightweight),
		BatteryCapacity: 12.9,
		State:           dto.IDLE,
	}, {
		SerialNumber:    uuid+"07",
		Model:           dto.Cruiserweight,
		WeightLimit:     lib.CalculateDroneWeightLimit(dto.Cruiserweight),
		BatteryCapacity: 91.3,
		State:           dto.LOADED,
	}, {
		SerialNumber:    uuid+"08",
		Model:           dto.Heavyweight,
		WeightLimit:     lib.CalculateDroneWeightLimit(dto.Heavyweight),
		BatteryCapacity: 73.6,
		State:           dto.LOADED,
	},{
		SerialNumber:    uuid+"09",
		Model:           dto.Lightweight,
		WeightLimit:     lib.CalculateDroneWeightLimit(dto.Lightweight),
		BatteryCapacity: 25,
		State:           dto.IDLE,
	},{
		SerialNumber:    uuid+"10",
		Model:           dto.Lightweight,
		WeightLimit:     lib.CalculateDroneWeightLimit(dto.Lightweight),
		BatteryCapacity: 25,
		State:           dto.IDLE,
	}}
	return drones
}

func fakeMedications() []dto.Medication {
	var medications = []dto.Medication{{
		Name:   gofakeit.Password(true, true, true, false, false, 12),
		Weight: 700,
		Code:   gofakeit.Password(false, true, true, false, false, 10),
		Image:  base64.StdEncoding.EncodeToString([]byte("fake_image")),
	}, {
		Name:   lib.NormalizeString(gofakeit.Company(), true),
		Weight: 210,
		Code:   gofakeit.Password(false, true, true, false, false, 10),
		Image:  base64.StdEncoding.EncodeToString([]byte("fake_image")),
	}, {
		Name:   lib.NormalizeString(gofakeit.Company(), true),
		Weight: 34,
		Code:   gofakeit.Password(false, true, true, false, false, 10),
		Image:  base64.StdEncoding.EncodeToString([]byte("fake_image")),
	}, {
		Name:   lib.NormalizeString(gofakeit.Company(), true),
		Weight: 115,
		Code:   gofakeit.Password(false, true, true, false, false, 10),
		Image:  base64.StdEncoding.EncodeToString([]byte("fake_image")),
	}, {
		Name:   lib.NormalizeString(gofakeit.Company(), true),
		Weight: 490,
		Code:   gofakeit.Password(false, true, true, false, false, 10),
		Image:  base64.StdEncoding.EncodeToString([]byte("fake_image")),
	}, {
		Name:   lib.NormalizeString(gofakeit.Company(), true),
		Weight: 226,
		Code:   gofakeit.Password(false, true, true, false, false, 10),
		Image:  base64.StdEncoding.EncodeToString([]byte("fake_image")),
	}, {
		Name:   lib.NormalizeString(gofakeit.Company(), true),
		Weight: 397,
		Code:   gofakeit.Password(false, true, true, false, false, 10),
		Image:  base64.StdEncoding.EncodeToString([]byte("fake_image")),
	}}
	return medications
}

// thereAreAll compares the request IDs (medicationItemIDs) with the collection
// obtained from the database (medicationIdsRealMap)
// if they all exist then it also returns the total weight
func thereAreAll(medicationIdsRealMap map[string]float64, medicationItemIDs []interface{}) (float64, bool) {
	var totalWeight = 0.0
	for _, v := range medicationItemIDs {
		_, exists := medicationIdsRealMap[v.(string)]
		if !exists {
			return 0.0, false
		} else {
			totalWeight += medicationIdsRealMap[v.(string)]
		}
	}
	return totalWeight, true
}

// endregion =============================================================================
