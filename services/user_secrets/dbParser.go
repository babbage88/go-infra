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
