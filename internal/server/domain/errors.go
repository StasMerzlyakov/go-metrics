package domain

import (
	"errors"
	"net/http"
)

var (
	ErrServerInternal    = errors.New("InternalError")        // Ошибка на сервере
	ErrDataFormat        = errors.New("DataFormatError")      // Ошибка в данных
	ErrDataDigestMismath = errors.New("ErrDataDigestMismath") // Несовпадение хэша
	ErrNotFound          = errors.New("NotFoundError")
	ErrDBConnection      = errors.New("DatabaseConnectionError")
	ErrMediaType         = errors.New("UnsupportedMediaTypeError")
)

func MapDomainErrorToHTTPStatusErr(err error) int {
	if errors.Is(err, ErrMediaType) {
		return http.StatusUnsupportedMediaType
	}

	if errors.Is(err, ErrDataFormat) {
		return http.StatusBadRequest
	}

	if errors.Is(err, ErrDataDigestMismath) {
		return http.StatusBadRequest
	}

	if errors.Is(err, ErrServerInternal) {
		return http.StatusBadRequest
	}

	if errors.Is(err, ErrDBConnection) {
		return http.StatusInternalServerError
	}

	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}
