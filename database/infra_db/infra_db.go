package infra_db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	cloudflaredns "github.com/babbage88/go-infra/cloud_providers/cloudflare"
	db_models "github.com/babbage88/go-infra/database/models"
	type_helper "github.com/babbage88/go-infra/utils/type_helper"
	"github.com/lib/pq"
)

type DatabaseConnection struct {
	DbHost     string `json:"dbHost"`
	DbPort     int32  `json:"dbPort"`
	DbUser     string `json:"dbUser"`
	DbPassword string `json:"dbPassword"`
	DbName     string `json:"database"`
}

type DatabaseConnectionOptions func(*DatabaseConnection)

// Global db instance
var (
	db     *sql.DB
	dbOnce sync.Once
	dbErr  error
)

func InitializeDbConnection(dbConn *DatabaseConnection) (*sql.DB, error) {
	dbOnce.Do(func() {
		// Connect to the PostgreSQL database
		slog.Info("Connecting to database: " + dbConn.DbHost)

		psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbConn.DbHost, type_helper.String(dbConn.DbPort), dbConn.DbUser, dbConn.DbPassword, dbConn.DbName)

		db, dbErr = sql.Open("postgres", psqlInfo)
		if dbErr != nil {
			slog.Error("Error connecting to the database", slog.String("Error", dbErr.Error()))
		}
	})
	return db, dbErr
}

func CloseDbConnection() error {
	if db != nil {
		slog.Info("Closing DB Connection")
		return db.Close()
	}
	return nil
}

func NewDatabaseConnection(opts ...DatabaseConnectionOptions) *DatabaseConnection {
	const (
		dbHost = "localhost"
		dbPort = 5432
		dbUser = "postgres"
		dbName = "go-infra"
	)
	dbPassword := os.Getenv("DB_PASSWORD")

	db := &DatabaseConnection{
		DbHost:     dbHost,
		DbPort:     dbPort,
		DbUser:     dbUser,
		DbPassword: dbPassword,
		DbName:     dbName,
	}

	for _, opt := range opts {
		opt(db)
	}

	return db
}

func WithDbHost(DbHostname string) DatabaseConnectionOptions {
	return func(c *DatabaseConnection) {
		c.DbHost = DbHostname
	}
}

func WithDbPort(dbPort int32) DatabaseConnectionOptions {
	return func(c *DatabaseConnection) {
		c.DbPort = dbPort
	}
}

func WithDbUser(dbUser string) DatabaseConnectionOptions {
	return func(c *DatabaseConnection) {
		c.DbUser = dbUser
	}
}

func WithDbPassword(dbPassword string) DatabaseConnectionOptions {
	return func(c *DatabaseConnection) {
		c.DbPassword = dbPassword
	}
}

func WithDbName(dbName string) DatabaseConnectionOptions {
	return func(c *DatabaseConnection) {
		c.DbName = dbName
	}
}

func InsertOrUpdateHostServer(db *sql.DB, records []db_models.HostServer) error {
	query := `
        INSERT INTO host_servers (
            hostname, ip_address, username, public_ssh_keyname, hosted_domains,
            ssl_key_path, is_container_host, is_vm_host, is_virtual_machine, id_db_host
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (hostname, ip_address)
        DO UPDATE SET
            username = EXCLUDED.username,
            public_ssh_keyname = EXCLUDED.public_ssh_keyname,
            hosted_domains = EXCLUDED.hosted_domains,
            ssl_key_path = EXCLUDED.ssl_key_path,
            is_container_host = EXCLUDED.is_container_host,
            is_vm_host = EXCLUDED.is_vm_host,
            is_virtual_machine = EXCLUDED.is_virtual_machine,
            id_db_host = EXCLUDED.id_db_host,
			last_modified = DEFAULT`

	for _, record := range records {
		slog.Info("Inserting Host Server record", slog.String("hostname", record.HostName), slog.String("IP", record.IpAddress))
		_, err := db.Exec(query, record.HostName, record.IpAddress, record.UserName, record.PublicSshKeyname, pq.Array(record.HostedDomains),
			record.SslKeyPath, record.IsContainerHost, record.IsVmHost, record.IsVirtualMachine, record.IsDbHost)
		if err != nil {
			slog.Error("Error inserting record", slog.String("Error", err.Error()))
			return err
		}
	}
	return nil
}

func ReadHostServer(db *sql.DB, id int64) (*db_models.HostServer, error) {
	query := `SELECT 
				id, 
				hostname, 
				ip_address, 
				username, 
				public_ssh_keyname, 
				hosted_domains, 
				ssl_key_path, 
				is_container_host, 
				is_vm_host, 
				is_virtual_machine, 
				is_db_host,
				created_at,
				last_modified
		FROM host_servers WHERE id = $1`

	var hostServer db_models.HostServer
	var hostedDomains pq.StringArray
	err := db.QueryRow(query, id).Scan(
		&hostServer.Id,
		&hostServer.HostName,
		&hostServer.IpAddress,
		&hostServer.UserName,
		&hostServer.PublicSshKeyname,
		&hostedDomains,
		&hostServer.SslKeyPath,
		&hostServer.IsContainerHost,
		&hostServer.IsVmHost,
		&hostServer.IsVirtualMachine,
		&hostServer.IsDbHost,
		&hostServer.CreatedAt,
		&hostServer.LastModified,
	)
	if err != nil {
		return nil, err
	}
	hostServer.HostedDomains = []string(hostedDomains)

	return &hostServer, nil
}

func DeleteHostServer(db *sql.DB, id int64) error {
	query := "DELETE FROM host_servers WHERE id = $1"
	_, err := db.Exec(query, id)
	return err
}

func InsertOrUpdateUser(db *sql.DB, user *db_models.User) error {
	query := `
		INSERT INTO users (username, password, email, role)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (username) DO UPDATE
		SET password = EXCLUDED.password, email = EXCLUDED.email
		RETURNING id`

	err := db.QueryRow(query,
		user.Username,
		user.Password,
		user.Email,
		user.Role,
	).Scan(&user.Id)

	slog.Info("Inserted or Upated User in DB.", slog.String("UserId", fmt.Sprint(user.Id)), slog.String("Username", user.Username))
	return err
}

func GetUserById(db *sql.DB, id int64) (*db_models.User, error) {
	query := `SELECT 
				id, username, password, email, role, created_at, last_modified
		FROM users WHERE id = $1`

	var user db_models.User
	slog.Info("Retrieving user from Database", slog.String("UserId", fmt.Sprint(id)))
	err := db.QueryRow(query, id).Scan(
		&user.Id,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.LastModified,
	)
	if err != nil {
		return nil, err
	}
	slog.Info("User found in Database.", slog.String("Username", user.Username))

	return &user, nil
}

func GetUserByUsername(db *sql.DB, username string) (*db_models.User, error) {
	query := `SELECT id, username, password, email, role, created_at, last_modified
		FROM users WHERE username = $1`

	var user db_models.User
	slog.Info("Retrieving user from Database", slog.String("UserName", username))
	err := db.QueryRow(query, username).Scan(
		&user.Id,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.LastModified,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("No user found with the given username", slog.String("UserName", username), slog.String("Error", err.Error()))
			return nil, fmt.Errorf("no user found with username: %s", username)
		}
		slog.Error("Error retrieving user from database", slog.String("Error", err.Error()))
		return nil, err
	}
	slog.Info("User found in Database.", slog.String("UserId", fmt.Sprint(user.Id)))

	return &user, nil
}

func DeleteUser(db *sql.DB, id int64) error {
	query := "DELETE FROM users WHERE id = $1"
	_, err := db.Exec(query, id)
	slog.Info("Deleting user from Database", slog.String("UserId", fmt.Sprint(id)))
	return err
}

func InsertAuthToken(db *sql.DB, authToken *db_models.AuthToken) error {
	query := `
		INSERT INTO auth_tokens (user_id, token, expiration)
		VALUES ($1, $2, $3)
		RETURNING id`

	err := db.QueryRow(query,
		authToken.UserId,
		authToken.Token,
		authToken.Expiration,
	).Scan(&authToken.Id)

	slog.Info("AuthToken added to Database for UserId", slog.String("UserId", fmt.Sprint(authToken.UserId)))
	return err
}

func GetAuthTokenFromDb(db *sql.DB, token string) (*db_models.AuthToken, error) {
	query := `SELECT 
				id, user_id, token, expiration, created_at, last_modified
			  FROM 
			  	auth_tokens WHERE id = $1`

	var authToken db_models.AuthToken
	err := db.QueryRow(query, token).Scan(
		&authToken.Id,
		&authToken.UserId,
		&authToken.Token,
		&authToken.Expiration,
	)
	if err != nil {
		return nil, err
	}
	slog.Info("Reading AuthToken form Database")

	return &authToken, nil
}

func DeleteAuthTokenById(db *sql.DB, id int64) error {
	query := "DELETE FROM auth_tokens WHERE id = $1"
	slog.Info("Deleting Auth toke with Id", slog.String("Id", fmt.Sprint(id)))
	_, err := db.Exec(query, id)
	return err
}

func deleteRecordsNotInList(db *sql.DB, czone cloudflaredns.CloudflareDnsZone) error {
	if len(czone.DnsRecords) == 0 {
		slog.Info("No Records to delete")
		return nil
	}

	// Declare a slice of empty interfaces to hold the IDs
	var ids []interface{}
	activerecords := make([]string, len(czone.DnsRecords))

	for i, record := range czone.DnsRecords {
		ids = append(ids, record.DnsRecordId)
		activerecords[i] = fmt.Sprintf("$%d", i+2) // Start from $2 since $1 is for zone_id
	}

	// Construct the query using the placeholders
	query := fmt.Sprintf(
		"DELETE FROM dns_records WHERE zone_id = $1 AND dns_record_id NOT IN (%s)",
		strings.Join(activerecords, ", "),
	)

	slog.Debug("Executing Delete query: ")
	ids = append([]interface{}{czone.ZoneId}, ids...)

	slog.Info("Running DELETE with WHERE Clause: ", slog.String("query string", query))
	_, err := db.Exec(query, ids...)
	return err
}

// InsertDnsRecords inserts a list of DnsRecords into the PostgreSQL database.
func InsertDnsRecords(db *sql.DB, czone cloudflaredns.CloudflareDnsZone) error {
	deleteRecordsNotInList(db, czone)

	query := `
		INSERT INTO public.dns_records( dns_record_id, zone_name, zone_id, name, content, proxied, type, comment, ttl)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (dns_record_id)
		DO UPDATE SET
			zone_name = EXCLUDED.zone_name,
			zone_id = EXCLUDED.zone_id,
			name = EXCLUDED.name,
			content = EXCLUDED.content,
			proxied = EXCLUDED.proxied,
			type = EXCLUDED.type,
			comment = EXCLUDED.comment,
			ttl = EXCLUDED.ttl,
			last_modified = DEFAULT;
	`

	for _, record := range czone.DnsRecords {
		fmt.Println("testing")
		fmt.Println(record.Name)
		slog.Info("Inserting record", slog.String("record_content", record.Content), slog.String("dns_record_id", record.DnsRecordId))
		_, err := db.Exec(query, record.DnsRecordId, record.ZoneName, record.ZoneId, record.Name, record.Content, record.Proxied, record.Type, record.Comment, record.Ttl)
		if err != nil {
			slog.Error("Error inserting record", slog.String("Error", err.Error()))
			return err
		}
	}
	return nil
}

// GetDnsRecordByName searches for a DNS record by name and returns it as a DnsRecordReq struct.
func GetDnsRecordByName(db *sql.DB, name string, rtype string) (*cloudflaredns.DnsRecordReq, error) {
	// slog.Info("Querying DNS record", slog.String("name", name), slog.String("type", rtype))

	query := `
		SELECT dns_record_id, zone_name, zone_id, name, content, proxied, type, comment, ttl, last_modified
		FROM public.dns_records
		WHERE name = $1 AND type = $2;
	`

	var record cloudflaredns.DnsRecordReq
	err := db.QueryRow(query, name, rtype).Scan(
		&record.DnsRecordId,
		&record.ZoneName,
		&record.ZoneId,
		&record.Name,
		&record.Content,
		&record.Proxied,
		&record.Type,
		&record.Comment,
		&record.Ttl,
		&record.LastModified,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Info("No record found with the specified name", slog.String("name", name))
			return nil, nil
		}
		slog.Error("Error querying record by name", slog.String("Error", err.Error()))
		return nil, err
	}

	return &record, nil
}

func GetAllDnsRecords(db *sql.DB) ([]cloudflaredns.DnsRecordReq, error) {
	query := `SELECT dns_record_id, zone_name, zone_id, name, content, proxied, type, comment, ttl, last_modified
			  FROM	public.dns_records;`

	rows, err := db.Query(query)
	if err != nil {
		slog.Error("Error running query", slog.String("Error", err.Error()))
	}
	var records []cloudflaredns.DnsRecordReq
	for rows.Next() {
		var record cloudflaredns.DnsRecordReq

		if err := rows.Scan(&record.DnsRecordId,
			&record.ZoneName,
			&record.ZoneId,
			&record.Name,
			&record.Content,
			&record.Proxied,
			&record.Type,
			&record.Comment,
			&record.Ttl,
			&record.LastModified); err != nil {
			slog.Error("Error parsing DB response", slog.String("Error", err.Error()))
		}
		slog.Info("Appending DNS Record", slog.String("dns_record_id", record.DnsRecordId), slog.String("name:", record.Name))
		records = append(records, record)
	}

	return records, nil
}

func GetDbDnsRecordByZoneId(db *sql.DB, czone *cloudflaredns.CloudflareDnsZone) ([]cloudflaredns.DnsRecordReq, error) {
	query := `SELECT dns_record_id, zone_name, zone_id, name, content, proxied, type, comment, ttl, last_modified
			  FROM	public.dns_records
			  WHERE	zone_id = $1;`

	rows, err := db.Query(query, czone.ZoneId)
	if err != nil {
		slog.Error("Error running query", slog.String("Error", err.Error()))
	}
	var records []cloudflaredns.DnsRecordReq
	for rows.Next() {
		var record cloudflaredns.DnsRecordReq

		if err := rows.Scan(&record.DnsRecordId,
			&record.ZoneName,
			&record.ZoneId,
			&record.Name,
			&record.Content,
			&record.Proxied,
			&record.Type,
			&record.Comment,
			&record.Ttl,
			&record.LastModified); err != nil {
			slog.Error("Error parsing DB response", slog.String("Error", err.Error()))
		}
		slog.Info("Appending DNS Record", slog.String("dns_record_id", record.DnsRecordId), slog.String("name:", record.Name))
		records = append(records, record)
	}

	return records, nil
}

func DeleteDnsRecord(db *sql.DB, record cloudflaredns.DnsRecordReq) error {
	query := `DELETE FROM public.dns_records
			  WHERE zone_id = $1 AND dns_record_id = $2`

	slog.Info("Deleting DNS record", slog.String("dns_record_id", record.DnsRecordId), slog.String("name:", record.Name))
	db.Exec(query, record.ZoneId, record.DnsRecordId)

	return nil
}

func DeleteDnsRecords(db *sql.DB, records []cloudflaredns.DnsRecordReq) error {

	for _, record := range records {
		DeleteDnsRecord(db, record)
	}

	return nil
}
