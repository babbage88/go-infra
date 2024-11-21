package db_access

import (
	"github.com/babbage88/go-infra/database/infra_db_pg"
)

type DbParser interface {
	ParseUserFromDb(dbuser infra_db_pg.User)
	ParseUserRowFromDb(dbRow infra_db_pg.GetUserLoginRow)
	ParseAuthTokenFromDb(token infra_db_pg.AuthToken)
}

func (u *UserDao) ParseUserRowFromDb(dbRow infra_db_pg.GetUserLoginRow) {
	u.Id = dbRow.ID
	u.UserName = dbRow.Username.String
	u.Email = dbRow.Email.String
	u.Enabled = dbRow.Enabled
	u.Role = dbRow.Role.String

}

func (u *UserDao) ParseUserFromDb(dbuser infra_db_pg.User) {
	u.Id = dbuser.ID
	u.UserName = dbuser.Username.String
	u.Email = dbuser.Email.String
	u.Role = dbuser.Role.String
	u.CreatedAt = dbuser.CreatedAt.Time
	u.LastModified = dbuser.LastModified.Time
	u.Enabled = dbuser.Enabled
	u.IsDeleted = dbuser.IsDeleted
}

func (t *AuthTokenDao) ParseAuthTokenFromDb(token infra_db_pg.AuthToken) {
	t.Id = token.ID
	t.Token = token.Token.String
	t.UserID = token.UserID.Int32
	t.CreatedAt = token.CreatedAt.Time
	t.Expiration = token.Expiration.Time
	t.LastModified = token.LastModified.Time
}
