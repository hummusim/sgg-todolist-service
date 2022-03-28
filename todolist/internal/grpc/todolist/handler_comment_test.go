package todolist

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/overridesh/sgg-todolist-service/internal/model"
	"github.com/overridesh/sgg-todolist-service/internal/repository"
	mockRepository "github.com/overridesh/sgg-todolist-service/pkg/mock"
	pbTodoList "github.com/overridesh/sgg-todolist-service/proto"
	"github.com/overridesh/sgg-todolist-service/tools"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
)

func TestGetComments(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*pbTodoList.GetCommentsResponse, error)
		output *status.Status
	}{
		{
			name: "GetComments_ErrStatusIdMustBeUUID",
			input: func() (*pbTodoList.GetCommentsResponse, error) {
				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetComments(ctx, &pbTodoList.GetCommentsRequest{
					Id: "ASD",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "GetComments_ErrStatusTaskNotFound",
			input: func() (*pbTodoList.GetCommentsResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(nil, repository.ErrTaskNotFound)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetComments(ctx, &pbTodoList.GetCommentsRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusTaskNotFound,
		},
		{
			name: "GetComments_GetCommentsByTaskIdErrStatusInternalServerError",
			input: func() (*pbTodoList.GetCommentsResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(nil, sql.ErrConnDone)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetComments(ctx, &pbTodoList.GetCommentsRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "GetComments_GetCommentsByTaskIdErrStatusInternalServerError",
			input: func() (*pbTodoList.GetCommentsResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				commentRepository := new(mockRepository.CommentRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				commentRepository.On("GetCommentsByTaskId", mock.Anything, tx.Id).Return(nil, sql.ErrConnDone)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, commentRepository, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetComments(ctx, &pbTodoList.GetCommentsRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "GetComments_Success",
			input: func() (*pbTodoList.GetCommentsResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				commentRepository := new(mockRepository.CommentRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				commentRepository.On("GetCommentsByTaskId", mock.Anything, tx.Id).Return([]*model.Comment{}, nil)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, commentRepository, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetComments(ctx, &pbTodoList.GetCommentsRequest{
					Id: tx.Id.String(),
				})
			},
			output: nil,
		},
		{
			name: "GetComments_SuccessWithValues",
			input: func() (*pbTodoList.GetCommentsResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				commentRepository := new(mockRepository.CommentRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				commentRepository.On("GetCommentsByTaskId", mock.Anything, tx.Id).Return([]*model.Comment{{
					Value: "Label",
				}}, nil)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, commentRepository, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetComments(ctx, &pbTodoList.GetCommentsRequest{
					Id: tx.Id.String(),
				})
			},
			output: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.input()
			if err != nil {
				if er, ok := status.FromError(err); ok {
					if er.Code() != tt.output.Code() {
						t.Errorf("error code: expected %v, received %v", codes.InvalidArgument, er.Code())
					}
					if er.Message() != tt.output.Message() {
						t.Errorf("error message: expected %v, received %v", tt.output.Message(), er.Message())
					}
				}
			}
		})
	}
}

func TestCreateComment(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*pbTodoList.CreateCommentResponse, error)
		output *status.Status
	}{
		{
			name: "CreateComment_ErrStatusIdMustBeUUID",
			input: func() (*pbTodoList.CreateCommentResponse, error) {
				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateComment(ctx, &pbTodoList.CreateCommentRequest{
					Id: "ASD",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "CreateComment_ErrStatusTaskNotFound",
			input: func() (*pbTodoList.CreateCommentResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(nil, repository.ErrTaskNotFound)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateComment(ctx, &pbTodoList.CreateCommentRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusTaskNotFound,
		},
		{
			name: "CreateComment_GetTaskStatusInternalServerError",
			input: func() (*pbTodoList.CreateCommentResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(nil, sql.ErrConnDone)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateComment(ctx, &pbTodoList.CreateCommentRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "CreateComment_CreateCommentInternalServerError",
			input: func() (*pbTodoList.CreateCommentResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				commentRepository := new(mockRepository.CommentRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				commentRepository.On("CreateComment", mock.Anything, model.Comment{
					TaskId: tx.Id,
					Value:  strings.ToLower(""),
				}).Return(nil, sql.ErrConnDone)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, commentRepository, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateComment(ctx, &pbTodoList.CreateCommentRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "CreateComment_Success",
			input: func() (*pbTodoList.CreateCommentResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				commentRepository := new(mockRepository.CommentRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				comment := model.Comment{
					TaskId: tx.Id,
					Value:  "level",
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				commentRepository.On("CreateComment", mock.Anything, comment).Return(&comment, nil)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, commentRepository, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateComment(ctx, &pbTodoList.CreateCommentRequest{
					Id:      tx.Id.String(),
					Comment: comment.Value,
				})
			},
			output: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.input()
			if err != nil {
				if er, ok := status.FromError(err); ok {
					if er.Code() != tt.output.Code() {
						t.Errorf("error code: expected %v, received %v", codes.InvalidArgument, er.Code())
					}
					if er.Message() != tt.output.Message() {
						t.Errorf("error message: expected %v, received %v", tt.output.Message(), er.Message())
					}
				}
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*emptypb.Empty, error)
		output *status.Status
	}{
		{
			name: "DeleteComment_TaskIdErrStatusIdMustBeUUID",
			input: func() (*emptypb.Empty, error) {
				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteComment(ctx, &pbTodoList.DeleteCommentRequest{
					Id: "ASD",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "DeleteComment_CommentIdErrStatusIdMustBeUUID",
			input: func() (*emptypb.Empty, error) {
				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteComment(ctx, &pbTodoList.DeleteCommentRequest{
					Id:        uuid.NewV4().String(),
					CommentId: "ASD",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "DeleteComment_DeleteCommentByTaskIdAndCommentIdErrStatusInternalServerError",
			input: func() (*emptypb.Empty, error) {
				var (
					taskId    uuid.UUID = uuid.NewV4()
					commentId uuid.UUID = uuid.NewV4()
				)

				commentRepository := new(mockRepository.CommentRepository)
				commentRepository.On("DeleteCommentByTaskIdAndCommentId", mock.Anything, taskId, commentId).Return(sql.ErrConnDone)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, commentRepository, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteComment(ctx, &pbTodoList.DeleteCommentRequest{
					Id:        taskId.String(),
					CommentId: commentId.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "DeleteComment_DeleteCommentByTaskIdAndCommentIdErrStatusCommentNotFound",
			input: func() (*emptypb.Empty, error) {
				var (
					taskId    uuid.UUID = uuid.NewV4()
					commentId uuid.UUID = uuid.NewV4()
				)

				commentRepository := new(mockRepository.CommentRepository)
				commentRepository.On("DeleteCommentByTaskIdAndCommentId", mock.Anything, taskId, commentId).Return(repository.ErrCommentNotFound)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, commentRepository, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteComment(ctx, &pbTodoList.DeleteCommentRequest{
					Id:        taskId.String(),
					CommentId: commentId.String(),
				})
			},
			output: ErrStatusCommentNotFound,
		},
		{
			name: "DeleteComment_Success",
			input: func() (*emptypb.Empty, error) {
				var (
					taskId    uuid.UUID = uuid.NewV4()
					commentId uuid.UUID = uuid.NewV4()
				)

				commentRepository := new(mockRepository.CommentRepository)
				commentRepository.On("DeleteCommentByTaskIdAndCommentId", mock.Anything, taskId, commentId).Return(nil)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, commentRepository, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteComment(ctx, &pbTodoList.DeleteCommentRequest{
					Id:        taskId.String(),
					CommentId: commentId.String(),
				})
			},
			output: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.input()
			if err != nil {
				if er, ok := status.FromError(err); ok {
					if er.Code() != tt.output.Code() {
						t.Errorf("error code: expected %v, received %v", codes.InvalidArgument, er.Code())
					}
					if er.Message() != tt.output.Message() {
						t.Errorf("error message: expected %v, received %v", tt.output.Message(), er.Message())
					}
				}
			}
		})
	}
}
