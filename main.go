package main

import (
	"crypto/rsa"
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"github.com/kaatinga/env_loader"
	"gitlab.com/group2prject_telehealth/scheduler_models"
	"io/ioutil"
	"os"

	_ "github.com/go-sql-driver/mysql"

	launcher "github.com/kaatinga/QuickHTTPServerLauncher"
)

const (
	publicKeyPath  = "keys/app.rsa.pub"
	privateKeyPath = "keys/app.rsa"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
	schedules Schedules
)

func main() {

	var err error

	// New web service
	config := launcher.NewConfig()

	// --- TODO удалить при подключении базы
	// Init schedules
	schedules.list = make(map[uint16]scheduler_models.Schedule)
	err = initSchedule()
	if err != nil {
		config.Logger.SubMsg.Err(err).Msg("schedule init error")
		os.Exit(1)
	}
	// --- TODO удалить при подключении базы

	// Читаем RSA-ключ
	var signBytes []byte
	signBytes, err = ioutil.ReadFile(privateKeyPath)
	if err != nil {
		config.Logger.SubMsg.Err(err).Msg("RSA private key file reading error")
		os.Exit(1)
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		config.Logger.SubMsg.Err(err).Msg("RSA private key parsing error")
		os.Exit(1)
	}
	config.Logger.SubMsg.Info().Msg("Ok!")

	// Загружаем публичный ключ для проверки JWT-кук
	config.Logger.Title.Info().Msg("Loading RSA public key...")
	var verifyBytes []byte
	verifyBytes, err = ioutil.ReadFile(publicKeyPath)
	if err != nil {
		config.Logger.SubMsg.Err(err).Msg("RSA public key file reading Error")
		os.Exit(1)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		config.Logger.SubMsg.Err(err).Msg("RSA public key parsing Error")
		os.Exit(1)
	}
	config.Logger.SubMsg.Info().Msg("Ok!")

	// Getting environment settings for the network port and all the necessary paths
	config.Logger.Title.Info().Msg("Loading environment settings...")
	var myEnvs EnvironmentSettings
	err = env_loader.LoadUsingReflect(&myEnvs)
	if err != nil {
		config.Logger.SubMsg.Err(err).Msg("Environment variables have not been read")
		os.Exit(1)
	}
	config.Logger.SubMsg.Info().Msg("Ok!")

	// создаём соединение с БД
	config.Logger.Title.Info().Msg("Establishing connection to the database...")
	config.DB, err = sql.Open("mysql", myEnvs.Database)
	if err != nil {
		config.Logger.SubMsg.Err(err).Msg("DB error")
		os.Exit(1)
	}
	defer config.DB.Close()
	config.Logger.SubMsg.Info().Msg("Ok!")

	config.SetDomain("kaatinga.ru")
	config.SetEmail("info@kaatinga.ru")
	config.SetLaunchMode("dev")
	config.SetPort(myEnvs.Port)

	err = config.Launch(SetUpHandlers)
	if err != nil {
		config.Logger.SubMsg.Err(err).Msg("The server stopped")
	}
}
