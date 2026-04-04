# Go-Infra - Backend API and Database Layer

## Overview

**go-infra** is the backend REST API and database layer for the **Infractl** platform. While `infra-cli` provides the command-line interface for infrastructure management, `go-infra` powers the centralized API that persists inventory, manages authentication, orchestrates deployments, and serves the web dashboard.

### Relationship to Infra-CLI

```
┌─────────────────────────────────────────────────────────────────┐
│                      End Users                                   │
├──────────────────────────┬──────────────────────────────────────┤
│  Developers              │  Infrastructure Teams                 │
│  (SSH + Git only)        │  (Web Dashboard)                      │
└──────────────┬───────────┴─────────────────┬────────────────────┘
               │                             │
               v                             v
        ┌────────────────┐         ┌──────────────────┐
        │  infra-cli     │         │  db-helper-ui    │
        │  (CLI Tool)    │         │  (React Web UI)  │
        └────────┬───────┘         └────────┬─────────┘
                 │                          │
                 └──────────────┬───────────┘
                                │
                                v
                        ┌───────────────────┐
                        │    go-infra API   │  ← You are here
                        │  (REST + WebSocket)
                        └────────┬──────────┘
                                 │
                                 v
                        ┌──────────────────┐
                        │  PostgreSQL DB   │
                        │  (Inventory Data)│
                        └──────────────────┘
                                 │
                                 v
                        ┌──────────────────────┐
                        │ Infrastructure       │
                        │ (VMs, DBs, Network, │
                        │  Storage, Secrets)   │
                        └──────────────────────┘
```

---

## Core Architecture

### Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| API Framework | Go + net/http, gorilla/mux | HTTP server and routing |
| Database | PostgreSQL | Persistent storage of infrastructure state |
| Migrations | goose | Database schema versioning |
| Code Generation | sqlc | Type-safe SQL queries |
| Authentication | JWT | API token-based authentication |
| Authorization | Role-based access control (RBAC) | User permission management |
| WebSocket | gorilla/websocket | Real-time communication for deployments/terminals |
| Logging | slog | Structured logging |
| Documentation | Swagger/OpenAPI | API documentation |

### Key Design Principles

1. **Single Source of Truth** - Database is the authoritative source for infrastructure state
2. **Event-Driven** - Deployments and operations trigger events logged in database
3. **Multi-Tenant** - Support multiple users and organizations
4. **Audit Trail** - All operations are logged with user, timestamp, and result
5. **Async Operations** - Long-running deployments handled asynchronously with WebSocket updates

---

## Package Documentation

### API Services (api/)

#### **api/api_server/** - REST API Server
*Purpose*: HTTP server, routing, and request handling

**Responsibilities:**
- Start HTTP server on configured port (default: 8993)
- Route requests to appropriate handlers
- CORS handling for browser clients
- Middleware chain (logging, authentication, panic recovery)
- Graceful shutdown

**Key Files:**
- `api_server.go` - Server initialization and routing
- `logging_middleware.go` - Request/response logging
- `api_server_structs.go` - Shared server structures

**Typical HTTP Handler Pattern:**
```go
func handleGetHosts(w http.ResponseWriter, r *http.Request) {
    // 1. Authenticate user
    // 2. Validate query parameters
    // 3. Query database
    // 4. Format response
    // 5. Write JSON response
}
```

---

#### **api/authapi/** - Authentication and Authorization
*Purpose*: User authentication, JWT token management, and authorization

**Key Components:**
- `auth_api_handlers.go` - Login, logout, token refresh endpoints
- `auth_api_middleware.go` - Middleware to enforce authentication on protected routes
- `auth_jwt.go` - JWT token creation and validation
- `auth_hashing.go` - Password hashing and verification
- `auth_local_login.go` - Local user authentication
- `auth_interface.go` - Authentication provider interfaces

**Authentication Flow:**
```
1. User provides credentials (username + password)
2. authapi validates credentials against database
3. If valid, generates JWT token valid for 24 hours
4. Client includes token in Authorization header: "Bearer <token>"
5. middleware verifies token on each request
6. If invalid/expired, returns 401 Unauthorized
```

**JWT Token Structure:**
```json
{
    "sub": "user-id-uuid",
    "name": "username",
    "email": "user@example.com",
    "roles": ["user", "admin"],
    "exp": 1234567890,
    "iat": 1234567800
}
```

**Planned Authentication Methods:**
- OAuth2 with GitHub
- OAuth2 with Microsoft (Entra 365)
- OAuth2 with Google
- LDAP directory integration

---

#### **api/user_api_handlers/** - User Management
*Purpose*: CRUD operations for user accounts and permissions

**Endpoints:**
- `GET /api/v1/users` - List all users (admin only)
- `GET /api/v1/users/:id` - Get user details
- `POST /api/v1/users` - Create new user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `POST /api/v1/users/:id/roles` - Assign roles

**Key User Attributes:**
- User ID (UUID)
- Username
- Email
- Hashed password
- Roles (user, admin, operator)
- Created/modified timestamps
- Active status

---

### Database Layer (database/)

#### **database/bootstrap/** - Initial Setup
*Purpose*: Demo data and initial configuration

**Responsibilities:**
- Create sample infrastructure resources
- Configure default settings
- Initialize test data

---

#### **database/infra_db/** - Database Pool Management
*Purpose*: PostgreSQL connection pooling and lifecycle

**Key Features:**
- Connection pooling for performance
- Graceful connection shutdown
- Connection health checks

---

#### **database/infra_db_pg/** - SQL Query Layer
*Purpose*: Auto-generated, type-safe SQL queries (via sqlc)

**Key Responsibilities:**
- Execute database queries safely with prepared statements
- Map database rows to Go structs
- Handle NULL values properly
- Provide compile-time type safety for SQL

**Generated Code Includes:**
- Query execution functions with required/optional parameters
- Struct definitions matching table schemas
- Error handling for SQL errors

**Example Generated Query:**
```go
// database/infra_db_pg/query.sql.go
func (q *Queries) GetHostByID(ctx context.Context, id uuid.UUID) (*HostServer, error) {
    // Auto-generated implementation
}
```

---

#### **database/models/** - Data Models
*Purpose*: Go struct definitions matching database schema

**Key Models:**
- `HostServer` - Physical or virtual server resource
- `HostServerType` - Server type/category (VM, Proxmox, VPS, AWS Instance, etc.)
- `PlatformType` - Service running on host (Kubernetes, Docker, PostgreSQL, etc.)
- `SSHKey` - SSH key storage and metadata
- `Application` - Registered application for deployment
- `Deployment` - Application deployment instance and history
- `User` - User account
- `UserSecret` - API tokens and secrets

---

### Services (services/)

Services handle complex business logic that may require interactions with multiple resources, external APIs, and database operations.

#### **services/host_servers/** - Infrastructure Inventory
*Purpose*: Manage registered infrastructure resources

**Key Responsibilities:**
- CRUD operations for host servers
- Host server type and platform type mappings
- Query hosts by IP, hostname, or ID
- Track host status and capabilities
- Associate SSH keys and secrets with hosts

**Key Models:**
```go
HostServer {
    ID                   uuid.UUID      // Unique identifier
    Hostname             string         // DNS name or label
    IPAddress            netip.Addr     // IP address
    Username             *string        // Default SSH user
    SSHKeyID             *uuid.UUID     // Associated SSH key
    SudoPasswordSecretID *uuid.UUID     // Sudo password secret
    HostServerTypes      []HostServerType  // Tags/categories
    PlatformTypes        []PlatformType    // Services available
    CreatedAt            time.Time
    LastModified         time.Time
}

HostServerType {
    ID           uuid.UUID  // e.g., "Database Server"
    Name         string
    LastModified time.Time
}

PlatformType {
    ID           uuid.UUID  // e.g., "Docker Host", "PostgreSQL"
    Name         string
    LastModified time.Time
}
```

---

#### **services/external_applications/** - Application Registry
*Purpose*: Register and manage applications that can be deployed

**Responsibilities:**
- Application metadata (name, git URL, container image)
- Application version tracking
- Deployment requirement configuration
- Application permission management

**Key Attributes:**
- Application ID and name
- Git repository URL or OCI container image
- Required resources (compute, database, storage, network)
- Configuration variables and secrets
- Deployment scripts or Helm charts

---

#### **services/user_crud_svc/** - User Management Service
*Purpose*: Business logic for user account management

**Responsibilities:**
- Create user accounts
- Update user profiles
- Delete users (with cascade cleanup)
- Manage user permissions and roles
- Track user activity

---

#### **services/user_secrets/** - Secrets Management
*Purpose*: Securely store and manage user secrets

**Stored Secret Types:**
- SSH private keys
- API tokens (Proxmox, Cloudflare, AWS, etc.)
- Database credentials
- OAuth tokens
- Custom application secrets

**Security Features:**
- Encrypted storage in database
- Audit trail of secret access
- Automatic secret rotation (planned)
- Access control by user role
---

#### **services/ssh_connections/** - WebSocket SSH Terminal
*Purpose*: Interactive SSH terminal via WebSocket

**Components:**
- SSH session management
- PTY (pseudo-terminal) allocation
- Real-time bidirectional data transfer
- Session logging and audit trail

**WebSocket Endpoint:**
```
ws://localhost:8080/ws/ssh/{hostID}
```

**Workflow:**
```
1. Client connects to WebSocket endpoint
2. Server authenticates using credentials from user_secrets
3. Server establishes SSH connection to target host
4. Client PtyRequest (column/row size)
5. Server allocates PTY on remote
6. Bidirectional data transfer (stdin/stdout/stderr)
7. Session closes, log recorded
```

---

#### **services/ssh_key_provider/** - SSH Key Management
*Purpose*: Retrieve and manage SSH keys for remote connections

**Responsibilities:**
- Store SSH keys securely
- Retrieve keys for authentication
- Support multiple key formats
- Key rotation and updates

---

#### **services/node_networking/** - Networking Configuration
*Purpose*: Manage network configuration and connectivity

**Planned Features:**
- VLAN configuration
- Firewall rule management
- Load balancer configuration
- Network segmentation

---

### Data Persistence (database/migrations and sqlc.yaml)

#### **Migrations** - Schema Versioning
*Location*: `/migrations/` directory, managed by goose

**Migration Workflow:**
```bash
# Create new migration
goose create add_new_table sql

# Run migrations
goose up

# Rollback
goose down

# View migration status
goose status
```

**Schema Includes Tables For:**
- Users and authentication
- Host servers and types
- Applications and deployments
- SSH keys and secrets
- Audit logs
- Session data

#### **sqlc** - SQL Code Generation
*Location*: `sqlc.yaml` configuration file

**Workflow:**
```
1. Write SQL queries in .sql files
2. Run: sqlc generate
3. Go code auto-generated with:
   - Prepared statements
   - Type-safe parameters
   - Return struct types
   - Error handling
```

---

### Internal Utilities (internal/)

#### **internal/middleware/** - HTTP Middleware
*Purpose*: Cross-cutting concerns for all HTTP requests

**Middleware Stack:**
1. Panic recovery - Gracefully handle panics
2. Request logging - Log all requests/responses
3. CORS - Handle cross-origin requests
4. Authentication - Verify JWT tokens
5. Authorization - Check user permissions

---

#### **internal/cors/** - CORS Configuration
*Purpose*: Allow browser clients from different origins

**Configured For:**
- localhost:3000 (development UI)
- Production frontends (configurable)

**Methods Allowed:** GET, POST, PUT, DELETE, OPTIONS  
**Headers Allowed:** Content-Type, Authorization

---

#### **internal/swaggerui/** - API Documentation
*Purpose*: Serve interactive API documentation

**Endpoint:**
```
http://localhost:8080/swagger
```

**Features:**
- Interactive endpoint testing
- Schema documentation
- Request/response examples
- Authentication playground

**Related Files:**
- `swagger.json` / `swagger.yaml` - OpenAPI specification
- Generated from code annotations

---

#### **internal/embedbin/** - Embedded Resources
*Purpose*: Embed migration schema and other resources in binary

**Benefits:**
- No external files required
- Migrations run automatically on startup
- Self-contained deployments

---

#### **internal/bumper/** - Version Management
*Purpose*: Semantic versioning for API and services

---

#### **internal/hashing/** - Cryptography
*Purpose*: Password hashing and verification

**Algorithm:** bcrypt with configurable cost factor  
**Security:** Salt generated automatically per password

---

#### **internal/pretty/** - Output Formatting
*Purpose*: Format responses and error messages

---

#### **internal/type_helper/** - Type Utilities
*Purpose*: Type conversions and validations

---

### Configuration

#### **cfg.yaml** - Application Configuration
```yaml
server:
  port: 8080
  host: 0.0.0.0
  
database:
  host: localhost
  port: 5432
  user: infractl
  password: ${DB_PASSWORD}
  name: infractl_db
  
jwt:
  secret: ${JWT_SECRET}
  expiry_hours: 24
  
auth:
  allowed_origins:
    - http://localhost:3000
    - https://example.com
    
logging:
  level: info
  format: json
```

#### **version.yaml** - Version Information
```yaml
version: 1.2.3
api_version: v1
compatible_cli_versions:
  - 1.2.x
```

---

## Database Schema Overview

### Core Tables

#### **users**
- id (UUID)
- username (string, unique)
- email (string)
- password_hash (string)
- roles (string array)
- created_at (timestamp)
- last_modified (timestamp)
- is_active (boolean)

#### **host_servers**
- id (UUID)
- hostname (string)
- ip_address (inet)
- username (string, nullable)
- ssh_key_id (UUID, nullable)
- sudo_password_secret_id (UUID, nullable)
- created_at (timestamp)
- last_modified (timestamp)

#### **host_server_types**
- id (UUID)
- name (string)
- last_modified (timestamp)

#### **host_server_type_mappings**
- host_server_id (UUID) → foreign key
- host_server_type_id (UUID) → foreign key

#### **platform_types**
- id (UUID)
- name (string) - "PostgreSQL", "Docker", "Kubernetes", etc.
- last_modified (timestamp)

#### **host_server_platform_mappings**
- host_server_id (UUID) → foreign key
- platform_type_id (UUID) → foreign key
- host_server_type_id (UUID) → foreign key (optional)

#### **ssh_keys**
- id (UUID)
- owner_id (UUID) → foreign key
- key_name (string)
- public_key (string)
- key_type (enum) - "ed25519", "rsa", etc.
- fingerprint (string)
- created_at (timestamp)

#### **user_secrets**
- id (UUID)
- owner_id (UUID) → foreign key
- secret_name (string)
- secret_type (enum) - "api_token", "password", "ssh_key", etc.
- encrypted_value (bytea)
- created_at (timestamp)
- accessed_at (timestamp, nullable)

#### **applications**
- id (UUID)
- owner_id (UUID) → foreign key
- app_name (string)
- git_repo_url (string, nullable)
- oci_image (string, nullable)
- version (string)
- created_at (timestamp)
- last_modified (timestamp)

#### **deployments**
- id (UUID)
- application_id (UUID) → foreign key
- target_host_id (UUID) → foreign key
- deployed_by (UUID) → foreign key (user)
- status (enum) - "pending", "running", "success", "failed"
- started_at (timestamp)
- completed_at (timestamp, nullable)
- logs (text)

#### **audit_log**
- id (UUID)
- user_id (UUID) → foreign key
- action (string) - "create_vm", "deploy_app", "update_secret"
- resource_type (string) - "host", "deployment", "user"
- resource_id (UUID, nullable)
- status (enum) - "success", "failure"
- details (jsonb)
- created_at (timestamp)

---

## API Endpoints

### Authentication
```
POST /api/v1/auth/login
    Request: {"username": "...", "password": "..."}
    Response: {"token": "...", "expires_in": 86400}

POST /api/v1/auth/refresh
    Header: Authorization: Bearer <token>
    Response: {"token": "...", "expires_in": 86400}

POST /api/v1/auth/logout
    Header: Authorization: Bearer <token>
```

### Users
```
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/:id
PUT    /api/v1/users/:id
DELETE /api/v1/users/:id
```

### Host Servers (Infrastructure Inventory)
```
GET    /api/v1/hosts
POST   /api/v1/hosts
GET    /api/v1/hosts/:id
PUT    /api/v1/hosts/:id
DELETE /api/v1/hosts/:id

GET    /api/v1/hosts/:id/platform-types
POST   /api/v1/hosts/:id/platform-types
```

### Applications
```
GET    /api/v1/applications
POST   /api/v1/applications
GET    /api/v1/applications/:id
PUT    /api/v1/applications/:id
DELETE /api/v1/applications/:id
```

### Deployments
```
GET    /api/v1/deployments
POST   /api/v1/applications/:id/deploy
GET    /api/v1/deployments/:id
GET    /api/v1/deployments/:id/logs

WS     /api/v1/ws/deployments/:id
       Real-time deployment logs via WebSocket
```

### SSH Terminal
```
WS     /api/v1/ws/ssh/:host-id
       Interactive SSH session via WebSocket
```

### Secrets
```
GET    /api/v1/secrets
POST   /api/v1/secrets
GET    /api/v1/secrets/:id
PUT    /api/v1/secrets/:id
DELETE /api/v1/secrets/:id
```

### Databases
```
GET    /api/v1/host/:id/databases
POST   /api/v1/host/:id/databases
```

---

## Key Workflows

### Workflow 1: User Registration and Authentication

```
1. API receives POST /auth/register
   └─> Validate email/username
   └─> Hash password with bcrypt
   └─> Insert into users table
   └─> Return success

2. User logs in with POST /auth/login
   └─> Look up user by username
   └─> Verify password with bcrypt
   └─> Generate JWT token
   └─> Return token

3. Client uses token for subsequent requests
   └─> Include: Authorization: Bearer <token>
   └─> Server validates JWT on each request
   └─> Extract user ID from token claims
   └─> Enforce user-level permissions
```

### Workflow 2: Register Host Server

```
1. User calls POST /api/v1/hosts
   └─> Request: {
       "hostname": "db-server-1",
       "ip_address": "192.168.1.100",
       "username": "admin",
       "ssh_key_id": "uuid-of-key"
     }

2. API validates input
   └─> Check IP address format
   └─> Verify SSH key exists
   └─> Check hostname not already registered
   
3. Insert into database
   └─> Create host_servers row
   └─> Create host_server_type_mappings
   └─> Create platform_type_mappings
   └─> Return host details

4. Host available for deployments
   └─> Can be target for application deployments
   └─> Can be queried for available resources
```

### Workflow 3: Deploy Application

```
1. User calls POST /api/v1/applications/:id/deploy
   └─> Request: {
       "target_host_id": "uuid",
       "configuration": {...}
     }

2. System creates deployment record
   └─> INSERT into deployments table
   └─> Status = "pending"

3. Background job processes deployment
   └─> Retrieve application details
   └─> Retrieve target host SSH credentials
   └─> Connect via SSH
   └─> Clone git repo OR pull container image
   └─> Install dependencies
   └─> Create systemd service
   └─> Start application
   └─> Update deployment status = "success"

4. Real-time updates to client
   └─> WebSocket sends deployment progress
   └─> Client displays live logs
   └─> Status changes push to UI

5. Deployment complete
   └─> Audit log recorded
   └─> User receives notification
   └─> Application accessible at configured address
```

---

## Deployment Modes

### Supported Deployment Targets

```
1. Direct SSH + Systemd
   - Fresh Linux VM/server
   - Create system user
   - Deploy binary/script
   - Create systemd service
   - Start service

2. Kubernetes Cluster
   - Push container to registry
   - Create Kubernetes manifests
   - Apply manifests (kubectl)
   - Verify pod startup
   - Configure ingress

3. Docker Host
   - Build/pull container image
   - Push to container registry
   - Run docker run with configuration
   - Configure networking
   - Set up volumes

4. Serverless (Planned)
   - AWS Lambda
   - Google Cloud Functions
   - Azure Functions
```

---

## Security Features

### Authentication
- JWT token-based API authentication
- Token expiration (default 24 hours)
- Refresh token support
- Future: OAuth2, LDAP, MFA

### Authorization
- Role-based access control (RBAC)
- Roles: user, operator, admin
- Admin: Full access
- Operator: Deploy/manage resources
- User: Read-only access

### Data Security
- Password hashing with bcrypt
- Encrypted storage of secrets
- TLS/HTTPS for all communications
- Database connection pooling with SSL
- Audit logging of all operations

### Network Security
- CORS configuration for trusted origins
- Rate limiting (planned)
- API key rotation support
- IP whitelisting (planned)

---

## Performance Considerations

### Database Optimization
- Connection pooling (pgx)
- Prepared statements (sqlc)
- Indexed queries for frequent lookups
- Pagination for large result sets

### Caching Strategy
- User session caching
- Host/server list caching
- Application metadata caching
- Invalidation on updates

### Async Processing
- Long-running deployments in background
- WebSocket for real-time updates
- Job queue for scalability (planned)

---

## Monitoring and Logging

### Structured Logging
```
{
    "timestamp": "2025-04-02T10:30:00Z",
    "level": "info",
    "service": "deployment",
    "action": "deploy_started",
    "deployment_id": "uuid",
    "user_id": "uuid",
    "host_id": "uuid",
    "status": "running",
    "duration_ms": 1200
}
```

### Metrics to Track
- API response times
- Database query performance
- Deployment success/failure rates
- User login attempts
- Resource utilization

---

## Development and Testing

### Running go-infra

```bash
# Set environment variables
export DB_PASSWORD=postgres
export JWT_SECRET=your-secret-key

# Load environment
source .env

# Run migrations
./config_goose.sh

# Start API server
go run main.go

# Server available at http://localhost:8080
```

### Testing

```bash
# Run unit tests
go test ./...

# Run specific package tests
go test ./services/host_servers -v

# Generate test coverage
go test -cover ./...
```

### Database Management

```bash
# Create new migration
goose create add_new_field sql

# Apply migrations
goose up

# Rollback last migration
goose down

# Check status
goose status
```

---

## Future Enhancements

### Planned Features

1. **Advanced Authentication**
   - Multi-factor authentication (MFA)
   - OAuth2/OIDC providers (GitHub, Azure, Google)
   - LDAP directory integration
   - API key management

2. **Deployment Enhancements**
   - Canary deployments
   - Blue-green deployments
   - Automatic rollback on failure
   - Helm chart support for Kubernetes

3. **Infrastructure Management**
   - Infrastructure as Code (Terraform) integration
   - Policy enforcement
   - Cost tracking and optimization
   - Resource quotas and limits

4. **Observability**
   - Distributed tracing (OpenTelemetry)
   - Metrics export (Prometheus)
   - Log aggregation integration
   - Alert management

5. **Scalability**
   - Horizontal scaling with load balancing
   - Database read replicas
   - Job queue for async operations (Redis)
   - Caching layer (Redis)

6. **Advanced Features**
   - Service mesh integration (Istio)
   - Multi-cluster management
   - GitOps integration
   - Compliance and audit reporting

---

## Integration with Infra-CLI

### How They Work Together

1. **CLI as Client**: `infra-cli` makes REST API calls to `go-infra`
2. **Shared Database**: Both read/write to same PostgreSQL instance
3. **Authentication**: CLI obtains JWT token from API, uses for subsequent calls
4. **Real-time Updates**: CLI subscribes to WebSocket for deployment progress
5. **Offline Capability**: CLI can work offline with cached credentials (planned)

### Data Flow Example: Deploy Application

```
infra-cli (user)
   ↓
infractl deploy remote-systemd app
   ↓
cli/cmd/deploy_remote_systemd_app.go
   ↓
POST /api/v1/applications/id/deploy
   (JWT token in Authorization header)
   ↓
go-infra api_server
   ↓
api/authapi (verify token)
   ↓
services/external_applications (get app details)
   ↓
INSERT deployment record with status="pending"
   ↓
Background job (async)
   ↓
services/host_servers (get host SSH creds)
   ↓
SSH deploy (via ssh package)
   ↓
UPDATE deployment status="success"
   ↓
WebSocket notification to client
   ↓
CLI displays success message
```

---

## Support

- **Issues**: GitHub Issues
- **Documentation**: [Wiki](https://github.com/babbage88/infractl/wiki)
- **API Docs**: Swagger UI at /swagger
- **SQL Queries**: See query.sql in migrations

---

**Last Updated**: April 2026  
**Version**: 1.0  
**Maintained By**: Infractl Team
