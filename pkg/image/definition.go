package image

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	TypeISO = "iso"
	TypeRAW = "raw"

	KubernetesTypeRKE2 = "rke2"
	KubernetesTypeK3s  = "k3s"
)

type Definition struct {
	APIVersion      string          `yaml:"apiVersion"`
	Image           Image           `yaml:"image"`
	OperatingSystem OperatingSystem `yaml:"operatingSystem"`
	Kubernetes      Kubernetes      `yaml:"kubernetes"`
}

type Image struct {
	ImageType       string `yaml:"imageType"`
	BaseImage       string `yaml:"baseImage"`
	OutputImageName string `yaml:"outputImageName"`
}

type OperatingSystem struct {
	KernelArgs []string              `yaml:"kernelArgs"`
	Users      []OperatingSystemUser `yaml:"users"`
	Systemd    Systemd               `yaml:"systemd"`
	Suma       Suma                  `yaml:"suma"`
}

type OperatingSystemUser struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSHKey   string `yaml:"sshKey"`
}

type Systemd struct {
	Enable  []string `yaml:"enable"`
	Disable []string `yaml:"disable"`
}

type Suma struct {
	Host          string `yaml:"host"`
	ActivationKey string `yaml:"activationKey"`
	GetSSL        bool   `yaml:"getSSL"`
}

type Kubernetes struct {
	Distribution string `yaml:"distribution"`
	Type         string `yaml:"type"`
	Channel      string `yaml:"channel"`
	Version      string `yaml:"version"`
	SELinux      bool   `yaml:"selinux"`
}

func ParseDefinition(data []byte) (*Definition, error) {
	var definition Definition

	if err := yaml.Unmarshal(data, &definition); err != nil {
		return nil, fmt.Errorf("could not parse the image definition: %w", err)
	}

	return &definition, nil
}
