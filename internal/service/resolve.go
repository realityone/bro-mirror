package service

import (
	"context"

	"buf.build/gen/go/bufbuild/buf/bufbuild/connect-go/buf/alpha/registry/v1alpha1/registryv1alpha1connect"
	registryv1alpha1 "buf.build/gen/go/bufbuild/buf/protocolbuffers/go/buf/alpha/registry/v1alpha1"
	"github.com/bufbuild/connect-go"
)

type ResolveService struct {
	Upstream

	upstreamClient registryv1alpha1connect.ResolveServiceClient
	registryv1alpha1connect.UnimplementedResolveServiceHandler
}

func NewResolveService(m *Mirror) *ResolveService {
	return &ResolveService{
		Upstream:       m,
		upstreamClient: registryv1alpha1connect.NewResolveServiceClient(m.GetClient()),
	}
}

func (r *ResolveService) GetModulePins(ctx context.Context, req *connect.Request[registryv1alpha1.GetModulePinsRequest]) (*connect.Response[registryv1alpha1.GetModulePinsResponse], error) {
	for _, ref := range req.Msg.ModuleReferences {
		if ref.Remote == "bapis.net" {
			ref.Remote = "buf.build"
		}
	}
	for _, ref := range req.Msg.CurrentModulePins {
		if ref.Remote == "bapis.net" {
			ref.Remote = "buf.build"
		}
	}

	resp, err := r.upstreamClient.GetModulePins(ctx, req)
	if err != nil {
		return nil, err
	}
	for _, pin := range resp.Msg.ModulePins {
		if pin.Remote == "buf.build" {
			pin.Remote = "bapis.net"
		}
	}
	return connect.NewResponse(resp.Msg), nil
}
