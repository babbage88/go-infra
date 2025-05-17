package services

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// Respose will return login result and the user info.
// swagger:response UserDao
// This text will appear as description of your response body.
type UserDao struct {
	// in:body
	Id           uuid.UUID  `json:"id"`
	UserName     string     `json:"username"`
	Email        string     `json:"email"`
	Roles        []string   `json:"roles"`
	RoleIds      uuid.UUIDs `json:"role_ids"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastModified time.Time  `json:"lastModified"`
	Enabled      bool       `json:"enabled"`
	IsDeleted    bool       `json:"isDeleted"`
}

type AuthTokenDao struct {
	Id           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Token        string    `json:"token"`
	Expiration   time.Time `json:"expiration"`
	CreatedAt    time.Time `json:"created_at"`
	LastModified time.Time `json:"last_modified"`
}

type HostServer struct {
	ID               int32
	Hostname         string
	IpAddress        netip.Addr
	Username         string
	PublicSshKeyname string
	HostedDomains    []string
	SslKeyPath       string
	IsContainerHost  bool
	IsVmHost         bool
	IsVirtualMachine bool
	IDDbHost         bool
	CreatedAt        time.Time
	LastModified     time.Time
}

type HostedDbPlatform struct {
	ID                int32
	PlatformName      string
	DefaultListenPort int32
}

type UserHostedDb struct {
	ID                   int32
	PriceTierCodeID      int32
	UserID               int32
	CurrentHostServerID  int32
	CurrentKubeClusterID int32
	UserApplicationIds   []int32
	DbPlatformID         int32
	Fqdn                 string
	PubIpAddress         netip.Addr
	ListenPort           int32
	PrivateIpAddress     *netip.Addr
	CreatedAt            time.Time
	LastModified         time.Time
}

type UserRoleDao struct {
	Id              uuid.UUID `json:"id"`
	RoleName        string    `json:"roleName"`
	RoleDescription string    `json:"roleDesc"`
	Enabled         bool      `json:"enabled"`
	IsDeleted       bool      `json:"isDeleted"`
	CreatedAt       time.Time `json:"createdAt"`
	LastModified    time.Time `json:"lastModified"`
}

type AppPermissionDao struct {
	Id                    uuid.UUID `json:"id"`
	PermissionName        string    `json:"permissionName"`
	PermissionDescription string    `json:"permissionDescription"`
}

type RolePermissionMappingDao struct {
	Id           uuid.UUID `json:"id"`
	RoleId       uuid.UUID `json:"roleId"`
	PermissionId uuid.UUID `json:"permissionId"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
	LastModified time.Time `json:"lastModified"`
}
