package user_secrets

import "github.com/babbage88/go-infra/database/infra_db_pg"

type DbStructParser interface {
	ParseExternalApplicationFromDb(extApp infra_db_pg.ExternalIntegrationApp)
	ParseExternalAuthTokenFromDb(token infra_db_pg.ExternalAuthToken)
}

func (t *ExternalApplicationAuthToken) ParseAuthTokenFromDb(token infra_db_pg.ExternalAuthToken) {
	t.Id = token.ID
	t.Token = token.Token
	t.ExternalApplicationId = token.ExternalAppID
	t.UserID = token.UserID
	t.CreatedAt = token.CreatedAt.Time
	t.Expiration = token.Expiration.Time
	t.LastModified = token.LastModified.Time
}

func (t *ExternalApplication) ParseExternalApplicationFromDb(extApp infra_db_pg.ExternalIntegrationApp) {
	t.Id = extApp.ID
	t.Name = extApp.Name
}

func (t *ExternalAppSecretMetadata) ParseExternalAppSecretMetadataFromDb(extAppToken infra_db_pg.ExternalAuthToken) {
	t.Id = extAppToken.ID
	t.UserId = extAppToken.UserID
	t.Expiration = extAppToken.Expiration.Time
}

func (t *UserSecretEntry) ParseExternalAppSecretMetadataFromDb(userSecretInfo infra_db_pg.GetUserSecretsByUserIdRow) {
	t.SecretMetadata.Id = userSecretInfo.AuthTokenID
	t.SecretMetadata.UserId = userSecretInfo.UserID
	t.SecretMetadata.Expiration = userSecretInfo.Expiration.Time
	t.SecretMetadata.CreatedAt = userSecretInfo.TokenCreatedAt.Time
	t.AppInfo.Id = userSecretInfo.ApplicationID
	t.AppInfo.Name = userSecretInfo.ApplicationName
	t.AppInfo.UrlEndpoint = userSecretInfo.EndpointUrl.String
}

func (t *UserSecretEntry) ParseExternalAppSecretMetadataFromAppId(userSecretInfo infra_db_pg.GetUserSecretsByAppIdRow) {
	t.SecretMetadata.Id = userSecretInfo.AuthTokenID
	t.SecretMetadata.UserId = userSecretInfo.UserID
	t.SecretMetadata.Expiration = userSecretInfo.Expiration.Time
	t.SecretMetadata.CreatedAt = userSecretInfo.TokenCreatedAt.Time
	t.AppInfo.Id = userSecretInfo.ApplicationID
	t.AppInfo.Name = userSecretInfo.ApplicationName
	t.AppInfo.UrlEndpoint = userSecretInfo.EndpointUrl.String
}
