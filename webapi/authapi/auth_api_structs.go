package authapi

type ParsedCertbotOutput struct {
	CertificateInfo string `json:"certificateInfo"`
	Warnings        string `json:"warnings"`
	DebugLog        string `json:"debugLog"`
}
