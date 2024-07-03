// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/StasMerzlyakov/go-metrics/internal/agent (interfaces: ResultSender,MetricStorage,Logger)

// Package agent_test is a generated GoMock package.
package agent_test

import (
	context "context"
	reflect "reflect"

	agent "github.com/StasMerzlyakov/go-metrics/internal/agent"
	gomock "github.com/golang/mock/gomock"
)

// MockResultSender is a mock of ResultSender interface.
type MockResultSender struct {
	ctrl     *gomock.Controller
	recorder *MockResultSenderMockRecorder
}

// MockResultSenderMockRecorder is the mock recorder for MockResultSender.
type MockResultSenderMockRecorder struct {
	mock *MockResultSender
}

// NewMockResultSender creates a new mock instance.
func NewMockResultSender(ctrl *gomock.Controller) *MockResultSender {
	mock := &MockResultSender{ctrl: ctrl}
	mock.recorder = &MockResultSenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResultSender) EXPECT() *MockResultSenderMockRecorder {
	return m.recorder
}

// SendMetrics mocks base method.
func (m *MockResultSender) SendMetrics(arg0 context.Context, arg1 []agent.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMetrics", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMetrics indicates an expected call of SendMetrics.
func (mr *MockResultSenderMockRecorder) SendMetrics(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMetrics", reflect.TypeOf((*MockResultSender)(nil).SendMetrics), arg0, arg1)
}

// MockMetricStorage is a mock of MetricStorage interface.
type MockMetricStorage struct {
	ctrl     *gomock.Controller
	recorder *MockMetricStorageMockRecorder
}

// MockMetricStorageMockRecorder is the mock recorder for MockMetricStorage.
type MockMetricStorageMockRecorder struct {
	mock *MockMetricStorage
}

// NewMockMetricStorage creates a new mock instance.
func NewMockMetricStorage(ctrl *gomock.Controller) *MockMetricStorage {
	mock := &MockMetricStorage{ctrl: ctrl}
	mock.recorder = &MockMetricStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricStorage) EXPECT() *MockMetricStorageMockRecorder {
	return m.recorder
}

// GetMetrics mocks base method.
func (m *MockMetricStorage) GetMetrics() []agent.Metrics {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetrics")
	ret0, _ := ret[0].([]agent.Metrics)
	return ret0
}

// GetMetrics indicates an expected call of GetMetrics.
func (mr *MockMetricStorageMockRecorder) GetMetrics() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetrics", reflect.TypeOf((*MockMetricStorage)(nil).GetMetrics))
}

// Refresh mocks base method.
func (m *MockMetricStorage) Refresh() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Refresh")
	ret0, _ := ret[0].(error)
	return ret0
}

// Refresh indicates an expected call of Refresh.
func (mr *MockMetricStorageMockRecorder) Refresh() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Refresh", reflect.TypeOf((*MockMetricStorage)(nil).Refresh))
}

// MockLogger is a mock of Logger interface.
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerMockRecorder
}

// MockLoggerMockRecorder is the mock recorder for MockLogger.
type MockLoggerMockRecorder struct {
	mock *MockLogger
}

// NewMockLogger creates a new mock instance.
func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogger) EXPECT() *MockLoggerMockRecorder {
	return m.recorder
}

// Infow mocks base method.
func (m *MockLogger) Infow(arg0 string, arg1 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Infow", varargs...)
}

// Infow indicates an expected call of Infow.
func (mr *MockLoggerMockRecorder) Infow(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Infow", reflect.TypeOf((*MockLogger)(nil).Infow), varargs...)
}
