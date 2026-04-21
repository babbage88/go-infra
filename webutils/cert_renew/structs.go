package cert_renew

import corecert "github.com/babbage88/infra-core/cert_renew"

const (
	TlsSecretKind       = corecert.TlsSecretKind
	TlsSecretApiVersion = corecert.TlsSecretApiVersion
	TlsSecretType       = corecert.TlsSecretType
)

type TlsSecretMetaData = corecert.TlsSecretMetaData
type KubeSecretManifest = corecert.KubeSecretManifest

func NewKubeTlsSecretManifest(tlsCrt string, tlsKey string, secretName string) *KubeSecretManifest {
	return corecert.NewKubeTlsSecretManifest(tlsCrt, tlsKey, secretName)
}
