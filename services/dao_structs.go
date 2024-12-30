package services

import (
	"net/netip"
	"time"
)

// Respose will return login result and the user info.
// swagger:response UserDao
// This text will appear as description of your response body.
type UserDao struct {
	// in:body
	Id           int32     `json:"id"`
	UserName     string    `json:"username"`
	Email        string    `json:"email"`
	Roles        []string  `json:"roles"`
	RoleIds      []int32   `json:"role_ids"`
	CreatedAt    time.Time `json:"createdAt"`
	LastModified time.Time `json:"lastModified"`
	Enabled      bool      `json:"enabled"`
	IsDeleted    bool      `json:"isDeleted"`
}

type AuthTokenDao struct {
	Id           int32     `json:"id"`
	UserID       int32     `json:"user_id"`
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
	Id              int32     `json:"id"`
	RoleName        string    `json:"roleName"`
	RoleDescription string    `json:"roleDesc"`
	Enabled         bool      `json:"enabled"`
	IsDeleted       bool      `json:"isDeleted"`
	CreatedAt       time.Time `json:"createdAt"`
	LastModified    time.Time `json:"lastModified"`
}

type AppPermissionDao struct {
	Id                    int32  `json:"id"`
	PermissionName        string `json:"permissionName"`
	PermissionDescription string `json:"permissionDescription"`
}

type RolePermissionMappingDao struct {
	Id           int32     `json:"id"`
	RoleId       int32     `json:"roleId"`
	PermissionId int32     `json:"permissionId"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
	LastModified time.Time `json:"lastModified"`
}
