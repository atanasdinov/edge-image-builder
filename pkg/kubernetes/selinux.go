package kubernetes

import (
	"context"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/http"
	"github.com/suse-edge/edge-image-builder/pkg/image"
)

var SELinuxRepository = image.AddRepo{
	URL:      "https://rpm.rancher.io/rke2/stable/common/slemicro/noarch",
	Unsigned: true,
}

var SELinuxPackage = "rke2-selinux"

func DownloadSELinuxRPMsSigningKey(gpgKeysDir string) error {
	const rancherSigningKeyURL = "https://rpm.rancher.io/public.key"
	var signingKeyPath = filepath.Join(gpgKeysDir, "rancher-public.key")

	return http.DownloadFile(context.Background(), rancherSigningKeyURL, signingKeyPath, nil)
}
