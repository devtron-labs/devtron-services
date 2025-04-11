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
	"github.com/devtron-labs/git-sensor/internals/logger"
	"github.com/devtron-labs/git-sensor/internals/sql"
	"github.com/devtron-labs/git-sensor/internals/sql/mocks"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"testing"
	"time"
)

func TestWebhookEventServiceImpl_GetWebhookParsedEventDataByEventIdAndUniqueId(t *testing.T) {

	type args struct {
		eventId  int
		uniqueId string
	}

	tests := []struct {
		name          string
		args          args
		want          *sql.WebhookEventParsedData
		mockExecution func(*mocks.WebhookEventParsedDataRepository)
		wantErr       assert.ErrorAssertionFunc
	}{
		{
			name: "WebhookEventParsedData_Found",
			args: args{
				eventId:  1,
				uniqueId: "uniqueId",
			},
			want: &sql.WebhookEventParsedData{
				Id:              1,
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
				CreatedOn: time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC),
				UpdatedOn: time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC),
			},
			wantErr: assert.NoError,
			mockExecution: func(mockImpl *mocks.WebhookEventParsedDataRepository) {
				mockImpl.
					On("GetWebhookParsedEventDataByEventIdAndUniqueId", 1, "uniqueId").
					Return(&sql.WebhookEventParsedData{
						Id:              1,
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
						CreatedOn: time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC),
						UpdatedOn: time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC),
					}, nil).Once()
			},
		},
		{
			name: "WebhookEventParsedData_NotFound",
			args: args{
				eventId:  1,
				uniqueId: "uniqueId",
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.EqualError(t, err, ErrWebhookEventParsedDataNotFound.Error(), msgAndArgs...)
			},
			mockExecution: func(mockImpl *mocks.WebhookEventParsedDataRepository) {
				mockImpl.
					On("GetWebhookParsedEventDataByEventIdAndUniqueId", 1, "uniqueId").
					Return(nil, pg.ErrNoRows).Once()
			},
		},
		{
			name: "Empty_UniqueId",
			args: args{
				eventId:  1,
				uniqueId: "",
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.EqualError(t, err, ErrWebhookEventParsedDataNotFound.Error(), msgAndArgs...)
			},
			mockExecution: func(mockImpl *mocks.WebhookEventParsedDataRepository) {
				mockImpl.AssertNotCalled(t, "GetWebhookParsedEventDataByEventIdAndUniqueId", 2)
			},
		},
		{
			name: "PG_Timeout_Error",
			args: args{
				eventId:  1,
				uniqueId: "uniqueId",
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.EqualError(t, err, os.ErrDeadlineExceeded.Error(), msgAndArgs...)
			},
			mockExecution: func(mockImpl *mocks.WebhookEventParsedDataRepository) {
				mockImpl.
					On("GetWebhookParsedEventDataByEventIdAndUniqueId", 1, "uniqueId").
					Return(nil, os.ErrDeadlineExceeded).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := getWebhookEventServiceImpl(t, tt.mockExecution, nil)
			got, err := impl.GetWebhookParsedEventDataByEventIdAndUniqueId(tt.args.eventId, tt.args.uniqueId)
			if !tt.wantErr(t, err, fmt.Sprintf("GetWebhookParsedEventDataByEventIdAndUniqueId(%v, %v)", tt.args.eventId, tt.args.uniqueId)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetWebhookParsedEventDataByEventIdAndUniqueId(%v, %v)", tt.args.eventId, tt.args.uniqueId)
		})
	}
}

func TestWebhookEventServiceImpl_getCiPipelineMaterialWebhookDataMapping(t *testing.T) {
	type args struct {
		ciPipelineMaterialId int
		webhookParsedDataId  int
	}
	tests := []struct {
		name          string
		args          args
		want          *sql.CiPipelineMaterialWebhookDataMapping
		isNewEntry    bool
		mockExecution func(*mocks.WebhookEventDataMappingRepository)
		wantErr       assert.ErrorAssertionFunc
	}{
		{
			name: "WebhookEventDataMapping_Found",
			args: args{
				ciPipelineMaterialId: 1,
				webhookParsedDataId:  1,
			},
			want: &sql.CiPipelineMaterialWebhookDataMapping{
				Id:                   1,
				CiPipelineMaterialId: 1,
				WebhookDataId:        1,
				ConditionMatched:     true,
				IsActive:             true,
				CreatedOn:            time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC),
				UpdatedOn:            time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC),
			},
			isNewEntry: false,
			wantErr:    assert.NoError,
			mockExecution: func(mockImpl *mocks.WebhookEventDataMappingRepository) {
				mockImpl.
					On("GetCiPipelineMaterialWebhookDataMapping", 1, 1).
					Return(&sql.CiPipelineMaterialWebhookDataMapping{
						Id:                   1,
						CiPipelineMaterialId: 1,
						WebhookDataId:        1,
						ConditionMatched:     true,
						IsActive:             true,
						CreatedOn:            time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC),
						UpdatedOn:            time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC),
					}, nil).Once()
			},
		},
		{
			name: "WebhookEventDataMapping_NotFound",
			args: args{
				ciPipelineMaterialId: 1,
				webhookParsedDataId:  1,
			},
			want:       nil,
			isNewEntry: true,
			wantErr:    assert.NoError,
			mockExecution: func(mockImpl *mocks.WebhookEventDataMappingRepository) {
				mockImpl.
					On("GetCiPipelineMaterialWebhookDataMapping", 1, 1).
					Return(nil, pg.ErrNoRows).Once()
			},
		},
		{
			name: "PG_Timeout_Error",
			args: args{
				ciPipelineMaterialId: 1,
				webhookParsedDataId:  1,
			},
			want:       nil,
			isNewEntry: false,
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.EqualError(t, err, os.ErrDeadlineExceeded.Error(), msgAndArgs...)
			},
			mockExecution: func(mockImpl *mocks.WebhookEventDataMappingRepository) {
				mockImpl.
					On("GetCiPipelineMaterialWebhookDataMapping", 1, 1).
					Return(nil, os.ErrDeadlineExceeded).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := getWebhookEventServiceImpl(t, nil, tt.mockExecution)
			got, isNewMapping, err := impl.getCiPipelineMaterialWebhookDataMapping(tt.args.ciPipelineMaterialId, tt.args.webhookParsedDataId)
			if !tt.wantErr(t, err, fmt.Sprintf("getCiPipelineMaterialWebhookDataMapping(%v, %v)", tt.args.ciPipelineMaterialId, tt.args.webhookParsedDataId)) {
				return
			}
			if !tt.wantErr(t, err, fmt.Sprintf("getCiPipelineMaterialWebhookDataMapping(%v, %v)", tt.args.ciPipelineMaterialId, tt.args.webhookParsedDataId)) {
				return
			}
			assert.Equalf(t, tt.isNewEntry, isNewMapping, "getCiPipelineMaterialWebhookDataMapping(%v, %v)", tt.args.ciPipelineMaterialId, tt.args.webhookParsedDataId)
			assert.Equalf(t, tt.want, got, "getCiPipelineMaterialWebhookDataMapping(%v, %v)", tt.args.ciPipelineMaterialId, tt.args.webhookParsedDataId)
		})
	}
}

func getWebhookEventServiceImpl(
	t *testing.T,
	webhookEventParsedDataMockExecution func(*mocks.WebhookEventParsedDataRepository),
	webhookEventDataMappingMockExecution func(*mocks.WebhookEventDataMappingRepository),
) *WebhookEventServiceImpl {
	sugaredLogger := getLogger()
	webhookEventDataMappingRepository := getWebhookEventDataMappingRepository(t)
	if webhookEventDataMappingMockExecution != nil {
		webhookEventDataMappingMockExecution(webhookEventDataMappingRepository)
	}
	webhookEventParsedDataRepositoryMock := getWebhookEventParsedDataRepositoryMock(t)
	if webhookEventParsedDataMockExecution != nil {
		webhookEventParsedDataMockExecution(webhookEventParsedDataRepositoryMock)
	}
	return NewWebhookEventServiceImpl(
		sugaredLogger,
		nil,
		webhookEventParsedDataRepositoryMock,
		webhookEventDataMappingRepository, nil, nil, nil, nil,
	)
}

func getLogger() *zap.SugaredLogger {
	return logger.NewSugaredLogger()
}

func getWebhookEventParsedDataRepositoryMock(t *testing.T) *mocks.WebhookEventParsedDataRepository {
	return mocks.NewWebhookEventParsedDataRepository(t)
}

func getWebhookEventDataMappingRepository(t *testing.T) *mocks.WebhookEventDataMappingRepository {
	return mocks.NewWebhookEventDataMappingRepository(t)
}
