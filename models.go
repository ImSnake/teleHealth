package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"gitlab.com/group2prject_telehealth/scheduler_models"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/kaatinga/assets"
	"github.com/kaatinga/bufferedlogger"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.com/group2prject_telehealth/backend/models"
)

type handlerData struct {
	db                   *sql.DB
	pageData             ViewData
	noRender             bool
	noAuth               bool
	formID               string
	formValue            string
	whereToRedirect      string
	additionalRedirectID string
	logger               bufferedlogger.Logger
	logData              bytes.Buffer // для буфферизации логов
	JSON                 bool         // определяем какой рендер будет использовать

	// поля для работы с расписанием с использованеим адаптеров
	Gap                scheduler_models.Gap
	DoctorID           uint16
	DoctorIDIsReceived bool
}

func (hd *handlerData) GatherUserData(r *http.Request) {
	var cookie *http.Cookie
	var err error

	cookie, err = r.Cookie(sessionName)
	if err != nil {
		hd.setError(http.StatusUnauthorized, errors.New("a session cookie has not been detected"))
		return
	}

	hd.logger.SubMsg.Info().Msg("A Session Cookie Detected")

	// validate the token
	var rawClaims ClaimBody
	var token *jwt.Token
	token, err = jwt.ParseWithClaims(cookie.Value, &rawClaims, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	switch err.(type) {
	case nil: // no Error
		if !token.Valid { // but may still be invalid
			hd.setError(http.StatusUnauthorized, errors.New("a session cookie is invalid"))
			return
		}
	case *jwt.ValidationError: // something was wrong during the validation
		vErr := err.(*jwt.ValidationError)

		switch vErr.Errors {
		case jwt.ValidationErrorExpired:
			hd.setError(http.StatusUnauthorized, errors.New("token expired"))
			return

		default:
			hd.setError(http.StatusInternalServerError, errors.New("session parsing Error #1"))
			return
		}
	default: // something else went wrong
		hd.setError(http.StatusInternalServerError, errors.New("session parsing Error #2"))
		return
	}

	// casting rawClaims into ClaimBody type
	claimBody, ok := token.Claims.(*ClaimBody)
	if !ok {
		hd.setError(http.StatusInternalServerError, errors.New("error when getting jwt token claims"))
		return
	}

	hd.pageData.Session = true

	hd.pageData.User.Name = claimBody.Name
	hd.pageData.User.Email = claimBody.Email
	hd.pageData.User.ID = claimBody.ID
	hd.pageData.User.Doctor = claimBody.Doctor
	hd.pageData.User.AvatarURI = claimBody.Avatar

	hd.logger.SubMsg.Info().Uint16("ID", hd.pageData.ID).Str("Email", hd.pageData.Email).Bool("Doctor", hd.pageData.Doctor).Msg("User Data")
}

// ViewData - модель данных страницы
type ViewData struct {
	Error    string
	Title    string
	Message  string
	Template string // Путь к файлу шаблона
	Status   uint16
	URL      string
	Method   string
	Session  bool
	MenuList []MenuData
	Data     *JSONResponse
	Token    string // хранит токен
	User
	Paginator // Для ленивой подгрузки данных
}

// User is a structure to process user data
type User struct {
	ID         uint16
	Name       string
	Surname    string
	Patronymic string
	Email      string
	Disabled   bool
	Pwd        string
	Doctor     bool
	Spec       byte
	Diploma    string
	Birthday   string
	Experience byte
	AvatarURI string
}

// MenuData - Модель данных ссылки на страницу
type MenuData struct {
	URL      string
	Name     string
	Selected bool
}

func AddMenuItem(menuData []MenuData, url, name, currentURL string) []MenuData {
	return append(menuData, MenuData{url, name, assets.CompareTwoStrings(currentURL, url)})
}

// Paginator

type Paginator struct {
	Paginator bool
	Total     int
	Page      uint32
	Title     string
	Buttons   []Button
}

type Button struct {
	Text string
	Page uint32
}

func (p *Paginator) FillOut() {

	// проверка допустимого диапазона чисел и исправление
	if p.Page < 1 {
		p.Page = 1
	}

	// определяем наличие пагинации
	if p.Page > 1 || p.Page == 1 && p.Total == 21 {

		// основная переменная для проверки есть ли пагинация
		p.Paginator = true

		// предварительный слайс с кнопками, макс. 2
		//p.Buttons = make([]Button, 1)

		p.Title = strconv.Itoa(int((p.Page-1)*20)) + " — " + strconv.Itoa(int(p.Page*20))
	}

	if p.Page > 1 {
		p.Buttons = append(p.Buttons, Button{
			Text: "Предыдущая",
			Page: p.Page - 1,
		})
	}

	if p.Total == 21 {
		p.Buttons = append(p.Buttons, Button{
			Text: "Следующая",
			Page: p.Page + 1,
		})
	}
}

type JSONResponse struct {
	Text       string
	Status     *uint16
	CustomData interface{}
}

func (response *JSONResponse) IsEmpty() bool {
	return response.Text == "" && response.CustomData == nil
}

//ReviewsInfo - returns the average rating and number of doctor reviews
func ReviewsInfo(docID models.Doctor, hd handlerData) (count int64, rating float32) {
	var summ float32
	ctx := context.Background()
	count, err := models.Reviews(qm.Where("doctor_id=?", docID.ID)).Count(ctx, hd.db)
	if err != nil {
		hd.setError(http.StatusNotFound, err)
		return
	}
	ratings, err := models.Reviews(qm.Where("doctor_id=?", docID.ID)).All(ctx, hd.db)
	if err != nil {
		hd.setError(http.StatusNotFound, err)
		return
	}
	for _, rating := range ratings {
		summ = summ + rating.Rating.Float32

	}
	var average float32
	if summ != 0 {
		average = summ / float32(count)
	} else {
		average = summ
	}
	return count, average
}

//AgeUser - returns the age of the doctor or patient
func AgeUser(date time.Time) int {
	sub := time.Now().Sub(date)
	age := sub.Hours() / 8766
	return int(age)
}
