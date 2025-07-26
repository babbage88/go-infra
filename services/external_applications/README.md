# External Applications Service

This package provides a complete CRUD interface for managing external integration applications in the system.

## Overview

The External Applications Service manages the `external_integration_apps` table, which stores information about external applications that can be integrated with the system. Each external application has a unique name and can optionally include an endpoint URL and description.

## Features

- **Create** external applications with name, endpoint URL, and description
- **Read** external applications by ID or name
- **Update** external application details
- **Delete** external applications by ID or name
- **List** all external applications
- **Lookup** application ID by name or name by ID

## Data Structures

### ExternalApplicationDao
The main data transfer object representing an external application:

```go
type ExternalApplicationDao struct {
    Id             uuid.UUID `json:"id"`
    Name           string    `json:"name"`
    CreatedAt      time.Time `json:"createdAt"`
    LastModified   time.Time `json:"lastModified"`
    EndpointUrl    string    `json:"endpointUrl,omitempty"`
    AppDescription string    `json:"appDescription,omitempty"`
}
```

### CreateExternalApplicationRequest
Request structure for creating a new external application:

```go
type CreateExternalApplicationRequest struct {
    Name           string `json:"name" validate:"required"`
    EndpointUrl    string `json:"endpointUrl,omitempty"`
    AppDescription string `json:"appDescription,omitempty"`
}
```

### UpdateExternalApplicationRequest
Request structure for updating an existing external application:

```go
type UpdateExternalApplicationRequest struct {
    Name           string `json:"name,omitempty"`
    EndpointUrl    string `json:"endpointUrl,omitempty"`
    AppDescription string `json:"appDescription,omitempty"`
}
```

## Interface

The service implements the `ExternalApplications` interface:

```go
type ExternalApplications interface {
    CreateExternalApplication(req CreateExternalApplicationRequest) (*ExternalApplicationDao, error)
    GetExternalApplicationById(id uuid.UUID) (*ExternalApplicationDao, error)
    GetExternalApplicationByName(name string) (*ExternalApplicationDao, error)
    GetAllExternalApplications() ([]ExternalApplicationDao, error)
    UpdateExternalApplication(id uuid.UUID, req UpdateExternalApplicationRequest) (*ExternalApplicationDao, error)
    DeleteExternalApplicationById(id uuid.UUID) error
    DeleteExternalApplicationByName(name string) error
    GetExternalApplicationIdByName(name string) (uuid.UUID, error)
    GetExternalApplicationNameById(id uuid.UUID) (string, error)
}
```

## Usage

### Initialization

```go
import (
    "github.com/babbage88/go-infra/services/external_applications"
    "github.com/jackc/pgx/v5/pgxpool"
)

// Initialize the service with a database connection
dbPool, err := pgxpool.New(context.Background(), databaseURL)
if err != nil {
    log.Fatal(err)
}

service := &external_applications.ExternalApplicationsService{
    DbConn: dbPool,
}
```

### Creating an External Application

```go
req := external_applications.CreateExternalApplicationRequest{
    Name:           "cloudflare",
    EndpointUrl:    "https://api.cloudflare.com",
    AppDescription: "Cloudflare DNS and CDN integration",
}

app, err := service.CreateExternalApplication(req)
if err != nil {
    log.Printf("Error creating external application: %v", err)
    return
}

fmt.Printf("Created application: %s (ID: %s)\n", app.Name, app.Id)
```

### Retrieving an External Application

```go
// By ID
app, err := service.GetExternalApplicationById(appId)
if err != nil {
    log.Printf("Error getting external application: %v", err)
    return
}

// By name
app, err := service.GetExternalApplicationByName("cloudflare")
if err != nil {
    log.Printf("Error getting external application: %v", err)
    return
}
```

### Updating an External Application

```go
req := external_applications.UpdateExternalApplicationRequest{
    EndpointUrl:    "https://api.cloudflare.com/v4",
    AppDescription: "Updated Cloudflare integration with v4 API",
}

updatedApp, err := service.UpdateExternalApplication(appId, req)
if err != nil {
    log.Printf("Error updating external application: %v", err)
    return
}
```

### Deleting an External Application

```go
// By ID
err := service.DeleteExternalApplicationById(appId)
if err != nil {
    log.Printf("Error deleting external application: %v", err)
    return
}

// By name
err := service.DeleteExternalApplicationByName("cloudflare")
if err != nil {
    log.Printf("Error deleting external application: %v", err)
    return
}
```

### Listing All External Applications

```go
apps, err := service.GetAllExternalApplications()
if err != nil {
    log.Printf("Error getting external applications: %v", err)
    return
}

for _, app := range apps {
    fmt.Printf("Application: %s (ID: %s)\n", app.Name, app.Id)
}
```

## Database Schema

The service operates on the `external_integration_apps` table:

```sql
CREATE TABLE public.external_integration_apps (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    endpoint_url text,
    app_description text
);
```

## Health Checks

The service includes health check functionality:

```go
healthCheck := &external_applications.ExternalApplicationsHealthCheck{
    DbConn: dbPool,
}

// Basic health check
err := healthCheck.HealthCheck()

// Database-specific health check
err := healthCheck.DatabaseHealthCheck()
```

## Error Handling

The service returns descriptive errors for common scenarios:

- **Not Found**: When trying to get, update, or delete a non-existent application
- **Database Errors**: When database operations fail
- **Validation Errors**: When required fields are missing

All errors are wrapped with context using `fmt.Errorf` and include relevant identifiers (ID, name) for debugging.

## Testing

Run the tests with:

```bash
go test ./services/external_applications/...
```

The test suite includes:
- Interface compliance tests
- Request structure validation
- DAO structure validation 