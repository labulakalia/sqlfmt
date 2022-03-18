// Code generated by MockGen. DO NOT EDIT.
// Source: sqlfmt/cockroach/pkg/cmd/roachtest/prometheus (interfaces: Cluster)

// Package prometheus is a generated GoMock package.
package prometheus

import (
	context "context"
	fs "io/fs"
	reflect "reflect"

	option "github.com/labulakalia/sqlfmt/cockroach/pkg/cmd/roachtest/option"
	logger "github.com/labulakalia/sqlfmt/cockroach/pkg/roachprod/logger"
	gomock "github.com/golang/mock/gomock"
)

// MockCluster is a mock of Cluster interface.
type MockCluster struct {
	ctrl     *gomock.Controller
	recorder *MockClusterMockRecorder
}

// MockClusterMockRecorder is the mock recorder for MockCluster.
type MockClusterMockRecorder struct {
	mock *MockCluster
}

// NewMockCluster creates a new mock instance.
func NewMockCluster(ctrl *gomock.Controller) *MockCluster {
	mock := &MockCluster{ctrl: ctrl}
	mock.recorder = &MockClusterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCluster) EXPECT() *MockClusterMockRecorder {
	return m.recorder
}

// ExternalIP mocks base method.
func (m *MockCluster) ExternalIP(arg0 context.Context, arg1 *logger.Logger, arg2 option.NodeListOption) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExternalIP", arg0, arg1, arg2)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExternalIP indicates an expected call of ExternalIP.
func (mr *MockClusterMockRecorder) ExternalIP(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExternalIP", reflect.TypeOf((*MockCluster)(nil).ExternalIP), arg0, arg1, arg2)
}

// Get mocks base method.
func (m *MockCluster) Get(arg0 context.Context, arg1 *logger.Logger, arg2, arg3 string, arg4 ...option.Option) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Get", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockClusterMockRecorder) Get(arg0, arg1, arg2, arg3 interface{}, arg4 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCluster)(nil).Get), varargs...)
}

// PutString mocks base method.
func (m *MockCluster) PutString(arg0 context.Context, arg1, arg2 string, arg3 fs.FileMode, arg4 ...option.Option) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "PutString", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutString indicates an expected call of PutString.
func (mr *MockClusterMockRecorder) PutString(arg0, arg1, arg2, arg3 interface{}, arg4 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutString", reflect.TypeOf((*MockCluster)(nil).PutString), varargs...)
}

// RunE mocks base method.
func (m *MockCluster) RunE(arg0 context.Context, arg1 option.NodeListOption, arg2 ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RunE", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// RunE indicates an expected call of RunE.
func (mr *MockClusterMockRecorder) RunE(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunE", reflect.TypeOf((*MockCluster)(nil).RunE), varargs...)
}
