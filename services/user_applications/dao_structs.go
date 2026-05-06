package user_applications

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// swagger:model InfraDependencyDao
type InfraDependencyDao struct {
	Id                 uuid.UUID              `json:"id"`
	DependencyType     string                 `json:"dependencyType"`
	DependencyName     string                 `json:"dependencyName"`
	HostServerTypeId   *uuid.UUID             `json:"hostServerTypeId,omitempty"`
	HostServerTypeName string                 `json:"hostServerTypeName,omitempty"`
	PlatformTypeId     *uuid.UUID             `json:"platformTypeId,omitempty"`
	PlatformTypeName   string                 `json:"platformTypeName,omitempty"`
	Config             map[string]interface{} `json:"config,omitempty"`
	CreatedAt          time.Time              `json:"createdAt"`
	LastModified       time.Time              `json:"lastModified"`
}

// swagger:model UserApplicationDao
type UserApplicationDao struct {
	Id                uuid.UUID              `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description,omitempty"`
	RepositoryUrl     string                 `json:"repositoryUrl"`
	ManifestPath      string                 `json:"manifestPath,omitempty"`
	SourceKind        string                 `json:"sourceKind"`
	ModuleName        string                 `json:"moduleName,omitempty"`
	PackageName       string                 `json:"packageName,omitempty"`
	PackageManager    string                 `json:"packageManager,omitempty"`
	DeployKind        string                 `json:"deployKind"`
	Registerable      bool                   `json:"registerable"`
	DeployConfig      map[string]interface{} `json:"deployConfig,omitempty"`
	BuildConfig       map[string]interface{} `json:"buildConfig,omitempty"`
	InfraDependencies []InfraDependencyDao   `json:"infraDependencies,omitempty"`
	CreatedAt         time.Time              `json:"createdAt"`
	LastModified      time.Time              `json:"lastModified"`
}

// swagger:model CreateInfraDependencyRequest
type CreateInfraDependencyRequest struct {
	DependencyType   string                 `json:"dependencyType"`
	DependencyName   string                 `json:"dependencyName"`
	HostServerTypeId *uuid.UUID             `json:"hostServerTypeId,omitempty"`
	PlatformTypeId   *uuid.UUID             `json:"platformTypeId,omitempty"`
	Config           map[string]interface{} `json:"config,omitempty"`
}

// swagger:model DiscoveredInfraDependency
type DiscoveredInfraDependency struct {
	DependencyType string                 `json:"dependencyType"`
	DependencyName string                 `json:"dependencyName"`
	Config         map[string]interface{} `json:"config,omitempty"`
}

// swagger:model DiscoverUserApplicationRequest
type DiscoverUserApplicationRequest struct {
	RepositoryUrl string `json:"repositoryUrl" validate:"required"`
	Branch        string `json:"branch,omitempty"`
	Tag           string `json:"tag,omitempty"`
}

// swagger:model DiscoveredUserApplicationCandidate
type DiscoveredUserApplicationCandidate struct {
	Name              string                      `json:"name"`
	Description       string                      `json:"description,omitempty"`
	RepositoryUrl     string                      `json:"repositoryUrl"`
	ManifestPath      string                      `json:"manifestPath,omitempty"`
	SourceKind        string                      `json:"sourceKind,omitempty"`
	ModuleName        string                      `json:"moduleName,omitempty"`
	PackageName       string                      `json:"packageName,omitempty"`
	PackageManager    string                      `json:"packageManager,omitempty"`
	DeployKind        string                      `json:"deployKind,omitempty"`
	ApplicationKind   string                      `json:"applicationKind,omitempty"`
	Registerable      bool                        `json:"registerable"`
	DeployConfig      map[string]interface{}      `json:"deployConfig,omitempty"`
	BuildConfig       map[string]interface{}      `json:"buildConfig,omitempty"`
	InfraDependencies []DiscoveredInfraDependency `json:"infraDependencies,omitempty"`
}

// swagger:model DiscoverUserApplicationResponse
type DiscoverUserApplicationResponse struct {
	Candidates []DiscoveredUserApplicationCandidate `json:"candidates"`
}

// swagger:model CreateUserApplicationRequest
type CreateUserApplicationRequest struct {
	Name              string                         `json:"name" validate:"required"`
	Description       string                         `json:"description,omitempty"`
	RepositoryUrl     string                         `json:"repositoryUrl" validate:"required"`
	ManifestPath      string                         `json:"manifestPath,omitempty"`
	SourceKind        string                         `json:"sourceKind,omitempty"`
	ModuleName        string                         `json:"moduleName,omitempty"`
	PackageName       string                         `json:"packageName,omitempty"`
	PackageManager    string                         `json:"packageManager,omitempty"`
	DeployKind        string                         `json:"deployKind" validate:"required"`
	Registerable      *bool                          `json:"registerable,omitempty"`
	DeployConfig      map[string]interface{}         `json:"deployConfig,omitempty"`
	BuildConfig       map[string]interface{}         `json:"buildConfig,omitempty"`
	InfraDependencies []CreateInfraDependencyRequest `json:"infraDependencies,omitempty"`
}

// swagger:model UpdateUserApplicationRequest
type UpdateUserApplicationRequest struct {
	Name              string                         `json:"name,omitempty"`
	Description       string                         `json:"description,omitempty"`
	RepositoryUrl     string                         `json:"repositoryUrl,omitempty"`
	ManifestPath      string                         `json:"manifestPath,omitempty"`
	SourceKind        string                         `json:"sourceKind,omitempty"`
	ModuleName        string                         `json:"moduleName,omitempty"`
	PackageName       string                         `json:"packageName,omitempty"`
	PackageManager    string                         `json:"packageManager,omitempty"`
	DeployKind        string                         `json:"deployKind,omitempty"`
	Registerable      *bool                          `json:"registerable,omitempty"`
	DeployConfig      map[string]interface{}         `json:"deployConfig,omitempty"`
	BuildConfig       map[string]interface{}         `json:"buildConfig,omitempty"`
	InfraDependencies []CreateInfraDependencyRequest `json:"infraDependencies,omitempty"`
}

func marshalOptionalJSON(value map[string]interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}

func unmarshalOptionalJSON(value []byte) map[string]interface{} {
	if len(value) == 0 {
		return nil
	}
	parsed := map[string]interface{}{}
	if err := json.Unmarshal(value, &parsed); err != nil {
		return map[string]interface{}{
			"_raw": string(value),
		}
	}
	return parsed
}
