// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	sitechecker "github.com/clambin/hostchecker/sitechecker"

	mock "github.com/stretchr/testify/mock"
)

// Checker is an autogenerated mock type for the Checker type
type Checker struct {
	mock.Mock
}

// Check provides a mock function with given fields:
func (_m *Checker) Check() (*sitechecker.Stats, error) {
	ret := _m.Called()

	var r0 *sitechecker.Stats
	if rf, ok := ret.Get(0).(func() *sitechecker.Stats); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sitechecker.Stats)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewChecker interface {
	mock.TestingT
	Cleanup(func())
}

// NewChecker creates a new instance of Checker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewChecker(t mockConstructorTestingTNewChecker) *Checker {
	mock := &Checker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
