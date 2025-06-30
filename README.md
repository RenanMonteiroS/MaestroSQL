# MaestroSQL

**MaestroSQL** is a powerful, user-friendly web application designed to simplify SQL Server database backup and restore operations. Built with Go and featuring a modern web interface, it provides an intuitive way to manage database operations with authentication support and concurrent processing capabilities.

## 🎯 Objective

MaestroSQL addresses the complexity of SQL Server database management by providing:

- **Simplified Database Operations**: Easy-to-use web interface for backup and restore operations
- **Concurrent Processing**: High-performance parallel processing for multiple database operations
- **Security**: Optional JWT-based authentication with MFA support
- **Cross-Platform**: Works on Windows, macOS, and Linux
- **Real-time Monitoring**: Comprehensive logging and progress tracking

## 🚀 Features

### Core Features
- 🔐 **Optional Authentication**: JWT-based authentication with MFA support
- 📊 **Database Discovery**: Automatic detection and listing of SQL Server databases
- 💾 **Backup Operations**: Concurrent backup of multiple databases with timestamp naming
- 🔄 **Restore Operations**: Intelligent restore from .bak files with automatic path resolution
- 📝 **Comprehensive Logging**: Detailed operation logs for backup, restore, and error tracking
- 🎨 **Modern UI**: Bootstrap-based responsive web interface with step-by-step wizard

### Technical Features
- ⚡ **Concurrent Processing**: Goroutine-based parallel operations for better performance
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
- **Authentication**: Configurable JWT authentication
- **Timeouts**: Operation-specific timeout settings
- **Paths**: Automatic SQL Server path detection

## 📡 API Endpoints

### Authentication
All endpoints (except `/`) require valid JWT token in `Authorization` header when authentication is enabled.

### Core Operations

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
    "password": "password"
  }
  ```
- **Response**: Connection object or error details
- **Features**: Connection validation, credential verification

#### `GET /databases`
**Description**: Retrieves all databases and their file information
- **Response**:
  ```json
  [
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
  ]
  ```
- **Features**: Automatic file classification, metadata extraction

#### `POST /backup`
**Description**: Performs concurrent backup operations
- **Request Body**:
  ```json
  {
    "databases": [
      {"name": "database1"},
      {"name": "database2"}
    ],
    "path": "/backup/directory/"
  }
  ```
- **Response**: Backup completion status and any errors
- **Features**: Concurrent processing, timestamp naming, progress logging

#### `POST /restore`
**Description**: Restores databases from backup files
- **Request Body**:
  ```json
  {
    "backupFilesPath": "/path/to/backup/files/"
  }
  ```
- **Response**: Restore completion status and any errors
- **Features**: Automatic file discovery, path resolution, concurrent processing

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

#### 3. Build the Application
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

#### 4. Run the Application
```bash
# Direct execution
./MaestroSQL

# Or via Go
go run main.go
```

The application will:
1. Start the web server on port 8000
2. Automatically open your default browser
3. Navigate to `http://localhost:8000`

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

### Authentication Configuration
Edit `config/config.go`:

```go
const (
    AuthenticatorUsage = true                    // Enable/disable authentication
    AuthenticatorURL  = "http://localhost:8080" // External auth service URL
)
```

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

1. Click user icon in navigation
2. Enter email and password
3. Provide MFA token (if required)
4. Receive JWT token for session
5. Token automatically included in API calls

### File Naming Convention

#### Backup Files
```
{database_name}={YYYY-MM-DD}_{HH-MM-SS}.bak
```
Example: `MyDatabase=2024-06-24_14-30-15.bak`

#### Log Files
- `backup.log`: Backup operation logs
- `restore.log`: Restore operation logs
- `fatal.log`: Critical error logs

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
- Optional JWT-based authentication
- MFA support through external authenticator
- Secure credential handling (passwords not logged)
- HTTPS support (configure reverse proxy)

## 📊 Monitoring and Logging

### Log Files Location
- Default: Current directory
- Configurable via environment variables
- Automatic log rotation recommended for production

### Log Formats
```
# Backup/Restore logs
2024-06-24 14:30:15 - Backup related to [DatabaseName] database completed
2024-06-24 14:30:15 - Backups total: 5
2024-06-24 14:30:15 - Tempo time: 2m30s

# Error logs
2024-06-24 14:30:15 - Error: Connection timeout after 30 seconds
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

- Environment Variables: Permit to set these environment variables for advanced configuration:
  ```bash
  # Server configuration
  MAESTRO_PORT=8000
  MAESTRO_HOST=localhost

  # Database defaults
  MAESTRO_DB_TIMEOUT=600  # seconds
  MAESTRO_BACKUP_TIMEOUT=600
  MAESTRO_RESTORE_TIMEOUT=900

  # Logging
  MAESTRO_LOG_PATH=/var/log/maestrosql/
  ```

- Support for MySQL and PostgreSQL
- Backup encryption
- Support for pt-br
---

**MaestroSQL** - Simplifying SQL Server Database Management, One Operation at a Time! 🎼
By: [Renan Monteiro](https://www.linkedin.com/in/renan-monteiro-de-souza-946a06214)