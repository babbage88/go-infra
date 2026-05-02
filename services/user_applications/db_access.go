package user_applications

import (
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/google/uuid"
)

func parseUserApplication(dbApp infra_db_pg.UserApplication, deps []InfraDependencyDao) UserApplicationDao {
	return UserApplicationDao{
		Id:                dbApp.ID,
		Name:              dbApp.Name,
		Description:       dbApp.Description.String,
		RepositoryUrl:     dbApp.RepositoryUrl,
		ManifestPath:      dbApp.ManifestPath.String,
		SourceKind:        dbApp.SourceKind,
		ModuleName:        dbApp.ModuleName.String,
		PackageName:       dbApp.PackageName.String,
		PackageManager:    dbApp.PackageManager.String,
		DeployKind:        dbApp.DeployKind,
		Registerable:      dbApp.Registerable,
		DeployConfig:      unmarshalOptionalJSON(dbApp.DeployConfig),
		BuildConfig:       unmarshalOptionalJSON(dbApp.BuildConfig),
		InfraDependencies: deps,
		CreatedAt:         dbApp.CreatedAt.Time,
		LastModified:      dbApp.LastModified.Time,
	}
}

func parseInfraDependencies(rows []infra_db_pg.GetUserApplicationInfraDependenciesByAppIdRow) []InfraDependencyDao {
	deps := make([]InfraDependencyDao, 0, len(rows))
	for _, row := range rows {
		dep := InfraDependencyDao{
			Id:                 row.ID,
			DependencyType:     row.DependencyType,
			DependencyName:     row.DependencyName,
			HostServerTypeName: row.HostServerTypeName.String,
			PlatformTypeName:   row.PlatformTypeName.String,
			Config:             unmarshalOptionalJSON(row.DependencyConfig),
			CreatedAt:          row.CreatedAt.Time,
			LastModified:       row.LastModified.Time,
		}
		if row.HostServerTypeID.Valid {
			id := uuid.UUID(row.HostServerTypeID.Bytes)
			dep.HostServerTypeId = &id
		}
		if row.PlatformTypeID.Valid {
			id := uuid.UUID(row.PlatformTypeID.Bytes)
			dep.PlatformTypeId = &id
		}
		deps = append(deps, dep)
	}
	return deps
}
