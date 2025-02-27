package cmd

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"

	"github.com/Numsina/tk_users/user_srv/config"
	"github.com/Numsina/tk_users/user_srv/dao"
	"github.com/Numsina/tk_users/user_srv/gen/users/v1"
	"github.com/Numsina/tk_users/user_srv/handler"
	"github.com/Numsina/tk_users/user_srv/initiallize"
	"github.com/Numsina/tk_users/user_srv/initiallize/tracing"
	logger "github.com/Numsina/tk_users/user_srv/logger"
	"github.com/Numsina/tk_users/user_srv/service"
	"github.com/Numsina/tk_users/user_srv/tools"
)

var (
	IP   string
	Port int
)

type App struct {
	db         *gorm.DB
	logger     *logger.Logger
	conf       *config.Config
	instanceId string
	client     *api.Client
}

func Execte() {
	app := new(App)
	app.Init()
	app.register()

}

func (a *App) register() {
	web := a.ioc()
	server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	users.RegisterUserServiceServer(server, web)

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", IP, Port))
	if err != nil {
		panic(err)
	}

	// 注册健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	ctx := context.Background()
	tp, err := tracing.InitJaeger(ctx)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = tp.Shutdown(ctx); err != nil {
			a.logger.Sugar().Fatalf(err.Error())
		}
	}()

	go func() {
		err = server.Serve(l)
		if err != nil {
			panic(err)
		}
	}()

	a.startConsul()

	// 优雅退出
	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	<-quit
	a.stopConsul()
	a.logger.Info("服务注销成功")
}

func (a *App) Init() {
	flag.StringVar(&IP, "IP", "0.0.0.0", "IP地址")
	flag.IntVar(&Port, "Port", 0, "应用启动端口")
	flag.Parse()
	if Port == 0 {
		port, err := tools.GetFreePort()
		if err != nil {
			panic(err)
		}
		Port = port
	}
	a.logger = initiallize.InitLogger()
	a.conf = initiallize.InitConfig()
	a.db = initiallize.InitDB()
}

func (a *App) ioc() *handler.UserHandler {
	err := dao.InitAutoMigrateTable(a.db)
	if err != nil {
		panic(err)
	}
	d := dao.NewUserDao(a.db, a.logger)
	srv := service.NewUserSvc(d, a.logger)
	return handler.NewUserHandler(srv)
}

func (a *App) startConsul() {
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", a.conf.ConsuleInfo.Host, a.conf.ConsuleInfo.Port)

	var err error
	a.client, err = api.NewClient(cfg)
	if err != nil {
		a.logger.Sugar().Panicf("初始化consul客户端失败, 失败原因为: %v\n", err)
	}

	a.instanceId = uuid.New().String()
	srv := &api.AgentServiceRegistration{
		ID:      a.instanceId,
		Name:    a.conf.ConsuleInfo.Name,
		Tags:    []string{"tkshop", "itgsyang", "user_srv"},
		Port:    Port,
		Address: a.conf.ConsuleInfo.Address,
		Check: &api.AgentServiceCheck{
			Interval: a.conf.ConsuleInfo.Interval,
			Timeout:  a.conf.ConsuleInfo.Timeout,
			GRPC:     fmt.Sprintf("%s:%d", a.conf.ConsuleInfo.Address, Port),
		},
	}

	err = a.client.Agent().ServiceRegister(srv)
	if err != nil {
		a.logger.Sugar().Panicf("user服务注册失败, err: %v", err)
	}

	a.logger.Info("user服务注册成功....")
}

func (a *App) stopConsul() {
	a.client.Agent().ServiceDeregister(a.instanceId)
}
