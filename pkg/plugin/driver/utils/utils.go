/*
Copyright 2017 Heptio Inc.

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

package utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/heptio/sonobuoy/pkg/plugin"
	"github.com/pkg/errors"
	gouuid "github.com/satori/go.uuid"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetSessionID generates a new session id.
// This is essentially an instance of a running plugin.
func GetSessionID() string {
	uuid := gouuid.NewV4()
	ret := make([]byte, hex.EncodedLen(8))
	hex.Encode(ret, uuid.Bytes()[0:8])
	return string(ret)
}

// IsPodFailing returns whether a plugin's pod is failing and isn't likely to
// succeed.
// TODO: this may require more revisions as we get more experience with
// various types of failures that can occur.
func IsPodFailing(pod *v1.Pod) (bool, string) {
	// Check if the pod is unschedulable
	for _, cond := range pod.Status.Conditions {
		if cond.Reason == "Unschedulable" {
			return true, fmt.Sprintf("Can't schedule pod: %v", cond.Message)
		}
	}

	for _, cstatus := range pod.Status.ContainerStatuses {
		// Check if a container in the pod is restarting multiple times
		if cstatus.RestartCount > 2 {
			errstr := fmt.Sprintf("Container %v has restarted unsuccessfully %v times", cstatus.Name, cstatus.RestartCount)
			return true, errstr
		}

		// Check if it can't fetch its image
		if waiting := cstatus.State.Waiting; waiting != nil {
			if waiting.Reason == "ImagePullBackOff" || waiting.Reason == "ErrImagePull" {
				errstr := fmt.Sprintf("Container %v is in state %v", cstatus.Name, waiting.Reason)
				return true, errstr
			}
		}
	}

	return false, ""
}

// MakeErrorResult constructs a plugin.Result given an error message and error
// data.  errdata is a map that will be placed in the sonobuoy results tarball
// for this plugin as a JSON file, so it's what users will see for why the
// plugin failed.  If errdata["error"] is not set, it will be filled in with an
// "Unknown error" string.
func MakeErrorResult(resultType string, errdata map[string]interface{}, nodeName string) *plugin.Result {
	errJSON, _ := json.Marshal(errdata)

	errstr := "Unknown error"
	if e, ok := errdata["error"]; ok {
		errstr = e.(string)
	}

	return &plugin.Result{
		Body:       bytes.NewReader(errJSON),
		Error:      errstr,
		ResultType: resultType,
		NodeName:   nodeName,
		MimeType:   "application/json",
	}
}

// GetCACertPEM extracts the CA cert from a tls.Certificate.
// If the provided Certificate has only one certificate in the chain, the CA
// will be the leaf cert.
func GetCACertPEM(cert *tls.Certificate) string {
	cacert := ""
	if len(cert.Certificate) > 0 {
		caCertDER := cert.Certificate[len(cert.Certificate)-1]
		cacert = string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caCertDER,
		}))
	}
	return cacert
}

// GetKeyPEM turns an RSA Private Key into a PEM-encoded string
func GetKeyPEM(key *ecdsa.PrivateKey) ([]byte, error) {
	derKEY, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: derKEY,
	}), nil
}

// MakeTLSSecret makes a Kubernetes secret object for the given TLS certificate.
func MakeTLSSecret(cert *tls.Certificate, namespace, secretName string) (*v1.Secret, error) {
	rsaKey, ok := cert.PrivateKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key not ECDSA")
	}

	if len(cert.Certificate) <= 0 {
		return nil, errors.New("no certs in tls.certificate")
	}

	certDER := cert.Certificate[0]
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	keyPEM, err := GetKeyPEM(rsaKey)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't PEM encode TLS key")
	}

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			v1.TLSPrivateKeyKey: keyPEM,
			v1.TLSCertKey:       certPEM,
		},
		Type: v1.SecretTypeTLS,
	}, nil

}
