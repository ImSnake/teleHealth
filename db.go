package main

import (
	// "context"
	"database/sql"
	"errors"

	// "fmt"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/volatiletech/sqlboiler/v4/queries/qm"
	// "gitlab.com/group2prject_telehealth/backend/models"
)

// считывает базовые данные о пользователе
func (hd *handlerData) getLoginData() (err error) {

	var query *sql.Rows
	query, err = hd.db.Query(`SELECT id, pwd, doctor, photo_url 
FROM patient
WHERE email=? AND deleted_at IS NULL;`, hd.pageData.Email)
	if err != nil {
		return
	}
	defer query.Close()

	if ok := query.Next(); ok {
		err = query.Scan(&hd.pageData.ID, &hd.pageData.Pwd, &hd.pageData.Doctor, &hd.pageData.AvatarURI)
		if err != nil {
			return
		}
	} else {
		return errors.New("no user data found")
	}
	return nil
}

// checkExistenceByEmail проверяет есть ли такой емейл уже в базе
func (hd *handlerData) checkExistenceByEmail() (existence bool, err error) {
	var query *sql.Rows
	query, err = hd.db.Query(`SELECT CASE WHEN EXISTS (SELECT ID FROM patient WHERE email=? AND deleted_at IS NULL)
THEN 1
ELSE 0 END;`, hd.pageData.User.Email)
	if err != nil {
		return
	}

	if query.Next() {
		err = query.Scan(&existence)
		if err != nil {
			return
		}
	} else {
		err = errors.New("request result is empty")
		return
	}

	defer query.Close()
	return
}

// newUser adds a new user to the db
func (hd *handlerData) newUser() (err error) {
	var result sql.Result

	result, err = hd.db.Exec(
		"INSERT INTO patient (email, pwd, doctor, photo_url) values (?, ?, ?, ?);",
		hd.pageData.Email,
		hd.pageData.Pwd,
		hd.pageData.Doctor,
		"/static/images/avatars/avatar.png")
	if err != nil {
		return err
	}
	var tempID int64
	tempID, err = result.LastInsertId()
	hd.pageData.User.ID = uint16(tempID)

	if hd.pageData.User.Doctor {
		_, err = hd.db.Exec(
			"INSERT INTO doctor (ID, name, surname, patronymic, birthdate, photo_url, biography) values (?, ?, ?, ?, ?, ?, ?);",
			hd.pageData.ID,
			hd.pageData.Name,
			hd.pageData.Surname,
			hd.pageData.Patronymic,
			hd.pageData.Birthday,
			"/static/images/avatars/avatar.png",
			"Введите текст")
			if err != nil {
				return err
			}

		_, err = hd.db.Exec("INSERT INTO specialization_kit (doctor_id, specialization_id, certificate_n, experience) values (?, ?, ?, ?);", hd.pageData.ID, hd.pageData.Spec, hd.pageData.Diploma, hd.pageData.Experience)
		if err != nil {
			return err
		}
	}

	return
}

// updateDoctor updates doctor's data in the DB
func (hd *handlerData) updateDoctor() (err error) {
	var result sql.Result
	result, err = hd.db.Exec("UPDATE doctor SET name=?, surname=?, patronymic=?, birthdate=? WHERE id=?;", hd.pageData.Name, hd.pageData.Surname, hd.pageData.Patronymic, hd.pageData.Birthday, hd.pageData.ID)
	if err != nil {
		return err
	}

	err = checkSQLUpdateQueryResult(&result)
	if err != nil {
		return err
	}

	result, err = hd.db.Exec("UPDATE specialization_kit SET certificate_n=?, experience=? WHERE doctor_id=?;", hd.pageData.Diploma, hd.pageData.Experience, hd.pageData.ID)
	if err != nil {
		return err
	}

	err = checkSQLUpdateQueryResult(&result)
	if err != nil {
		return err
	}

	return
}

// checkSQLUpdateQueryResult checks SQL UPDATE query result
func checkSQLUpdateQueryResult(result *sql.Result) error {
	affected, err := (*result).RowsAffected()
	if err != nil {
		return err
	}

	if affected < 1 {
		return errors.New("no such a record in the database")
	}

	return nil
}
