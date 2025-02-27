package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/Numsina/tk_users/user_web/api"
	"github.com/Numsina/tk_users/user_web/config"
	"github.com/Numsina/tk_users/user_web/gen/users/v1"
	"github.com/Numsina/tk_users/user_web/initialize"
	prom "github.com/Numsina/tk_users/user_web/initialize/metrics"
	"github.com/Numsina/tk_users/user_web/initialize/tracing"
	"github.com/Numsina/tk_users/user_web/logger"
	"github.com/Numsina/tk_users/user_web/middleware"
	"github.com/Numsina/tk_users/user_web/middleware/metrics"
	"github.com/Numsina/tk_users/user_web/service"
	"github.com/Numsina/tk_users/user_web/tools"
)

type App struct {
	conf       *config.Config
	logger     *logger.Logger
	client     users.UserServiceClient
	jhl        *middleware.JWT
	instanceId string
	consulApi  *consul.Client
	port       int
	conn       *grpc.ClientConn
}

func (a *App) Init() {
	a.conf = initialize.InitConfig()
	initialize.InitRedis()
	a.logger = initialize.InitLogger()
	a.jhl = middleware.NewJWT([]byte(a.conf.JwtInfo.Key))

}

func Run() {
	a := new(App)
	a.Init()
	r := gin.Default()
	a.dial()
	a.use(r)
	a.ioc(r)
	ctx := context.Background()
	tp, err := tracing.InitJaeger(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = tp.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown tracer: %v", err)
		}
	}()
	a.port, _ = tools.GetFreePort()
	a.registerConsul()
	prom.InitPrometheus()
	go func() {
		r.Run(fmt.Sprintf(":%d", a.port))

	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt)
	<-quit
	a.stopConsul()
	a.conn.Close()
}

func (a *App) dial() {
	var err error
	var dsn = fmt.Sprintf("consul://%s:%d/%s?wait=14s", a.conf.ConsuleInfo.Host, a.conf.ConsuleInfo.Port, "tk_user_srv")
	a.conn, err = grpc.NewClient(
		dsn,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			md := metadata.Pairs(
				"timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10),
				"client_id", "tk_user_web",
			)
			ctx = metadata.NewOutgoingContext(ctx, md)
			tracer := otel.Tracer("conn tk_user_srv")
			ctx, span := tracer.Start(ctx, method)
			defer span.End()
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)

	if err != nil {
		a.logger.Sugar().Panicf("连接用户服务失败, 原因：%s", err.Error())
	}

	a.client = users.NewUserServiceClient(a.conn)
}

func (a *App) registerConsul() {
	cfg := consul.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", a.conf.ConsuleInfo.Host, a.conf.ConsuleInfo.Port)
	var err error
	a.consulApi, err = consul.NewClient(cfg)
	if err != nil {
		a.logger.Sugar().Panicf("user_web consul客户端初始化失败, 失败原因：%s", err)
	}

	a.instanceId = uuid.New().String()
	chk := &consul.AgentServiceCheck{
		Interval: "10s",
		Timeout:  "10s",
		HTTP:     fmt.Sprintf("http://%s:%d/health", a.conf.ConsuleInfo.Address, a.port),
	}
	svc := &consul.AgentServiceRegistration{
		ID:      a.instanceId,
		Address: a.conf.ConsuleInfo.Address,
		Port:    a.port,
		Name:    a.conf.ConsuleInfo.Name,
		Tags:    []string{"tk_user_web", "itgsyang"},
		Check:   chk,
	}
	err = a.consulApi.Agent().ServiceRegister(svc)
	if err != nil {
		a.logger.Sugar().Panicf("user_web注册微服务失败, 失败原因：%s", err)
	}

	a.logger.Sugar().Info("user_web注册微服务成功")
}

func (a *App) stopConsul() {
	err := a.consulApi.Agent().ServiceDeregister(a.instanceId)
	if err != nil {
		a.logger.Sugar().Errorf("注销consul实例失败, 失败原因:%s", err)
		for i := 0; i < 3; i++ {
			err = a.consulApi.Agent().ServiceDeregister(a.instanceId)
			if err == nil {
				break
			}
		}
	}
	a.logger.Sugar().Info("consul注销成功")
}

func (a *App) ioc(r *gin.Engine) {
	svc := service.NewService(a.client)
	userhandler := api.NewUserHandler(svc, a.logger, a.jhl)
	userhandler.RegisterRouters(r)

}

func (a *App) use(r *gin.Engine) {

	r.Use(middleware.Cors(),
		middleware.NewLoginJWTMiddleWareBuilder(a.jhl).IngorePaths("/v1/users/login", "/v1/users/signup", "/metrics", "/health").Build(),
		metrics.NewMetrics(a.conf.NacosInfo.DataId, a.instanceId, a.conf.ConsuleInfo.Name, "tk_user_web", "统计请求的响应，请求的活跃数， 请求总数").Build(),
		//trace.Trace(),
		otelgin.Middleware("tk_user_web", otelgin.WithFilter(func(request *http.Request) bool {
			if strings.Contains(request.URL.Path, "health") {
				return false
			}
			//fmt.Println(request.Method)
			//
			//if strings.Contains(request.URL.Path, "grpc.health.v1") {
			//	return false
			//}

			return true
		})),
	)
}
