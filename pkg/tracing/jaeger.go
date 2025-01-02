package tracing

import (
	"context"

	"github.com/uber/jaeger-client-go"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type JaegerPropagator struct{}

func (p JaegerPropagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	// SKIP extract the Jaeger trace context from the incoming request.
	return ctx
}

func (p JaegerPropagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return
	}

	jTraceID, err := jaeger.TraceIDFromString(sc.TraceID().String())
	if err != nil {
		return
	}
	jSpanID, err := jaeger.SpanIDFromString(sc.SpanID().String())
	if err != nil {
		return
	}
	jsc := jaeger.NewSpanContext(
		jTraceID, jSpanID,
		jaeger.SpanID(0),
		sc.IsSampled(),
		map[string]string{},
	)
	jTP := jsc.String()

	carrier.Set(jaeger.TraceContextHeaderName, jTP)
}

func (p JaegerPropagator) Fields() []string {
	return []string{jaeger.TraceContextHeaderName}
}
