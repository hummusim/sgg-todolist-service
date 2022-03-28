package todolist

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/overridesh/sgg-todolist-service/internal/model"
	"github.com/overridesh/sgg-todolist-service/internal/repository"
	pbTodoList "github.com/overridesh/sgg-todolist-service/proto"
	"github.com/overridesh/sgg-todolist-service/tools"
)

// NewTodoListGRPC implements the protobuf interface
type todoListGRPC struct {
	taskRepository    repository.TaskRepository
	commentRepository repository.CommentRepository
	labelRepository   repository.LabelRepository
}

// New initializes a new NewTodoListGRPC struct.
func NewGRPC(taskRepository repository.TaskRepository, commentRepository repository.CommentRepository, labelRepository repository.LabelRepository) pbTodoList.TodoListServiceServer {
	return &todoListGRPC{
		taskRepository:    taskRepository,
		commentRepository: commentRepository,
		labelRepository:   labelRepository,
	}
}

func (svc *todoListGRPC) GetTask(ctx context.Context, in *pbTodoList.GetTaskRequest) (*pbTodoList.GetTaskResponse, error) {
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

	response := pbTodoList.GetTaskResponse{
		Id:        task.Id.String(),
		Value:     task.Value,
		Completed: task.Completed,
		CreatedAt: tools.FormatDate(task.CreatedAt),
		UpdatedAt: tools.FormatDate(task.UpdatedAt),
	}

	if task.DueDate.Valid {
		response.DueDate = tools.FormatDate(task.DueDate.Time)
	}

	comments, err := svc.commentRepository.GetCommentsByTaskId(ctx, taskId)
	if err != nil && err != repository.ErrCommentNotFound {
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

	labels, err := svc.labelRepository.GetLabelsByTaskId(ctx, taskId)
	if err != nil && err != repository.ErrLabelNotFound {
		zap.S().Errorf("cannot get task", zap.Error(err))
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

func (svc *todoListGRPC) GetTasks(ctx context.Context, in *pbTodoList.GetTasksRequest) (*pbTodoList.GetTasksResponse, error) {
	var page int32 = in.GetPage()

	tasks, err := svc.taskRepository.GetTasks(ctx, page)
	if err != nil {
		zap.S().Errorf("cannot get tasks", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	response := pbTodoList.GetTasksResponse{}
	for _, task := range tasks {
		taskList := pbTodoList.Task{
			Id:        task.Id.String(),
			Value:     task.Value,
			Completed: task.Completed,
			CreatedAt: tools.FormatDate(task.CreatedAt),
			UpdatedAt: tools.FormatDate(task.UpdatedAt),
		}

		if task.DueDate.Valid {
			taskList.DueDate = tools.FormatDate(task.DueDate.Time)
		}

		response.Tasks = append(response.Tasks, &taskList)
	}

	return &response, nil
}

func (svc *todoListGRPC) CreateTask(ctx context.Context, in *pbTodoList.CreateTaskRequest) (*pbTodoList.CreateTaskResponse, error) {
	var dueDate sql.NullTime

	if len(strings.TrimSpace(in.GetDueDate())) > 0 {
		dueDateTime, err := time.Parse(tools.TimeLayout, in.GetDueDate())
		if err != nil {
			zap.S().Errorf("cannot parse timelayout", zap.Error(err))
			return nil, ErrStatusCannotParseTimeLayout.Err()
		}
		dueDate.Time = dueDateTime
		dueDate.Valid = true
	}

	task, err := svc.taskRepository.CreateTask(ctx, model.Task{
		Value:   in.GetValue(),
		DueDate: dueDate,
	})
	if err != nil {
		zap.S().Errorf("cannot create task", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}
	response := pbTodoList.CreateTaskResponse{
		Task: &pbTodoList.Task{
			Id:        task.Id.String(),
			Value:     task.Value,
			Completed: task.Completed,
			CreatedAt: tools.FormatDate(task.CreatedAt),
			UpdatedAt: tools.FormatDate(task.UpdatedAt),
		},
	}

	if task.DueDate.Valid {
		response.Task.DueDate = tools.FormatDate(task.DueDate.Time)
	}

	if err := tools.SetStatusCode(ctx, http.StatusCreated); err != nil {
		zap.S().Errorf("cannot set new status_code", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	return &response, nil
}

func (svc *todoListGRPC) UpdateTask(ctx context.Context, in *pbTodoList.UpdateTaskRequest) (*pbTodoList.UpdateTaskResponse, error) {
	taskId, err := tools.GetValidUUID(in.GetId())
	if err != nil {
		return nil, err
	}

	task, err := svc.taskRepository.GetTask(ctx, taskId)
	if err != nil {
		if err == repository.ErrTaskNotFound {
			return nil, ErrStatusTaskNotFound.Err()
		}
		zap.S().Errorf("cannot update task", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	task.Completed = in.Completed
	task.Value = in.Value

	if task.DueDate.Valid {
		dueDateTime, err := time.Parse(tools.TimeLayout, in.GetDueDate())
		if err != nil {
			zap.S().Errorf("cannot parse timelayout", zap.Error(err))
			return nil, ErrStatusCannotParseTimeLayout.Err()
		}
		task.DueDate.Time = dueDateTime
	}

	if err := svc.taskRepository.UpdateTask(ctx, task); err != nil {
		zap.S().Errorf("cannot update task", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	response := pbTodoList.UpdateTaskResponse{
		Task: &pbTodoList.Task{
			Id:        task.Id.String(),
			Value:     task.Value,
			Completed: task.Completed,
			CreatedAt: tools.FormatDate(task.CreatedAt),
			UpdatedAt: tools.FormatDate(task.UpdatedAt),
		},
	}

	if task.DueDate.Valid {
		response.Task.DueDate = tools.FormatDate(task.DueDate.Time)
	}

	return &response, nil
}

func (svc *todoListGRPC) DeleteTask(ctx context.Context, in *pbTodoList.DeleteTaskRequest) (*emptypb.Empty, error) {
	taskId, err := tools.GetValidUUID(in.GetId())
	if err != nil {
		return nil, err
	}

	if err := svc.taskRepository.DeleteTask(ctx, taskId); err != nil {
		if err == repository.ErrTaskNotFound {
			return nil, ErrStatusTaskNotFound.Err()
		}
		zap.S().Errorf("cannot delete task", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	if err := tools.SetStatusCode(ctx, http.StatusNoContent); err != nil {
		zap.S().Errorf("cannot set new status_code", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	return &emptypb.Empty{}, nil
}

func (svc *todoListGRPC) UpdateTaskStatus(ctx context.Context, in *pbTodoList.UpdateTaskStatusRequest) (*emptypb.Empty, error) {
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

	task.Completed = in.GetCompleted()

	if err := svc.taskRepository.UpdateTask(ctx, task); err != nil {
		zap.S().Errorf("cannot update task", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	if err := tools.SetStatusCode(ctx, http.StatusNoContent); err != nil {
		zap.S().Errorf("cannot set new status_code", zap.Error(err))
		return nil, ErrStatusInternalServerError.Err()
	}

	return &emptypb.Empty{}, nil
}
