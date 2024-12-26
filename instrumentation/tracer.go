package app

import (
	"context"
	"log"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var tracer tracer.Tracer

func newExporter(ctx context.Context) (trace.SpanExporter, error) { // 使用OTLP exporter
	return otlptracehttp.New(ctx)
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("ExampleService"),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)

}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "hello-span") // 创建了一个span，名字是hello-span
	defer span.End()                                     // 结束span

	// 业务逻辑
}

func main() {
	ctx := context.Background()
	exp, err := newExporter(ctx)
	if err != nil {
		log.Fatalf("初始化exporter失败: %v", err)
	}

	// Create a new tracer provider with a batch span processor and the given exporter.
	tp := newTraceProvider(exp)

	// Handle shutdown properly so nothing leaks.
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)

	// 为这个请求创建一个tracer
	// tracer = otel.GetTracerProvider().Tracer("example.com/basic")
	tracer = otel.GetTracerProvider().Tracer("/rolldice")
	// tracer = otel.GetTracerProvider().Tracer("/rolldice/{player}")

}
