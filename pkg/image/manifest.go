/*
Copyright 2017 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// NOTE: This is manually replicated from: https://github.com/kubernetes/kubernetes/blob/master/test/utils/image/manifest.go

package image

import (
	"fmt"
	"io/ioutil"

	version "github.com/hashicorp/go-version"
	yaml "gopkg.in/yaml.v2"
)

const (
	// Agnhost image
	Agnhost = iota
	// AgnhostPrivate image
	AgnhostPrivate
	// APIServer image
	APIServer
	// AppArmorLoader image
	AppArmorLoader
	// AuthenticatedAlpine image
	AuthenticatedAlpine
	// AuthenticatedWindowsNanoServer image
	AuthenticatedWindowsNanoServer
	// BusyBox image
	BusyBox
	// CheckMetadataConcealment image
	CheckMetadataConcealment
	// CudaVectorAdd image
	CudaVectorAdd
	// CudaVectorAdd2 image
	CudaVectorAdd2
	// Dnsutils image
	Dnsutils
	// DebianIptables Image
	DebianIptables
	// EchoServer image
	EchoServer
	// Etcd image
	Etcd
	// GlusterDynamicProvisioner image
	GlusterDynamicProvisioner
	// Httpd image
	Httpd
	// HttpdNew image
	HttpdNew
	// InvalidRegistryImage image
	InvalidRegistryImage
	// IpcUtils image
	IpcUtils
	// JessieDnsutils image
	JessieDnsutils
	// Kitten image
	Kitten
	// Mounttest image
	Mounttest
	// MounttestUser image
	MounttestUser
	// Nautilus image
	Nautilus
	// NFSProvisioner image
	NFSProvisioner
	// Nginx image
	Nginx
	// NginxNew image
	NginxNew
	// Nonewprivs image
	Nonewprivs
	// NonRoot runs with a default user of 1234
	NonRoot
	// Pause - when these values are updated, also update cmd/kubelet/app/options/container_runtime.go
	// Pause image
	Pause
	// Perl image
	Perl
	// PrometheusDummyExporter image
	PrometheusDummyExporter
	// PrometheusToSd image
	PrometheusToSd
	// Redis image
	Redis
	// RegressionIssue74839 image
	RegressionIssue74839
	// ResourceConsumer image
	ResourceConsumer
	// ResourceController image
	ResourceController
	// SdDummyExporter image
	SdDummyExporter
	// StartupScript image
	StartupScript
	// TestWebserver image
	TestWebserver
	// VolumeNFSServer image
	VolumeNFSServer
	// VolumeISCSIServer image
	VolumeISCSIServer
	// VolumeGlusterServer image
	VolumeGlusterServer
	// VolumeRBDServer image
	VolumeRBDServer
)

const (
	buildImageRegistry      = "k8s.gcr.io/build-image"
	dockerGluster           = "docker.io/gluster"
	dockerLibraryRegistry   = "mirror.gcr.io/library"
	e2eRegistry             = "gcr.io/kubernetes-e2e-test-images"
	e2eVolumeRegistry       = "gcr.io/kubernetes-e2e-test-images/volume"
	etcdRegistry            = "quay.io/coreos"
	gcAuthenticatedRegistry = "gcr.io/authenticated-image-pulling"
	gcRegistry              = "k8s.gcr.io"
	gcrReleaseRegistry      = "gcr.io/gke-release"
	googleContainerRegistry = "gcr.io/google-containers"
	invalidRegistry         = "invalid.com/invalid"
	privateRegistry         = "gcr.io/k8s-authenticated-test"
	promoterE2eRegistry     = "k8s.gcr.io/e2e-test-images"
	quayIncubator           = "quay.io/kubernetes_incubator"
	quayK8sCSI              = "quay.io/k8scsi"
	sampleRegistry          = "gcr.io/google-samples"
	sigStorageRegistry      = "k8s.gcr.io/sig-storage"
)

// RegistryList holds public and private image registries
type RegistryList struct {
	BuildImageRegistry      string `yaml:"buildImageRegistry"`
	DockerGluster           string `yaml:"dockerGluster,omitempty"`
	DockerLibraryRegistry   string `yaml:"dockerLibraryRegistry,omitempty"`
	E2eRegistry             string `yaml:"e2eRegistry,omitempty"`
	E2eVolumeRegistry       string `yaml:"e2eVolumeRegistry"`
	EtcdRegistry            string `yaml:"etcdRegistry,omitempty"`
	GcAuthenticatedRegistry string `yaml:"gcAuthenticatedRegistry,omitempty"`
	GcRegistry              string `yaml:"gcRegistry,omitempty"`
	GcrReleaseRegistry      string `yaml:"gcrReleaseRegistry,omitempty"`
	GoogleContainerRegistry string `yaml:"googleContainerRegistry,omitempty"`
	InvalidRegistry         string `yaml:"invalidRegistry,omitempty"`
	PrivateRegistry         string `yaml:"privateRegistry,omitempty"`
	PromoterE2eRegistry     string `yaml:"promoterE2eRegistry"`
	QuayIncubator           string `yaml:"quayIncubator,omitempty"`
	QuayK8sCSI              string `yaml:"quayK8sCSI,omitempty"`
	SampleRegistry          string `yaml:"sampleRegistry,omitempty"`
	SigStorageRegistry      string `yaml:"sigStorageRegistry"`

	K8sVersion *version.Version `yaml:"-"`
	Images     map[int]Config   `yaml:"-"`
}

// Config holds an image's fully qualified name components registry, name, and tag
type Config struct {
	registry string
	name     string
	tag      string
}

// NewRegistryList returns a default registry or one that matches a config file passed
func NewRegistryList(repoConfig, k8sVersion string) (*RegistryList, error) {
	registry := &RegistryList{
		BuildImageRegistry:      buildImageRegistry,
		DockerGluster:           dockerGluster,
		DockerLibraryRegistry:   dockerLibraryRegistry,
		E2eRegistry:             e2eRegistry,
		E2eVolumeRegistry:       e2eVolumeRegistry,
		EtcdRegistry:            etcdRegistry,
		GcAuthenticatedRegistry: gcAuthenticatedRegistry,
		GcRegistry:              gcRegistry,
		GcrReleaseRegistry:      gcrReleaseRegistry,
		GoogleContainerRegistry: googleContainerRegistry,
		InvalidRegistry:         invalidRegistry,
		PrivateRegistry:         privateRegistry,
		QuayIncubator:           quayIncubator,
		QuayK8sCSI:              quayK8sCSI,
		PromoterE2eRegistry:     promoterE2eRegistry,
		SampleRegistry:          sampleRegistry,
		SigStorageRegistry:      sigStorageRegistry,
	}

	// Load in a config file
	if repoConfig != "" {

		fileContent, err := ioutil.ReadFile(repoConfig)
		if err != nil {
			return nil, fmt.Errorf("Error reading '%v' file contents: %v", repoConfig, err)
		}

		err = yaml.Unmarshal(fileContent, &registry)
		if err != nil {
			return nil, fmt.Errorf("Error unmarshalling '%v' YAML file: %v", repoConfig, err)
		}
	}

	// Init images for k8s version & repos configured
	version, err := validateVersion(k8sVersion)
	if err != nil {
		return nil, err
	}

	registry.K8sVersion = version

	return registry, nil
}

// GetDefaultImageRegistries returns the default default image registries used for
// a given version of the Kubernetes E2E tests
func GetDefaultImageRegistries(version string) (*RegistryList, error) {
	// Init images for k8s version & repos configured
	v, err := validateVersion(version)
	if err != nil {
		return nil, err
	}

	switch v.Segments()[0] {
	case 1:
		switch v.Segments()[1] {
		case 13, 14, 15, 16:
			return nil, fmt.Errorf("version not supported for this build: %s", version)
		case 17:
			return &RegistryList{
				E2eRegistry:             e2eRegistry,
				DockerLibraryRegistry:   dockerLibraryRegistry,
				GcRegistry:              gcRegistry,
				GoogleContainerRegistry: googleContainerRegistry,
				DockerGluster:           dockerGluster,
				QuayIncubator:           quayIncubator,

				// The following keys are used in the v1.17 registry list however their images
				// cannot be pulled as they are used as part of tests for checking image pull
				// behavior. They are omitted from the resulting config.
				// InvalidRegistry:         invalidRegistry,
				// GcAuthenticatedRegistry: gcAuthenticatedRegistry,
				// PrivateRegistry:         privateRegistry,
			}, nil
		case 18:
			return &RegistryList{
				E2eRegistry:             e2eRegistry,
				DockerLibraryRegistry:   dockerLibraryRegistry,
				GcRegistry:              gcRegistry,
				GoogleContainerRegistry: googleContainerRegistry,
				DockerGluster:           dockerGluster,
				QuayIncubator:           quayIncubator,
				PromoterE2eRegistry:     promoterE2eRegistry,

				// The following keys are used in the v1.18 registry list however their images
				// cannot be pulled as they are used as part of tests for checking image pull
				// behavior. They are omitted from the resulting config.
				// InvalidRegistry:         invalidRegistry,
				// GcAuthenticatedRegistry: gcAuthenticatedRegistry,
				// PrivateRegistry:         privateRegistry,
			}, nil
		case 19:
			return &RegistryList{
				BuildImageRegistry:    buildImageRegistry,
				E2eRegistry:           e2eRegistry,
				E2eVolumeRegistry:     e2eVolumeRegistry,
				DockerLibraryRegistry: dockerLibraryRegistry,
				GcRegistry:            gcRegistry,
				DockerGluster:         dockerGluster,
				PromoterE2eRegistry:   promoterE2eRegistry,
				SigStorageRegistry:    sigStorageRegistry,

				// The following keys are used in the v1.19 registry list however their images
				// cannot be pulled as they are used as part of tests for checking image pull
				// behavior. They are omitted from the resulting config.
				// InvalidRegistry:         invalidRegistry,
				// GcAuthenticatedRegistry: gcAuthenticatedRegistry,
				// PrivateRegistry:         privateRegistry,
			}, nil
		}
	}
	return nil, fmt.Errorf("No matching configuration for k8s version: %v", v)
}

// GetFullyQualifiedImageName returns the fully qualified URI to an image (including tag)
func (i Config) GetFullyQualifiedImageName() string {
	return fmt.Sprintf("%s/%s:%s", i.registry, i.name, i.tag)
}
