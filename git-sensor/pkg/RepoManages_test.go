package pkg

import (
	"fmt"
	"github.com/devtron-labs/git-sensor/internals/logger"
	"github.com/devtron-labs/git-sensor/internals/sql"
	"github.com/devtron-labs/git-sensor/internals/sql/mocks"
	"github.com/devtron-labs/git-sensor/pkg/git"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestRepoManagerImpl_GetWebhookAndCiDataById(t *testing.T) {
	type args struct {
		id                   int
		ciPipelineMaterialId int
	}
	tests := []struct {
		name                                 string
		args                                 args
		want                                 *git.WebhookAndCiData
		webhookEventParsedDataMockExecution  func(*mocks.WebhookEventParsedDataRepository)
		webhookEventDataMappingMockExecution func(*mocks.WebhookEventDataMappingRepository)
		wantErr                              assert.ErrorAssertionFunc
	}{
		{
			name: "WebhookEvent_ParsedData_And_DataMappings_Found",
			args: args{
				id:                   1,
				ciPipelineMaterialId: 1,
			},
			want: &git.WebhookAndCiData{
				ExtraEnvironmentVariables: map[string]string{
					"VAR1": "envValue1",
					"VAR2": "envValue2",
					"VAR3": "filterValue3",
					"VAR4": "filterValue4",
				},
				WebhookData: &git.WebhookData{
					Id:              1,
					EventActionType: "PUSH",
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			wantErr: assert.NoError,
			webhookEventParsedDataMockExecution: func(mockImpl *mocks.WebhookEventParsedDataRepository) {
				mockImpl.
					On("GetWebhookEventParsedDataById", 1).
					Return(&sql.WebhookEventParsedData{
						Id:              1,
						EventActionType: "PUSH",
						Data: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
						CiEnvVariableData: map[string]string{
							"VAR1": "envValue1",
							"VAR2": "envValue2",
						},
					}, nil).Once()
			},
			webhookEventDataMappingMockExecution: func(mockImpl *mocks.WebhookEventDataMappingRepository) {
				mockImpl.
					On("GetWebhookPayloadFilterDataForPipelineMaterialId", 1, 1).
					Return(&sql.CiPipelineMaterialWebhookDataMapping{
						Id:                   1,
						CiPipelineMaterialId: 1,
						FilterResults: []*sql.CiPipelineMaterialWebhookDataMappingFilterResult{
							{
								MatchedGroups: map[string]string{
									"VAR1": "filterValue1",
									"VAR2": "filterValue2",
								},
							},
							{
								MatchedGroups: map[string]string{
									"VAR3": "filterValue3",
									"VAR4": "filterValue4",
								},
							},
						},
					}, nil).Once()
			},
		},
		{
			name: "WebhookEvent_ParsedData_Found_But_DataMappings_NotFound",
			args: args{
				id:                   1,
				ciPipelineMaterialId: 1,
			},
			want: &git.WebhookAndCiData{
				WebhookData: &git.WebhookData{
					Id:              1,
					EventActionType: "PUSH",
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			wantErr: assert.NoError,
			webhookEventParsedDataMockExecution: func(mockImpl *mocks.WebhookEventParsedDataRepository) {
				mockImpl.
					On("GetWebhookEventParsedDataById", 1).
					Return(&sql.WebhookEventParsedData{
						Id:              1,
						EventActionType: "PUSH",
						Data: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
						CiEnvVariableData: map[string]string{
							"VAR1": "envValue1",
							"VAR2": "envValue2",
						},
					}, nil).Once()
			},
			webhookEventDataMappingMockExecution: func(mockImpl *mocks.WebhookEventDataMappingRepository) {
				mockImpl.
					On("GetWebhookPayloadFilterDataForPipelineMaterialId", 1, 1).
					Return(nil, pg.ErrNoRows).Once()
			},
		},
		{
			name: "WebhookEvent_ParsedData_And_DataMappings_NotFound",
			args: args{
				id:                   1,
				ciPipelineMaterialId: 1,
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.EqualError(t, err, pg.ErrNoRows.Error(), msgAndArgs...)
			},
			webhookEventParsedDataMockExecution: func(mockImpl *mocks.WebhookEventParsedDataRepository) {
				mockImpl.
					On("GetWebhookEventParsedDataById", 1).
					Return(nil, pg.ErrNoRows).Once()
			},
			webhookEventDataMappingMockExecution: func(mockImpl *mocks.WebhookEventDataMappingRepository) {
				mockImpl.
					AssertNotCalled(t, "GetWebhookPayloadFilterDataForPipelineMaterialId", 1, 1)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := getRepoManagerImpl(t, tt.webhookEventParsedDataMockExecution, tt.webhookEventDataMappingMockExecution)
			got, err := impl.GetWebhookAndCiDataById(tt.args.id, tt.args.ciPipelineMaterialId)
			if !tt.wantErr(t, err, fmt.Sprintf("GetWebhookAndCiDataById(%v, %v)", tt.args.id, tt.args.ciPipelineMaterialId)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetWebhookAndCiDataById(%v, %v)", tt.args.id, tt.args.ciPipelineMaterialId)
		})
	}
}

func getRepoManagerImpl(
	t *testing.T,
	webhookEventParsedDataMockExecution func(*mocks.WebhookEventParsedDataRepository),
	webhookEventDataMappingMockExecution func(*mocks.WebhookEventDataMappingRepository),
) *RepoManagerImpl {
	sugaredLogger := getLogger()
	webhookEventDataMappingRepository := getWebhookEventDataMappingRepository(t)
	if webhookEventDataMappingMockExecution != nil {
		webhookEventDataMappingMockExecution(webhookEventDataMappingRepository)
	}
	webhookEventParsedDataRepositoryMock := getWebhookEventParsedDataRepositoryMock(t)
	if webhookEventParsedDataMockExecution != nil {
		webhookEventParsedDataMockExecution(webhookEventParsedDataRepositoryMock)
	}
	webhookEventBeanConverterImpl := git.NewWebhookEventBeanConverterImpl()
	return NewRepoManagerImpl(
		sugaredLogger,
		nil, nil, nil, nil, nil, nil, nil, nil,
		webhookEventParsedDataRepositoryMock,
		webhookEventDataMappingRepository,
		webhookEventBeanConverterImpl, nil, nil,
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
