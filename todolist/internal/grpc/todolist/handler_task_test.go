package todolist

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/overridesh/sgg-todolist-service/internal/model"
	"github.com/overridesh/sgg-todolist-service/internal/repository"
	mockRepository "github.com/overridesh/sgg-todolist-service/pkg/mock"
	pbTodoList "github.com/overridesh/sgg-todolist-service/proto"
	"github.com/overridesh/sgg-todolist-service/tools"
)

func TestGetTask(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*pbTodoList.GetTaskResponse, error)
		output *status.Status
	}{
		{
			name: "GetTask_Success",
			input: func() (*pbTodoList.GetTaskResponse, error) {
				commentRepository := new(mockRepository.CommentRepository)
				labelRepository := new(mockRepository.LabelRepository)
				taskRepository := new(mockRepository.TaskRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				commentRepository.On("GetCommentsByTaskId", mock.Anything, tx.Id).Return(nil, repository.ErrCommentNotFound)
				labelRepository.On("GetLabelsByTaskId", mock.Anything, tx.Id).Return(nil, repository.ErrLabelNotFound)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, commentRepository, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetTask(ctx, &pbTodoList.GetTaskRequest{
					Id: tx.Id.String(),
				})
			},
			output: nil,
		},
		{
			name: "GetTask_LabelsErrStatusInternalServerError",
			input: func() (*pbTodoList.GetTaskResponse, error) {
				commentRepository := new(mockRepository.CommentRepository)
				labelRepository := new(mockRepository.LabelRepository)
				taskRepository := new(mockRepository.TaskRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				commentRepository.On("GetCommentsByTaskId", mock.Anything, tx.Id).Return(nil, repository.ErrCommentNotFound)
				labelRepository.On("GetLabelsByTaskId", mock.Anything, tx.Id).Return(nil, sql.ErrConnDone)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, commentRepository, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetTask(ctx, &pbTodoList.GetTaskRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "GetTask_CommentsErrStatusInternalServerError",
			input: func() (*pbTodoList.GetTaskResponse, error) {
				commentRepository := new(mockRepository.CommentRepository)
				labelRepository := new(mockRepository.LabelRepository)
				taskRepository := new(mockRepository.TaskRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				commentRepository.On("GetCommentsByTaskId", mock.Anything, tx.Id).Return(nil, sql.ErrConnDone)
				labelRepository.On("GetLabelsByTaskId", mock.Anything, tx.Id).Return(nil, repository.ErrLabelNotFound)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, commentRepository, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetTask(ctx, &pbTodoList.GetTaskRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "GetTask_IdMustBeUUID",
			input: func() (*pbTodoList.GetTaskResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetTask(context.Background(), &pbTodoList.GetTaskRequest{
					Id: "ABC",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "GetTask_IdMusBeValidUUID",
			input: func() (*pbTodoList.GetTaskResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetTask(context.Background(), &pbTodoList.GetTaskRequest{
					Id: uuid.Nil.String(),
				})
			},
			output: tools.ErrStatusIdMusBeValidUUID,
		},
		{
			name: "GetTask_InternalErr",
			input: func() (*pbTodoList.GetTaskResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(nil, errors.New("unknow_error"))

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetTask(context.Background(), &pbTodoList.GetTaskRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "GetTask_NotFound",
			input: func() (*pbTodoList.GetTaskResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(nil, repository.ErrTaskNotFound)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetTask(context.Background(), &pbTodoList.GetTaskRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusTaskNotFound,
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

func TestGetTasks(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*pbTodoList.GetTasksResponse, error)
		output *status.Status
	}{
		{
			name: "GetTasks_Success",
			input: func() (*pbTodoList.GetTasksResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				var page int32 = 1
				taskRepository.On("GetTasks", mock.Anything, page).Return([]*model.Task{}, nil)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetTasks(context.Background(), &pbTodoList.GetTasksRequest{
					Page: page,
				})
			},
			output: nil,
		},
		{
			name: "GetTasks_SuccessWithValues",
			input: func() (*pbTodoList.GetTasksResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				var page int32 = 1
				taskRepository.On("GetTasks", mock.Anything, page).Return([]*model.Task{{
					Id: uuid.NewV4(),
				}}, nil)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetTasks(context.Background(), &pbTodoList.GetTasksRequest{
					Page: page,
				})
			},
			output: nil,
		},
		{
			name: "GetTasks_LabelsErrStatusInternalServerError",
			input: func() (*pbTodoList.GetTasksResponse, error) {
				var page int32 = 1
				taskRepository := new(mockRepository.TaskRepository)
				taskRepository.On("GetTasks", mock.Anything, page).Return(nil, errors.New("unknow_error"))

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetTasks(context.Background(), &pbTodoList.GetTasksRequest{
					Page: page,
				})
			},
			output: ErrStatusInternalServerError,
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

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*pbTodoList.CreateTaskResponse, error)
		output *status.Status
	}{
		{
			name: "CreateTask_Success",
			input: func() (*pbTodoList.CreateTaskResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				task := model.Task{}
				taskRepository.On("CreateTask", mock.Anything, task).Return(&task, nil)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateTask(context.Background(), &pbTodoList.CreateTaskRequest{
					Value: task.Value,
				})
			},
			output: nil,
		},
		{
			name: "CreateTask_ErrStatusCannotParseTimeLayout",
			input: func() (*pbTodoList.CreateTaskResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateTask(context.Background(), &pbTodoList.CreateTaskRequest{
					Value:   uuid.NewV4().String(),
					DueDate: uuid.NewV4().String(),
				})
			},
			output: ErrStatusCannotParseTimeLayout,
		},
		{
			name: "CreateTask_ErrStatusInternalServerError",
			input: func() (*pbTodoList.CreateTaskResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				task := model.Task{}
				taskRepository.On("CreateTask", mock.Anything, task).Return(&task, errors.New("uknow error"))

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateTask(context.Background(), &pbTodoList.CreateTaskRequest{
					Value: task.Value,
				})
			},
			output: ErrStatusInternalServerError,
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

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*pbTodoList.UpdateTaskResponse, error)
		output *status.Status
	}{
		{
			name: "UpdateTask_ErrGetValidUUID",
			input: func() (*pbTodoList.UpdateTaskResponse, error) {
				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTask(context.Background(), &pbTodoList.UpdateTaskRequest{
					Id: "ASD",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "UpdateTask_ErrStatusTaskNotFound",
			input: func() (*pbTodoList.UpdateTaskResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(nil, repository.ErrTaskNotFound)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTask(context.Background(), &pbTodoList.UpdateTaskRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusTaskNotFound,
		},
		{
			name: "UpdateTask_ErrStatusTaskNotFoundOnUpdate",
			input: func() (*pbTodoList.UpdateTaskResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				taskRepository.On("UpdateTask", mock.Anything, &tx).Return(nil, repository.ErrTaskNotFound)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTask(context.Background(), &pbTodoList.UpdateTaskRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusTaskNotFound,
		},
		{
			name: "UpdateTask_ErrStatusInternalServerError",
			input: func() (*pbTodoList.UpdateTaskResponse, error) {
				task := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository := new(mockRepository.TaskRepository)
				taskRepository.On("GetTask", mock.Anything, task.Id).Return(nil, errors.New("unknown_error"))

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTask(context.Background(), &pbTodoList.UpdateTaskRequest{
					Id: task.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "UpdateTask_ErrStatusInternalServerErrorOnUpdate",
			input: func() (*pbTodoList.UpdateTaskResponse, error) {
				task := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository := new(mockRepository.TaskRepository)
				taskRepository.On("GetTask", mock.Anything, task.Id).Return(&task, nil)
				taskRepository.On("UpdateTask", mock.Anything, &task).Return(errors.New("unknown_error"))

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTask(context.Background(), &pbTodoList.UpdateTaskRequest{
					Id: task.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "UpdateTask_Success",
			input: func() (*pbTodoList.UpdateTaskResponse, error) {
				task := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository := new(mockRepository.TaskRepository)
				taskRepository.On("GetTask", mock.Anything, task.Id).Return(&task, nil)
				taskRepository.On("UpdateTask", mock.Anything, &task).Return(nil)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTask(context.Background(), &pbTodoList.UpdateTaskRequest{
					Id: task.Id.String(),
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

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*emptypb.Empty, error)
		output *status.Status
	}{
		{
			name: "DeleteTask_ErrGetValidUUID",
			input: func() (*emptypb.Empty, error) {
				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteTask(context.Background(), &pbTodoList.DeleteTaskRequest{
					Id: "ASD",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "DeleteTask_ErrStatusTaskNotFound",
			input: func() (*emptypb.Empty, error) {
				taskRepository := new(mockRepository.TaskRepository)
				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("DeleteTask", mock.Anything, tx.Id).Return(repository.ErrTaskNotFound)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteTask(context.Background(), &pbTodoList.DeleteTaskRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusTaskNotFound,
		},
		{
			name: "DeleteTask_ErrStatusInternalServerError",
			input: func() (*emptypb.Empty, error) {
				task := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository := new(mockRepository.TaskRepository)
				taskRepository.On("DeleteTask", mock.Anything, task.Id).Return(errors.New("unknown_error"))

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteTask(context.Background(), &pbTodoList.DeleteTaskRequest{
					Id: task.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "DeleteTask_Success",
			input: func() (*emptypb.Empty, error) {
				task := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository := new(mockRepository.TaskRepository)
				taskRepository.On("DeleteTask", mock.Anything, task.Id).Return(nil)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteTask(context.Background(), &pbTodoList.DeleteTaskRequest{
					Id: task.Id.String(),
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

func TestUpdateTaskStatus(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*emptypb.Empty, error)
		output *status.Status
	}{
		{
			name: "UpdateTask_ErrGetValidUUID",
			input: func() (*emptypb.Empty, error) {
				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTaskStatus(context.Background(), &pbTodoList.UpdateTaskStatusRequest{
					Id: "ASD",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "UpdateTask_ErrStatusTaskNotFound",
			input: func() (*emptypb.Empty, error) {
				taskRepository := new(mockRepository.TaskRepository)
				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(nil, repository.ErrTaskNotFound)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTaskStatus(context.Background(), &pbTodoList.UpdateTaskStatusRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusTaskNotFound,
		},
		{
			name: "UpdateTask_ErrStatusTaskNotFoundOnUpdate",
			input: func() (*emptypb.Empty, error) {
				taskRepository := new(mockRepository.TaskRepository)
				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				taskRepository.On("UpdateTask", mock.Anything, &tx).Return(nil, repository.ErrTaskNotFound)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTaskStatus(context.Background(), &pbTodoList.UpdateTaskStatusRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusTaskNotFound,
		},
		{
			name: "UpdateTask_ErrStatusInternalServerErrorOnGet",
			input: func() (*emptypb.Empty, error) {
				task := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository := new(mockRepository.TaskRepository)
				taskRepository.On("GetTask", mock.Anything, task.Id).Return(nil, errors.New("unknown_error"))

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTaskStatus(context.Background(), &pbTodoList.UpdateTaskStatusRequest{
					Id: task.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "UpdateTask_ErrStatusInternalServerErroOnUpdater",
			input: func() (*emptypb.Empty, error) {
				task := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository := new(mockRepository.TaskRepository)
				taskRepository.On("GetTask", mock.Anything, task.Id).Return(&task, nil)
				taskRepository.On("UpdateTask", mock.Anything, &task).Return(errors.New("unknown_error"))

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTaskStatus(context.Background(), &pbTodoList.UpdateTaskStatusRequest{
					Id: task.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "UpdateTask_ErrStatusInternalServerErrorOnUpdate",
			input: func() (*emptypb.Empty, error) {
				task := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository := new(mockRepository.TaskRepository)
				taskRepository.On("GetTask", mock.Anything, task.Id).Return(&task, nil)
				taskRepository.On("UpdateTask", mock.Anything, &task).Return(errors.New("unknown_error"))

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTaskStatus(context.Background(), &pbTodoList.UpdateTaskStatusRequest{
					Id: task.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "UpdateTask_Success",
			input: func() (*emptypb.Empty, error) {
				task := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository := new(mockRepository.TaskRepository)
				taskRepository.On("GetTask", mock.Anything, task.Id).Return(&task, nil)
				taskRepository.On("UpdateTask", mock.Anything, &task).Return(nil)

				conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.UpdateTaskStatus(context.Background(), &pbTodoList.UpdateTaskStatusRequest{
					Id: task.Id.String(),
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
