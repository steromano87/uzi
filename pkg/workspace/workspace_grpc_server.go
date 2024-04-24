package workspace

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

/////////////////////////
// GRPC implementation //
/////////////////////////

func (w *Workspace) Initialize(_ context.Context, request *InitializationRequest) (*InitializationResponse, error) {
	if err := w.ExtractFromArchive(request.GetArchive(), request.GetCompressionAlgorithm()); err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &InitializationResponse{RemoteWorkingFolder: w.location}, nil
}

func (w *Workspace) Reset(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := w.DeleteContent(); err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &emptypb.Empty{}, nil
}
