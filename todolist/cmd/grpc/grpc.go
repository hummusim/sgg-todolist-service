package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	errors_stack "github.com/go-errors/errors"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.uber.org/zap"
	"go.uber.org/zap/zapgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	healthcheck "github.com/overridesh/sgg-todolist-service/internal/grpc/healthcheck"
	"github.com/overridesh/sgg-todolist-service/internal/grpc/todolist"
	"github.com/overridesh/sgg-todolist-service/internal/repository"
	"github.com/overridesh/sgg-todolist-service/pkg/storage/sql"
	pbTodoList "github.com/overridesh/sgg-todolist-service/proto"
	"github.com/overridesh/sgg-todolist-service/tools"
)

type app struct {
	config     *Config
	grpcServer *grpc.Server
	sql        sql.DB
}

// Config secrets for app
// In the future could be inject from Vault or something like that.
type Config struct {
	Database struct {
		User     string `envconfig:"DATABASE_USERNAME" required:"true"`
		Password string `envconfig:"DATABASE_PASSWORD" required:"true"`
		Name     string `envconfig:"DATABASE_NAME" required:"true"`
		Host     string `envconfig:"DATABASE_HOSTNAME" required:"true"`
		Port     int32  `envconfig:"DATABASE_PORT" required:"true"`
	}
	Certfile string `envconfig:"CERT_FILE" required:"true"`
	Keyfile  string `envconfig:"KEY_FILE" required:"true"`
	Host     string `default:"0.0.0.0" envconfig:"HOST"`
	Port     int    `default:"10000" envconfig:"PORT"`
}

func main() {
	var (
		err error
		prg app = app{}
	)

	go prg.listenExit()

	var config Config = Config{}
	if err = tools.GetConfig("", &config); err != nil {
		log.Fatal(err)
	}

	// Load Config
	prg.config = &config

	// Db Connection
	prg.sql, err = sql.NewConnection(fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
		prg.config.Database.User,
		prg.config.Database.Password,
		prg.config.Database.Host,
		prg.config.Database.Port,
		prg.config.Database.Name,
	))
	if err != nil {
		log.Fatal(err)
	}

	// start app
	if err = prg.Start(); err != nil {
		log.Fatal(err)
	}
}

func (p *app) Start() error {
	// For now I won't dwell on something custom for logs.
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("cannot create new logger, error %v", err)
	}
	defer logger.Sync()

	// Replace global logger of zap, so we can use "zap.L()" and Set GRPC Logger
	zap.ReplaceGlobals(logger)

	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Sugar().Fatalf("failed to listen: %v", err)
	}

	p.grpcServer, err = p.startGRPCServer(zapgrpc.NewLogger(zap.L()))
	if err != nil {
		logger.Sugar().Fatalf("failed to start grpc server: %v", err)
	}

	// Register grpc service
	pbTodoList.RegisterTodoListServiceServer(
		p.grpcServer,
		todolist.NewGRPC(
			repository.NewTaskRepository(p.sql),
			repository.NewCommentRepository(p.sql),
			repository.NewLabelRepository(p.sql),
		),
	)
	pbTodoList.RegisterHealthcheckServiceServer(p.grpcServer, healthcheck.NewGRPC())

	return p.grpcServer.Serve(lis)
}

// startGRPCServer start a new grpc server
func (p *app) startGRPCServer(logger grpclog.LoggerV2) (*grpc.Server, error) {
	// Define customfunc to handle panic
	var customFunc = func(err interface{}) error {
		logger.Error(
			"panic",
			zap.Any("raw", err),
			errors_stack.Wrap(err, 2).ErrorStack(),
		)
		return errors.New("internal server error")
	}

	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []grpcRecovery.Option{
		grpcRecovery.WithRecoveryHandler(customFunc),
	}

	transport, err := credentials.NewServerTLSFromFile(p.config.Certfile, p.config.Keyfile)
	if err != nil {
		return nil, err
	}

	options := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpcMiddleware.ChainUnaryServer(
				// Just for recovery from the panic
				grpcRecovery.UnaryServerInterceptor(opts...),
			),
		),
		// Just for recovery from the panic
		grpcMiddleware.WithStreamServerChain(
			grpcRecovery.StreamServerInterceptor(opts...),
		),
		grpc.Creds(transport),
	}

	grpclog.SetLoggerV2(logger)

	// Create new grpc server
	return grpc.NewServer(options...), nil
}

func (p *app) stop() {
	if p.grpcServer != nil {
		zap.L().Warn("stopping grpc server")
		p.grpcServer.Stop()
	}
	if p.sql != nil {
		zap.L().Warn("stopping db connection")
		if err := p.sql.Close(); err != nil {
			zap.S().Errorw("cannot close connection, error: %v", err)
		}
	}
}

func (p *app) listenExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		p.stop()
		os.Exit(0)
	}()
}
