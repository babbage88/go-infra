package db_models

import (
	"time"
)

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

type User struct {
	Id       int64    `json:"id"`
	Username string   `json:"userName"`
	Password string   `json:"password"`
	Email    string   `json:"email"`
	ApiTokes []string `json:"apiTokens"`
}

type AuthToken struct {
	Id         int64     `json:"id"`
	UserId     int64     `json:"UserId"`
	Token      string    `json:"token"`
	Expiration time.Time `json:"expiration"`
}
