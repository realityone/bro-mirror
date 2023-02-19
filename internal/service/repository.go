package service

import (
	"context"

	"buf.build/gen/go/bufbuild/buf/bufbuild/connect-go/buf/alpha/registry/v1alpha1/registryv1alpha1connect"
	registryv1alpha1 "buf.build/gen/go/bufbuild/buf/protocolbuffers/go/buf/alpha/registry/v1alpha1"
	"github.com/bufbuild/connect-go"
)

type RepositoryService struct {
	Upstream

	upstreamClient registryv1alpha1connect.RepositoryServiceClient
	registryv1alpha1connect.UnimplementedRepositoryServiceHandler
}

func NewRepositoryService(m *Mirror) *RepositoryService {
	return &RepositoryService{
		Upstream:       m,
		upstreamClient: registryv1alpha1connect.NewRepositoryServiceClient(m.GetClient()),
	}
}

func (r *RepositoryService) GetRepositoriesByFullName(ctx context.Context, req *connect.Request[registryv1alpha1.GetRepositoriesByFullNameRequest]) (*connect.Response[registryv1alpha1.GetRepositoriesByFullNameResponse], error) {
	resp, err := r.upstreamClient.GetRepositoriesByFullName(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp.Msg), nil
}

func (r *RepositoryService) GetRepositoryByFullName(ctx context.Context, req *connect.Request[registryv1alpha1.GetRepositoryByFullNameRequest]) (*connect.Response[registryv1alpha1.GetRepositoryByFullNameResponse], error) {
	resp, err := r.upstreamClient.GetRepositoryByFullName(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp.Msg), nil
}

func (r *RepositoryService) ListRepositories(ctx context.Context, req *connect.Request[registryv1alpha1.ListRepositoriesRequest]) (*connect.Response[registryv1alpha1.ListRepositoriesResponse], error) {
	resp, err := r.upstreamClient.ListRepositories(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp.Msg), nil
}

func (r *RepositoryService) ListUserRepositories(ctx context.Context, req *connect.Request[registryv1alpha1.ListUserRepositoriesRequest]) (*connect.Response[registryv1alpha1.ListUserRepositoriesResponse], error) {
	resp, err := r.upstreamClient.ListUserRepositories(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp.Msg), nil
}

func (r *RepositoryService) DeleteRepository(ctx context.Context, req *connect.Request[registryv1alpha1.DeleteRepositoryRequest]) (*connect.Response[registryv1alpha1.DeleteRepositoryResponse], error) {
	resp, err := r.upstreamClient.DeleteRepository(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp.Msg), nil
}
