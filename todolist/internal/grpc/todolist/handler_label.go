package todolist

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/overridesh/sgg-todolist-service/internal/model"
	"github.com/overridesh/sgg-todolist-service/internal/repository"
	pbTodoList "github.com/overridesh/sgg-todolist-service/proto"
	"github.com/overridesh/sgg-todolist-service/tools"
)

func (svc *todoListGRPC) GetLabels(ctx context.Context, in *pbTodoList.GetLabelsRequest) (*pbTodoList.GetLabelsResponse, error) {
	taskId, err := tools.GetValidUUID(in.GetId())
	if err != nil {
		return nil, err
	}

	task, err := svc.taskRepository.GetTask(ctx, taskId)
	if err != nil {
		if err == repository.ErrTaskNotFound {
			return nil, ErrStatusTaskNotFound.Err()
		}
		zap.S().Errorf("cannot get task", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	response := pbTodoList.GetLabelsResponse{}

	labels, err := svc.labelRepository.GetLabelsByTaskId(ctx, task.Id)
	if err != nil {
		zap.S().Errorf("cannot get labels", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	for _, label := range labels {
		response.Labels = append(response.Labels, &pbTodoList.Label{
			Id:        label.Id.String(),
			Name:      label.Value,
			CreatedAt: tools.FormatDate(label.CreatedAt),
		})
	}

	return &response, nil
}

func (svc *todoListGRPC) CreateLabel(ctx context.Context, in *pbTodoList.CreateLabelRequest) (*pbTodoList.CreateLabelResponse, error) {
	taskId, err := tools.GetValidUUID(in.GetId())
	if err != nil {
		return nil, err
	}

	task, err := svc.taskRepository.GetTask(ctx, taskId)
	if err != nil {
		if err == repository.ErrTaskNotFound {
			return nil, ErrStatusTaskNotFound.Err()
		}
		zap.S().Errorf("cannot get task", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	label, err := svc.labelRepository.CreateLabel(ctx, model.Label{
		TaskId: task.Id,
		Value:  strings.ToLower(in.GetLabel()),
	})
	if err != nil {
		if err == repository.ErrLabelAlreadyExists {
			return nil, ErrStatusLabelAlreadyExists.Err()
		}
		zap.S().Errorf("cannot create label", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	response := pbTodoList.CreateLabelResponse{
		Label: &pbTodoList.Label{
			Id:        label.Id.String(),
			Name:      label.Value,
			CreatedAt: tools.FormatDate(label.CreatedAt),
		},
	}

	if err := tools.SetStatusCode(ctx, http.StatusCreated); err != nil {
		zap.S().Errorf("cannot set new status_code", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}
	return &response, nil
}

func (svc *todoListGRPC) DeleteLabel(ctx context.Context, in *pbTodoList.DeleteLabelRequest) (*emptypb.Empty, error) {
	taskId, err := tools.GetValidUUID(in.GetId())
	if err != nil {
		return nil, err
	}

	labelId, err := tools.GetValidUUID(in.GetLabelId())
	if err != nil {
		return nil, err
	}

	if err := svc.labelRepository.DeleteLabelByTaskIdAndLabelId(ctx, taskId, labelId); err != nil {
		if err == repository.ErrLabelNotFound {
			return nil, ErrStatusErrLabelNotFound.Err()
		}
		zap.S().Errorf("cannot delete comment", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	if err := tools.SetStatusCode(ctx, http.StatusNoContent); err != nil {
		zap.S().Errorf("cannot set new status_code", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	return &emptypb.Empty{}, nil
}
