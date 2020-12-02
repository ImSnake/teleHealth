package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/kaatinga/bufferedlogger"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// routes
func SetUpHandlers(r *httprouter.Router, db *sql.DB) {

	// страницы

	r.GET("/", Adapt(Welcome, InitPage(false, db)))
	r.GET("/doctor/:id", Adapt(DoctorPage, InitPage(false, db)))
	r.GET("/catalog/:id", Adapt(Catalog, InitPage(false, db)))
	r.GET("/profile", Adapt(DoctorProfile, InitPage(false, db)))

	// API

	// возвращает список специализаций

	r.GET("/specs", Adapt(Specs, InitPage(true, db)))
	r.GET("/specs_exists", Adapt(SpecsExists, InitPage(true, db))) // возвращает список в котором есть врачи

	// отзывы
	r.GET("/reviews/:id", Adapt(GetReviews, InitPage(true, db))) // возврашает массив c дотзывы с doctor_id == id

	// доктор
	r.GET("/api/doctor/:id", Adapt(DoctorJSON, InitPage(true, db)))
	r.POST("/api/doctor/add_spec", Adapt(AddSpecDoctor, CheckPostBytesLimit(3000), InitPage(false, db)))                 //добавляет специализацию
	r.POST("/api/doctor/add_photo", Adapt(AddPhotoProfile, CheckPostBytesLimit(300000), InitPage(false, db)))            // менятет аватар
	r.GET("/get-avatar/:id", Adapt(GetAvatar, CheckPostBytesLimit(300000), InitPage(false, db)))                         // возвращает путь к аватару доктора
	r.POST("/api/doctor/update_biography/:id", Adapt(UpdateBiography, CheckPostBytesLimit(300000), InitPage(false, db))) // менятет аватар

	// каталог
	r.GET("/catalog", Adapt(Catalog, InitPage(true, db)))
	r.GET("/specs/:id", Adapt(AllSpecialists, InitPage(true, db))) // возвращает массив врачей со specialization_id == :id

	// авторизация, регистрация и выход из личного кабинета
	r.POST("/login", Adapt(loginHandler, InitPage(true, db)))
	r.POST("/register", Adapt(registerHandler, InitPage(true, db)))
	r.POST("/logout", Adapt(Logout, InitPage(true, db)))

	r.POST("/api/editdoctor", Adapt(EditDoctor, InitPage(false, db)))

	// schedule handlers
	r.POST("/api/newschedule", Adapt(NewSchedule, InitPage(false, db)))
	r.POST("/api/makeavailable", Adapt(MakeAvailable, GetGapTime(), InitPage(false, db)))
	r.POST("/api/makeavailablebyhour/:day", Adapt(MakeAvailableByHour, InitPage(false, db)))
	r.POST("/api/makeunavailable", Adapt(MakeUnavailable, GetGapTime(), InitPage(false, db)))
	r.POST("/api/cancel", Adapt(Cancel, GetGapTime(), InitPage(false, db))) // только для доктора
	r.POST("/api/enrol", Adapt(Enrol, GetGapTime(), InitPage(false, db)))

	// получаем всё расписание доктора
	r.GET("/api/schedule/:id", Adapt(Schedule, InitPage(true, db)))

	// получаем список часов за день с отметкой наличия записи или открытой для записи 15-минутки
	r.GET("/api/schedule/:id/:month", Adapt(GetMonthDayList, InitPage(true, db)))
	r.GET("/api/dayschedule/:id/:day", Adapt(GetDayHourList, InitPage(true, db)))

	//r.GET("/api/dayschedule/:id/:day", Adapt(GetDayCalendar, InitPage(true, db)))
	r.GET("/api/hourschedule/:id/:date", Adapt(GetHourCalendar, InitPage(true, db)))

	// Регистрируем хандлер для созданного файлсервера статики.
	r.ServeFiles("/static/*filepath", http.Dir("ui/static")) // Relative path is not supported!

	// Обработчик favicon.ico
	r.GET("/favicon.ico", faviconHandler())
}

type Adapter func(httprouter.Handle) httprouter.Handle

func Adapt(next httprouter.Handle, adapters ...Adapter) httprouter.Handle {
	for _, adapter := range adapters {
		next = adapter(next)
	}
	return next
}

func InitPage(noAuth bool, db *sql.DB) Adapter {
	return func(next httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, actions httprouter.Params) {

			// Create handlerData struct
			hd := handlerData{db: db, noAuth: noAuth}
			hd.pageData.Status = http.StatusOK // by default
			hd.pageData.Data = &JSONResponse{Status: &hd.pageData.Status}

			// Create Logger
			// TODO использовать общий логгер можно, тут он для возможной буфферизации логов
			hd.logger = bufferedlogger.InitLog(&hd.logData)

			hd.logger.Title.Info().Str("IP", r.RemoteAddr).Str("Method", r.Method).Str("URL", r.URL.String()).Msg("== A new request is received:")

			// рендер и логер работают отложенно с проверкой условия
			defer func() {

				hd.logger.SubMsg.Info().Bool("hd.JSON", hd.JSON).Msg("JSON Mode")

				if hd.JSON {
					hd.RenderJSON(w)
				} else {
					hd.RenderHTML(w)
				}

				// Flush the log
				if hd.pageData.Status > 399 {
					os.Stderr.Write(hd.logData.Bytes())
				} else {
					os.Stdout.Write(hd.logData.Bytes())
				}
			}()

			hd.pageData.URL = r.URL.String()
			hd.pageData.Method = r.Method

			// Кастыль чтобы не проверять статус аутентификации при логауте
			if !hd.noAuth {
				hd.GatherUserData(r) // проверяем пользователя и заполняем его данные
			}

			if hd.pageData.Status == 419 {
				hd.logger.SubMsg.Warn().Msg("deleting a bugged cookie")
				clearSession(w) // Раз в куки глюк, то её можно удалить
			}

			if hd.pageData.Status == http.StatusOK || hd.pageData.Status == http.StatusUnauthorized {
				ctx := context.WithValue(r.Context(), "hd", &hd)
				next(w, r.WithContext(ctx), actions)
			} else {
				hd.logger.SubMsg.Warn().Msg("следующий хэндлер исключён")
			}
		}
	}
}

func CheckPostBytesLimit(maxBytesLimit int64) Adapter {
	return func(next httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, actions httprouter.Params) {

			// Get HandlerData struct
			hd := r.Context().Value("hd").(*handlerData)

			// check the request body size limit
			r.Body = http.MaxBytesReader(w, r.Body, maxBytesLimit)
			err := r.ParseForm()
			if err != nil {
				hd.setError(http.StatusBadRequest, err)
				return
			}

			// send the handlerData to the next handler
			ctx := context.WithValue(r.Context(), "hd", hd)
			next(w, r.WithContext(ctx), actions)
		}
	}
}

// Функция исполняет типовые действия в случае ошибки. Вызывается из formRequest() в случае любой ошибки
func (hd *handlerData) setError(status uint16, err error) {

	// устанавливаем статус
	hd.pageData.Status = status

	// записываем ошибку в модель
	if err != nil {
		hd.pageData.Error = err.Error()

		// выводим ошибку в лог
		hd.logger.SubMsg.Error().Msg(hd.pageData.Error)

		// set the error template
		if hd.pageData.Status != http.StatusUnauthorized {
			hd.pageData.Template = getPath("error.html")
		}
	}
}

//RenderJSON - Функция для вывода страницы пользователю
func (hd *handlerData) RenderJSON(w http.ResponseWriter) {

	// проверяем что есть ошибка и сообщаем в лог
	var body []byte
	var err error

	if hd.pageData.Status != http.StatusOK {
		hd.pageData.Data.Text = hd.pageData.Error
		hd.logger.SubMsg.Warn().Uint16("code", hd.pageData.Status).Msg("The code is not 200, the Status")
	} else if hd.pageData.Data.IsEmpty() {
		hd.pageData.Status = http.StatusNoContent
	} else {
		hd.pageData.Data.Text = hd.pageData.Message
	}

	if !hd.pageData.Data.IsEmpty() {
		body, err = json.Marshal(hd.pageData.Data)
		if err != nil {
			hd.logger.SubMsg.Err(err).Msg("JSON Marshalling Error")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(int(hd.pageData.Status)) // Добавляем в заголовок сообщение об ошибке

		_, err = w.Write(body)
		if err != nil {
			hd.logger.SubMsg.Err(err).Msg("RenderJSON Error")
			return
		}

		// если ошибки нет
		hd.logger.SubSubMsg.Info().Msg("JSON is rendered successfully")
	} else {
		hd.logger.SubMsg.Warn().Msg("JSON response is turned off")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(int(hd.pageData.Status)) // Добавляем в заголовок сообщение об ошибке
	}
}

//RenderHTML - Функция для вывода страницы пользователю
func (hd *handlerData) RenderHTML(w http.ResponseWriter) {

	// проверяем что есть ошибка и сообщаем в лог
	if hd.pageData.Status != http.StatusOK {

		if hd.pageData.Title == "" {
			hd.pageData.Title = strconv.Itoa(int((*hd).pageData.Status))
		}

		if hd.pageData.Message == "" {
			hd.pageData.Message = http.StatusText(int((*hd).pageData.Status))
		}

		if hd.pageData.Message == "" { // Может так быть что StatusText() ничего не вернёт, тогда дополняем текст сами
			hd.pageData.Message = strings.Join([]string{"Ошибка обработки запроса, код ошибки", (*hd).pageData.Title}, " ")
		}

		w.WriteHeader(int(hd.pageData.Status)) // Добавляем в заголовок сообщение об ошибке
		hd.logger.SubMsg.Warn().Uint16("code", hd.pageData.Status).Msg("The code is not 200, the Status")
	}

	// путь к основному шаблону
	layout := getPath("base/base.html")

	if hd.pageData.Template == "" {
		hd.logger.SubMsg.Warn().Msg("Template is not set")
		hd.pageData.Template = getPath("index.html") // дефолтный контент
	}

	hd.logger.SubMsg.Info().Str("template", (*hd).pageData.Template).Msg("The page will be rendered by")

	var tmpl *template.Template
	var err error
	tmpl, err = template.ParseFiles(layout, (*hd).pageData.Template)
	if err != nil {
		// Вываливаем в лог кучу хлама для анализа
		hd.logger.SubMsg.Err(err).Msg("Template Error")
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", hd.pageData)
	if err != nil {
		hd.logger.SubMsg.Err(err).Msg("Template Error")
		return
	}

	// если ошибки нет
	hd.logger.SubSubMsg.Info().Msg("Ошибки при формировании страницы по шаблону не обнаружено")
}

// getPath takes the input string with a '/' ot '\' separators and returns the OS specific path to the file
func getPath(fileName string) string {
	return filepath.Join("ui", "templates", fileName)
}
