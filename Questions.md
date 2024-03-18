- АРХИТЕКТУРА!!! 
   - убрал картинку - пришел к гексогональной архитектуре но с use-кейсами и доменной областью, как средством деления логики приложения
     в учебном приложинии получиличь области
         - работа с бэкапом
         - администривные операции(пока только ping)
         - работа с метриками
         

- storage/wrapper/retriable.go - может есть более изящное решение? (дублирование кода, создание функциональные оберток)

- defer req.Body.Close()  
  В документации (https://pkg.go.dev/net/http#Request):
    The Server will close the request body. The ServeHTTP Handler does not need to.
  Получается что можно не заморачиваться?
  
- DOCKER - было бы логично добавить тестирование storage на postgres, но смущает реализация unit-тестов и для локального запуска и в связке с github metricstest.  (вылазят вопросы типа - можно ли завязаться на переменные окружение  POSTGRES_PASSWORD: postgres, POSTGRES_DB: praktikum и т.д). Будет ли использование в кусре далее? Стоит ли сейчас заморичиваться?

- (ОТВЕТ ПОЛУЧЕН; TODO) w.WriteHeader(http.StatusOK) - где писать, надо ли вообще писать? TODO поправить хэндлеры

- (ОТВЕТ ПОЛУЧЕН; TODO ) Загрузка конфигурации (flag, env) (см internal/config/server.go) - можно ли покрыть тестами (ссылка на пример, если есть; надо ли вообще?  - TODO пример - https://github.com/golang/go/blob/master/src/flag/example_test.go)

