package todolist

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/overridesh/sgg-todolist-service/internal/model"
	"github.com/overridesh/sgg-todolist-service/internal/repository"
	pbTodoList "github.com/overridesh/sgg-todolist-service/proto"
	"github.com/overridesh/sgg-todolist-service/tools"
)

func (svc *todoListGRPC) GetComments(ctx context.Context, in *pbTodoList.GetCommentsRequest) (*pbTodoList.GetCommentsResponse, error) {
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

	response := pbTodoList.GetCommentsResponse{}

	comments, err := svc.commentRepository.GetCommentsByTaskId(ctx, task.Id)
	if err != nil {
		zap.S().Errorf("cannot get task", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	for _, comment := range comments {
		response.Comments = append(response.Comments, &pbTodoList.Comment{
			Id:        comment.Id.String(),
			Message:   comment.Value,
			CreatedAt: tools.FormatDate(comment.CreatedAt),
		})
	}

	return &response, nil
}

func (svc *todoListGRPC) CreateComment(ctx context.Context, in *pbTodoList.CreateCommentRequest) (*pbTodoList.CreateCommentResponse, error) {
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

	comment, err := svc.commentRepository.CreateComment(ctx, model.Comment{
		TaskId: task.Id,
		Value:  in.GetComment(),
	})
	if err != nil {
		zap.S().Errorf("cannot create comment", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}
	response := pbTodoList.CreateCommentResponse{
		Comment: &pbTodoList.Comment{
			Id:        comment.Id.String(),
			Message:   comment.Value,
			CreatedAt: tools.FormatDate(comment.CreatedAt),
		},
	}

	if err := tools.SetStatusCode(ctx, http.StatusCreated); err != nil {
		zap.S().Errorf("cannot set new status_code", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	return &response, nil
}

func (svc *todoListGRPC) DeleteComment(ctx context.Context, in *pbTodoList.DeleteCommentRequest) (*emptypb.Empty, error) {
	taskId, err := tools.GetValidUUID(in.GetId())
	if err != nil {
		return nil, err
	}

	commentId, err := tools.GetValidUUID(in.GetCommentId())
	if err != nil {
		return nil, err
	}

	if err := svc.commentRepository.DeleteCommentByTaskIdAndCommentId(ctx, taskId, commentId); err != nil {
		if err == repository.ErrCommentNotFound {
			return nil, ErrStatusCommentNotFound.Err()
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
