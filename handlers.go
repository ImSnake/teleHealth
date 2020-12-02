package main

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"net/http"
	"path/filepath"

	"gitlab.com/group2prject_telehealth/backend/models"
)

const (
	sessionName = "teleHeathSession"
)

// === HTML PAGE HANDLERS ===

// Welcome is the service homepage handler
func Welcome(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.pageData.Message = "Добро пожаловать в нашу Телемедецину!"
	hd.pageData.Title = "Главная страница"
}

// Catalog is the catalog handler
func Catalog(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.pageData.Template = getPath("catalog.html")
}

// DoctorPage is the doctor handler
func DoctorPage(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.pageData.Template = getPath("doctor.html")
	ctx := context.Background()

	doctorID := ps.ByName("id")

	doctor, err := models.Doctors(qm.Where("id=?", doctorID)).One(ctx, hd.db)
	if err != nil {
		hd.setError(http.StatusNotFound, err)
		return
	}

	hd.pageData.Data.CustomData = doctor
}

// DoctorProfile is the doctor profile handler
func DoctorProfile(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hd := r.Context().Value("hd").(*handlerData)
	hd.pageData.Template = getPath("profile.html")
}

// About is the About page handler
//func About(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//
//	hd := r.Context().Value("hd").(*handlerData)
//
//	hd.pageData.Message = "Это учебный проект Телемедецина"
//	hd.pageData.Title = "О проекте"
//}

// === OTHER HANDLERS ===

// Favicon Handler
func faviconHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.ServeFile(w, r, filepath.Join("ui", "static", "favicon.ico"))
	}
}
