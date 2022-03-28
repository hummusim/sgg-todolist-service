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

func TestGetLabels(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*pbTodoList.GetLabelsResponse, error)
		output *status.Status
	}{
		{
			name: "GetLabels_ErrStatusIdMustBeUUID",
			input: func() (*pbTodoList.GetLabelsResponse, error) {
				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetLabels(ctx, &pbTodoList.GetLabelsRequest{
					Id: "ASD",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "GetLabels_ErrStatusTaskNotFound",
			input: func() (*pbTodoList.GetLabelsResponse, error) {
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

				return client.GetLabels(ctx, &pbTodoList.GetLabelsRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusTaskNotFound,
		},
		{
			name: "GetLabels_GetLabelsByTaskIdErrStatusInternalServerError",
			input: func() (*pbTodoList.GetLabelsResponse, error) {
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

				return client.GetLabels(ctx, &pbTodoList.GetLabelsRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "GetLabels_GetLabelsByTaskIdErrStatusInternalServerError",
			input: func() (*pbTodoList.GetLabelsResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				labelRepository := new(mockRepository.LabelRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				labelRepository.On("GetLabelsByTaskId", mock.Anything, tx.Id).Return(nil, sql.ErrConnDone)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetLabels(ctx, &pbTodoList.GetLabelsRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "GetLabels_Success",
			input: func() (*pbTodoList.GetLabelsResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				labelRepository := new(mockRepository.LabelRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				labelRepository.On("GetLabelsByTaskId", mock.Anything, tx.Id).Return([]*model.Label{}, nil)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetLabels(ctx, &pbTodoList.GetLabelsRequest{
					Id: tx.Id.String(),
				})
			},
			output: nil,
		},
		{
			name: "GetLabels_SuccessWithValues",
			input: func() (*pbTodoList.GetLabelsResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				labelRepository := new(mockRepository.LabelRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				labelRepository.On("GetLabelsByTaskId", mock.Anything, tx.Id).Return([]*model.Label{{
					Value: "Label",
				}}, nil)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.GetLabels(ctx, &pbTodoList.GetLabelsRequest{
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

func TestCreateLabel(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*pbTodoList.CreateLabelResponse, error)
		output *status.Status
	}{
		{
			name: "CreateLabel_ErrStatusIdMustBeUUID",
			input: func() (*pbTodoList.CreateLabelResponse, error) {
				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateLabel(ctx, &pbTodoList.CreateLabelRequest{
					Id: "ASD",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "CreateLabel_ErrStatusTaskNotFound",
			input: func() (*pbTodoList.CreateLabelResponse, error) {
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

				return client.CreateLabel(ctx, &pbTodoList.CreateLabelRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusTaskNotFound,
		},
		{
			name: "CreateLabel_GetTaskStatusInternalServerError",
			input: func() (*pbTodoList.CreateLabelResponse, error) {
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

				return client.CreateLabel(ctx, &pbTodoList.CreateLabelRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "CreateLabel_CreateLabelInternalServerError",
			input: func() (*pbTodoList.CreateLabelResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				labelRepository := new(mockRepository.LabelRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				labelRepository.On("CreateLabel", mock.Anything, model.Label{
					TaskId: tx.Id,
					Value:  strings.ToLower(""),
				}).Return(nil, sql.ErrConnDone)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateLabel(ctx, &pbTodoList.CreateLabelRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "CreateLabel_CreateLabelErrStatusLabelAlreadyExists",
			input: func() (*pbTodoList.CreateLabelResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				labelRepository := new(mockRepository.LabelRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				labelRepository.On("CreateLabel", mock.Anything, model.Label{
					TaskId: tx.Id,
					Value:  strings.ToLower(""),
				}).Return(nil, repository.ErrLabelAlreadyExists)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateLabel(ctx, &pbTodoList.CreateLabelRequest{
					Id: tx.Id.String(),
				})
			},
			output: ErrStatusLabelAlreadyExists,
		},
		{
			name: "CreateLabel_Success",
			input: func() (*pbTodoList.CreateLabelResponse, error) {
				taskRepository := new(mockRepository.TaskRepository)
				labelRepository := new(mockRepository.LabelRepository)

				tx := model.Task{
					Id: uuid.NewV4(),
				}

				label := model.Label{
					TaskId: tx.Id,
					Value:  "level",
				}

				taskRepository.On("GetTask", mock.Anything, tx.Id).Return(&tx, nil)
				labelRepository.On("CreateLabel", mock.Anything, label).Return(&label, nil)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(taskRepository, nil, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.CreateLabel(ctx, &pbTodoList.CreateLabelRequest{
					Id:    tx.Id.String(),
					Label: label.Value,
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

func TestDeleteLabel(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*emptypb.Empty, error)
		output *status.Status
	}{
		{
			name: "DeleteLabel_TaskIdErrStatusIdMustBeUUID",
			input: func() (*emptypb.Empty, error) {
				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteLabel(ctx, &pbTodoList.DeleteLabelRequest{
					Id: "ASD",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "DeleteLabel_LabelIdErrStatusIdMustBeUUID",
			input: func() (*emptypb.Empty, error) {
				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, nil)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteLabel(ctx, &pbTodoList.DeleteLabelRequest{
					Id:      uuid.NewV4().String(),
					LabelId: "ASD",
				})
			},
			output: tools.ErrStatusIdMustBeUUID,
		},
		{
			name: "DeleteLabel_DeleteLabelByTaskIdAndLabelIdErrStatusInternalServerError",
			input: func() (*emptypb.Empty, error) {
				var (
					taskId  uuid.UUID = uuid.NewV4()
					labelId uuid.UUID = uuid.NewV4()
				)

				labelRepository := new(mockRepository.LabelRepository)
				labelRepository.On("DeleteLabelByTaskIdAndLabelId", mock.Anything, taskId, labelId).Return(sql.ErrConnDone)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteLabel(ctx, &pbTodoList.DeleteLabelRequest{
					Id:      taskId.String(),
					LabelId: labelId.String(),
				})
			},
			output: ErrStatusInternalServerError,
		},
		{
			name: "DeleteLabel_DeleteLabelByTaskIdAndLabelIdErrStatusErrLabelNotFound",
			input: func() (*emptypb.Empty, error) {
				var (
					taskId  uuid.UUID = uuid.NewV4()
					labelId uuid.UUID = uuid.NewV4()
				)

				labelRepository := new(mockRepository.LabelRepository)
				labelRepository.On("DeleteLabelByTaskIdAndLabelId", mock.Anything, taskId, labelId).Return(repository.ErrLabelNotFound)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteLabel(ctx, &pbTodoList.DeleteLabelRequest{
					Id:      taskId.String(),
					LabelId: labelId.String(),
				})
			},
			output: ErrStatusErrLabelNotFound,
		},
		{
			name: "DeleteLabel_Success",
			input: func() (*emptypb.Empty, error) {
				var (
					taskId  uuid.UUID = uuid.NewV4()
					labelId uuid.UUID = uuid.NewV4()
				)

				labelRepository := new(mockRepository.LabelRepository)
				labelRepository.On("DeleteLabelByTaskIdAndLabelId", mock.Anything, taskId, labelId).Return(nil)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil, nil, labelRepository)))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				client := pbTodoList.NewTodoListServiceClient(conn)

				return client.DeleteLabel(ctx, &pbTodoList.DeleteLabelRequest{
					Id:      taskId.String(),
					LabelId: labelId.String(),
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
