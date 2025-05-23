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

package k8s_test

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/common-lib/utils/k8s/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/rest"
	"testing"
)

type k8sServiceTest struct {
	mockDefaultTransport *mocks.HttpTransportInterface
	mockCustomTransport  *mocks.HttpTransportInterface
	mockKubeConfig       *mocks.KubeConfigInterface
	k8sService           *k8s.K8sServiceImpl
}

func initK8sServiceTest(t *testing.T) *k8sServiceTest {
	// Mock the dependencies
	mockDefaultTransport := mocks.NewHttpTransportInterface(t)
	mockCustomTransport := mocks.NewHttpTransportInterface(t)

	k8sService := intiK8sUtil().
		SetCustomHttpClientConfig(mockCustomTransport).
		SetDefaultHttpClientConfig(mockDefaultTransport)

	return &k8sServiceTest{
		mockDefaultTransport: mockDefaultTransport,
		mockCustomTransport:  mockCustomTransport,
		k8sService:           k8sService,
	}
}

func TestK8sService_AllMethodsWithOpts(t *testing.T) {
	tests := []struct {
		name           string
		method         func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error
		useDefault     bool
		expectedCalled string
	}{
		{
			name: "GetCoreV1ClientInCluster with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.GetCoreV1ClientInCluster(opts...)
				return err
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetCoreV1ClientInCluster with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.GetCoreV1ClientInCluster(opts...)
				return err
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sDiscoveryClientInCluster with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.GetK8sDiscoveryClientInCluster(opts...)
				return err
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sDiscoveryClientInCluster with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.GetK8sDiscoveryClientInCluster(opts...)
				return err
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "DeletePodByLabel with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				return k8sService.DeletePodByLabel("namespace", "label", getDefaultClusterConfig(), opts...)
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "DeletePodByLabel with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				return k8sService.DeletePodByLabel("namespace", "label", getDefaultClusterConfig(), opts...)
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "CreateJob with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				return k8sService.CreateJob("namespace", "name", getDefaultClusterConfig(), nil, opts...)
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "CreateJob with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				return k8sService.CreateJob("namespace", "name", getDefaultClusterConfig(), nil, opts...)
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "DiscoveryClientGetLiveZCall with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.DiscoveryClientGetLiveZCall(getDefaultClusterConfig(), opts...)
				return err
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "DiscoveryClientGetLiveZCall with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.DiscoveryClientGetLiveZCall(getDefaultClusterConfig(), opts...)
				return err
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "DeleteJob with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				return k8sService.DeleteJob("namespace", "name", getDefaultClusterConfig(), opts...)
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "DeleteJob with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				return k8sService.DeleteJob("namespace", "name", getDefaultClusterConfig(), opts...)
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sInClusterRestConfig with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.GetK8sInClusterRestConfig(opts...)
				return err
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sInClusterRestConfig with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.GetK8sInClusterRestConfig(opts...)
				return err
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sConfigAndClients with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, _, _, err := k8sService.GetK8sConfigAndClients(getDefaultClusterConfig(), opts...)
				return err
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sConfigAndClients with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, _, _, err := k8sService.GetK8sConfigAndClients(getDefaultClusterConfig(), opts...)
				return err
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sInClusterConfigAndDynamicClients with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, _, _, err := k8sService.GetK8sInClusterConfigAndDynamicClients(opts...)
				return err
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sInClusterConfigAndDynamicClients with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, _, _, err := k8sService.GetK8sInClusterConfigAndDynamicClients(opts...)
				return err
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetRestConfigByCluster with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.GetRestConfigByCluster(getDefaultClusterConfig(), opts...)
				return err
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetRestConfigByCluster with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.GetRestConfigByCluster(getDefaultClusterConfig(), opts...)
				return err
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sInClusterConfigAndClients with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, _, _, err := k8sService.GetK8sInClusterConfigAndClients(opts...)
				return err
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sInClusterConfigAndClients with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, _, _, err := k8sService.GetK8sInClusterConfigAndClients(opts...)
				return err
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sDiscoveryClient with default transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.GetK8sDiscoveryClient(getDefaultClusterConfig(), opts...)
				return err
			},
			useDefault:     true,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
		{
			name: "GetK8sDiscoveryClient with overridden transport",
			method: func(k8sService *k8s.K8sServiceImpl, opts ...k8s.K8sServiceOpts) error {
				_, err := k8sService.GetK8sDiscoveryClient(getDefaultClusterConfig(), opts...)
				return err
			},
			useDefault:     false,
			expectedCalled: "OverrideConfigWithCustomTransport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the dependencies
			impl := initK8sServiceTest(t)

			// Arrange
			if tt.useDefault {
				impl.mockDefaultTransport.On("OverrideConfigWithCustomTransport", mock.Anything).Return(&rest.Config{}, nil).Once()
			} else {
				impl.mockCustomTransport.On("OverrideConfigWithCustomTransport", mock.Anything).Return(&rest.Config{}, nil).Once()
			}
			// Act
			var opts []k8s.K8sServiceOpts
			if tt.useDefault {
				opts = append(opts, k8s.WithDefaultHttpTransport())
			} else {
				opts = append(opts, k8s.WithOverriddenHttpTransport())
			}
			_ = tt.method(impl.k8sService, opts...)

			// Assert
			if tt.useDefault {
				impl.mockDefaultTransport.AssertCalled(t, tt.expectedCalled, mock.Anything)
				impl.mockDefaultTransport.AssertNumberOfCalls(t, tt.expectedCalled, 1)
				impl.mockCustomTransport.AssertNotCalled(t, tt.expectedCalled, mock.Anything)
			} else {
				impl.mockCustomTransport.AssertCalled(t, tt.expectedCalled, mock.Anything)
				impl.mockCustomTransport.AssertNumberOfCalls(t, tt.expectedCalled, 1)
				impl.mockDefaultTransport.AssertNotCalled(t, tt.expectedCalled, mock.Anything)
			}
		})
	}
}
