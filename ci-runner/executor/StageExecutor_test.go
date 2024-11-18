/*
 * Copyright (c) 2024. Devtron Inc.
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

package executor

import (
	"github.com/devtron-labs/ci-runner/helper"
	commonBean "github.com/devtron-labs/common-lib/ci-runner/bean"
	"reflect"
	"testing"
)

func Test_deduceVariables(t *testing.T) {
	type args struct {
		desiredVars          []*commonBean.VariableObject
		globalVars           map[string]string
		preeCiStageVariable  map[int]map[string]*commonBean.VariableObject
		postCiStageVariables map[int]map[string]*commonBean.VariableObject
	}
	tests := []struct {
		name    string
		args    args
		want    []*commonBean.VariableObject
		wantErr bool
	}{
		{name: "only value type",
			args: args{
				desiredVars: []*commonBean.VariableObject{&commonBean.VariableObject{Name: "age", Value: "20", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeNumber},
					&commonBean.VariableObject{Name: "name", Value: "test", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeString},
					&commonBean.VariableObject{Name: "status", Value: "true", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeBool}},
				globalVars:           nil,
				preeCiStageVariable:  nil,
				postCiStageVariables: nil,
			},
			wantErr: false,
			want: []*commonBean.VariableObject{&commonBean.VariableObject{Name: "age", Value: "20", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeNumber},
				&commonBean.VariableObject{Name: "name", Value: "test", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeString},
				&commonBean.VariableObject{Name: "status", Value: "true", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeBool}},
		}, {name: "from global",
			args: args{
				desiredVars: []*commonBean.VariableObject{&commonBean.VariableObject{Name: "age", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeNumber, ReferenceVariableName: "age"},
					&commonBean.VariableObject{Name: "name", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeString, ReferenceVariableName: "my-name"},
					&commonBean.VariableObject{Name: "status", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeBool, ReferenceVariableName: "status"}},
				globalVars:           map[string]string{"age": "20", "my-name": "test", "status": "true"},
				preeCiStageVariable:  nil,
				postCiStageVariables: nil,
			},
			wantErr: false,
			want: []*commonBean.VariableObject{&commonBean.VariableObject{Name: "age", Value: "20", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeNumber, TypedValue: float64(20), ReferenceVariableName: "age"},
				&commonBean.VariableObject{Name: "name", Value: "test", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeString, TypedValue: "test", ReferenceVariableName: "my-name"},
				&commonBean.VariableObject{Name: "status", Value: "true", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeBool, TypedValue: true, ReferenceVariableName: "status"}},
		}, {name: "REF_PRE_CI",
			args: args{
				desiredVars: []*commonBean.VariableObject{&commonBean.VariableObject{Name: "age", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeNumber, ReferenceVariableName: "age", ReferenceVariableStepIndex: 1},
					&commonBean.VariableObject{Name: "name", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeString, ReferenceVariableName: "my-name", ReferenceVariableStepIndex: 1},
					&commonBean.VariableObject{Name: "status", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeBool, ReferenceVariableName: "status", ReferenceVariableStepIndex: 1}},
				globalVars: map[string]string{"age": "22", "my-name": "test1", "status": "false"},
				preeCiStageVariable: map[int]map[string]*commonBean.VariableObject{1: {"age": &commonBean.VariableObject{Name: "age", Value: "20"},
					"my-name": &commonBean.VariableObject{Name: "my-name", Value: "test"},
					"status":  &commonBean.VariableObject{Name: "status", Value: "true"},
				}},
				postCiStageVariables: nil,
			},
			wantErr: false,
			want: []*commonBean.VariableObject{&commonBean.VariableObject{Name: "age", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeNumber, ReferenceVariableName: "age", Value: "20", TypedValue: float64(20), ReferenceVariableStepIndex: 1},
				&commonBean.VariableObject{Name: "name", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeString, ReferenceVariableName: "my-name", Value: "test", TypedValue: "test", ReferenceVariableStepIndex: 1},
				&commonBean.VariableObject{Name: "status", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeBool, ReferenceVariableName: "status", Value: "true", TypedValue: true, ReferenceVariableStepIndex: 1}},
		}, {name: "VARIABLE_TYPE_REF_POST_CI",
			args: args{
				desiredVars: []*commonBean.VariableObject{&commonBean.VariableObject{Name: "age", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeNumber, ReferenceVariableName: "age", ReferenceVariableStepIndex: 1},
					&commonBean.VariableObject{Name: "name", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeString, ReferenceVariableName: "my-name", ReferenceVariableStepIndex: 1},
					&commonBean.VariableObject{Name: "status", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeBool, ReferenceVariableName: "status", ReferenceVariableStepIndex: 1}},
				globalVars: map[string]string{"age": "22", "my-name": "test1", "status": "false"},
				postCiStageVariables: map[int]map[string]*commonBean.VariableObject{1: {"age": &commonBean.VariableObject{Name: "age", Value: "20"},
					"my-name": &commonBean.VariableObject{Name: "my-name", Value: "test"},
					"status":  &commonBean.VariableObject{Name: "status", Value: "true"},
				}},
				preeCiStageVariable: nil,
			},
			wantErr: false,
			want: []*commonBean.VariableObject{&commonBean.VariableObject{Name: "age", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeNumber, ReferenceVariableName: "age", Value: "20", TypedValue: float64(20), ReferenceVariableStepIndex: 1},
				&commonBean.VariableObject{Name: "name", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeString, ReferenceVariableName: "my-name", Value: "test", TypedValue: "test", ReferenceVariableStepIndex: 1},
				&commonBean.VariableObject{Name: "status", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeBool, ReferenceVariableName: "status", Value: "true", TypedValue: true, ReferenceVariableStepIndex: 1}},
		},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deduceVariables(tt.args.desiredVars, tt.args.globalVars, tt.args.preeCiStageVariable, tt.args.postCiStageVariables, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("deduceVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deduceVariables() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunCiSteps(t *testing.T) {
	type args struct {
		stageType                  helper.StepType
		req                        *helper.CommonWorkflowRequest
		globalEnvironmentVariables map[string]string
		preeCiStageVariable        map[int]map[string]*commonBean.VariableObject
	}
	tests := []struct {
		name                       string
		args                       args
		wantPreeCiStageVariableOut map[int]map[string]*commonBean.VariableObject
		wantPostCiStageVariable    map[int]map[string]*commonBean.VariableObject
		wantErr                    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		stageExecutor := NewStageExecutorImpl(nil, nil)
		t.Run(tt.name, func(t *testing.T) {
			gotPreeCiStageVariableOut, gotPostCiStageVariable, _, err := stageExecutor.RunCiCdSteps(tt.args.stageType, nil, tt.args.req.PreCiSteps, nil, tt.args.globalEnvironmentVariables, tt.args.preeCiStageVariable)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCiCdSteps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPreeCiStageVariableOut, tt.wantPreeCiStageVariableOut) {
				t.Errorf("RunCiCdSteps() gotPreeCiStageVariableOut = %v, want %v", gotPreeCiStageVariableOut, tt.wantPreeCiStageVariableOut)
			}
			if !reflect.DeepEqual(gotPostCiStageVariable, tt.wantPostCiStageVariable) {
				t.Errorf("RunCiCdSteps() gotPostCiStageVariable = %v, want %v", gotPostCiStageVariable, tt.wantPostCiStageVariable)
			}
		})
	}
}
