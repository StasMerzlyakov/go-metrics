package domain

import (
	"errors"
)

var (
	ErrServerInternal = errors.New("InternalError")   // Ошибка на сервере
	ErrDataFormat     = errors.New("DataFormatError") // Ошибка в данных
	ErrDBConnection   = errors.New("DatabaseConnectionError")
)
