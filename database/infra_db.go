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
