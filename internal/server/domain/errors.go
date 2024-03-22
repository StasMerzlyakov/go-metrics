package domain

import (
	"errors"
)

var (
	ErrServerInternal = errors.New("InternalError")   // Ошибка на сервере
	ErrDataFormat     = errors.New("DataFormatError") // Ошибка в данных
	ErrNotFound       = errors.New("NotFoundError")
	ErrDBConnection   = errors.New("DatabaseConnectionError")
)
