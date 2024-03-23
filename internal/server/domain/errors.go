package domain

import (
	"errors"
)

var (
	ErrServerInternal    = errors.New("InternalError")        // Ошибка на сервере
	ErrDataFormat        = errors.New("DataFormatError")      // Ошибка в данных
	ErrDataDigestMismath = errors.New("ErrDataDigestMismath") // Несовпадение хэша
	ErrNotFound          = errors.New("NotFoundError")
	ErrDBConnection      = errors.New("DatabaseConnectionError")
)
