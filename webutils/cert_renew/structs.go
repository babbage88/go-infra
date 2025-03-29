package cert_renew

import "encoding/base64"

const TlsSecretKind string = "Secret"
const TlsSecretApiVersion string = "v1"
const TlsSecretType string = "kubernetes.io/tls"

type TlsSecretMetaData struct {
	Name string `json:"name" yaml:"name"`
}

type KubeSecretManifest struct {
	ApiVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Data       map[string]string `json:"data" yaml:"data"`
	Kind       string            `json:"kind" yaml:"kind"`
	Metadata   TlsSecretMetaData `json:"metadata" yaml:"metadata"`
	Type       string            `json:"type" yaml:"type"`
}

func NewKubeTlsSecretManifest(tlsCrt string, tlsKey string, secretName string) *KubeSecretManifest {
	base64EncodedTlsCrt := base64.StdEncoding.EncodeToString([]byte(tlsCrt))
	base64EncodedTlsKey := base64.StdEncoding.EncodeToString([]byte(tlsKey))
	var data map[string]string = make(map[string]string)
	data["tls.crt"] = base64EncodedTlsCrt
	data["tls.key"] = base64EncodedTlsKey

	manifest := &KubeSecretManifest{
		ApiVersion: TlsSecretApiVersion,
		Data:       data,
		Kind:       TlsSecretKind,
		Metadata:   TlsSecretMetaData{Name: secretName},
		Type:       TlsSecretType,
	}

	return manifest
}
