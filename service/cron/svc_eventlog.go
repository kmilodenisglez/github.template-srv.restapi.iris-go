package cron

import (
	"github.com/go-co-op/gocron"
	"log"
	"restapi.app/repo/db"
	"restapi.app/service/utils"
	"time"
)

// ISvcEventLog EventLog request service interface
type ISvcEventLog interface {
	MeinerCronJob() error
}

type svcEventLogReqs struct {
	svcConf     *utils.SvcConfig
	reposDrones *db.RepoDrones
}

// endregion =============================================================================

// NewSvcRepoEventLog instantiate the Drones request services
func NewSvcRepoEventLog(svcConf *utils.SvcConfig) ISvcEventLog {
	reposDrones := db.NewRepoDrones(svcConf)
	return &svcEventLogReqs{svcConf, &reposDrones}
}

// MeinerCronJob periodic task to check drones battery levels and create history/audit event log for this
func (e svcEventLogReqs) MeinerCronJob() error {
	// cron job is started only if it is active in configuration
	if e.svcConf.CronEnabled {
		log.Printf("schedules a new periodic Job with an interval: %d seconds", e.svcConf.EveryTime)
		cron := gocron.NewScheduler(time.UTC)

		_, err := cron.Every(e.svcConf.EveryTime).Seconds().WaitForSchedule().Do(e.doFunc)
		if err != nil {
			return err
		}
		// starts the scheduler asynchronously
		cron.StartAsync()
	}
	return nil
}

func (e svcEventLogReqs) doFunc() {
	log.Println("cron job executing")

	log.Println("cron job ending")
}
