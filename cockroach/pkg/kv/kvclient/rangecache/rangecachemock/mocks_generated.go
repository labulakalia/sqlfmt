// Code generated by MockGen. DO NOT EDIT.
// Source: sqlfmt/cockroach/pkg/kv/kvclient/rangecache (interfaces: RangeDescriptorDB)

// Package rangecachemock is a generated GoMock package.
package rangecachemock

import (
	context "context"
	reflect "reflect"

	roachpb "github.com/labulakalia/sqlfmt/cockroach/pkg/roachpb"
	gomock "github.com/golang/mock/gomock"
)

// MockRangeDescriptorDB is a mock of RangeDescriptorDB interface.
type MockRangeDescriptorDB struct {
	ctrl     *gomock.Controller
	recorder *MockRangeDescriptorDBMockRecorder
}

// MockRangeDescriptorDBMockRecorder is the mock recorder for MockRangeDescriptorDB.
type MockRangeDescriptorDBMockRecorder struct {
	mock *MockRangeDescriptorDB
}

// NewMockRangeDescriptorDB creates a new mock instance.
func NewMockRangeDescriptorDB(ctrl *gomock.Controller) *MockRangeDescriptorDB {
	mock := &MockRangeDescriptorDB{ctrl: ctrl}
	mock.recorder = &MockRangeDescriptorDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRangeDescriptorDB) EXPECT() *MockRangeDescriptorDBMockRecorder {
	return m.recorder
}

// FirstRange mocks base method.
func (m *MockRangeDescriptorDB) FirstRange() (*roachpb.RangeDescriptor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FirstRange")
	ret0, _ := ret[0].(*roachpb.RangeDescriptor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FirstRange indicates an expected call of FirstRange.
func (mr *MockRangeDescriptorDBMockRecorder) FirstRange() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FirstRange", reflect.TypeOf((*MockRangeDescriptorDB)(nil).FirstRange))
}

// RangeLookup mocks base method.
func (m *MockRangeDescriptorDB) RangeLookup(arg0 context.Context, arg1 roachpb.RKey, arg2 bool) ([]roachpb.RangeDescriptor, []roachpb.RangeDescriptor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RangeLookup", arg0, arg1, arg2)
	ret0, _ := ret[0].([]roachpb.RangeDescriptor)
	ret1, _ := ret[1].([]roachpb.RangeDescriptor)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// RangeLookup indicates an expected call of RangeLookup.
func (mr *MockRangeDescriptorDBMockRecorder) RangeLookup(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RangeLookup", reflect.TypeOf((*MockRangeDescriptorDB)(nil).RangeLookup), arg0, arg1, arg2)
}
