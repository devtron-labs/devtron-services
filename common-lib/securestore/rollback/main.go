/*
 * Copyright (c) 2024. Devtron Inc.
 */

package main

import (
	"flag"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Command line flags
	var (
		//databaseName = flag.String("database", "orchestrator", "Database name to connect to")
		table    = flag.String("table", "all", "Table to rollback (cluster, gitops_config, docker_artifact_store, git_provider, remote_connection_config, all)")
		recordId = flag.String("id", "", "Specific record ID to rollback (optional)")
		validate = flag.Bool("validate", false, "Validate rollback results")
		help     = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// Set log level
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.Info("Starting cluster config rollback utility")

	// Create rollback service
	service, err := NewRollbackService()
	if err != nil {
		log.Fatalf("Failed to create rollback service: %v", err)
	}

	// Execute based on flags
	if *validate {
		err = executeValidation(service, *table)
		if err != nil {
			log.Fatalf("Validation failed: %v", err)
		}
		log.Info("Validation completed successfully")
		return
	}

	if *recordId != "" {
		// Rollback specific record
		err = executeSpecificRollback(service, *table, *recordId)
		if err != nil {
			log.Fatalf("Failed to rollback specific record: %v", err)
		}
		log.Info("Successfully rolled back specific record")
	} else {
		// Rollback table(s)
		err = executeTableRollback(service, *table)
		if err != nil {
			log.Fatalf("Failed to rollback: %v", err)
		}
		log.Info("Successfully completed rollback")
	}

	log.Info("Rollback utility completed")
}

func executeValidation(service *RollbackServiceImpl, table string) error {
	switch table {
	case "cluster":
		return service.ValidateRollback()
	case "gitops_config":
		return service.ValidateGitOpsConfigRollback()
	case "docker_artifact_store":
		return service.ValidateDockerArtifactStoreRollback()
	case "git_provider":
		return service.ValidateGitProviderRollback()
	case "remote_connection_config":
		return service.ValidateRemoteConnectionConfigRollback()
	case "all":
		return service.ValidateAllRollbacks()
	default:
		return fmt.Errorf("unsupported table: %s", table)
	}
}

func executeSpecificRollback(service *RollbackServiceImpl, table string, recordId string) error {
	switch table {
	case "cluster":
		id, err := strconv.Atoi(recordId)
		if err != nil {
			return fmt.Errorf("invalid cluster ID: %v", err)
		}
		return service.RollbackSpecificCluster(id)
	case "gitops_config":
		id, err := strconv.Atoi(recordId)
		if err != nil {
			return fmt.Errorf("invalid gitops config ID: %v", err)
		}
		return service.RollbackSpecificGitOpsConfig(id)
	case "docker_artifact_store":
		return service.RollbackSpecificDockerArtifactStore(recordId)
	case "git_provider":
		id, err := strconv.Atoi(recordId)
		if err != nil {
			return fmt.Errorf("invalid git provider ID: %v", err)
		}
		return service.RollbackSpecificGitProvider(id)
	case "remote_connection_config":
		id, err := strconv.Atoi(recordId)
		if err != nil {
			return fmt.Errorf("invalid remote connection config ID: %v", err)
		}
		return service.RollbackSpecificRemoteConnectionConfig(id)
	default:
		return fmt.Errorf("unsupported table for specific rollback: %s", table)
	}
}

func executeTableRollback(service *RollbackServiceImpl, table string) error {
	switch table {
	case "cluster":
		return service.RollbackEncryptedConfig()
	case "gitops_config":
		return service.RollbackGitOpsConfigEncryptedFields()
	case "docker_artifact_store":
		return service.RollbackDockerArtifactStoreEncryptedFields()
	case "git_provider":
		return service.RollbackGitProviderEncryptedFields()
	case "remote_connection_config":
		return service.RollbackRemoteConnectionConfigEncryptedFields()
	case "all":
		return service.RollbackAllTables()
	default:
		return fmt.Errorf("unsupported table: %s", table)
	}
}

func printHelp() {
	fmt.Println("Database Encryption Rollback Utility")
	fmt.Println("====================================")
	fmt.Println()
	fmt.Println("This utility reverts encrypted data in database tables by:")
	fmt.Println("1. Reading encrypted data from specified columns")
	fmt.Println("2. Decrypting it using the stored encryption key")
	fmt.Println("3. Storing the decrypted data back as plain text")
	fmt.Println()
	fmt.Println("Supported Tables:")
	fmt.Println("  - cluster (config column - EncryptedMap)")
	fmt.Println("  - gitops_config (token column - EncryptedString)")
	fmt.Println("  - docker_artifact_store (aws_secret_accesskey, password - EncryptedString)")
	fmt.Println("  - git_provider (password, ssh_private_key, access_token - EncryptedString)")
	fmt.Println("  - remote_connection_config (ssh_password, ssh_auth_key - EncryptedString)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run *.go [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -database string")
	fmt.Println("        Database name to connect to (default: orchestrator)")
	fmt.Println("  -table string")
	fmt.Println("        Table to rollback (cluster, gitops_config, docker_artifact_store,")
	fmt.Println("        git_provider, remote_connection_config, all) (default: all)")
	fmt.Println("  -id string")
	fmt.Println("        Specific record ID to rollback (optional)")
	fmt.Println("  -validate")
	fmt.Println("        Validate rollback results")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Rollback all tables")
	fmt.Println("  go run *.go")
	fmt.Println()
	fmt.Println("  # Rollback specific table")
	fmt.Println("  go run *.go -table=cluster")
	fmt.Println("  go run *.go -table=gitops_config")
	fmt.Println()
	fmt.Println("  # Rollback specific record")
	fmt.Println("  go run *.go -table=cluster -id=123")
	fmt.Println("  go run *.go -table=docker_artifact_store -id=abc-def-123")
	fmt.Println()
	fmt.Println("  # Validate rollback results")
	fmt.Println("  go run *.go -validate")
	fmt.Println("  go run *.go -table=cluster -validate")
	fmt.Println()
	fmt.Println("  # Use different database")
	fmt.Println("  go run *.go -database=mydb")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  PG_ADDR     - PostgreSQL address (default: 127.0.0.1)")
	fmt.Println("  PG_PORT     - PostgreSQL port (default: 5432)")
	fmt.Println("  PG_USER     - PostgreSQL username")
	fmt.Println("  PG_PASSWORD - PostgreSQL password")
	fmt.Println("  PG_DATABASE - PostgreSQL database (overridden by -database flag)")
	fmt.Println()
}
