package infra_db

import (
	"database/sql"
	"log/slog"

	cloudflaredns "git.trahan.dev/go-infra/cloud_providers/cloudflare"
	_ "github.com/lib/pq"
)

// InsertDnsRecords inserts a list of DnsRecords into the PostgreSQL database.
func InsertDnsRecords(db *sql.DB, records []cloudflaredns.DnsRecordReq) error {
	query := `
		INSERT INTO public.dns_records( dns_record_id, zone_name, zone_id, name, content, proxied, type, comment, ttl)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (dns_record_id) DO NOTHING;
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
