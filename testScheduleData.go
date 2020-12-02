package main

import (
	"gitlab.com/group2prject_telehealth/scheduler_models"
	"time"
)

func addRange(currentTime, endTime time.Time, schedule *scheduler_models.Schedule) (err error) {
	for {
		err = schedule.MakeAvailable(scheduler_models.Gap{Time: currentTime})
		if err != nil {
			return
		}

		currentTime = currentTime.Add(15 * time.Minute)

		if endTime == currentTime {
			break
		}
	}

	return
}

type timeGap struct {
	startDateTime time.Time
	endDateTime   time.Time
}

func initSchedule() error {
	var (
		tempSchedule *scheduler_models.Schedule
		err          error
		doctorID     uint16 = 26
	)

	tempSchedule = &scheduler_models.Schedule{
		Activated: true,
		Gaps:      make(map[scheduler_models.Gap]*scheduler_models.Enrolment),
		UserID:    doctorID,
	}

	var timeGaps = []timeGap{
		{
			startDateTime: time.Date(2020, 12, 15, 7, 0, 0, 0, time.UTC),
			endDateTime:   time.Date(2020, 12, 15, 21, 0, 0, 0, time.UTC),
		},
		{
			startDateTime: time.Date(2020, 12, 16, 8, 0, 0, 0, time.UTC),
			endDateTime:   time.Date(2020, 12, 16, 12, 0, 0, 0, time.UTC),
		},
		{
			startDateTime: time.Date(2020, 12, 17, 12, 15, 0, 0, time.UTC),
			endDateTime:   time.Date(2020, 12, 17, 12, 30, 0, 0, time.UTC),
		},
		{
			startDateTime: time.Date(2020, 12, 17, 12, 45, 0, 0, time.UTC),
			endDateTime:   time.Date(2020, 12, 17, 13, 0, 0, 0, time.UTC),
		},
		{
			startDateTime: time.Date(2020, 12, 17, 20, 45, 0, 0, time.UTC),
			endDateTime:   time.Date(2020, 12, 17, 21, 0, 0, 0, time.UTC),
		},
	}

	for _, value := range timeGaps {
		err = addRange(
			value.startDateTime,
			value.endDateTime,
			tempSchedule,
		)
		if err != nil {
			return err
		}
	}

	err = schedules.AddSchedule(doctorID, tempSchedule)
	return nil
}
