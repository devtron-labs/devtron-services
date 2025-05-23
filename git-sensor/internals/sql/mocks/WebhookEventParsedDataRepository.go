// Code generated by mockery v2.42.0. DO NOT EDIT.

package mocks

import (
	sql "github.com/devtron-labs/git-sensor/internals/sql"
	mock "github.com/stretchr/testify/mock"
)

// WebhookEventParsedDataRepository is an autogenerated mock type for the WebhookEventParsedDataRepository type
type WebhookEventParsedDataRepository struct {
	mock.Mock
}

// GetWebhookEventParsedDataById provides a mock function with given fields: id
func (_m *WebhookEventParsedDataRepository) GetWebhookEventParsedDataById(id int) (*sql.WebhookEventParsedData, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for GetWebhookEventParsedDataById")
	}

	var r0 *sql.WebhookEventParsedData
	var r1 error
	if rf, ok := ret.Get(0).(func(int) (*sql.WebhookEventParsedData, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(int) *sql.WebhookEventParsedData); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sql.WebhookEventParsedData)
		}
	}

	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetWebhookEventParsedDataByIds provides a mock function with given fields: ids, limit
func (_m *WebhookEventParsedDataRepository) GetWebhookEventParsedDataByIds(ids []int, limit int) ([]*sql.WebhookEventParsedData, error) {
	ret := _m.Called(ids, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetWebhookEventParsedDataByIds")
	}

	var r0 []*sql.WebhookEventParsedData
	var r1 error
	if rf, ok := ret.Get(0).(func([]int, int) ([]*sql.WebhookEventParsedData, error)); ok {
		return rf(ids, limit)
	}
	if rf, ok := ret.Get(0).(func([]int, int) []*sql.WebhookEventParsedData); ok {
		r0 = rf(ids, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*sql.WebhookEventParsedData)
		}
	}

	if rf, ok := ret.Get(1).(func([]int, int) error); ok {
		r1 = rf(ids, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetWebhookParsedEventDataByEventIdAndUniqueId provides a mock function with given fields: eventId, uniqueId
func (_m *WebhookEventParsedDataRepository) GetWebhookParsedEventDataByEventIdAndUniqueId(eventId int, uniqueId string) (*sql.WebhookEventParsedData, error) {
	ret := _m.Called(eventId, uniqueId)

	if len(ret) == 0 {
		panic("no return value specified for GetWebhookParsedEventDataByEventIdAndUniqueId")
	}

	var r0 *sql.WebhookEventParsedData
	var r1 error
	if rf, ok := ret.Get(0).(func(int, string) (*sql.WebhookEventParsedData, error)); ok {
		return rf(eventId, uniqueId)
	}
	if rf, ok := ret.Get(0).(func(int, string) *sql.WebhookEventParsedData); ok {
		r0 = rf(eventId, uniqueId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sql.WebhookEventParsedData)
		}
	}

	if rf, ok := ret.Get(1).(func(int, string) error); ok {
		r1 = rf(eventId, uniqueId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveWebhookParsedEventData provides a mock function with given fields: webhookEventParsedData
func (_m *WebhookEventParsedDataRepository) SaveWebhookParsedEventData(webhookEventParsedData *sql.WebhookEventParsedData) error {
	ret := _m.Called(webhookEventParsedData)

	if len(ret) == 0 {
		panic("no return value specified for SaveWebhookParsedEventData")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*sql.WebhookEventParsedData) error); ok {
		r0 = rf(webhookEventParsedData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateWebhookParsedEventData provides a mock function with given fields: webhookEventParsedData
func (_m *WebhookEventParsedDataRepository) UpdateWebhookParsedEventData(webhookEventParsedData *sql.WebhookEventParsedData) error {
	ret := _m.Called(webhookEventParsedData)

	if len(ret) == 0 {
		panic("no return value specified for UpdateWebhookParsedEventData")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*sql.WebhookEventParsedData) error); ok {
		r0 = rf(webhookEventParsedData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewWebhookEventParsedDataRepository creates a new instance of WebhookEventParsedDataRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWebhookEventParsedDataRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *WebhookEventParsedDataRepository {
	mock := &WebhookEventParsedDataRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
