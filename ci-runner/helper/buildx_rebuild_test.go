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

package helper

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"testing"
	"time"

	cicxt "github.com/devtron-labs/ci-runner/executor/context"
	bean2 "github.com/devtron-labs/ci-runner/helper/bean"
	"github.com/devtron-labs/common-lib/utils/retryFunc"
)

// ---------------------------------------------------------------------------
// MockBuildxK8sInterface
// ---------------------------------------------------------------------------

type MockBuildxK8sInterface struct {
	RestartErr  error
	RegisterErr error
	WaitBlocks  bool          // true = WaitUntilBuilderPodLive never sends on done
	WaitDelay   time.Duration // 0 = immediate send; >0 = delay before send
	CatchErr    error         // error CatchBuilderPodLivenessError returns
}

func (m *MockBuildxK8sInterface) PatchOwnerReferenceInBuilders() {}

func (m *MockBuildxK8sInterface) RegisterBuilderPods(_ context.Context) error {
	return m.RegisterErr
}

func (m *MockBuildxK8sInterface) RestartBuilders(_ context.Context) error {
	return m.RestartErr
}

func (m *MockBuildxK8sInterface) CatchBuilderPodLivenessError(ctx context.Context) error {
	if m.CatchErr != nil {
		return m.CatchErr
	}
	<-ctx.Done()
	return nil
}

func (m *MockBuildxK8sInterface) WaitUntilBuilderPodLive(ctx context.Context, done chan<- bool) {
	if m.WaitBlocks {
		<-ctx.Done()
		return
	}
	if m.WaitDelay > 0 {
		select {
		case <-time.After(m.WaitDelay):
		case <-ctx.Done():
			return
		}
	}
	done <- true
}

// ---------------------------------------------------------------------------
// MockCommandExecutor — returns a configurable error without running docker
// ---------------------------------------------------------------------------

type MockCommandExecutor struct {
	Err error
}

func (m *MockCommandExecutor) RunCommand(_ cicxt.CiContext, _ *exec.Cmd) error {
	return m.Err
}

func (m *MockCommandExecutor) RunCommandWithCtx(_ cicxt.CiContext, _ *exec.Cmd) error {
	return m.Err
}

// ---------------------------------------------------------------------------
// Factory helpers
// ---------------------------------------------------------------------------

func mockFactory(mock BuildxK8sInterface) BuildxK8sClientFactory {
	return func(_ []string) (BuildxK8sInterface, error) {
		return mock, nil
	}
}

func errorFactory(err error) BuildxK8sClientFactory {
	return func(_ []string) (BuildxK8sInterface, error) {
		return nil, err
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newTestImpl(factory BuildxK8sClientFactory) *DockerHelperImpl {
	return &DockerHelperImpl{
		DockerCommandEnv: []string{},
		cmdExecutor:      &MockCommandExecutor{Err: errors.New("docker not available in test")},
		k8sClientFactory: factory,
	}
}

func makeCiContext() cicxt.CiContext {
	return cicxt.BuildCiContext(context.Background(), false)
}

// computeBuilderPodWaitDuration mirrors the logic in BuildArtifact.
func computeBuilderPodWaitDuration(secs int) time.Duration {
	if secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return 2 * time.Minute
}

// ---------------------------------------------------------------------------
// Group A: builderPodWaitDuration computation
// ---------------------------------------------------------------------------

func TestBuilderPodWaitDuration(t *testing.T) {
	tests := []struct {
		name string
		secs int
		want time.Duration
	}{
		{name: "A1: zero → default 2m", secs: 0, want: 2 * time.Minute},
		{name: "A2: negative → default 2m", secs: -1, want: 2 * time.Minute},
		{name: "A3: 300 → 5m", secs: 300, want: 5 * time.Minute},
		{name: "A4: 1 → 1s", secs: 1, want: 1 * time.Second},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeBuilderPodWaitDuration(tc.secs)
			if got != tc.want {
				t.Errorf("computeBuilderPodWaitDuration(%d) = %v, want %v", tc.secs, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Group B: executeDockerReBuild
// ---------------------------------------------------------------------------

func TestBuildxRebuild(t *testing.T) {
	emptyCiCtx := makeCiContext()
	emptyMetadata := bean2.DockerBuildStageMetadata{}

	t.Run("B1: useBuildxK8sDriver=false returns nil immediately", func(t *testing.T) {
		impl := newTestImpl(errorFactory(errors.New("should not be called")))
		err := impl.executeDockerReBuild(
			emptyCiCtx,
			nil,
			false, // useBuildxK8sDriver=false
			"docker buildx build",
			[]string{},
			emptyMetadata,
			[]any{},
			5*time.Second,
		)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("B2: RestartErr is propagated", func(t *testing.T) {
		restartErr := errors.New("restart fail")
		mock := &MockBuildxK8sInterface{RestartErr: restartErr}
		impl := newTestImpl(mockFactory(mock))
		err := impl.executeDockerReBuild(
			emptyCiCtx,
			mock,
			true,
			"docker buildx build",
			[]string{},
			emptyMetadata,
			[]any{},
			5*time.Second,
		)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, restartErr) {
			t.Errorf("expected restartErr, got %v", err)
		}
	})

	t.Run("B3: factory error is propagated", func(t *testing.T) {
		factoryErr := errors.New("factory error")
		mock := &MockBuildxK8sInterface{} // RestartErr=nil so Restart succeeds
		impl := newTestImpl(errorFactory(factoryErr))
		err := impl.executeDockerReBuild(
			emptyCiCtx,
			mock,
			true,
			"docker buildx build",
			[]string{},
			emptyMetadata,
			[]any{},
			5*time.Second,
		)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, factoryErr) {
			t.Errorf("expected factoryErr, got %v", err)
		}
	})

	t.Run("B4: RegisterErr is propagated", func(t *testing.T) {
		registerErr := errors.New("register fail")
		mock := &MockBuildxK8sInterface{RegisterErr: registerErr}
		impl := newTestImpl(mockFactory(mock))
		err := impl.executeDockerReBuild(
			emptyCiCtx,
			mock,
			true,
			"docker buildx build",
			[]string{},
			emptyMetadata,
			[]any{},
			5*time.Second,
		)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, registerErr) {
			t.Errorf("expected registerErr, got %v", err)
		}
	})

	t.Run("B5: pod ready immediately → not BuilderPodDeletedError", func(t *testing.T) {
		mock := &MockBuildxK8sInterface{WaitBlocks: false, WaitDelay: 0}
		impl := newTestImpl(mockFactory(mock))
		err := impl.executeDockerReBuild(
			emptyCiCtx,
			mock,
			true,
			"docker buildx build .",
			[]string{},
			emptyMetadata,
			[]any{},
			5*time.Second,
		)
		// docker is not available in test env; error will be a docker exec error, NOT BuilderPodDeletedError
		if errors.Is(err, BuilderPodDeletedError) {
			t.Errorf("expected non-BuilderPodDeletedError, got BuilderPodDeletedError")
		}
	})

	t.Run("B6: pod ready after 50ms → not BuilderPodDeletedError", func(t *testing.T) {
		mock := &MockBuildxK8sInterface{WaitBlocks: false, WaitDelay: 50 * time.Millisecond}
		impl := newTestImpl(mockFactory(mock))
		err := impl.executeDockerReBuild(
			emptyCiCtx,
			mock,
			true,
			"docker buildx build .",
			[]string{},
			emptyMetadata,
			[]any{},
			5*time.Second,
		)
		if errors.Is(err, BuilderPodDeletedError) {
			t.Errorf("expected non-BuilderPodDeletedError, got BuilderPodDeletedError")
		}
	})

	t.Run("B7: WaitBlocks=true with short timeout → BuilderPodDeletedError (wrapped in RetryableError)", func(t *testing.T) {
		mock := &MockBuildxK8sInterface{WaitBlocks: true}
		impl := newTestImpl(mockFactory(mock))
		err := impl.executeDockerReBuild(
			emptyCiCtx,
			mock,
			true,
			"docker buildx build .",
			[]string{},
			emptyMetadata,
			[]any{},
			100*time.Millisecond,
		)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// executeDockerReBuild wraps BuilderPodDeletedError in RetryableError
		// which does not implement Unwrap, so we check via IsRetryableError and error message
		if !retryFunc.IsRetryableError(err) {
			t.Errorf("expected RetryableError wrapping BuilderPodDeletedError, got %T: %v", err, err)
		}
		if !contains(err.Error(), BuilderPodDeletedError.Error()) {
			t.Errorf("expected error message to contain %q, got %q", BuilderPodDeletedError.Error(), err.Error())
		}
	})
}

// ---------------------------------------------------------------------------
// Group C: waitForBuilderPods (package-level function)
// ---------------------------------------------------------------------------

func TestWaitForBuilderPods(t *testing.T) {
	t.Run("C1: pod ready immediately → nil", func(t *testing.T) {
		mock := &MockBuildxK8sInterface{WaitBlocks: false, WaitDelay: 0}
		ctx := context.Background()
		err := waitForBuilderPods(ctx, mock, 5*time.Second)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("C2: pod ready after 50ms → nil", func(t *testing.T) {
		mock := &MockBuildxK8sInterface{WaitBlocks: false, WaitDelay: 50 * time.Millisecond}
		ctx := context.Background()
		err := waitForBuilderPods(ctx, mock, 5*time.Second)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("C3: WaitBlocks with 100ms timeout → error with duration", func(t *testing.T) {
		mock := &MockBuildxK8sInterface{WaitBlocks: true}
		ctx := context.Background()
		timeout := 100 * time.Millisecond
		err := waitForBuilderPods(ctx, mock, timeout)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		errMsg := err.Error()
		if !contains(errMsg, "did not reach Running state") {
			t.Errorf("error %q should contain 'did not reach Running state'", errMsg)
		}
		if !contains(errMsg, "100ms") {
			t.Errorf("error %q should contain '100ms'", errMsg)
		}
	})

	t.Run("C4: pre-cancelled context → non-nil error", func(t *testing.T) {
		mock := &MockBuildxK8sInterface{WaitBlocks: true}
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel before calling
		err := waitForBuilderPods(ctx, mock, 5*time.Second)
		if err == nil {
			t.Error("expected non-nil error for pre-cancelled context, got nil")
		}
	})
}

// contains is a simple helper since strings.Contains is from standard library.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// ---------------------------------------------------------------------------
// Group F: CommonWorkflowRequest JSON field
// ---------------------------------------------------------------------------

func TestCommonWorkflowRequestJSON(t *testing.T) {
	t.Run("F1: JSON missing field → defaults to 0", func(t *testing.T) {
		jsonStr := `{"workflowNamePrefix": "test"}`
		var req CommonWorkflowRequest
		if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if req.BuildxBuilderPodWaitDurationSecs != 0 {
			t.Errorf("expected 0, got %d", req.BuildxBuilderPodWaitDurationSecs)
		}
	})

	t.Run("F2: JSON includes field=300 → parsed correctly", func(t *testing.T) {
		jsonStr := `{"buildxBuilderPodWaitDurationSecs": 300}`
		var req CommonWorkflowRequest
		if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if req.BuildxBuilderPodWaitDurationSecs != 300 {
			t.Errorf("expected 300, got %d", req.BuildxBuilderPodWaitDurationSecs)
		}
	})

	t.Run("F3: round-trip marshal/unmarshal preserves value", func(t *testing.T) {
		orig := CommonWorkflowRequest{BuildxBuilderPodWaitDurationSecs: 180}
		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		var parsed CommonWorkflowRequest
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if parsed.BuildxBuilderPodWaitDurationSecs != 180 {
			t.Errorf("expected 180, got %d", parsed.BuildxBuilderPodWaitDurationSecs)
		}
	})
}
