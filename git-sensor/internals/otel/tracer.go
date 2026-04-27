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

package otel

import (
	"context"
	stdlog "log"

	"github.com/caarlos0/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	globallog "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config holds the two exposed env vars for OTel.
type Config struct {
	Endpoint    string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:""`
	ServiceName string `env:"OTEL_SERVICE_NAME" envDefault:"git-sensor"`
}

// Enabled is set to true after a successful Init; checked by logger and DB hook.
var Enabled bool

// LP is the SDK log provider; exported so the otelzap bridge can reference it.
var LP *sdklog.LoggerProvider

func GetConfig() (*Config, error) {
	cfg := &Config{}
	return cfg, env.Parse(cfg)
}

// Init sets the global TracerProvider and LoggerProvider via OTLP gRPC.
// Returns a shutdown function that flushes pending telemetry on process exit.
// When Endpoint is empty, tracing/logging is disabled and a no-op is returned.
// Uses stdlib log for messages because it runs before the zap logger is created.
func Init(cfg *Config) (func(context.Context) error, error) {
	noop := func(ctx context.Context) error { return nil }

	if cfg.Endpoint == "" {
		stdlog.Println("OTEL_EXPORTER_OTLP_ENDPOINT not set — OTel tracing/logging disabled")
		return noop, nil
	}

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceNameKey.String(cfg.ServiceName)),
	)
	if err != nil {
		return noop, err
	}

	// --- Trace provider ---
	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return noop, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// --- Log provider ---
	logExp, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(cfg.Endpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return noop, err
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	globallog.SetLoggerProvider(lp)
	LP = lp
	Enabled = true

	stdlog.Printf("OTel initialized — endpoint=%s service=%s\n", cfg.Endpoint, cfg.ServiceName)

	return func(ctx context.Context) error {
		_ = tp.Shutdown(ctx)
		return lp.Shutdown(ctx)
	}, nil
}
