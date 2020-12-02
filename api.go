package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/kaatinga/assets"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.com/group2prject_telehealth/backend/models"
	"gitlab.com/group2prject_telehealth/scheduler_models"
)

// === API HANDLERS ===

// GetDayHourList is an API Handler to get day`s busy hours
func GetDayHourList(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	dateFormat := "02Jan2006"

	day, err := time.Parse(dateFormat, ps.ByName("day"))
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	doctorID, ok := assets.StUint16(ps.ByName("id"))
	if !ok {
		hd.setError(http.StatusBadRequest, IncorrectDoctorID)
		return
	}

	// Получаем все записи
	var schedule *scheduler_models.Schedule
	schedule, err = schedules.GetSchedule(doctorID)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	// Отмечаем дни где есть запись
	var hours = make(map[string]bool)
	for gap, _ := range schedule.Gaps {
		if (gap.Time.Month() == day.Month()) && (gap.Time.Year() == day.Year()) && (gap.Time.Day() == day.Day()) {
			hours[strconv.Itoa(gap.Time.Hour())] = true
		}
	}

	hd.pageData.Data.CustomData = hours
}

// GetMonthDayList is an API Handler to get month's busy days
func GetMonthDayList(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	// Получаем год и месяц в формате RFC1123
	month, err := time.Parse("Jan2006", ps.ByName("month"))
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	doctorID, ok := assets.StUint16(ps.ByName("id"))
	if !ok {
		hd.setError(http.StatusBadRequest, IncorrectDoctorID)
		return
	}

	// Получаем все записи
	var schedule *scheduler_models.Schedule
	schedule, err = schedules.GetSchedule(doctorID)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	// Отмечаем дни где есть запись
	var days = make(map[string]bool)
	for gap, _ := range schedule.Gaps {
		if (gap.Time.Month() == month.Month()) && (gap.Time.Year() == month.Year()) {
			days[strconv.Itoa(gap.Time.Day())] = true
		}
	}

	hd.pageData.Data.CustomData = days
}

// Deprecated: GetDayCalendar is an API Handler to get day schedule
func GetDayCalendar(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	dateFormat := "02Jan2006"

	day := ps.ByName("day")
	_, err := time.Parse(dateFormat, day)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	doctorID, ok := assets.StUint16(ps.ByName("id"))
	if !ok {
		hd.setError(http.StatusBadRequest, IncorrectDoctorID)
		return
	}

	// Получаем все записи
	var schedule *scheduler_models.Schedule
	schedule, err = schedules.GetSchedule(doctorID)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	response := scheduler_models.NewSchedule(doctorID)

	// Убираем из списка те записи, что не относятся к запрошенному дню
	for gap, schedule := range schedule.Gaps {
		if gap.Time.Format(dateFormat) == day {
			response.Gaps[gap] = schedule
		}
	}

	hd.pageData.Data.CustomData = &response
}

// GetHourCalendar is an API Handler to get hour schedule
func GetHourCalendar(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	dateFormat := "02Jan2006T15"

	day := ps.ByName("date")
	_, err := time.Parse(dateFormat, day)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	doctorID, ok := assets.StUint16(ps.ByName("id"))
	if !ok {
		hd.setError(http.StatusBadRequest, IncorrectDoctorID)
		return
	}

	// Получаем все записи
	var schedule *scheduler_models.Schedule
	schedule, err = schedules.GetSchedule(doctorID)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	response := scheduler_models.NewSchedule(doctorID)

	// Убираем из списка те записи, что не относятся к запрошенному дню
	for gap, schedule := range schedule.Gaps {
		if gap.Time.Format(dateFormat) == day {
			response.Gaps[gap] = schedule
		}
	}

	hd.pageData.Data.CustomData = &response
}

// Specs is a handler that returns the specialization list
func Specs(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	ctx := context.Background()

	allSpecs, err := models.Specializations().All(ctx, hd.db)
	if err != nil {
		hd.setError(http.StatusInternalServerError, err)
		return
	}

	hd.pageData.Data.CustomData = allSpecs
}

//SpecsExists - returns the existing specializations
func SpecsExists(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	var existSpecs []int
	var contentSpecs []int
	var slice []models.Specialization

	ctx := context.Background()

	allSpecs, err := models.SpecializationKits().All(ctx, hd.db)
	if err != nil {
		hd.setError(http.StatusNotFound, err)
		return
	}

	for _, spec := range allSpecs {
		contentSpecs = append(contentSpecs, spec.SpecializationID.Int)
	}

	existSpecs = unique(contentSpecs)
	sort.Ints(existSpecs)

	for _, spec := range existSpecs {
		row, err := models.Specializations(qm.WhereIn("ID=?", spec)).One(ctx, hd.db)
		if err != nil {
			hd.setError(http.StatusNotFound, err)
			return
		}
		slice = append(slice, *row)

	}

	hd.pageData.Data.CustomData = slice
}

// убирает повторяющиеся значения
func unique(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// AllSpecialists is a handler that returns doctor list by specialization
func AllSpecialists(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	ctx := context.Background()

	type catalogDoctor struct {
		ID          int
		Surname     string
		Name        string
		Patronymic  string
		Experience  int
		Rating      float32
		CountReview int64
		PhotoURL    string
	}
	var slice []catalogDoctor

	idSpecs := ps.ByName("id")

	allSpec, err := models.SpecializationKits(qm.Where("specialization_id=?", idSpecs)).All(ctx, hd.db)
	if err != nil {
		hd.setError(http.StatusNotFound, err)
		return
	}

	for _, one := range allSpec {
		doctor, err := models.Doctors(qm.Where("ID=?", one.DoctorID.Int)).One(ctx, hd.db)
		if err != nil {
			hd.setError(http.StatusNotFound, err)
			return
		}

		reviewsCount, reviewsAverage := ReviewsInfo(*doctor, *hd)

		oneCatalogDoctor := catalogDoctor{
			ID:          one.DoctorID.Int,
			Surname:     doctor.Surname.String,
			Name:        doctor.Name.String,
			Patronymic:  doctor.Patronymic.String,
			Experience:  one.Experience.Int,
			Rating:      reviewsAverage,
			CountReview: reviewsCount,
			PhotoURL:    doctor.PhotoURL.String,
		}
		slice = append(slice, oneCatalogDoctor)
		sort.Slice(slice, func(i, j int) (less bool) {
			return slice[i].Rating > slice[j].Rating
		})
	}
	hd.pageData.Data.CustomData = slice
}

// GetReviews returns reviews by doctor ID. The ID is retrieved from URI
func GetReviews(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	type fullReview struct {
		CreatedAt time.Time
		Email     string
		Name      string
		Text      string
		Rating    float32
	}

	ctx := context.Background()

	var slice []fullReview
	idDoctor := ps.ByName("id")

	reviewsDoctor, err := models.Reviews(qm.Where("doctor_id=?", idDoctor)).All(ctx, hd.db)
	if err != nil {
		hd.setError(http.StatusNotFound, err)
		return
	}

	for _, review := range reviewsDoctor {
		patient, err := models.Patients(qm.Where("ID=?", review.PatientID)).One(ctx, hd.db)
		if err != nil {
			hd.setError(http.StatusNotFound, err)
			return
		}
		oneReview := fullReview{
			CreatedAt: review.CreatedAt.Time,
			Email:     patient.Email.String,
			Name:      patient.Name.String,
			Text:      review.Text.String,
			Rating:    review.Rating.Float32,
		}

		slice = append(slice, oneReview)
	}

	hd.pageData.Data.CustomData = slice
}

// Schedule handlers

// Schedule хэндлер Возвращает расписание доктора
func Schedule(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	doctorID, ok := assets.StUint16(ps.ByName("id"))
	if !ok {
		hd.setError(http.StatusBadRequest, IncorrectDoctorID)
		return
	}

	schedule, err := schedules.GetSchedule(doctorID)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	hd.pageData.Data.CustomData = *schedule
}

// NewSchedule is a handler that creates new schedule
func NewSchedule(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	hd.CreateNewSchedule()
}

func (hd *handlerData) CreateNewSchedule() {
	newSchedule := scheduler_models.NewSchedule(hd.pageData.User.ID)

	_, ok := schedules.list[hd.pageData.User.ID]
	if ok {
		return
	}

	err := schedules.AddSchedule(hd.pageData.User.ID, newSchedule)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}
}

// MakeAvailable is a handler that makes a time gap available
func MakeAvailable(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	err := schedules.list[hd.pageData.User.ID].MakeAvailable(hd.Gap)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}
}

// MakeAvailable is a handler that makes a time gap available
func MakeAvailableByHour(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	if hd.pageData.User.ID == 0 {
		hd.setError(http.StatusBadRequest, IncorrectDoctorID)
		return
	}

	schedule, ok := schedules.list[hd.pageData.User.ID]
	if !ok {
		hd.setError(http.StatusBadRequest, DoctorHasNoSchedule)
		return
	}

	var hours []byte
	var err error
	dateFormat := "02Jan2006"

	hd.Gap.Time, err = time.Parse(dateFormat, ps.ByName("day"))
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&hours)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	if len(hours) == 0 {
		hd.setError(http.StatusBadRequest, EmptyData)
		return
	}

	var gap scheduler_models.Gap
	for _, value := range hours {

		gap.Time = hd.Gap.Time.Add(time.Duration(value) * time.Hour)

		err = addRange(gap.Time, gap.Time.Add(1*time.Hour), &schedule)
		if err != nil {
			hd.logger.SubMsg.Info().Time("gap", gap.Time).Msg("error on add a gap")
			return
		}
	}
}

// Enrol is a handler that enrols a patient in a time gap
func Enrol(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	// проверка что нет запись на самого себя
	if hd.DoctorID == hd.pageData.User.ID {
		hd.setError(http.StatusBadRequest, errors.New("impossible to enrol to yourself"))
		return
	}

	err := schedules.list[hd.DoctorID].Enrol(hd.Gap, hd.pageData.User.ID)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}
}

func MakeUnavailable(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	hd.logger.SubMsg.Info().Time("time", hd.Gap.Time).Msg("Received Gap")

	err := schedules.list[hd.pageData.User.ID].MakeUnavailable(hd.Gap)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	//fmt.Println(schedules.list[hd.pageData.User.ID].Gaps)
}

func GetGapTime() Adapter {
	return func(next httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, actions httprouter.Params) {
			hd := r.Context().Value("hd").(*handlerData)
			hd.JSON = true // устанавливаем JSON Render

			if hd.pageData.User.ID == 0 {
				hd.setError(http.StatusBadRequest, IncorrectDoctorID)
				return
			}

			err := hd.Gap.SetTime(r.PostFormValue("time"))
			if err != nil {
				hd.setError(http.StatusBadRequest, err)
				return
			}

			// мы вычитываем doctor id и если оно вычитывается, то мы работаем с ID доктора,
			// иначе с ID аутентифицированного пользователя
			var ok bool
			hd.DoctorID, hd.DoctorIDIsReceived = assets.StUint16(r.PostFormValue("doctor"))
			if !hd.DoctorIDIsReceived {
				_, ok = schedules.list[hd.pageData.User.ID]
			} else {
				if hd.DoctorID == 0 {
					hd.setError(http.StatusBadRequest, IncorrectDoctorID)
					return
				}
				_, ok = schedules.list[hd.DoctorID]
			}

			// проверяем наличие расписания, иначе словим null pointer
			if !ok {
				hd.setError(http.StatusNotFound, DoctorHasNoSchedule)
				return
			}

			ctx := context.WithValue(r.Context(), "hd", hd)
			next(w, r.WithContext(ctx), actions)
		}
	}
}

func Cancel(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	var id uint16
	if hd.DoctorIDIsReceived {
		id = hd.pageData.User.ID
	} else {
		id = hd.DoctorID
	}

	err := schedules.list[id].Cancel(hd.Gap)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}
}

// DoctorJSON is the doctor handler
func DoctorJSON(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render
	ctx := context.Background()

	type DoctorProfile struct {
		DoctorData      models.Doctor
		Age             int
		Rating          float32
		ReviewsData     models.ReviewSlice
		Specializations models.SpecializationKitSlice
	}

	doctorID := ps.ByName("id")

	doctor, err := models.Doctors(qm.Where("id=?", doctorID)).One(ctx, hd.db)
	if err != nil {
		hd.setError(http.StatusNotFound, err)
		return
	}

	reviews, err := models.Reviews(qm.Where("doctor_id=?", doctor.ID)).All(ctx, hd.db)
	if err != nil {
		hd.setError(http.StatusNotFound, err)
		return
	}

	specs, err := models.SpecializationKits(qm.Where("doctor_id=?", doctor.ID)).All(ctx, hd.db)
	if err != nil {
		hd.setError(http.StatusNotFound, err)
		return
	}

	_, rating := ReviewsInfo(*doctor, *hd)

	data := DoctorProfile{
		DoctorData:      *doctor,
		Age:             AgeUser(doctor.Birthdate.Time),
		Rating:          rating,
		ReviewsData:     reviews,
		Specializations: specs,
	}

	hd.pageData.Data.CustomData = data
}

// DoctorSpecs is a handler to get all specializations of a doctor
func DoctorSpecs(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	doctorID := ps.ByName("id")

	doctorData, err := models.SpecializationKits(qm.Where("doctor_id=?", doctorID)).One(r.Context(), hd.db)
	if err != nil {
		hd.setError(http.StatusNotFound, err)
		return
	}

	hd.pageData.Data.CustomData = doctorData
}

//AddSpecDoctor - this handler inserts a new line in DB. Specialization Kit
func AddSpecDoctor(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	req, err := ioutil.ReadAll(r.Body)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	docNewSpec := models.SpecializationKit{}

	if err = json.Unmarshal(req, &docNewSpec); err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	existingSpecs, err := models.SpecializationKits(qm.And("doctor_id=? AND specialization_id=?", docNewSpec.DoctorID.Int, docNewSpec.SpecializationID.Int)).Exists(r.Context(), hd.db)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		fmt.Println("******11111***********")
		return
	}

	if !existingSpecs {
		if err = docNewSpec.Insert(r.Context(), hd.db, boil.Whitelist(models.SpecializationKitColumns.DoctorID,
			models.SpecializationKitColumns.SpecializationID,
			models.SpecializationKitColumns.CertificateN,
			models.SpecializationKitColumns.Experience)); err != nil {
			hd.setError(http.StatusBadRequest, err)
			return
		}
	} else {
		hd.setError(http.StatusBadRequest, errors.New("this specialization exist"))
		return
	}
}

//AddPhotoProfile this function adds a photo to profile
func AddPhotoProfile(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	r.ParseMultipartForm(0)
	src, hdr, err := r.FormFile("my-file")
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}
	docID := r.FormValue("doctor_id")
	defer r.Body.Close()
	ID, _ := strconv.Atoi(docID)

	hdr.Filename = time.Now().Format("2006-01-02_15-04-05") + "-" + hdr.Filename

	dst, err := os.Create(filepath.Join(".\\ui\\static\\images\\avatars", hdr.Filename))
	if err != nil {
		hd.logger.Title.Error().Err(err).Msg("Can't created a new file")
		return
	}

	addURL, _ := models.FindDoctor(r.Context(), hd.db, uint(ID))

	oldPhoto := addURL.PhotoURL.String
	if oldPhoto != "/static/images/avatars/avatar.png" {
		if err := os.Remove(filepath.Join(".\\ui", filepath.FromSlash(oldPhoto))); err != nil {
			hd.logger.Title.Error().Err(err).Msg("Can't remove file")
			return
		}
	}

	addURL.PhotoURL.String = "/static/images/avatars/" + hdr.Filename

	_, err = addURL.Update(r.Context(), hd.db, boil.Infer())
	if err != nil {
		hd.logger.Title.Error().Err(err).Msg("Can't update database")
		return
	}

	defer dst.Close()

	io.Copy(dst, src)
}

//GetAvatar returns the path to the avatar file from doctor_id
func GetAvatar(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	doctorID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	urlPhoto, err := models.FindDoctor(r.Context(), hd.db, uint(doctorID), "photo_url")
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	hd.pageData.Data.CustomData = urlPhoto.PhotoURL.String
}

//UpdateBiography updates doctor profile biography
func UpdateBiography(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	doctorID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	newBiographyBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	var newText models.Doctor
	newText.ID = uint(doctorID)

	if err := json.Unmarshal(newBiographyBody, &newText); err != nil {
		hd.setError(http.StatusNoContent, err)
		return
	}

	_, err = newText.Update(r.Context(), hd.db, boil.Whitelist("biography"))
	if err != nil {
		hd.setError(http.StatusNotImplemented, err)
		return
	}
}
