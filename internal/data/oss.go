package data

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/bufbuild/buf/private/bufpkg/bufmodule/bufmoduleref"
	registryv1alpha1 "github.com/bufbuild/buf/private/gen/proto/go/buf/alpha/registry/v1alpha1"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"github.com/realityone/bro-mirror/internal/conf"
	"github.com/tencentyun/cos-go-sdk-v5"
	"google.golang.org/protobuf/proto"
)

type ObjectStorage struct {
	cos    *cos.Client
	logger log.Logger
}

func NewObjectStorage(cfg *conf.Data_TencentOSS, logger log.Logger) (*ObjectStorage, func(), error) {
	data := &ObjectStorage{
		logger: logger,
	}

	bucketURL, err := url.Parse(cfg.BucketUrl)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	ossTimeout := time.Second * 20
	if cfg.Timeout != nil {
		ossTimeout = cfg.Timeout.AsDuration()
	}
	client := cos.NewClient(&cos.BaseURL{
		BucketURL: bucketURL,
	}, &http.Client{
		Timeout: ossTimeout,
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretId,
			SecretKey: cfg.SecretKey,
		},
	})
	data.cos = client

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return data, cleanup, nil
}

func SnapshotKey(ref bufmoduleref.ModuleReference) string {
	return fmt.Sprintf("%s.gz", path.Join(ref.IdentityString(), ref.Reference()))
}

func (o *ObjectStorage) HasModuleSnapshot(ctx context.Context, ref bufmoduleref.ModuleReference) (bool, error) {
	if !bufmoduleref.IsCommitModuleReference(ref) {
		return false, errors.Errorf("reference is not a commit: %s", ref)
	}
	key := SnapshotKey(ref)
	exist, err := o.cos.Object.IsExist(ctx, key)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return exist, nil
}

func (o *ObjectStorage) CreateModuleSnapshot(ctx context.Context, ref bufmoduleref.ModuleReference, snapshot *registryv1alpha1.DownloadResponse) error {
	key := SnapshotKey(ref)
	rawData, err := proto.Marshal(snapshot)
	if err != nil {
		return errors.WithStack(err)
	}
	gzipped, err := gzippedBuffer(rawData)
	if err != nil {
		return err
	}
	if _, err := o.cos.Object.Put(ctx, key, gzipped, nil); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (o *ObjectStorage) FetchModuleSnapshot(ctx context.Context, ref bufmoduleref.ModuleReference) (*registryv1alpha1.DownloadResponse, error) {
	key := SnapshotKey(ref)
	resp, err := o.cos.Object.Get(ctx, key, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	data, err := ungzippedBuffer(resp.Body)
	if err != nil {
		return nil, err
	}
	snapshot := &registryv1alpha1.DownloadResponse{}
	if err := proto.Unmarshal(data, snapshot); err != nil {
		return nil, errors.WithStack(err)
	}
	return snapshot, nil
}

func gzippedBuffer(rawData []byte) (*bytes.Buffer, error) {
	gzipped := bytes.NewBuffer(nil)
	writer := gzip.NewWriter(gzipped)
	if _, err := writer.Write(rawData); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := writer.Close(); err != nil {
		return nil, errors.WithStack(err)
	}
	return gzipped, nil
}

func ungzippedBuffer(r io.Reader) ([]byte, error) {
	ungzipped, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer ungzipped.Close()
	data, err := io.ReadAll(ungzipped)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return data, nil
}
