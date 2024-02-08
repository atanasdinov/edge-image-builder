package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/suse-edge/edge-image-builder/pkg/http"
	"github.com/suse-edge/edge-image-builder/pkg/image"
	"go.uber.org/zap"
)

const (
	rke2SELinuxPolicyReleaseURL = "https://github.com/rancher/rke2-selinux/releases/download/%s/%s"
	k3sSELinuxPolicyReleaseURL  = "https://github.com/k3s-io/k3s-selinux/releases/download/%s/%s"

	rancherSigningKeyURL = "https://rpm.rancher.io/public.key"
)

type selinuxPolicy struct {
	downloadURL string
	rpmName     string
}

func DownloadSELinuxRPMs(kubernetes *image.Kubernetes, rpmDir, gpgKeysDir string) error {
	policy := newSELinuxPolicy(kubernetes.Version)
	policyPath := filepath.Join(rpmDir, policy.rpmName)

	if _, err := os.Stat(policyPath); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			zap.S().Errorf("Looking for SELinux RPM '%s' failed unexpectedly: %s", policyPath, err)
		}

		if err = http.DownloadFile(context.Background(), policy.downloadURL, policyPath, nil); err != nil {
			return fmt.Errorf("downloading selinux policy: %w", err)
		}
	}

	signingKeyPath := filepath.Join(gpgKeysDir, "rancher-public.key")
	if err := http.DownloadFile(context.Background(), rancherSigningKeyURL, signingKeyPath, nil); err != nil {
		return fmt.Errorf("downloading rancher signing key: %w", err)
	}

	return nil
}

func newSELinuxPolicy(kubernetesVersion string) selinuxPolicy {
	const (
		rke2PolicyVersion       = "0.17"
		rke2PolicyChannel       = "stable"
		rke2PolicyReleaseNumber = "1"

		k3sPolicyVersion       = "1.4" // TODO: Bump to 1.5 once the release assets are fixed
		k3sPolicyChannel       = "stable"
		k3sPolicyReleaseNumber = "1"

		rke2PolicyRPM = "rke2-selinux-%s-%s.slemicro.noarch.rpm"
		k3sPolicyRPM  = "k3s-selinux-%s-%s.slemicro.noarch.rpm"
	)

	if strings.Contains(kubernetesVersion, image.KubernetesDistroRKE2) {
		rpm := fmt.Sprintf(rke2PolicyRPM, rke2PolicyVersion, rke2PolicyReleaseNumber)
		version := fmt.Sprintf("v%s.%s.%s", rke2PolicyVersion, rke2PolicyChannel, rke2PolicyReleaseNumber)

		return selinuxPolicy{
			downloadURL: fmt.Sprintf(rke2SELinuxPolicyReleaseURL, version, rpm),
			rpmName:     rpm,
		}
	}

	rpm := fmt.Sprintf(k3sPolicyRPM, k3sPolicyVersion, k3sPolicyReleaseNumber)
	version := fmt.Sprintf("v%s.%s.%s", k3sPolicyVersion, k3sPolicyChannel, k3sPolicyReleaseNumber)

	return selinuxPolicy{
		downloadURL: fmt.Sprintf(k3sSELinuxPolicyReleaseURL, version, rpm),
		rpmName:     rpm,
	}
}
