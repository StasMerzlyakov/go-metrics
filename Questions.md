- chi.Use хочет handler - можно ли подставить в chi http.HandlerFunc  (с handlerFunc работать удобней, чем с обычным http.Handler; см compress_utils_test.go); есть ли преобразование HandlerFunc в Handler

- Куда спрятать функции для тестирования (compress_utils_test.go)

- Формат логов - что писать, где писть, как писать

- w.WriteHeader(http.StatusOK) - где писать, надо ли вообще писать или http сервер сам добавляет