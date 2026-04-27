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

package sql

import (
	"context"
	"time"

	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// OtelQueryHook creates an OTel span for every go-pg query.
// Registered via db.OnQueryProcessed in NewDbConnection.
//
// Note: go-pg v6 OnQueryProcessed does not carry request context, so these
// spans are roots (not nested under parent HTTP/gRPC traces). They still give
// per-query duration and error visibility in SigNoz.
func OtelQueryHook(event *pg.QueryProcessedEvent) {
	endTime := time.Now()

	tracer := otel.Tracer("git-sensor/db")
	_, span := tracer.Start(context.Background(), "db.query",
		trace.WithTimestamp(event.StartTime),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End(trace.WithTimestamp(endTime))

	query, _ := event.FormattedQuery()
	if len(query) > 500 {
		query = query[:500]
	}
	span.SetAttributes(
		semconv.DBSystemPostgreSQL,
		attribute.String("db.statement", query),
		attribute.String("db.operation", event.Func),
	)
	if event.Error != nil {
		span.RecordError(event.Error)
		span.SetStatus(codes.Error, event.Error.Error())
	}
}
