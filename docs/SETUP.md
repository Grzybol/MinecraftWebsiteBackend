# Setup Instructions

## Configuration

### 1. Database Configuration

The application uses MongoDB and MariaDB. Configure the connection strings in `config.go`:

```go
const (
    DefaultMongoURI  = "mongodb://username:password@localhost:27017"
    DefaultMariaBase = "username:password@tcp(127.0.0.1:3306)/"
    DatabaseName     = "your_database_name"
)
```

### 2. Server Tokens

Configure server authentication tokens in `config.go`:

```go
var ServerTokens = map[string]string{
    "boxpvp":   "your_boxpvp_token_here",
    "survival": "your_survival_token_here", 
    "skygen":   "your_skygen_token_here",
    "global":   "your_global_token_here",
}
```

### 3. Environment Variables

You can override configuration using environment variables:

- `MONGO_URI` - MongoDB connection string
- `MARIADB_BASE_URI` - MariaDB connection string
- `BARCODE_CACHE_TTL_HOURS` - Barcode cache TTL in hours (default: 24)

### 4. Socket Paths

Configure socket paths for equipment and eco balance:

```go
const (
    DefaultBackpackSock = "/tmp/equipment_backpack.sock"
    DefaultArmorSock    = "/tmp/equipment_armor.sock"
    DefaultHotbarSock   = "/tmp/equipment_hotbar.sock"
)
```

### 5. Backend Authentication

Set the backend authentication token:

```go
const BackendAuthToken = "your_backend_token_here"
```

## Security Notes

- Never commit `config.go` to version control
- Use environment variables for sensitive data in production
- Keep server tokens secure and rotate them regularly
- The `config.go` file is already added to `.gitignore`

## Installation

1. Copy `config.go` and replace placeholder values with real configuration
2. Ensure all required databases are running
3. Set up Unix sockets for Minecraft server communication
4. Run the application with `go run .` 