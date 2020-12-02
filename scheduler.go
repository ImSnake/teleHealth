package main

import (
	"errors"
	"gitlab.com/group2prject_telehealth/scheduler_models"
	"sync"
)

type Schedules struct {
	sync.RWMutex
	list map[uint16]scheduler_models.Schedule
}

func (s *Schedules) GetSchedule(userID uint16) (*scheduler_models.Schedule, error) {
	s.Lock()
	defer s.Unlock()

	_, ok := s.list[userID]
	if ok {
		schedule := s.list[userID]
		return &schedule, nil
	}

	return nil, errors.New("the user has no schedule")
}

func (s *Schedules) AddSchedule(userID uint16, schedule *scheduler_models.Schedule) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.list[userID]
	if !ok {
		s.list[userID] = *schedule
		return nil
	}

	return errors.New("the user already has a schedule")
}
