/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package git

import (
	"fmt"
	"github.com/devtron-labs/git-sensor/internals/sql"
	mocks "github.com/devtron-labs/git-sensor/pkg/git/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
	"time"
)

func TestWebhookHandlerImpl_upsertWebhookEventParsedData(t *testing.T) {
	type args struct {
		eventId                int
		webhookEventParsedData *sql.WebhookEventParsedData
	}
	tests := []struct {
		name          string
		args          args
		mockExecution func(service *mocks.WebhookEventService)
		wantErr       assert.ErrorAssertionFunc
	}{
		{
			name: "WebhookEventParsedData_Upsert_Success",
			args: args{
				eventId: 1,
				webhookEventParsedData: &sql.WebhookEventParsedData{
					EventId:         1,
					PayloadDataId:   1,
					UniqueId:        "uniqueId",
					EventActionType: "PUSH",
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					CiEnvVariableData: map[string]string{
						"envKey1": "envValue1",
						"envKey2": "envValue2",
					},
				},
			},
			mockExecution: func(service *mocks.WebhookEventService) {
				service.On("GetWebhookParsedEventDataByEventIdAndUniqueId", 1, "uniqueId").
					Return(&sql.WebhookEventParsedData{
						Id:              1,
						EventId:         1,
						PayloadDataId:   1,
						UniqueId:        "uniqueId",
						EventActionType: "PUSH",
						Data: map[string]string{
							"key1": "value1",
						},
						CiEnvVariableData: map[string]string{
							"envKey1": "envValue1",
						},
						CreatedOn: time.Date(2023, time.April, 10, 0, 0, 0, 0, time.UTC),
						UpdatedOn: time.Date(2023, time.April, 10, 0, 0, 0, 0, time.UTC),
					}, nil).Once()
				service.On("UpdateWebhookParsedEventData", mock.AnythingOfType("*sql.WebhookEventParsedData")).
					Return(nil).Once()
				service.AssertNotCalled(t, "SaveWebhookParsedEventData", mock.Anything)
			},
			wantErr: assert.NoError,
		},
		{
			name: "WebhookEventParsedData_Upsert_Error",
			args: args{
				eventId: 1,
				webhookEventParsedData: &sql.WebhookEventParsedData{
					EventId:         1,
					PayloadDataId:   1,
					UniqueId:        "uniqueId",
					EventActionType: "PUSH",
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					CiEnvVariableData: map[string]string{
						"envKey1": "envValue1",
						"envKey2": "envValue2",
					},
				},
			},
			mockExecution: func(service *mocks.WebhookEventService) {
				service.On("GetWebhookParsedEventDataByEventIdAndUniqueId", 1, "uniqueId").
					Return(&sql.WebhookEventParsedData{
						Id:              1,
						EventId:         1,
						PayloadDataId:   1,
						UniqueId:        "uniqueId",
						EventActionType: "PUSH",
						Data: map[string]string{
							"key1": "value1",
						},
						CiEnvVariableData: map[string]string{
							"envKey1": "envValue1",
						},
						CreatedOn: time.Date(2023, time.April, 10, 0, 0, 0, 0, time.UTC),
						UpdatedOn: time.Date(2023, time.April, 10, 0, 0, 0, 0, time.UTC),
					}, nil).Once()
				service.On("UpdateWebhookParsedEventData", mock.AnythingOfType("*sql.WebhookEventParsedData")).
					Return(os.ErrDeadlineExceeded).Once()
				service.AssertNotCalled(t, "SaveWebhookParsedEventData", mock.Anything)
			},
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.EqualError(t, err, os.ErrDeadlineExceeded.Error(), msgAndArgs...)
			},
		},
		{
			name: "WebhookEventParsedData_Save_Success",
			args: args{
				eventId: 1,
				webhookEventParsedData: &sql.WebhookEventParsedData{
					EventId:         1,
					PayloadDataId:   1,
					UniqueId:        "uniqueId",
					EventActionType: "PUSH",
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					CiEnvVariableData: map[string]string{
						"envKey1": "envValue1",
						"envKey2": "envValue2",
					},
				},
			},
			mockExecution: func(service *mocks.WebhookEventService) {
				service.On("GetWebhookParsedEventDataByEventIdAndUniqueId", 1, "uniqueId").
					Return(nil, ErrWebhookEventParsedDataNotFound).Once()
				service.On("SaveWebhookParsedEventData", mock.AnythingOfType("*sql.WebhookEventParsedData")).
					Return(nil).Once()
				service.AssertNotCalled(t, "UpdateWebhookParsedEventData", mock.Anything)
			},
			wantErr: assert.NoError,
		},
		{
			name: "WebhookEventParsedData_Save_Error",
			args: args{
				eventId: 1,
				webhookEventParsedData: &sql.WebhookEventParsedData{
					EventId:         1,
					PayloadDataId:   1,
					UniqueId:        "uniqueId",
					EventActionType: "PUSH",
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					CiEnvVariableData: map[string]string{
						"envKey1": "envValue1",
						"envKey2": "envValue2",
					},
				},
			},
			mockExecution: func(service *mocks.WebhookEventService) {
				service.On("GetWebhookParsedEventDataByEventIdAndUniqueId", 1, "uniqueId").
					Return(nil, ErrWebhookEventParsedDataNotFound).Once()
				service.On("SaveWebhookParsedEventData", mock.AnythingOfType("*sql.WebhookEventParsedData")).
					Return(os.ErrDeadlineExceeded).Once()
				service.AssertNotCalled(t, "UpdateWebhookParsedEventData", mock.Anything)
			},
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.EqualError(t, err, os.ErrDeadlineExceeded.Error(), msgAndArgs...)
			},
		},
		{
			name: "GetWebhookParsedEventDataByEventIdAndUniqueId_PG_Connection_Error",
			args: args{
				eventId: 1,
				webhookEventParsedData: &sql.WebhookEventParsedData{
					EventId:         1,
					PayloadDataId:   1,
					UniqueId:        "uniqueId",
					EventActionType: "PUSH",
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					CiEnvVariableData: map[string]string{
						"envKey1": "envValue1",
						"envKey2": "envValue2",
					},
				},
			},
			mockExecution: func(service *mocks.WebhookEventService) {
				service.On("GetWebhookParsedEventDataByEventIdAndUniqueId", 1, "uniqueId").
					Return(nil, os.ErrDeadlineExceeded).Once()
				service.AssertNotCalled(t, "SaveWebhookParsedEventData", mock.Anything)
				service.AssertNotCalled(t, "UpdateWebhookParsedEventData", mock.Anything)
			},
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.EqualError(t, err, os.ErrDeadlineExceeded.Error(), msgAndArgs...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := getNewWebhookHandlerImplImpl(t, tt.mockExecution)
			tt.wantErr(t, impl.upsertWebhookEventParsedData(tt.args.eventId, tt.args.webhookEventParsedData), fmt.Sprintf("upsertWebhookEventParsedData(%v, %v)", tt.args.eventId, tt.args.webhookEventParsedData))
		})
	}
}

func getNewWebhookHandlerImplImpl(
	t *testing.T,
	WebhookEventServiceMockExecution func(service *mocks.WebhookEventService),
) *WebhookHandlerImpl {
	sugaredLogger := getLogger()
	webhookEventService := getWebhookEventServiceMock(t)
	if WebhookEventServiceMockExecution != nil {
		WebhookEventServiceMockExecution(webhookEventService)
	}
	return NewWebhookHandlerImpl(
		sugaredLogger,
		webhookEventService,
		nil,
	)
}

func getWebhookEventServiceMock(t *testing.T) *mocks.WebhookEventService {
	return mocks.NewWebhookEventService(t)
}
