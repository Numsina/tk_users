package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"time"
)

var (
	endpoint = "192.168.84.10:4318"
	tracer   = otel.Tracer("gin-server")
)

func main() {
	ctx := context.Background()

	tp, err := initTracer(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = tp.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown tracer: %v", err)
		}
	}()

	r := gin.New()

	r.Use(otelgin.Middleware("gin-name"))
	// 在响应头记录 TRACE-ID
	r.Use(func(c *gin.Context) {
		c.Header("Trace-Id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String())
	})

	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		name := getUser(c, id)
		c.JSON(http.StatusOK, gin.H{
			"name": name,
			"id":   id,
		})
	})
	_ = r.Run(":8008")
}

func newJaegerTraceProvider(ctx context.Context) (*sdkTrace.TracerProvider, error) {
	// 创建导出器， 将追踪到的数据发送到server
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(endpoint), otlptracehttp.WithInsecure())

	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName("demo")))
	if err != nil {
		return nil, err
	}

	// 创建追踪提供器，用来管理和处理追踪的数据
	traceProvider := sdkTrace.NewTracerProvider(
		//将之前创建的资源对象添加到追踪提供器，以便在追踪数据中包含应用程序的元数据
		sdkTrace.WithResource(res),
		sdkTrace.WithSampler(sdkTrace.AlwaysSample()),                     //采样
		sdkTrace.WithBatcher(exp, sdkTrace.WithBatchTimeout(time.Second))) // 设置批量处理器，将追踪到的数据批量发送，提高性能
	return traceProvider, nil
}

/*
*
initTracer:

	1.初始化jaeger追踪提供器
	2.将其设置为全局追踪提供器
	3.设置文本映射传播器(用于在不同服务之间传递追踪上下文和其他信息)
*/
func initTracer(ctx context.Context) (*sdkTrace.TracerProvider, error) {
	tp, err := newJaegerTraceProvider(ctx)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	)
	return tp, nil
}

func getUser(ctx *gin.Context, id string) string {
	_, span := tracer.Start(
		ctx.Request.Context(), "getUser", trace.WithAttributes(attribute.String("id", id)),
	)
	defer span.End()

	time.Sleep(100 * time.Millisecond)
	// 业务逻辑....
	if id == "7" {
		return "qm"
	}
	return "unknown"
}
