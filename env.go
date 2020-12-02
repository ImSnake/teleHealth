package main

// Модель данных для параметров окружения
type EnvironmentSettings struct {
	Port       string `env:"PORT" validate:"required"`
	Database   string `env:"DATABASE" validate:"required"`
}
