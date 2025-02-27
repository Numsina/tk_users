package tracing

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/Numsina/tk_users/user_web/initialize"
)

func newJaegerTracer(ctx context.Context) (*sdkTrace.TracerProvider, error) {
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(
		fmt.Sprintf("%s:%d", initialize.Conf.JaegerInfo.Host, initialize.Conf.JaegerInfo.Port),
	), otlptracehttp.WithInsecure())

	if err != nil {
		log.Println("链接jaeger失败, :", err)
		return nil, err
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName("user_web")))
	if err != nil {
		log.Println("初始化资源失败, 失败原因: ", err)
		return nil, err
	}

	tracerProvider := sdkTrace.NewTracerProvider(sdkTrace.WithResource(res), sdkTrace.WithSampler(sdkTrace.AlwaysSample()),
		sdkTrace.WithBatcher(exp, sdkTrace.WithBatchTimeout(time.Second)),
	)
	return tracerProvider, nil
}

func InitJaeger(ctx context.Context) (*sdkTrace.TracerProvider, error) {
	tp, err := newJaegerTracer(ctx)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil

}
