package external_applications

import (
	"github.com/babbage88/go-infra/database/infra_db_pg"
)

type DbParser interface {
	ParseExternalApplicationFromDb(dbApp infra_db_pg.ExternalIntegrationApp)
	ParseExternalApplicationFromGetAllRow(dbRow infra_db_pg.GetAllExternalAppsRow)
}

func (ea *ExternalApplicationDao) ParseExternalApplicationFromDb(dbApp infra_db_pg.ExternalIntegrationApp) {
	ea.Id = dbApp.ID
	ea.Name = dbApp.Name
	ea.CreatedAt = dbApp.CreatedAt.Time
	ea.LastModified = dbApp.LastModified.Time

	if dbApp.EndpointUrl.Valid {
		ea.EndpointUrl = dbApp.EndpointUrl.String
	}

	if dbApp.AppDescription.Valid {
		ea.AppDescription = dbApp.AppDescription.String
	}
}

func (ea *ExternalApplicationDao) ParseExternalApplicationFromGetAllRow(dbRow infra_db_pg.GetAllExternalAppsRow) {
	ea.Id = dbRow.ID
	ea.Name = dbRow.Name
	// Note: GetAllExternalApps only returns id and name, so other fields remain zero values
}
