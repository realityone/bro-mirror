package service

import (
	"context"

	"github.com/bufbuild/buf/private/gen/proto/connect/buf/alpha/registry/v1alpha1/registryv1alpha1connect"
	registryv1alpha1 "github.com/bufbuild/buf/private/gen/proto/go/buf/alpha/registry/v1alpha1"
	"github.com/bufbuild/connect-go"
)

type DownloadService struct {
	Mirror

	upstreamClient registryv1alpha1connect.DownloadServiceClient
	registryv1alpha1connect.UnimplementedDownloadServiceHandler
}

func NewDownloadService(m Mirror) *DownloadService {
	return &DownloadService{
		Mirror:         m,
		upstreamClient: registryv1alpha1connect.NewDownloadServiceClient(m.GetClient()),
	}
}

func (d *DownloadService) Download(ctx context.Context, req *connect.Request[registryv1alpha1.DownloadRequest]) (*connect.Response[registryv1alpha1.DownloadResponse], error) {
	// oss := d.ObjectStorage()

	resp, err := d.upstreamClient.Download(ctx, req)
	if err != nil {
		return nil, err
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
