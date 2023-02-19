package service

import (
	"context"

	"github.com/bufbuild/buf/private/gen/proto/connect/buf/alpha/registry/v1alpha1/registryv1alpha1connect"
	registryv1alpha1 "github.com/bufbuild/buf/private/gen/proto/go/buf/alpha/registry/v1alpha1"
	"github.com/bufbuild/connect-go"
)

type RepositoryCommitService struct {
	Mirror

	upstreamClient registryv1alpha1connect.RepositoryCommitServiceClient
	registryv1alpha1connect.UnimplementedRepositoryCommitServiceHandler
}

func NewRepositoryCommitService(m Mirror) *RepositoryCommitService {
	return &RepositoryCommitService{
		Mirror:         m,
		upstreamClient: registryv1alpha1connect.NewRepositoryCommitServiceClient(m.GetClient()),
	}
}

func (r *RepositoryCommitService) GetRepositoryCommitByReference(ctx context.Context, req *connect.Request[registryv1alpha1.GetRepositoryCommitByReferenceRequest]) (*connect.Response[registryv1alpha1.GetRepositoryCommitByReferenceResponse], error) {
	resp, err := r.upstreamClient.GetRepositoryCommitByReference(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp.Msg), nil
}
