package infra_db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"sync"

	cloudflaredns "github.com/babbage88/go-infra/cloud_providers/cloudflare"
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

type HostServer struct {
	HostName         string   `json:"hostname"`
	IpAddress        string   `json:"ip_address"`
	UserName         string   `json:"username"`
	PublicSshKeyname string   `json:"public_ssh_key"`
	HostedDomains    []string `json:"hosted_domains"`
	SslKeyPath       string   `json:"ssl_key_path"`
	IsContainerHost  bool     `json:"is_container_host"`
	IsVmHost         bool     `json:"is_vm_host"`
	IsVirtualMachine bool     `json:"is_virtual_machine"`
	IsDbHost         bool     `json:"is_db_host"`
}

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

		fmt.Println(psqlInfo)
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

func InsertOrUpdateHostServer(db *sql.DB, records []HostServer) error {
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
            id_db_host = EXCLUDED.id_db_host`

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

// InsertDnsRecords inserts a list of DnsRecords into the PostgreSQL database.
func InsertDnsRecords(db *sql.DB, records []cloudflaredns.DnsRecordReq) error {

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
			ttl = EXCLUDED.ttl;
	`

	for _, record := range records {
		slog.Info("Inserting record", slog.String("dns_record_id", record.DnsRecordId))
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
		SELECT dns_record_id, zone_name, zone_id, name, content, proxied, type, comment, ttl
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
