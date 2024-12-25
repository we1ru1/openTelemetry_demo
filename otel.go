package main

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown函数调用通过shutdownFuncs注册的清理函数，并将错误连接在一起。
	// 每个注册的清理函数将被调用一次。
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr函数调用shutdown函数以进行清理，并确保返回所有错误。
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// 设置传播器。
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// // 设置tracer提供程序。
	// tracerProvider, err := newTraceProvider()
	// if err != nil {
	// 	handleErr(err)
	// 	return
	// }
	// shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	// otel.SetTracerProvider(tracerProvider)

	// 设置meter提供程序。
	meterProvider, err := newMeterProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// // 设置logger提供程序
	// loggerProvider, err := newLoggerProvider()
	// if err != nil {
	// 	handleErr(err)
	// 	return
	// }
	// shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	// global.SetLoggerProvider(loggerProvider)

	return
}

// newPropagator函数创建一个传播器，该传播器支持TraceContext和Baggage格式。
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newTraceProvider函数创建一个新的trace.TracerProvider。
func newTraceProvider() (*trace.TracerProvider, error) {
	traceExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// 默认是5秒，这里设置为1秒以演示目的。
			trace.WithBatchTimeout(time.Second)),
	)
	return traceProvider, nil
}

// newMeterProvider函数创建一个新的metric.MeterProvider。
func newMeterProvider() (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// 默认是1分钟，这里设置为3秒以演示目的。
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider, nil
}

func newLoggerProvider() (*log.LoggerProvider, error) {
	logExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}
