package tools

import (
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrStatusIdMustBeUUID     *status.Status = status.New(codes.InvalidArgument, "the id must be uuid")
	ErrStatusIdMusBeValidUUID *status.Status = status.New(codes.InvalidArgument, "the id must be a valid uuid")
)

func GetValidUUID(id string) (uuid.UUID, error) {
	validUUID, err := uuid.FromString(id)
	if err != nil {
		return uuid.Nil, ErrStatusIdMustBeUUID.Err()
	}

	if validUUID == uuid.Nil {
		return uuid.Nil, ErrStatusIdMusBeValidUUID.Err()
	}

	return validUUID, nil
}
