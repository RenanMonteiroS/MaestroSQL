# MaestroSQL

**MaestroSQL** is a powerful, user-friendly web application designed to simplify SQL Server database backup and restore operations. Built with Go and featuring a modern web interface, it provides an intuitive way to manage database operations with authentication support and concurrent processing capabilities.

## 🎯 Objective

MaestroSQL addresses the complexity of SQL Server database management by providing:

- **Simplified Database Operations**: Easy-to-use web interface for backup and restore operations
- **Concurrent Processing**: High-performance parallel processing for multiple database operations
- **Security**: Optional session-based authentication with OAuth2 (Google, Microsoft) and MFA support, CSRF and CORS protection, and SSL/TLS encryption.
- **Cross-Platform**: Works on Windows, macOS, and Linux
- **Real-time Monitoring**: Comprehensive and structured logging and progress tracking

## 🚀 Features

### Core Features
- 🔐 **Optional Authentication**: Session-based authentication with support for multiple methods, including [OSI](https://github.com/RenanMonteiroS/OSI), Google OAuth2, and Microsoft OAuth2.
- 📊 **Database Discovery**: Automatic detection and listing of SQL Server databases
- 💾 **Backup Operations**: Concurrent backup of multiple databases with timestamp naming
- 🔄 **Restore Operations**: Intelligent restore from .bak files with automatic path resolution
- 📝 **Structured Logging**: Detailed and structured operation logs for backup, restore, and error tracking
- 🎨 **Modern UI**: Bootstrap-based responsive web interface with step-by-step wizard
- 🌐 **Multi-language Support**: Support for English (en-US) and Portuguese (pt-BR).
- 📄 **Custom 404 Page**: A user-friendly 404 page to handle invalid routes.

### Technical Features
- ⚡ **Concurrent Processing**: Goroutine-based parallel operations for better performance
- 🛡️ **Security**: CSRF and CORS protection, and SSL/TLS encryption
- ⚔️ **Sanitized Queries**: All SQL queries are sanitized to prevent SQL injection attacks.
- 📦 **Standardized JSON Response**: All API responses follow a standard JSON format.
- ⏱️ **Timeout Handling**: Configurable timeouts (10 min backup, 15 min restore)
- 🔧 **Automatic Path Resolution**: Uses SQL Server's default data and log paths
- 📁 **Smart File Handling**: Supports both .bak and .BAK file extensions
- 🌐 **Cross-Platform Browser Support**: Automatic browser opening on application start

## 🏗️ Architecture

MaestroSQL follows a clean architecture pattern with clear separation of concerns:

```
┌───────────────────────────────────────────────────────────────┐
│                         Web Layer                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐│
│  │   HTML Forms    │  │  REST API       │  │  Static Assets  ││
│  │  (Bootstrap)    │  │  (Gin Router)   │  │  (Embedded)     ││
│  └─────────────────┘  └─────────────────┘  └─────────────────┘│
└───────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌───────────────────────────────────────────────────────────────┐
│                      Controller Layer                         │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                  DatabaseController                     │  │
│  │  • ConnectDatabase()             • BackupDatabase()     │  │
│  │  • GetDatabases()                • RestoreDatabase()    │  │
│  └─────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌───────────────────────────────────────────────────────────────┐
│                         Service Layer                         │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                     DatabaseService                     │  │
│  │  • Authentication Logic    • Error Handling             │  │
│  │  • Business Rules         • Logging Coordination        │  │
│  │  • Data Transformation    • Concurrency Management      │  │
│  └─────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌───────────────────────────────────────────────────────────────┐
│                       Repository Layer                        │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                   DatabaseRepository                    │  │
│  │  • SQL Query Execution      • Connection Management     │  │
│  │  • Transaction Handling     • Concurrent Operations     │  │
│  │  • Result Processing        • Timeout Management        │  │
│  └─────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌───────────────────────────────────────────────────────────────┐
│                           Data Layer                          │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                        SQL Server                       │  │
│  │  • Database Files                 • System Catalogs     │  │
│  │  • Backup/Restore Engine          • Default Paths       │  │
│  │  • Transaction Logs               • File Metadata       │  │
│  └─────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
```

### Component Details

#### Models
- **ConnInfo**: Database connection parameters
- **Database**: Database metadata with file information
- **DatabaseFile**: Individual database file details (data/log)
- **RestoreDb**: Restore operation configuration
- **BackupDataFile**: Comprehensive backup file metadata

#### Configuration
- **Authentication**: Configurable session-based authentication.
- **Timeouts**: Operation-specific timeout settings
- **Paths**: Automatic SQL Server path detection

## 📡 API Endpoints

### Authentication
Authentication is handled via a session-based cookie store. When authentication is enabled, most endpoints are protected by a middleware (`middleware/auth_middleware.go`) that verifies the user's session. The session status can be checked at the `/session` endpoint.

#### `GET /login`
**Description**: Initiates the authentication process. The authentication method is specified via a query parameter.
- **Query Parameters**:
    - `method`: `osi`, `google`, or `microsoft`.
- **Behavior**:
    - For `google` and `microsoft`, it redirects the user to the respective OAuth2 provider's login page.
    - For `osi`, it requires a `POST` request with user credentials.

#### `POST /login?method=osi`
**Description**: Authenticates the user using the OSI method.
- **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "password",
    "mfaKey": "123456"
  }
  ```
- **Response (success)**: 
  ```json
  {
      "status": "success",
      "code": 200,
      "message": "Login done successfully",
      "data": {
          "user": "example@example.com"
      },
      "timestamp": "2025-07-16T10:40:11-03:00",
      "path": "/login"
  }
  ```
- **Response (fail)**: 
  ```json
  {
    "status": "error",
    "code": 400,
    "message": "OSI Response is not OK",
    "errors": {
        "osiMsg": "Wrong TOTP value"
    },
    "timestamp": "2025-07-16T10:41:59-03:00",
    "path": "/login"
  }
  ```

#### `GET /logout`
**Description**: Clears the user's session cookie, effectively logging them out. If the logout was successful, it will return a status 200.

#### `GET /session`
**Description**: Retrieves session information. If the user have a session, it returns a 200 status code. If not, it returns a 401 status code.
- **Response (success)**:
  ```json
  {
    "success": true,
    "code": 200,
    "message": "Session found",
    "data": {
      "user": "user@example.com"
    },
    "timestamp": "2025-07-16T10:23:47-03:00",
    "path": "/session"
  }
  ```
- **Response (fail)**:
  ```json
  {
    "status": "error",
    "code": 401,
    "message": "None session was found",
    "errors": {
        "session": "None session was found"
    },
    "timestamp": "2025-07-16T10:26:23-03:00",
    "path": "/session"
  }
  ```

#### `GET /auth/google/callback`
**Description**: Callback URL for Google OAuth2. Handles the authorization code from Google, exchanges it for a token, retrieves user information, and creates a session.

#### `GET /auth/microsoft/callback`
**Description**: Callback URL for Microsoft OAuth2. Handles the authorization code from Microsoft, exchanges it for a token, retrieves user information, and creates a session.

### Core Operations
When authentication is enabled, the following endpoints are protected and require a valid user session.

#### `GET /`
**Description**: Serves the main web interface
- **Response**: HTML form with step-by-step wizard
- **Features**: Responsive design, real-time validation

#### `POST /connect`
**Description**: Establishes connection to SQL Server
- **Request Body**:
  ```json
  {
    "host": "server_address",
    "port": "1433",
    "user": "username",
    "password": "password",
    "instance": "instance",
    "encryption":"encryption",
    "trustServerCertificate": true
  }
  ```
- **Response (success)**:
  ```json
  {
    "status": "success",
    "code": 200,
    "message": "Connection done successfully",
    "data": {
        "server": "localhost"
    },
    "timestamp": "2025-07-16T10:43:43-03:00",
    "path": "/connect"
  }
  ```
- **Response (fail)**:
  ```json
  {
    "status": "error",
    "code": 500,
    "message": "Cannot connect to the server",
    "errors": {
        "connect": "TLS Handshake failed: tls: failed to verify certificate: x509: certificate signed by unknown authority"
    },
    "timestamp": "2025-07-16T10:43:43-03:00",
    "path": "/connect"
  }
  ```

#### `GET /databases`
**Description**: Retrieves all databases and their file information
- **Response (success)**:
  ```json
  {
    "status": "success",
    "code": 200,
    "message": "Databases collected successfully",
    "data": [
      {
        "id": "database_id",
        "name": "database_name",
        "files": [
          {
            "LogicalName": "logical_name",
            "PhysicalName": "physical_path",
            "FileType": "ROWS|LOG"
          }
        ]
      }
    ],
    "timestamp": "2025-07-16T10:48:48-03:00",
    "path": "/databases"    
  }
  ```
- **Response (fail)**:
  ```json
  {
    "status": "error",
    "code": 500,
    "message": "Cannot get databases",
    "errors": {
        "databases": "failed to send SQL Batch: write tcp 127.0.0.1:58411..."
    },
    "timestamp": "2025-07-16T10:50:42-03:00",
    "path": "/databases"
  }
  ```

#### `POST /backup`
**Description**: Performs concurrent backup operations
- **Request Body**:
  ```json
  {
    "databases": [
      {"name": "database1"},
      {"name": "database2"}
    ],
    "path": "/backup/directory/",
    "concurrentOpe": 4
  }
  ```
- **Response (success)**:
  ```json
  {
    "status": "success",
    "code": 200,
    "message": "Backup done successfully.",
    "data": {
      "backupDone": [
          {
            "name": "database1",
            "name": "database2"
          }
      ],
      "backupPath": "/backup/directory/",
      "totalBackup": 2,
      "totalTime": "0h0m1s"
    },
    "timestamp": "2025-07-16T10:52:18-03:00",
    "path": "/backup"
  }
  ```
- **Response (done with errors)**:
  ```json
  {
    "status": "error",
    "code": 207,
    "message": "Backup completed with errors.",
    "data": {
      "backupDone": [
          {
            "name": "database1",
          }
      ],
      "backupPath": "/backup/directory/",
      "totalBackup": 1,
      "totalTime": "0h0m1s"
    },
    "errors": {
        "backupErrors": [
            {
                "database": "database2",
                "error": "mssql: BACKUP DATABASE is being terminated abnormally."
            }
        ],
        "totalBackupErrors": 1
    },
    "timestamp": "2025-07-16T10:52:18-03:00",
    "path": "/backup"
  }
  ```
- **Response (error)**:
  ```json
  {
    "status": "error",
    "code": 207,
    "message": "Backup completed with errors.",
    "data": {
      "backupDone": [
          {
            "name": "database1",
          }
      ],
      "backupPath": "/backup/directory/",
      "totalBackup": 1,
      "totalTime": "0h0m1s"
    },
    "errors": {
        "backupErrors": [
            {
                "database": "database2",
                "error": "mssql: BACKUP DATABASE is being terminated abnormally."
            }
        ],
        "totalBackupErrors": 1
    },
    "timestamp": "2025-07-16T10:52:18-03:00",
    "path": "/backup"
  }
  ```

#### `POST /restore`
**Description**: Restores databases from backup files
- **Request Body**:
  ```json
  {
    "backupFilesPath": "/path/to/backup/files/",
    "concurrentOpe": 4
  }
  ```
- **Response (success)**:
  ```json
  {
    "status": "success",
    "code": 200,
    "message": "Restore done successfully.",
    "data": {
        "backupPath": "/path/to/backup/files/",
        "restoreDone": [
            {
                "backupPath": "/path/to/backup/files/database.bak",
                "database": {
                    "name": "database",
                    "files": [
                        {
                          "LogicalName": "logical_name",
                          "PhysicalName": "physical_path",
                          "FileType": "ROWS|LOG"
                        }
                    ]
                }
            }
        ],
        "totalRestore": 1,
        "totalTime": "0h0m1s"
    },
    "timestamp": "2025-07-16T13:58:46-03:00",
    "path": "/restore"
  }
  ```

- **Response (done with errors)**:
  ```json
  {
    "status": "error",
    "code": 207,
    "message": "Restore operation completed with errors",
    "errors": {
        "restoreErrors": [
            {
                "database": "database1",
                "error": "read tcp [::1]:42211-\u003e[::1]:1433: wsarecv: Cancellation of an existing connection was requested by the remote host."
            }
        ],
        "totalRestoreErrors": 1
    },
    "data": {
        "backupPath": "/path/to/backup/files/",
        "restoreDone": [
            {
                "backupPath": "/path/to/backup/files/database2.bak",
                "database": {
                    "name": "database2",
                    "files": [
                        {
                          "LogicalName": "logical_name",
                          "PhysicalName": "physical_path",
                          "FileType": "ROWS|LOG"
                        }
                    ]
                }
            }
        ],
        "totalRestore": 1,
        "totalTime": "0h2m8s"
    },
    "timestamp": "2025-07-16T14:09:11-03:00",
    "path": "/restore"
  }
  ```

- **Response (fail)**:
  ```json
  {
    "status": "success",
    "code": 500,
    "message": "No restore was completed.",
    "errors": {
        "restoreErrors": [
            {
                "database": "SVS_CPR_ALU_APARECIDA",
                "error": "mssql: RESTORE DATABASE está sendo encerrado de forma anormal."
            }
        ],
        "totalRestoreErrors": 1
    },
    "timestamp": "2025-07-16T13:58:46-03:00",
    "path": "/restore"
  }
  ```

## 🛠️ Building and Installation

### Prerequisites
- Go 1.23.2 or later
- SQL Server instance (local or remote)
- Web browser (Chrome, Firefox, Safari, Edge)

### Build Instructions

#### 1. Clone the Repository
```bash
git clone https://github.com/RenanMonteiroS/MaestroSQLWeb.git
cd MaestroSQLWeb
```

#### 2. Install Dependencies
```bash
go mod tidy
```

#### 3. Generate Assets
```bash
go generate
```

#### 4. Build the Application
```bash
# For current platform
go build -o MaestroSQL

# For Windows
GOOS=windows GOARCH=amd64 go build -o MaestroSQL.exe

# For Linux
GOOS=linux GOARCH=amd64 go build -o MaestroSQL

# For macOS
GOOS=darwin GOARCH=amd64 go build -o MaestroSQL
```

#### 5. Run the Application
```bash
# Direct execution
./MaestroSQL

# Or via Go
go run main.go
```

The application will:
1. Start the web server on port defined in `config/config.go`
2. Automatically open your default browser (if config.AppOpenOnceRunned is true)
3. Navigate to `http://localhost:8000`, or another value defined in `config/config.go` (config.AppHost and config.AppPort)

### Development Commands

```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Vet code for issues
go vet ./...

# Run with hot reload (using air)
air

# Build for all platforms
./build.sh  # Create this script for cross-compilation
```

## ⚙️ Configuration

All configuration is done in the `config/config.go` file.

| Parameter | Description |
| --- | --- |
| `AuthenticationMethods` | A list of enabled authentication methods (e.g., `[]string{"OSI", "OAUTH2GOOGLE"}`). |
| `AuthenticatorURL` | The URL of the external OSI authentication service. |
| `AppHost` | The host where the application will run. Use `0.0.0.0` to listen on all interfaces. |
| `AppPort` | The port where the application will run. |
| `AppOpenOnceRunned` | Open the browser automatically when the application starts. |
| `AppCertificateUsage` | Enable or disable HTTPS. |
| `AppCertificateLocation` | The path to the SSL certificate file. |
| `AppCertificateKeyLocation` | The path to the SSL key file. |
| `AppSessionSecret` | A secret key for the session cookie store. |
| `AppCSRFTokenUsage` | Enable or disable CSRF protection. |
| `AppCSRFCookieSecret` | A secret for the cookie used for CSRF token verification. |
| `AppCSRFTokenSecret` | A secret for the token used for CSRF token verification. |
| `CORSUsage` | Enable or disable CORS. |
| `CORSAllowOrigins` | A list of allowed origins for CORS. |
| `GoogleOAuth2ClientID` | Client ID for Google OAuth2. |
| `GoogleOAuth2ClientSecret` | Client Secret for Google OAuth2. |
| `GoogleOAuth2RedirectURL` | Redirect URL for Google OAuth2. |
| `MicrosoftOAuth2ClientID` | Client ID for Microsoft OAuth2. |
| `MicrosoftOAuth2ClientSecret` | Client Secret for Microsoft OAuth2. |
| `MicrosoftOAuth2RedirectURL` | Redirect URL for Microsoft OAuth2. |
| `MicrosoftOAuth2AzureADEndpoint` | Azure AD Endpoint for Microsoft OAuth2. |

## 📋 Usage Guide

### Step-by-Step Operation

#### 1. **Database Connection**
- Enter SQL Server host and port
- Provide authentication credentials
- Test connection before proceeding

#### 2. **Operation Selection**
- Choose between Backup or Restore
- Each operation has specific requirements and options

#### 3. **Database Selection** (Backup only)
- View all available databases
- Select multiple databases for batch operations
- Use "Select All" for convenience

#### 4. **Path Configuration**
- Specify backup destination (backup) or source (restore)
- Use absolute paths for reliability
- Ensure proper permissions

#### 5. **Execution**
- Review operation summary
- Confirm before execution
- Monitor progress through logs

### Authentication Flow (if enabled)

1.  The user navigates to the application.
2.  If a protected resource is accessed without a valid session, they are considered unauthenticated.
3.  The user can initiate login through the UI, which will redirect them to `/login?method=<provider>`.
4.  **For OAuth2 (Google/Microsoft):**
    - The user is redirected to the provider's login page.
    - After successful authentication, the provider redirects back to the application's callback URL (`/auth/google/callback` or `/auth/microsoft/callback`).
    - The application validates the response, retrieves user information, and creates a session, storing the user's email in a secure cookie.
5.  **For OSI:**
    - The user provides their credentials in a login form.
    - The application sends a `POST` request to `/login?method=osi`.
    - The application validates the credentials and creates a session.
6.  Once the session is created, the user can access the protected routes. The session is automatically verified by a middleware on each request.
7.  To log out, the user can access the `/logout` endpoint, which clears the session cookie.

### File Naming Convention

#### Backup Files
```
{database_name}={YYYY-MM-DD}_{HH-MM-SS}.bak
```
Example: `MyDatabase=2024-06-24_14-30-15.bak`

#### Log Files
- `backup.log`: Backup operation logs
- `restore.log`: Restore operation logs
- `app.log`: All app related logs

## 🔧 Advanced Features

### Concurrent Processing
- Backup and restore operations run in parallel using goroutines
- Configurable worker pools for different operation types
- Channel-based communication for result aggregation
- Proper resource cleanup and error handling

### Timeout Management
- Context-based timeouts for database operations
- Configurable timeout values per operation type
- Graceful cancellation of long-running operations

### Error Handling
- Comprehensive error logging with timestamps
- Partial success handling (some operations succeed, others fail)
- User-friendly error messages in web interface
- Detailed technical logs for troubleshooting

### Security Features
- Optional session-based authentication with OAuth2 and OSI support.
- MFA support through external authenticator (OSI).
- Secure credential handling (passwords not logged).
- HTTPS support with SSL/TLS certificates.
- CSRF and CORS protection.

## 📊 Monitoring and Logging

### Log Files Location
- Default: Current directory
- Configurable via environment variables
- Automatic log rotation recommended for production

### Log Formats
The application uses a structured logging format (JSON) for easy parsing and analysis.
```json
{"level":"info","time":"2024-07-01T14:30:15-03:00","message":"Backup related to [DatabaseName] database completed"}
{"level":"info","time":"2024-07-01T14:30:15-03:00","message":"Backups total: 5"}
{"level":"info","time":"2024-07-01T14:30:15-03:00","message":"Tempo time: 2m30s"}
{"level":"error","time":"2024-07-01T14:30:15-03:00","message":"Error: Connection timeout after 30 seconds"}
```

### Performance Metrics
- Total operation time
- Individual database processing time
- Success/failure rates
- Resource utilization

## 🚨 Troubleshooting

### Common Issues

#### Connection Problems
```bash
# Check SQL Server is running
sqlcmd -S server_name -U username -Q "SELECT @@VERSION"

# Verify network connectivity
telnet server_name 1433

```

#### Permission Issues
```sql
-- Grant necessary permissions to restore
ALTER SERVER ROLE dbcreator ADD MEMBER [username]

-- Grant necessary permissions to backup
USE db_example;
ALTER ROLE db_backupoperator ADD MEMBER [username];
```

#### File Path Issues
- Use absolute paths
- Ensure directory exists and is writable
- Check SQL Server service account permissions

## 🔐 Security Considerations

### Production Deployment
1. **Authentication**: Always enable in production
2. **Network Security**: Restrict access to necessary IPs
3. **File Permissions**: Secure backup directories
4. **Regular Updates**: Keep dependencies current

### Best Practices
- Use dedicated service account for SQL Server connections
- Implement backup encryption for sensitive data
- Regular security audits and penetration testing
- Monitor access logs for suspicious activity
- Implement backup retention policies

## 🤝 Contributing

### Development Setup
1. Fork the repository
2. Create feature branch
3. Follow Go coding standards
4. Add tests for new features
5. Update documentation
6. Submit pull request

### Code Style
- Follow `go fmt` formatting
- Use meaningful variable names
- Add comments for complex logic
- Implement proper error handling
- Write unit tests

## 📄 License

This project is licensed under CC BY-NC 4.0 - see the [LICENSE](LICENSE.md) file for details.

## 🆘 Support

For support and questions:
- **GitHub Issues**: [Project Issues](https://github.com/RenanMonteiroS/MaestroSQLWeb/issues)
- **Documentation**: This README

## 🏆 Acknowledgments

- **Bootstrap**: For the responsive UI framework
- **Gin Framework**: For the robust HTTP routing
- **Microsoft SQL Server**: For the database engine
- **Go**: For performance and reliability

### TODOs

- Login into the database server with instance, instead of port
- Choose every database name, for each .bak file in restore
- Built in authentication
- Disable the same connection pool for all operations, making all requests independent
- Support for MySQL and PostgreSQL
- Backup encryption
---

**MaestroSQL** - Simplifying SQL Server Database Management, One Operation at a Time! 🎼
By: [Renan Monteiro](https://www.linkedin.com/in/renan-monteiro-de-souza-946a06214)
