package main

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/kaatinga/assets"
	"net/http"
)

// === AUTH HANDLERS ===

//registerHandler ...
func registerHandler(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	var err error

	hd.pageData.Email = r.PostFormValue("email")
	hd.pageData.Pwd = r.PostFormValue("add-password")
	hd.pageData.Doctor = assets.StBool(r.FormValue("doctor-true"))

	hd.logger.SubMsg.Info().Str("Email", hd.pageData.Email).Str("Password", hd.pageData.Pwd).Bool("DoctorJSON", hd.pageData.Doctor).Msg("Form data retrieved")

	// проверяем емейл и пароль
	hd.logger.SubMsg.Info().Msg("Validating the input data...")
	if hd.pageData.Email == "" || hd.pageData.Pwd == "" {
		hd.setError(http.StatusBadRequest, EmptyData)
		return
	}

	if len(hd.pageData.Pwd) < 6 {
		hd.setError(http.StatusBadRequest, ShortPassword)
		return
	}

	if !assets.IsEmailValid(hd.pageData.Email) {
		hd.setError(http.StatusBadRequest, IncorrectInputLoginData)
		return
	}

	if hd.pageData.Doctor {
		hd.GetDoctorDataFromForm(r)
	} else {
		hd.logger.SubMsg.Info().Msg("The future user is NOT a doctor!")
	}

	// At this point we consider the request as correct
	hd.logger.SubMsg.Info().Msg("Validation passed. Creating new user...")
	err = hd.newUser()
	if err != nil {
		if len(err.Error()) > 6 && err.Error()[:6] == "UNIQUE" {
			hd.setError(http.StatusBadRequest, err)
			return
		} else {
			hd.setError(http.StatusInternalServerError, err)
			return
		}
	}

	// создаём расписание доктору при регистрации
	if hd.pageData.User.Doctor {
		hd.CreateNewSchedule()
	}
}

// loginHandler returns a token for the user
func loginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	var err error

	hd.pageData.Email = r.PostFormValue("email")
	password := r.PostFormValue("password")

	if hd.pageData.Email == "" || password == "" {
		hd.setError(http.StatusBadRequest, EmptyData)
		return
	}

	if !assets.IsEmailValid(hd.pageData.Email) {
		hd.setError(http.StatusBadRequest, IncorrectInputLoginData)
		return
	}

	err = hd.getLoginData()
	if err != nil {
		hd.setError(http.StatusBadRequest, err)
		return
	}

	if hd.pageData.User.Doctor {
		hd.logger.SubMsg.Info().Msg("The user is a doctor")

		// TODO можно добавить имя пользователя в куку
	}

	if hd.pageData.User.Pwd != password {
		hd.setError(http.StatusBadRequest, errors.New("incorrect login or password"))
		return
	}

	// At this point we consider the request as correct
	hd.pageData.Token, err = hd.pageData.CreateToken()
	if err != nil {
		hd.setError(http.StatusInternalServerError, err)
		return
	}

	setCookie(w, sessionName, hd.pageData.Token)

	if hd.pageData.User.Doctor {
		hd.CreateNewSchedule()
	}
}

// Logout is to clear the session cookie (to logout eventually)
func Logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	clearSession(w) // Очищаем куку сессии

	hd.noAuth = true // Кастыль чтобы не проверять статус аутентификации при логауте
}

func (hd *handlerData) GetDoctorDataFromForm(r *http.Request) {
	hd.pageData.Surname = r.PostFormValue("surname")
	hd.pageData.Patronymic = r.PostFormValue("patronymic")
	hd.pageData.Name = r.PostFormValue("name")

	hd.logger.SubMsg.Info().Str("Surname", hd.pageData.Surname).Str("Name", hd.pageData.Name).Str("Patronymic", hd.pageData.Patronymic).Msg("Form data retrieved")

	if hd.pageData.Name == "" || hd.pageData.Surname == "" {
		hd.setError(http.StatusBadRequest, EmptyData)
		return
	}

	// забираем специализацию из формы
	var ok bool
	hd.pageData.User.Spec, ok = assets.StByte(r.PostFormValue("specialization"))
	if !ok {
		hd.setError(http.StatusBadRequest, errors.New("incorrect specialization id"))
		return
	}
	hd.logger.SubMsg.Info().Uint8("ID", hd.pageData.User.Spec).Msg("Specialization")

	// забираем стаж
	hd.pageData.User.Experience, ok = assets.StByte(r.PostFormValue("experience"))
	if !ok {
		hd.setError(http.StatusBadRequest, errors.New("incorrect experience"))
		return
	}
	hd.logger.SubMsg.Info().Uint8("Years", hd.pageData.User.Experience).Msg("Experience")

	// забираем номер диплома
	hd.pageData.User.Diploma = r.PostFormValue("document")
	if hd.pageData.User.Diploma == "" {
		hd.setError(http.StatusBadRequest, EmptyData)
		return
	}
	hd.logger.SubMsg.Info().Str("№", hd.pageData.User.Diploma).Msg("Diploma")

	// забираем ДР
	hd.pageData.User.Birthday = r.PostFormValue("birthday")
	if hd.pageData.User.Birthday == "" {
		hd.setError(http.StatusBadRequest, EmptyData)
		return
	}
	hd.logger.SubMsg.Info().Str("Day", hd.pageData.User.Birthday).Msg("Birthday")
}

// EditDoctor
func EditDoctor(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.JSON = true // устанавливаем JSON Render

	// ID доктора берём из куки и проверяем
	if !hd.pageData.User.Doctor {
		hd.setError(http.StatusNotFound, IncorrectDoctorID)
		return
	}

	// начинаем получать данные из формы
	hd.GetDoctorDataFromForm(r)

	err := hd.updateDoctor()
	if err != nil {
		hd.setError(http.StatusInternalServerError, err)
	}
}
