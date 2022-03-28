package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"go.uber.org/zap/zapgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	pbTodoList "github.com/overridesh/sgg-todolist-service/proto"
	"github.com/overridesh/sgg-todolist-service/third_party"
	"github.com/overridesh/sgg-todolist-service/tools"
)

type app struct {
	config     *Config
	grpcServer *grpc.Server
}

// Config secrets for app
// In the future could be inject from Vault or something like that.
type Config struct {
	Host                   string `default:"0.0.0.0" envconfig:"HOST"`
	Port                   int    `default:"11000" envconfig:"PORT"`
	APIPrefix              string `default:"/api" envconfig:"API_PREFIX"`
	PaymentGatewayGRPCHost string `envconfig:"TODOLIST_GRPC_HOST" required:"true"`
	PaymentGatewayGRPCCert string `envconfig:"TODOLIST_GRPC_CERT" required:"true"`
	Certfile               string `envconfig:"CERT_FILE" required:"true"`
	Keyfile                string `envconfig:"KEY_FILE" required:"true"`
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

	server, err := p.startHTTPServer()
	if err != nil {
		log.Fatalf("cannot create new server, error %v", err)
	}

	return server.ListenAndServeTLS(p.config.Certfile, p.config.Keyfile)
}

func (p *app) startHTTPServer() (*http.Server, error) {
	transport, err := credentials.NewClientTLSFromFile(p.config.PaymentGatewayGRPCCert, "")
	if err != nil {
		return nil, err
	}

	// Adds gRPC internal logs. This is quite verbose, so adjust as desired!
	grpclog.SetLoggerV2(zapgrpc.NewLogger(zap.L()))

	// Create a client connection to the gRPC Server we just started.
	// This is where the gRPC-Gateway proxies the requests.
	conn, err := grpc.DialContext(
		context.Background(),
		p.config.PaymentGatewayGRPCHost,
		grpc.WithTransportCredentials(transport),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}

	gwmux := runtime.NewServeMux(
		runtime.WithErrorHandler(handlerError),
		runtime.WithForwardResponseOption(httpResponseModifier),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	// Register handlers like a router
	err = pbTodoList.RegisterTodoListServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}
	// Register handlers like a router
	err = pbTodoList.RegisterHealthcheckServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	// OpenAPI Swagger Documentation
	openAPIRoute := getOpenAPIHandler()

	gatewayAddr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)

	gwServer := &http.Server{
		Addr: gatewayAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, p.config.APIPrefix) {
				gwmux.ServeHTTP(w, r)
				return
			}
			openAPIRoute.ServeHTTP(w, r)
		}),
	}

	zap.S().Infof("Serving gRPC-Gateway and OpenAPI Documentation on https://", gatewayAddr)
	return gwServer, nil
}

func (p *app) stop() {
	if p.grpcServer != nil {
		zap.L().Warn("stopping grpc server")
		p.grpcServer.Stop()
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

// getOpenAPIHandler serves an OpenAPI UI.
// Adapted from https://github.com/philips/grpc-gateway-example/blob/a269bcb5931ca92be0ceae6130ac27ae89582ecc/cmd/serve.go#L63
func getOpenAPIHandler() http.Handler {
	mime.AddExtensionType(".svg", "image/svg+xml")

	// Use subdirectory in embedded files
	subFS, err := fs.Sub(third_party.OpenAPI, "OpenAPI")
	if err != nil {
		zap.S().Fatalf("cannot read openapi %v", err)
	}

	return http.FileServer(http.FS(subFS))
}

func httpResponseModifier(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	// set http status code
	if vals := md.HeaderMD.Get("x-http-code"); len(vals) > 0 {
		code, err := strconv.Atoi(vals[0])
		if err != nil {
			return err
		}

		// delete the headers to not expose any grpc-metadata in http response
		delete(md.HeaderMD, "x-http-code")
		delete(w.Header(), "Grpc-Metadata-X-Http-Code")
		w.WriteHeader(code)

		if code == http.StatusNoContent {
			w.Write(nil)
		}
	}

	return nil
}

func handlerError(ctx context.Context, n *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	grpcCode := status.Code(err)
	var statusCode int = runtime.HTTPStatusFromCode(grpcCode)

	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		runtime.DefaultHTTPErrorHandler(ctx, n, marshaler, w, r, err)
		return
	}

	// set http status code
	if vals := md.HeaderMD.Get("x-http-code"); len(vals) > 0 {
		code, errConvert := strconv.Atoi(vals[0])
		if errConvert != nil {
			runtime.DefaultHTTPErrorHandler(ctx, n, marshaler, w, r, err)
			return
		}
		delete(md.HeaderMD, "x-http-code")
		delete(w.Header(), "Grpc-Metadata-X-Http-Code")
		statusCode = code
	}

	//creating a new HTTTPStatusError with a custom status, and passing error
	newError := runtime.HTTPStatusError{
		HTTPStatus: statusCode,
		Err:        err,
	}

	// using default handler to do the rest of heavy lifting of marshaling error and adding headers
	runtime.DefaultHTTPErrorHandler(ctx, n, marshaler, w, r, &newError)
}
