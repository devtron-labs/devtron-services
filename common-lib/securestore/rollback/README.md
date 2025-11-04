# Database Encryption Rollback Utility

This utility provides functionality to revert encrypted data in multiple database tables back to plain text format.

## Overview

The rollback utility performs the following operations:
1. Reads encrypted data from specified columns in supported tables
2. Decrypts the data using the stored encryption key from the `attributes` table
3. Stores the decrypted data back as plain text in the same columns

## Supported Tables

| Table | Encrypted Columns | Data Type |
|-------|------------------|-----------|
| `cluster` | `config` | EncryptedMap |
| `gitops_config` | `token` | EncryptedString |
| `docker_artifact_store` | `aws_secret_accesskey`, `password` | EncryptedString |
| `git_provider` | `password`, `ssh_private_key`, `access_token` | EncryptedString |
| `remote_connection_config` | `ssh_password`, `ssh_auth_key` | EncryptedString |

## Files

- `rollback_service.go` - Core rollback service with methods for all supported tables
- `main.go` - Command-line executable for running rollback operations
- `rollback_service_test.go` - Unit tests
- `run_rollback.sh` - Convenient shell script with safety checks
- `README.md` - This documentation file

## Usage

### Prerequisites

1. Ensure the encryption key exists in the `attributes` table
2. Set up the required environment variables for database connection
3. Have appropriate database permissions to read and update the cluster table

### Environment Variables

```bash
export PG_ADDR="127.0.0.1"          # PostgreSQL address
export PG_PORT="5432"               # PostgreSQL port
export PG_USER="your_username"      # PostgreSQL username
export PG_PASSWORD="your_password"  # PostgreSQL password
export PG_DATABASE="orchestrator"   # PostgreSQL database name
```

### Running the Utility

#### Rollback All Tables

```bash
cd common-lib/securestore/rollback
go run *.go
```

#### Rollback Specific Table

```bash
go run *.go -table=cluster
go run *.go -table=gitops_config
go run *.go -table=docker_artifact_store
go run *.go -table=git_provider
go run *.go -table=remote_connection_config
```

#### Rollback Specific Record

```bash
go run *.go -table=cluster -id=123
go run *.go -table=docker_artifact_store -id=abc-def-123
go run *.go -table=gitops_config -id=456
```

#### Validate Rollback Results

```bash
go run *.go -validate                    # Validate all tables
go run *.go -table=cluster -validate     # Validate specific table
```

#### Use Different Database

```bash
go run *.go -database=mydb
```

#### Show Help

```bash
go run *.go -help
```

### Command Line Options

- `-database string` - Database name to connect to (default: "orchestrator")
- `-table string` - Table to rollback (cluster, gitops_config, docker_artifact_store, git_provider, remote_connection_config, all) (default: "all")
- `-id string` - Specific record ID to rollback (optional)
- `-validate` - Validate rollback results
- `-help` - Show help message

### Using the Shell Script

The shell script provides additional safety features:

```bash
# Interactive rollback with confirmation
./run_rollback.sh

# Rollback specific table
./run_rollback.sh -t cluster

# Rollback specific record
./run_rollback.sh -t cluster -i 123

# Validate results
./run_rollback.sh -v

# Dry run (show what would be executed)
./run_rollback.sh --dry-run
```

## Code Structure

The rollback utility is implemented as a single Go package with the following files:

- `rollback_service.go` - Core rollback service implementation for all tables
- `main.go` - Command-line interface and main function
- `rollback_service_test.go` - Unit tests
- `go.mod` - Go module definition

All files are in the same `package main` to create a single executable.

## How It Works

### Encryption Detection

The utility uses the `EncryptedMap.Scan()` method to detect if data is encrypted:
- If the data can be successfully scanned as an `EncryptedMap`, it's considered encrypted
- If scanning fails, the data is assumed to be already in plain text format

### Decryption Process

1. The utility loads the encryption key from the `attributes` table
2. For each cluster with config data:
   - Attempts to scan the config as an `EncryptedMap`
   - If successful, the `Scan()` method automatically decrypts the data
   - The decrypted data is then marshaled to JSON and stored back

### Safety Features

- **Non-destructive**: If data is already in plain text, it's left unchanged
- **Validation**: Provides validation functionality to verify rollback success
- **Logging**: Comprehensive logging for monitoring progress and debugging
- **Error handling**: Continues processing other clusters even if one fails

## Error Handling

The utility handles various error scenarios:
- Database connection failures
- Missing encryption keys
- Invalid encrypted data
- JSON marshaling errors
- Database update failures

Each error is logged with appropriate context, and the utility continues processing remaining clusters.

## Validation

The validation feature checks that all cluster configs are valid JSON after rollback:
- Attempts to unmarshal each config as JSON
- Reports any configs that are not valid JSON
- Provides summary of validation results

## Security Considerations

- The utility requires access to the encryption key stored in the database
- Ensure proper database permissions are in place
- Consider backing up the database before running the rollback
- The rollback operation converts encrypted data to plain text permanently

## Troubleshooting

### Common Issues

1. **"encryption key not found"**
   - Ensure the encryption key exists in the attributes table
   - Run the encryption key setup if needed

2. **Database connection errors**
   - Verify environment variables are set correctly
   - Check database connectivity and permissions

3. **"Failed to scan as encrypted data"**
   - This is usually normal and indicates data is already in plain text
   - Check logs to confirm the data is being handled correctly

### Logging

The utility provides detailed logging at different levels:
- `INFO`: General progress and summary information
- `WARN`: Non-critical issues (e.g., data already in plain text)
- `ERROR`: Critical errors that prevent processing
- `DEBUG`: Detailed information for troubleshooting (when enabled)
