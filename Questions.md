- АРХИТЕКТУРА!!! 
   - убрал картинку - после нескольких итераций пришел к гексогональной архитектуре но с use-кейсами и доменной областью, как средством деления логики приложения
     в учебном приложении получилиcь области
         - работа с бэкапом
         - администривные операции(пока только ping)
         - работа с метриками
         

- storage/wrapper/retriable.go - может есть более изящное решение? (дублирование кода, создание функциональные оберток)

- middleware/testutils/http_handler.go - добавил Handler интерфейс ради генерации mock. Можно ли как-то заставить gomock сгенерировать mock для стандартного интерфейса http.Handler.

- defer req.Body.Close() в http-хэндлере: 
  В документации (https://pkg.go.dev/net/http#Request):
    The Server will close the request body. The ServeHTTP Handler does not need to.
  Получается что можно не заморачиваться с закрытием?
  
- DOCKER - было бы логично добавить тестирование storage на postgres, но смущает реализация unit-тестов и для локального запуска и в связке с github metricstest.  (появляются вопросы по реализации - можно ли завязаться на переменные окружение  POSTGRES_PASSWORD: postgres, POSTGRES_DB: praktikum и т.д). Возможно будет использование docker в кусре далее? Стоит ли сейчас заниматься?

- (ОТВЕТ ПОЛУЧЕН; TODO) w.WriteHeader(http.StatusOK) - где писать, надо ли вообще писать? TODO поправить хэндлеры

- (ОТВЕТ ПОЛУЧЕН; TODO ) Загрузка конфигурации (flag, env) (см internal/config/server.go) - можно ли покрыть тестами (ссылка на пример, если есть; надо ли вообще?  - TODO пример - https://github.com/golang/go/blob/master/src/flag/example_test.go)

- TODO - оборнуть через fmt.Errorf  места типа "if err := mc.storage.SetMetrics(ctx, gaugeList); err != nil {return err}"

- TODO Убрать логику из http.Handlers


