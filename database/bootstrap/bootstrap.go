package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/babbage88/go-infra/internal/pretty"
	"github.com/jackc/pgx/v5"
)

func configureDbConn() (*pgx.Conn, func(), error) {
	dBUrl := os.Getenv("DATABASE_BOOTSTRAP_URL")

	conn, err := pgx.Connect(context.Background(), dBUrl)
	if err != nil {
		slog.Error("Unable to connect to the database: ", slog.String("Error", err.Error()))
		return conn, nil, err
	}
	closeFunc := func() {
		conn.Close(context.Background())
	}
	return conn, closeFunc, err
}

func CreateDbIfNotExist() error {
	newDbName := os.Getenv("NEW_INFRA_DB")
	if newDbName == "" {
		return fmt.Errorf("NEW_INFRA_DB environment variable is not set")
	}

	dbExists, err := CheckDbExists(newDbName)
	if err != nil {
		pretty.PrintErrorf("%s", err.Error())
		return err
	}

	if dbExists {
		pretty.Printf("The Databse %s already exists.", newDbName)
		return err
	} else {
		err = CreateNewDb(newDbName)
	}

	return err
}

func CreateInfradbUser(name string) error {
	dBUrl := os.Getenv("DATABASE_URL")
	fmt.Println(dBUrl)
	conn, err := pgx.Connect(context.Background(), dBUrl)
	if err != nil {
		pretty.PrintErrorf("Unable to connect to the database: %s", err.Error())
		return err
	}
	defer conn.Close(context.Background())
	roleExists, err := CheckRoleExists(name)
	if err != nil {
		pretty.PrintErrorf("Error checking if role exists %s", err.Error())
	}
	if roleExists {
		pretty.Printf("Role already esxits %s", name)
	} else {
		pretty.Print("Attempting Create Role Query")
		query := fmt.Sprintf(`CREATE ROLE %s;`, pgx.Identifier{name}.Sanitize())
		_, err := conn.Exec(context.Background(), query)
		if err != nil {
			pretty.PrintErrorf("Error exec create user query %s", err.Error())
		}
	}

	permQry := fmt.Sprintf(`GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %s;`, pgx.Identifier{name}.Sanitize())

	pretty.Print("Executing grant privs query")
	_, err = conn.Exec(context.Background(), permQry)
	if err != nil {
		pretty.PrintErrorf("Error executing grant privs qry %s", err.Error())
		return err
	}
	pretty.Print("Success configuring user %s", name)
	return nil
}

func CheckDbExists(name string) (bool, error) {
	conn, closeConn, err := configureDbConn()
	defer closeConn()
	if err != nil {
		pretty.PrintErrorf("%s", err.Error())
		return false, err
	}
	var retVal bool

	qry := `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`

	row := conn.QueryRow(context.Background(), qry, name)
	err = row.Scan(&retVal)
	if err != nil {
		pretty.PrintErrorf("Error scanning row %s", err.Error())
		slog.Error("Error scanning row ", slog.String("Error", err.Error()))
		return false, err

	}

	return retVal, err
}

func CheckRoleExists(name string) (bool, error) {
	dBUrl := os.Getenv("DATABASE_URL")
	fmt.Println(dBUrl)
	conn, err := pgx.Connect(context.Background(), dBUrl)
	if err != nil {
		pretty.PrintErrorf("Unable to connect to the database: %s", err.Error())
		return false, err
	}
	defer conn.Close(context.Background())

	var retVal bool
	qry := `SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)`

	row := conn.QueryRow(context.Background(), qry, name)
	err = row.Scan(&retVal)
	if err != nil {
		pretty.PrintErrorf("Error scanning row %s", err.Error())
		slog.Error("Error scanning row ", slog.String("Error", err.Error()))
		return false, err

	}

	return retVal, err

}

func CreateNewDb(name string) error {
	conn, closeConn, err := configureDbConn()
	defer closeConn()
	if err != nil {
		pretty.PrintErrorf("Error creating database connection: %s", err.Error())
		return err
	}

	// Safely escape the database name
	qry := fmt.Sprintf("CREATE DATABASE %s;", pgx.Identifier{name}.Sanitize())

	pretty.Print("Attempting to execute Db Create query.")
	_, err = conn.Exec(context.Background(), qry)
	if err != nil {
		pretty.PrintErrorf("Error executing query to create new database: %s", err.Error())
		return err
	}

	return nil
}

func NewDb() {
	err := CreateDbIfNotExist()
	if err != nil {
		pretty.PrintErrorf("%s", err.Error())
		slog.Error("Error creating Database", slog.String("error", err.Error()))
	}
}
