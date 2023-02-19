package data

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/bufbuild/buf/private/bufpkg/bufmodule/bufmoduleref"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"github.com/realityone/bro-mirror/internal/conf"
	"github.com/tencentyun/cos-go-sdk-v5"
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

func (o *ObjectStorage) HasModuleSnapshot(ctx context.Context, ref bufmoduleref.ModuleReference) (bool, error) {
	key := path.Join(ref.IdentityString(), ref.Reference())
	exist, err := o.cos.Object.IsExist(ctx, key)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return exist, nil
}

// func (o *ObjectStorage) CreateModuleSnapshot(ctx context.Context, ref bufmoduleref.ModuleReference) (bool, error) {
// 	key := path.Join(ref.IdentityString(), ref.Reference())
// 	exist, err := o.cos.Object.IsExist(ctx, key)
// 	if err != nil {
// 		return false, errors.WithStack(err)
// 	}
// 	return exist, nil
// }
