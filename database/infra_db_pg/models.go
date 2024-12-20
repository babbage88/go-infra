// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package infra_db_pg

import (
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"
)

type AuthToken struct {
	ID           int32
	UserID       pgtype.Int4
	Token        pgtype.Text
	Expiration   pgtype.Timestamp
	CreatedAt    pgtype.Timestamptz
	LastModified pgtype.Timestamptz
}

type DnsRecord struct {
	ID           int32
	DnsRecordID  string
	ZoneName     pgtype.Text
	ZoneID       pgtype.Text
	Name         pgtype.Text
	Content      pgtype.Text
	Proxied      pgtype.Bool
	Type         pgtype.Text
	Comment      pgtype.Text
	Ttl          pgtype.Int4
	CreatedAt    pgtype.Timestamptz
	LastModified pgtype.Timestamptz
}

type HostServer struct {
	ID               int32
	Hostname         string
	IpAddress        netip.Addr
	Username         pgtype.Text
	PublicSshKeyname pgtype.Text
	HostedDomains    []string
	SslKeyPath       pgtype.Text
	IsContainerHost  pgtype.Bool
	IsVmHost         pgtype.Bool
	IsVirtualMachine pgtype.Bool
	IDDbHost         pgtype.Bool
	CreatedAt        pgtype.Timestamptz
	LastModified     pgtype.Timestamptz
}

type HostedDbPlatform struct {
	ID                int32
	PlatformName      string
	DefaultListenPort pgtype.Int4
}

type User struct {
	ID           int32
	Username     pgtype.Text
	Password     pgtype.Text
	Email        pgtype.Text
	Role         pgtype.Text
	CreatedAt    pgtype.Timestamptz
	LastModified pgtype.Timestamptz
	Enabled      bool
	IsDeleted    bool
}

type UserHostedDb struct {
	ID                   int32
	PriceTierCodeID      int32
	UserID               int32
	CurrentHostServerID  int32
	CurrentKubeClusterID pgtype.Int4
	UserApplicationIds   []int32
	DbPlatformID         int32
	Fqdn                 string
	PubIpAddress         netip.Addr
	ListenPort           int32
	PrivateIpAddress     *netip.Addr
	CreatedAt            pgtype.Timestamptz
	LastModified         pgtype.Timestamptz
}

type UserHostedK8 struct {
	ID                   int32
	PriceTierCodeID      int32
	UserID               int32
	OrganizationID       pgtype.Int4
	CurrentHostServerIds []int32
	UserApplicationIds   []int32
	UserCertificateIds   []int32
	K8Type               string
	ApiEndpointFqdn      string
	ClusterName          string
	PubIpAddress         netip.Addr
	ListenPort           int32
	PrivateIpAddress     *netip.Addr
	CreatedAt            pgtype.Timestamptz
	LastModified         pgtype.Timestamptz
}

type UsersAudit struct {
	AuditID   int32
	UserID    pgtype.Int4
	Username  pgtype.Text
	Email     pgtype.Text
	DeletedAt pgtype.Timestamptz
	DeletedBy pgtype.Text
}
