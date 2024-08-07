// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/StasMerzlyakov/go-metrics/internal/server/app (interfaces: Pinger,AllMetricsStorage,BackupFormatter,Storage,MetricsChecker)

// Package app_test is a generated GoMock package.
package app_test

import (
	context "context"
	reflect "reflect"

	domain "github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	gomock "github.com/golang/mock/gomock"
)

// MockPinger is a mock of Pinger interface.
type MockPinger struct {
	ctrl     *gomock.Controller
	recorder *MockPingerMockRecorder
}

// MockPingerMockRecorder is the mock recorder for MockPinger.
type MockPingerMockRecorder struct {
	mock *MockPinger
}

// NewMockPinger creates a new mock instance.
func NewMockPinger(ctrl *gomock.Controller) *MockPinger {
	mock := &MockPinger{ctrl: ctrl}
	mock.recorder = &MockPingerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPinger) EXPECT() *MockPingerMockRecorder {
	return m.recorder
}

// Ping mocks base method.
func (m *MockPinger) Ping(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockPingerMockRecorder) Ping(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockPinger)(nil).Ping), arg0)
}

// MockAllMetricsStorage is a mock of AllMetricsStorage interface.
type MockAllMetricsStorage struct {
	ctrl     *gomock.Controller
	recorder *MockAllMetricsStorageMockRecorder
}

// MockAllMetricsStorageMockRecorder is the mock recorder for MockAllMetricsStorage.
type MockAllMetricsStorageMockRecorder struct {
	mock *MockAllMetricsStorage
}

// NewMockAllMetricsStorage creates a new mock instance.
func NewMockAllMetricsStorage(ctrl *gomock.Controller) *MockAllMetricsStorage {
	mock := &MockAllMetricsStorage{ctrl: ctrl}
	mock.recorder = &MockAllMetricsStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAllMetricsStorage) EXPECT() *MockAllMetricsStorageMockRecorder {
	return m.recorder
}

// GetAllMetrics mocks base method.
func (m *MockAllMetricsStorage) GetAllMetrics(arg0 context.Context) ([]domain.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllMetrics", arg0)
	ret0, _ := ret[0].([]domain.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllMetrics indicates an expected call of GetAllMetrics.
func (mr *MockAllMetricsStorageMockRecorder) GetAllMetrics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllMetrics", reflect.TypeOf((*MockAllMetricsStorage)(nil).GetAllMetrics), arg0)
}

// SetAllMetrics mocks base method.
func (m *MockAllMetricsStorage) SetAllMetrics(arg0 context.Context, arg1 []domain.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetAllMetrics", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetAllMetrics indicates an expected call of SetAllMetrics.
func (mr *MockAllMetricsStorageMockRecorder) SetAllMetrics(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAllMetrics", reflect.TypeOf((*MockAllMetricsStorage)(nil).SetAllMetrics), arg0, arg1)
}

// MockBackupFormatter is a mock of BackupFormatter interface.
type MockBackupFormatter struct {
	ctrl     *gomock.Controller
	recorder *MockBackupFormatterMockRecorder
}

// MockBackupFormatterMockRecorder is the mock recorder for MockBackupFormatter.
type MockBackupFormatterMockRecorder struct {
	mock *MockBackupFormatter
}

// NewMockBackupFormatter creates a new mock instance.
func NewMockBackupFormatter(ctrl *gomock.Controller) *MockBackupFormatter {
	mock := &MockBackupFormatter{ctrl: ctrl}
	mock.recorder = &MockBackupFormatterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBackupFormatter) EXPECT() *MockBackupFormatterMockRecorder {
	return m.recorder
}

// Read mocks base method.
func (m *MockBackupFormatter) Read(arg0 context.Context) ([]domain.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", arg0)
	ret0, _ := ret[0].([]domain.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockBackupFormatterMockRecorder) Read(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockBackupFormatter)(nil).Read), arg0)
}

// Write mocks base method.
func (m *MockBackupFormatter) Write(arg0 context.Context, arg1 []domain.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Write indicates an expected call of Write.
func (mr *MockBackupFormatterMockRecorder) Write(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockBackupFormatter)(nil).Write), arg0, arg1)
}

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockStorage) Add(arg0 context.Context, arg1 *domain.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockStorageMockRecorder) Add(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockStorage)(nil).Add), arg0, arg1)
}

// AddMetrics mocks base method.
func (m *MockStorage) AddMetrics(arg0 context.Context, arg1 []domain.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddMetrics", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddMetrics indicates an expected call of AddMetrics.
func (mr *MockStorageMockRecorder) AddMetrics(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddMetrics", reflect.TypeOf((*MockStorage)(nil).AddMetrics), arg0, arg1)
}

// Get mocks base method.
func (m *MockStorage) Get(arg0 context.Context, arg1 string, arg2 domain.MetricType) (*domain.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1, arg2)
	ret0, _ := ret[0].(*domain.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockStorageMockRecorder) Get(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStorage)(nil).Get), arg0, arg1, arg2)
}

// GetAllMetrics mocks base method.
func (m *MockStorage) GetAllMetrics(arg0 context.Context) ([]domain.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllMetrics", arg0)
	ret0, _ := ret[0].([]domain.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllMetrics indicates an expected call of GetAllMetrics.
func (mr *MockStorageMockRecorder) GetAllMetrics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllMetrics", reflect.TypeOf((*MockStorage)(nil).GetAllMetrics), arg0)
}

// Set mocks base method.
func (m *MockStorage) Set(arg0 context.Context, arg1 *domain.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockStorageMockRecorder) Set(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockStorage)(nil).Set), arg0, arg1)
}

// SetMetrics mocks base method.
func (m *MockStorage) SetMetrics(arg0 context.Context, arg1 []domain.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetMetrics", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetMetrics indicates an expected call of SetMetrics.
func (mr *MockStorageMockRecorder) SetMetrics(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetMetrics", reflect.TypeOf((*MockStorage)(nil).SetMetrics), arg0, arg1)
}

// MockMetricsChecker is a mock of MetricsChecker interface.
type MockMetricsChecker struct {
	ctrl     *gomock.Controller
	recorder *MockMetricsCheckerMockRecorder
}

// MockMetricsCheckerMockRecorder is the mock recorder for MockMetricsChecker.
type MockMetricsCheckerMockRecorder struct {
	mock *MockMetricsChecker
}

// NewMockMetricsChecker creates a new mock instance.
func NewMockMetricsChecker(ctrl *gomock.Controller) *MockMetricsChecker {
	mock := &MockMetricsChecker{ctrl: ctrl}
	mock.recorder = &MockMetricsCheckerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricsChecker) EXPECT() *MockMetricsCheckerMockRecorder {
	return m.recorder
}

// CheckMetrics mocks base method.
func (m *MockMetricsChecker) CheckMetrics(arg0 *domain.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckMetrics", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckMetrics indicates an expected call of CheckMetrics.
func (mr *MockMetricsCheckerMockRecorder) CheckMetrics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckMetrics", reflect.TypeOf((*MockMetricsChecker)(nil).CheckMetrics), arg0)
}
