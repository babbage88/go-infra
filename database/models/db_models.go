package db_models

import (
	"time"
)

type HostServer struct {
	Id               int64     `json:"id"`
	HostName         string    `json:"hostname"`
	IpAddress        string    `json:"ip_address"`
	UserName         string    `json:"username"`
	PublicSshKeyname string    `json:"public_ssh_key"`
	HostedDomains    []string  `json:"hosted_domains"`
	SslKeyPath       string    `json:"ssl_key_path"`
	IsContainerHost  bool      `json:"is_container_host"`
	IsVmHost         bool      `json:"is_vm_host"`
	IsVirtualMachine bool      `json:"is_virtual_machine"`
	IsDbHost         bool      `json:"is_db_host"`
	CreatedAt        time.Time `json:"created"`
	LastModified     time.Time `json:"last_modified"`
}

type User struct {
	Id           int64     `json:"id"`
	Username     string    `json:"userName"`
	Password     string    `json:"password"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created"`
	LastModified time.Time `json:"last_modified"`
}

type AuthToken struct {
	Id           int64     `json:"id"`
	UserId       int64     `json:"UserId"`
	Token        string    `json:"token"`
	Expiration   time.Time `json:"expiration"`
	CreatedAt    time.Time `json:"created"`
	LastModified time.Time `json:"last_modified"`
}
