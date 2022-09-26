package main

import (
	"encoding/base64"

	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/repo/db"

	"github.com/asaskevich/govalidator"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/lib"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/schema"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/schema/dto"

	"os"
	"testing"

	"github.com/kataras/iris/v12/httptest"
)

func TestNewApp(t *testing.T) {
	// set environment variable
	_ = os.Setenv(schema.EnvConfigPath, "./conf/conf.yaml")
	app, config := newApp()
	e := httptest.New(t, app)

	repo := db.NewRepoDrones(config)

	isPopulated := repo.IsPopulated()
	if !isPopulated {
		// populate database
		err := repo.PopulateDB()
		if err != nil {
			t.Errorf("error populating the database")
		}
	}
	// check server status
	e.GET("/status").Expect().Status(httptest.StatusOK)

	// without basic auth
	e.GET("/api/v1/drones").Expect().Status(httptest.StatusUnauthorized)
	e.GET("/api/v1/medications").Expect().Status(httptest.StatusUnauthorized)

	// with valid JWT auth
	cred := dto.UserCredIn{
		Username: "richard.sargon@meinermail.com",
		Password: "password1",
	}

	_ = e.POST("/api/v1/auth").WithJSON(cred).Expect().Status(httptest.StatusOK)

	// with invalid JWT auth
	cred = dto.UserCredIn{
		Username: "noexist@meinermail.com",
		Password: "fakepasswd",
	}

	_ = e.POST("/api/v1/auth").WithJSON(cred).Expect().Status(httptest.StatusUnauthorized)


	// drone valid
	droneValid := dto.Drone{
		SerialNumber:    lib.GenerateUUIDStr(),
		Model:           dto.Cruiserweight,
		WeightLimit:     lib.CalculateDroneWeightLimit(dto.Cruiserweight),
		BatteryCapacity: 45,
		State:           dto.IDLE,
	}
	// validate drone fields
	ok, _ := govalidator.ValidateStruct(droneValid)
	if !ok {
		t.Errorf("drone %s must be valid", droneValid.SerialNumber)
	}

	// drone invalid
	droneInvalid := dto.Drone{
		SerialNumber:    lib.GenerateUUIDStr(),
		Model:           dto.Cruiserweight,
		WeightLimit:     701,
		BatteryCapacity: 45,
		State:           dto.IDLE,
	}
	// validate drone fields
	ok, _ = govalidator.ValidateStruct(droneInvalid)
	if ok {
		t.Errorf("drone %s must be invalid, the weight limit is greater than 500gr", droneValid.SerialNumber)
	}

	// medication valid
	medicationValid := dto.Medication{
		Name:   gofakeit.Password(true, true, true, false, false, 12),
		Weight: 700,
		Code:   gofakeit.Password(false, true, true, false, false, 10),
		Image:  base64.StdEncoding.EncodeToString([]byte("fake_image")),
	}
	// validate medication fields
	ok, _ = govalidator.ValidateStruct(medicationValid)
	if !ok {
		t.Errorf("medication %s must be valid", medicationValid.Code)
	}
}