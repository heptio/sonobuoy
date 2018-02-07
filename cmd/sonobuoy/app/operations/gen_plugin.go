/*
Copyright 2018 Heptio Inc.

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

package operations

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"

	"github.com/pkg/errors"

	"github.com/heptio/sonobuoy/pkg/plugin"
	"github.com/heptio/sonobuoy/pkg/plugin/loader"
)

const (
	placeholderHostname      = "<hostname>"
	placeholderNamespace     = "sonobuoy"
	placeholderSonobuoyImage = "gcr.io/heptio-images/sonobuoy:master"
)

// GenPluginConfig are the input options for running
type GenPluginConfig struct {
	Paths      []string
	PluginName string
}

// GeneratePluginManifest partially initialises a plugin, then dumps out plugin's manifest
func GeneratePluginManifest(cfg GenPluginConfig) ([]byte, error) {
	plugins, err := loader.LoadAllPlugins(
		placeholderNamespace,
		placeholderSonobuoyImage,
		cfg.Paths,
		[]plugin.Selection{{Name: cfg.PluginName}},
	)
	if err != nil {
		return nil, err
	}

	if len(plugins) != 1 {
		return nil, fmt.Errorf("expected 1 plugin, got %v", len(plugins))
	}

	cert, err := genCert()
	if err != nil {
		return nil, err
	}

	return plugins[0].FillTemplate(placeholderHostname, cert)
}

func genCert() (*tls.Certificate, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't generate private key")
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(0),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create certificate")
	}

	return &tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privKey,
	}, nil
}
