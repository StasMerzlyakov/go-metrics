package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/handler"
	"github.com/go-resty/resty/v2"
)

/*
По документации
Godoc uses a naming convention to associate an example function with a package-level identifier.

func ExampleFoo()     // documents the Foo function or type
func ExampleBar_Qux() // documents the Qux method of type Bar
func Example()        // documents the package as a whole

В нашем случаем metricOperationAdapter - private; методом проб нашел рабочий вариант - добавить '_' перед типом, а сразу за типом писать имя функции.
Не очень ясно - что делать, когда для запуска примера нужны разные вспомогательные объекты.
*/

func Example_metricOperationAdapterPostCounter() {

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(mux)

	// Request example start

	counterName := "TestCounter"
	testValue := int64(132)
	testValueStr := fmt.Sprintf("%v", testValue)

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/update/counter/" + counterName + "/" + testValueStr
	req.Header.Add("Content-Type", handler.TextPlain)
	_, err := req.Send()

	fmt.Printf("%v", err)

	// Output:
	// <nil>
}
