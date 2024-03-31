// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/StasMerzlyakov/go-metrics/internal/agent (interfaces: ResultSender)

// Package mocks is a generated GoMock package.
package mocks

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
