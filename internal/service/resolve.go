package service

import (
	"context"

	"buf.build/gen/go/bufbuild/buf/bufbuild/connect-go/buf/alpha/registry/v1alpha1/registryv1alpha1connect"
	registryv1alpha1 "buf.build/gen/go/bufbuild/buf/protocolbuffers/go/buf/alpha/registry/v1alpha1"
	"github.com/bufbuild/connect-go"
	"github.com/pkg/errors"
)

type ResolveService struct {
	Mirror

	upstreamClient registryv1alpha1connect.ResolveServiceClient
	registryv1alpha1connect.UnimplementedResolveServiceHandler
}

func NewResolveService(m Mirror) *ResolveService {
	return &ResolveService{
		Mirror:         m,
		upstreamClient: registryv1alpha1connect.NewResolveServiceClient(m.GetClient()),
	}
}

func (r *ResolveService) GetModulePins(ctx context.Context, req *connect.Request[registryv1alpha1.GetModulePinsRequest]) (*connect.Response[registryv1alpha1.GetModulePinsResponse], error) {
	for _, ref := range req.Msg.ModuleReferences {
		if ref.Remote == r.ServerName() {
			ref.Remote = r.Remote()
		}
	}
	for _, ref := range req.Msg.CurrentModulePins {
		if ref.Remote == r.ServerName() {
			ref.Remote = r.Remote()
		}
	}

	resp, err := r.upstreamClient.GetModulePins(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, pin := range resp.Msg.ModulePins {
		if pin.Remote == r.Remote() {
			pin.Remote = r.ServerName()
		}
	}
	return connect.NewResponse(resp.Msg), nil
}
