// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package infra_db_pg

import (
	"net/netip"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AppPermission struct {
	ID                    uuid.UUID
	PermissionName        string
	PermissionDescription pgtype.Text
}

type AuthToken struct {
	ID           uuid.UUID
	UserID       uuid.UUID
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

type HealthCheck struct {
	ID           int32
	Status       pgtype.Text
	CheckType    pgtype.Text
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

type RolePermissionMapping struct {
	ID           uuid.UUID
	RoleID       uuid.UUID
	PermissionID uuid.UUID
	Enabled      bool
	CreatedAt    pgtype.Timestamptz
	LastModified pgtype.Timestamptz
}

type RolePermissionsView struct {
	RoleId       uuid.UUID
	Role         string
	PermissionId pgtype.UUID
	Permission   pgtype.Text
}

type TempAdminInfo struct {
	DevAdminroleid pgtype.UUID
	DevuserID      pgtype.UUID
}

type User struct {
	ID           uuid.UUID
	Username     pgtype.Text
	Password     pgtype.Text
	Email        pgtype.Text
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

type UserPermissionsView struct {
	UserId       pgtype.UUID
	Username     pgtype.Text
	PermissionId pgtype.UUID
	Permission   pgtype.Text
	Role         pgtype.Text
	LastModified pgtype.Timestamptz
}

type UserRole struct {
	ID              uuid.UUID
	RoleName        string
	RoleDescription pgtype.Text
	CreatedAt       pgtype.Timestamptz
	LastModified    pgtype.Timestamptz
	Enabled         bool
	IsDeleted       bool
}

type UserRoleMapping struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	RoleID       uuid.UUID
	Enabled      bool
	CreatedAt    pgtype.Timestamptz
	LastModified pgtype.Timestamptz
}

type UserRolesActive struct {
	RoleId          uuid.UUID
	RoleName        string
	RoleDescription pgtype.Text
	CreatedAt       pgtype.Timestamptz
	LastModified    pgtype.Timestamptz
	Enabled         bool
	IsDeleted       bool
}

type UsersAudit struct {
	AuditID   int32
	UserID    pgtype.UUID
	Username  pgtype.Text
	Email     pgtype.Text
	DeletedAt pgtype.Timestamptz
	DeletedBy pgtype.Text
}

type UsersWithRole struct {
	ID           uuid.UUID
	Username     pgtype.Text
	Password     pgtype.Text
	Email        pgtype.Text
	Roles        []string
	RoleIds      uuid.UUIDs
	CreatedAt    pgtype.Timestamptz
	LastModified pgtype.Timestamptz
	Enabled      bool
	IsDeleted    bool
}
