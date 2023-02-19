package service

import (
	"context"

	"github.com/bufbuild/buf/private/bufpkg/bufmodule/bufmoduleref"
	"github.com/bufbuild/buf/private/gen/proto/connect/buf/alpha/registry/v1alpha1/registryv1alpha1connect"
	registryv1alpha1 "github.com/bufbuild/buf/private/gen/proto/go/buf/alpha/registry/v1alpha1"
	"github.com/bufbuild/connect-go"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
)

type DownloadService struct {
	Mirror

	upstreamClient registryv1alpha1connect.DownloadServiceClient
	logger         log.Logger
	registryv1alpha1connect.UnimplementedDownloadServiceHandler
}

func NewDownloadService(m Mirror, logger log.Logger) *DownloadService {
	return &DownloadService{
		Mirror:         m,
		upstreamClient: registryv1alpha1connect.NewDownloadServiceClient(m.GetClient()),
		logger:         logger,
	}
}

func (d *DownloadService) Download(ctx context.Context, req *connect.Request[registryv1alpha1.DownloadRequest]) (*connect.Response[registryv1alpha1.DownloadResponse], error) {
	ref, err := bufmoduleref.NewModuleReference(d.Remote(), req.Msg.Owner, req.Msg.Repository, req.Msg.Reference)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	oss := d.ObjectStorage()
	exist, err := oss.HasModuleSnapshot(ctx, ref)
	if err != nil {
		return nil, err
	}
	if exist {
		cached, err := oss.FetchModuleSnapshot(ctx, ref)
		if err == nil {
			return connect.NewResponse(cached), nil
		}
		log.NewHelper(d.logger).Errorf("failed to fetch module snapshot: %v, %+v", ref, err)
	}

	resp, err := d.upstreamClient.Download(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := oss.CreateModuleSnapshot(ctx, ref, resp.Msg); err != nil {
		log.NewHelper(d.logger).Errorf("failed to create module snapshot: %v, %+v", ref, err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (d *DownloadService) DownloadManifestAndBlobs(ctx context.Context, req *connect.Request[registryv1alpha1.DownloadManifestAndBlobsRequest]) (*connect.Response[registryv1alpha1.DownloadManifestAndBlobsResponse], error) {
	resp, err := d.upstreamClient.DownloadManifestAndBlobs(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp.Msg), nil
}
