package main

import "errors"

var (
	IncorrectDoctorID   = errors.New("incorrect doctor id")
	DoctorHasNoSchedule = errors.New("doctor has no schedule")
	//IncorrectInputTime  = errors.New("incorrect input time")
	IncorrectInputLoginData  = errors.New("incorrect input login data")
	ShortPassword  = errors.New("the password is too short")
	EmptyData  = errors.New("the input data is empty")
)