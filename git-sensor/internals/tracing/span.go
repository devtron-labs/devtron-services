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

// Package tracing provides lightweight, loosely-coupled OpenTelemetry helpers
// for git-sensor.  Business-logic packages (pkg/git, pkg) do NOT import the
// OTel SDK directly — they either call these helpers or use the hook types
// defined here, which are nil-safe when OTel is disabled.
package tracing

import (
	"context"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "git-sensor"

// Start creates a child span from ctx.
// When OTel is disabled the global provider is a noop — zero overhead.
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, spanName, opts...)
}

// StartCaller creates a child span whose name is derived from the calling
// function via runtime.Caller — no hardcoded string literals needed.
//
// Example: (*tracedRepoManager).GetCommitMetadata calling StartCaller produces
// span name "(*tracedRepoManager).GetCommitMetadata".
func StartCaller(ctx context.Context, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindInternal)}
	if len(attrs) > 0 {
		opts = append(opts, trace.WithAttributes(attrs...))
	}
	return Start(ctx, callerName(2), opts...)
}

// callerName returns the short function name skip frames up the call stack.
// skip=1 → direct caller of callerName; skip=2 → caller's caller; etc.
func callerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	// fn.Name() = "github.com/devtron-labs/git-sensor/pkg.(*tracedRepoManager).GetCommitMetadata"
	// Strip the module path prefix (everything up to and including the last /)
	name := fn.Name()
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		name = name[i+1:]
	}
	// Strip the top-level package qualifier ("pkg." → drop it)
	if i := strings.IndexByte(name, '.'); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// End records err on the span (when non-nil) and calls span.End().
// Designed for use in a named-return defer:
//
//	func Foo() (err error) {
//	    ctx, span := tracing.StartCaller(ctx)
//	    defer func() { tracing.End(span, err) }()
//	    ...
//	}
func End(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
