package domain

//go:generate mockgen -destination "./generated_mocks_test.go" -package ${GOPACKAGE}_test . Logger

type Logger interface {
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}
