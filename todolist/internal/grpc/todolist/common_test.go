package todolist

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/overridesh/sgg-todolist-service/internal/repository"
	pbTodoList "github.com/overridesh/sgg-todolist-service/proto"
)

// dialer func for test grpc server
func dialer(
	taskRepository repository.TaskRepository,
	commentRepository repository.CommentRepository,
	labelRepository repository.LabelRepository,
) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	pbTodoList.RegisterTodoListServiceServer(
		server,
		NewGRPC(
			taskRepository,
			commentRepository,
			labelRepository,
		),
	)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}
