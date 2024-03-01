package domain

import "errors"

var (
	InternalError = errors.New("internalError") // Ошибка на сервере
	DataError     = errors.New("dataError")     // Ошибка в данных
)
