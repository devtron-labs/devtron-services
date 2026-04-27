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

package tracing

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	internalotel "github.com/devtron-labs/git-sensor/internals/otel"
)

// CommandHookFn wraps a git CLI command execution with a tracing span.
// pkg/git imports this type without pulling in any OTel SDK dependency.
// nil means tracing is disabled.
type CommandHookFn func(ctx context.Context, cmdArgs string) (context.Context, func(error))

// PollHookFn wraps a repository poll cycle with a tracing span.
// nil means tracing is disabled.
type PollHookFn func(ctx context.Context, materialId int) (context.Context, func(error))

var (
	activeCommandHook CommandHookFn
	activePollHook    PollHookFn
)

// Register wires up the active hook implementations.
// Must be called from main.go after internalotel.Init().
// When OTel is disabled all hooks remain nil; callers guard with nil checks
// so there is zero runtime overhead when tracing is turned off.
func Register() {
	if !internalotel.Enabled {
		return
	}

	// git command spans — span name = "git.<subcommand>" (e.g. git.fetch, git.log, git.clone).
	// The subcommand is extracted from the cmdArgs string at runtime; no hardcoded names.
	cmdTracer := otel.Tracer(tracerName + "/git")
	activeCommandHook = func(ctx context.Context, cmdArgs string) (context.Context, func(error)) {
		if len(cmdArgs) > 300 {
			cmdArgs = cmdArgs[:300]
		}
		ctx, span := cmdTracer.Start(ctx, gitSubcmd(cmdArgs),
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attribute.String("git.args", cmdArgs)),
		)
		return ctx, func(err error) {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.End()
		}
	}

	// poll spans — one span per material poll cycle in the background watcher.
	pollTracer := otel.Tracer(tracerName + "/watcher")
	activePollHook = func(ctx context.Context, materialId int) (context.Context, func(error)) {
		ctx, span := pollTracer.Start(ctx, "watcher.poll",
			trace.WithAttributes(attribute.Int("material.id", materialId)),
		)
		return ctx, func(err error) {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.End()
		}
	}
}

// gitSubcmd extracts the git subcommand from a full command args string.
// "/usr/bin/git -c ... fetch origin" → "git.fetch"
// Falls back to "git.exec" if extraction fails.
func gitSubcmd(cmdArgs string) string {
	fields := strings.Fields(cmdArgs)
	for i, f := range fields {
		if strings.HasSuffix(f, "git") && i+1 < len(fields) {
			for _, sub := range fields[i+1:] {
				if !strings.HasPrefix(sub, "-") {
					return "git." + sub
				}
			}
		}
	}
	return "git.exec"
}

// CommandHook returns the active git command hook (nil when tracing is off).
// Called once per git CLI execution; nil check is a single pointer comparison.
func CommandHook() CommandHookFn { return activeCommandHook }

// PollHook returns the active poll hook (nil when tracing is off).
func PollHook() PollHookFn { return activePollHook }
