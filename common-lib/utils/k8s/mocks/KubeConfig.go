// Code generated by mockery v2.42.0. DO NOT EDIT.

package mocks

import (
	http "net/http"

	dynamic "k8s.io/client-go/dynamic"

	k8s "github.com/devtron-labs/common-lib/utils/k8s"

	kubernetes "k8s.io/client-go/kubernetes"

	mock "github.com/stretchr/testify/mock"

	rest "k8s.io/client-go/rest"
)

// KubeConfigInterface is an autogenerated mock type for the KubeConfigInterface type
type KubeConfigInterface struct {
	mock.Mock
}

// GetK8sConfigAndClients provides a mock function with given fields: clusterConfig
func (_m *KubeConfigInterface) GetK8sConfigAndClients(clusterConfig *k8s.ClusterConfig) (*rest.Config, *http.Client, *kubernetes.Clientset, error) {
	ret := _m.Called(clusterConfig)

	if len(ret) == 0 {
		panic("no return value specified for GetK8sConfigAndClients")
	}

	var r0 *rest.Config
	var r1 *http.Client
	var r2 *kubernetes.Clientset
	var r3 error
	if rf, ok := ret.Get(0).(func(*k8s.ClusterConfig) (*rest.Config, *http.Client, *kubernetes.Clientset, error)); ok {
		return rf(clusterConfig)
	}
	if rf, ok := ret.Get(0).(func(*k8s.ClusterConfig) *rest.Config); ok {
		r0 = rf(clusterConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rest.Config)
		}
	}

	if rf, ok := ret.Get(1).(func(*k8s.ClusterConfig) *http.Client); ok {
		r1 = rf(clusterConfig)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*http.Client)
		}
	}

	if rf, ok := ret.Get(2).(func(*k8s.ClusterConfig) *kubernetes.Clientset); ok {
		r2 = rf(clusterConfig)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*kubernetes.Clientset)
		}
	}

	if rf, ok := ret.Get(3).(func(*k8s.ClusterConfig) error); ok {
		r3 = rf(clusterConfig)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// GetK8sConfigAndClientsByRestConfig provides a mock function with given fields: restConfig
func (_m *KubeConfigInterface) GetK8sConfigAndClientsByRestConfig(restConfig *rest.Config) (*http.Client, *kubernetes.Clientset, error) {
	ret := _m.Called(restConfig)

	if len(ret) == 0 {
		panic("no return value specified for GetK8sConfigAndClientsByRestConfig")
	}

	var r0 *http.Client
	var r1 *kubernetes.Clientset
	var r2 error
	if rf, ok := ret.Get(0).(func(*rest.Config) (*http.Client, *kubernetes.Clientset, error)); ok {
		return rf(restConfig)
	}
	if rf, ok := ret.Get(0).(func(*rest.Config) *http.Client); ok {
		r0 = rf(restConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Client)
		}
	}

	if rf, ok := ret.Get(1).(func(*rest.Config) *kubernetes.Clientset); ok {
		r1 = rf(restConfig)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*kubernetes.Clientset)
		}
	}

	if rf, ok := ret.Get(2).(func(*rest.Config) error); ok {
		r2 = rf(restConfig)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetK8sInClusterConfigAndClients provides a mock function with given fields:
func (_m *KubeConfigInterface) GetK8sInClusterConfigAndClients() (*rest.Config, *http.Client, *kubernetes.Clientset, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetK8sInClusterConfigAndClients")
	}

	var r0 *rest.Config
	var r1 *http.Client
	var r2 *kubernetes.Clientset
	var r3 error
	if rf, ok := ret.Get(0).(func() (*rest.Config, *http.Client, *kubernetes.Clientset, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *rest.Config); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rest.Config)
		}
	}

	if rf, ok := ret.Get(1).(func() *http.Client); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*http.Client)
		}
	}

	if rf, ok := ret.Get(2).(func() *kubernetes.Clientset); ok {
		r2 = rf()
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*kubernetes.Clientset)
		}
	}

	if rf, ok := ret.Get(3).(func() error); ok {
		r3 = rf()
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// GetK8sInClusterConfigAndDynamicClients provides a mock function with given fields:
func (_m *KubeConfigInterface) GetK8sInClusterConfigAndDynamicClients() (*rest.Config, *http.Client, dynamic.Interface, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetK8sInClusterConfigAndDynamicClients")
	}

	var r0 *rest.Config
	var r1 *http.Client
	var r2 dynamic.Interface
	var r3 error
	if rf, ok := ret.Get(0).(func() (*rest.Config, *http.Client, dynamic.Interface, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *rest.Config); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rest.Config)
		}
	}

	if rf, ok := ret.Get(1).(func() *http.Client); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*http.Client)
		}
	}

	if rf, ok := ret.Get(2).(func() dynamic.Interface); ok {
		r2 = rf()
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(dynamic.Interface)
		}
	}

	if rf, ok := ret.Get(3).(func() error); ok {
		r3 = rf()
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// GetK8sInClusterRestConfig provides a mock function with given fields:
func (_m *KubeConfigInterface) GetK8sInClusterRestConfig() (*rest.Config, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetK8sInClusterRestConfig")
	}

	var r0 *rest.Config
	var r1 error
	if rf, ok := ret.Get(0).(func() (*rest.Config, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *rest.Config); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rest.Config)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRestConfigByCluster provides a mock function with given fields: clusterConfig
func (_m *KubeConfigInterface) GetRestConfigByCluster(clusterConfig *k8s.ClusterConfig) (*rest.Config, error) {
	ret := _m.Called(clusterConfig)

	if len(ret) == 0 {
		panic("no return value specified for GetRestConfigByCluster")
	}

	var r0 *rest.Config
	var r1 error
	if rf, ok := ret.Get(0).(func(*k8s.ClusterConfig) (*rest.Config, error)); ok {
		return rf(clusterConfig)
	}
	if rf, ok := ret.Get(0).(func(*k8s.ClusterConfig) *rest.Config); ok {
		r0 = rf(clusterConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rest.Config)
		}
	}

	if rf, ok := ret.Get(1).(func(*k8s.ClusterConfig) error); ok {
		r1 = rf(clusterConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OverrideRestConfigWithCustomTransport provides a mock function with given fields: restConfig
func (_m *KubeConfigInterface) OverrideRestConfigWithCustomTransport(restConfig *rest.Config) (*rest.Config, error) {
	ret := _m.Called(restConfig)

	if len(ret) == 0 {
		panic("no return value specified for OverrideRestConfigWithCustomTransport")
	}

	var r0 *rest.Config
	var r1 error
	if rf, ok := ret.Get(0).(func(*rest.Config) (*rest.Config, error)); ok {
		return rf(restConfig)
	}
	if rf, ok := ret.Get(0).(func(*rest.Config) *rest.Config); ok {
		r0 = rf(restConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rest.Config)
		}
	}

	if rf, ok := ret.Get(1).(func(*rest.Config) error); ok {
		r1 = rf(restConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewKubeConfigInterface creates a new instance of KubeConfigInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewKubeConfigInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *KubeConfigInterface {
	mock := &KubeConfigInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
