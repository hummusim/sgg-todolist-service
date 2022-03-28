package todolist

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/overridesh/sgg-todolist-service/internal/repository"
)

var (
	ErrStatusInternalServerError   *status.Status = status.New(codes.Internal, "internal server error")
	ErrStatusTaskNotFound          *status.Status = status.New(codes.NotFound, repository.ErrTaskNotFound.Error())
	ErrStatusCommentNotFound       *status.Status = status.New(codes.NotFound, repository.ErrCommentNotFound.Error())
	ErrStatusErrLabelNotFound      *status.Status = status.New(codes.NotFound, repository.ErrLabelNotFound.Error())
	ErrStatusLabelAlreadyExists    *status.Status = status.New(codes.AlreadyExists, repository.ErrLabelAlreadyExists.Error())
	ErrStatusCannotParseTimeLayout *status.Status = status.New(codes.InvalidArgument, "cannot parse timelayout")
)
