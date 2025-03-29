package cert_renew

const TlsSecretKind string = "Secret"
const TlsSecretApiVersion string = "v1"
const TlsSecretType string = "kubernetes.io/tls"

type TlsSecretData struct {
	TlsCrt string `json:"tls.crt" yaml:"tls.crt"`
	TlsKey string `json:"tls.key" yaml:"tls.key"`
}

type TlsSecretMetaData struct {
	Name string `json:"name" yaml:"name"`
}

type KubeSecretManifest struct {
	ApiVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Data       TlsSecretData     `json:"data" yaml:"data"`
	Kind       string            `json:"kind" yaml:"kind"`
	Metadata   TlsSecretMetaData `json:"metadata" yaml:"metadata"`
	Type       string            `json:"type" yaml:"type"`
}

func NewKubeTlsSecretManifest(tlsCrt string, tlsKey string, secretName string) *KubeSecretManifest {
	manifest := &KubeSecretManifest{
		ApiVersion: TlsSecretApiVersion,
		Data: TlsSecretData{
			TlsCrt: tlsCrt,
			TlsKey: tlsKey,
		},
		Kind:     TlsSecretKind,
		Metadata: TlsSecretMetaData{Name: secretName},
		Type:     TlsSecretType,
	}

	return manifest
}
