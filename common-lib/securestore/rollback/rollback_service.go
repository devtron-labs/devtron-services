/*
 * Copyright (c) 2024. Devtron Inc.
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/securestore"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
	log "github.com/sirupsen/logrus"
)

// AuditLog contains audit fields
type AuditLog struct {
	CreatedOn time.Time `sql:"created_on,type:timestamptz"`
	CreatedBy int32     `sql:"created_by"`
	UpdatedOn time.Time `sql:"updated_on,type:timestamptz"`
	UpdatedBy int32     `sql:"updated_by"`
}

// Cluster represents the cluster table structure for rollback operations
type Cluster struct {
	tableName struct{} `sql:"cluster" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Config    string   `sql:"config"` // Using string instead of EncryptedMap for direct manipulation
	AuditLog
}

// GitOpsConfig represents the gitops_config table structure for rollback operations
type GitOpsConfig struct {
	tableName struct{} `sql:"gitops_config" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Token     string   `sql:"token"` // Using string instead of EncryptedString for direct manipulation
	AuditLog
}

// DockerArtifactStore represents the docker_artifact_store table structure for rollback operations
type DockerArtifactStore struct {
	tableName          struct{} `sql:"docker_artifact_store" pg:",discard_unknown_columns"`
	Id                 string   `sql:"id,pk"`
	AwsSecretAccessKey string   `sql:"aws_secret_accesskey"` // Using string instead of EncryptedString
	Password           string   `sql:"password"`             // Using string instead of EncryptedString
	AuditLog
}

// GitProvider represents the git_provider table structure for rollback operations
type GitProvider struct {
	tableName     struct{} `sql:"git_provider" pg:",discard_unknown_columns"`
	Id            int      `sql:"id,pk"`
	Password      string   `sql:"password"`        // Using string instead of EncryptedString
	SshPrivateKey string   `sql:"ssh_private_key"` // Using string instead of EncryptedString
	AccessToken   string   `sql:"access_token"`    // Using string instead of EncryptedString
	AuditLog
}

// RemoteConnectionConfig represents the remote_connection_config table structure for rollback operations
type RemoteConnectionConfig struct {
	tableName   struct{} `sql:"remote_connection_config" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	SshPassword string   `sql:"ssh_password"` // Using string instead of EncryptedString
	SshAuthKey  string   `sql:"ssh_auth_key"` // Using string instead of EncryptedString
	AuditLog
}

// Attributes represents the attributes table for encryption key storage
type Attributes struct {
	tableName struct{} `sql:"attributes" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Key       string   `sql:"key,notnull"`
	Value     string   `sql:"value,notnull"`
	Active    bool     `sql:"active, notnull"`
	AuditLog
}

// Constants
const (
	ENCRYPTED_KEY  = "encrypted_data"
	ENCRYPTION_KEY = "encryptionKey"
)

type RollbackService interface {
	// Cluster table methods
	RollbackEncryptedConfig() error
	RollbackSpecificCluster(clusterId int) error
	ValidateRollback() error

	// GitOps Config table methods
	RollbackGitOpsConfigEncryptedFields() error
	RollbackSpecificGitOpsConfig(id int) error
	ValidateGitOpsConfigRollback() error

	// Docker Artifact Store table methods
	RollbackDockerArtifactStoreEncryptedFields() error
	RollbackSpecificDockerArtifactStore(id string) error
	ValidateDockerArtifactStoreRollback() error

	// Git Provider table methods
	RollbackGitProviderEncryptedFields() error
	RollbackSpecificGitProvider(id int) error
	ValidateGitProviderRollback() error

	// Remote Connection Config table methods
	RollbackRemoteConnectionConfigEncryptedFields() error
	RollbackSpecificRemoteConnectionConfig(id int) error
	ValidateRemoteConnectionConfigRollback() error

	// Combined operations
	RollbackAllTables() error
	ValidateAllRollbacks() error
}

type RollbackServiceImpl struct {
	dbConnection *pg.DB
}

func NewRollbackService() (*RollbackServiceImpl, error) {
	dbConn, err := newDbConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}
	return &RollbackServiceImpl{
		dbConnection: dbConn,
	}, nil
}

// ============================================================================
// CLUSTER TABLE ROLLBACK METHODS
// ============================================================================

// RollbackClusterEncryptedConfig reads encrypted config data from cluster table, decrypts it and stores it back as plain text
func (impl *RollbackServiceImpl) RollbackEncryptedConfig() error {
	log.Info("Starting rollback of encrypted config in cluster table")

	// Initialize encryption key
	err := securestore.SetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to set encryption key: %w", err)
	}

	// Get all clusters with non-empty config
	var clusters []Cluster
	err = impl.dbConnection.Model(&clusters).
		Where("config IS NOT NULL").
		Select()
	if err != nil {
		return fmt.Errorf("failed to fetch clusters: %w", err)
	}

	log.Infof("Found %d clusters with config data to rollback", len(clusters))

	successCount := 0
	errorCount := 0

	for _, cluster := range clusters {
		err := impl.rollbackSingleCluster(&cluster)
		if err != nil {
			log.Errorf("Failed to rollback cluster ID %d: %v", cluster.Id, err)
			errorCount++
		} else {
			log.Infof("Successfully rolled back cluster ID %d", cluster.Id)
			successCount++
		}
	}

	log.Infof("Rollback completed. Success: %d, Errors: %d", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("rollback completed with %d errors out of %d clusters", errorCount, len(clusters))
	}

	return nil
}

func (impl *RollbackServiceImpl) rollbackSingleCluster(cluster *Cluster) error {
	// Try to parse the config as encrypted data
	var encryptedMap securestore.EncryptedMap
	err := encryptedMap.Scan(cluster.Config)
	if err != nil {
		// If scanning fails, it might already be plain text
		log.Warnf("Cluster ID %d: Failed to scan as encrypted data, might already be plain text: %v", cluster.Id, err)
		return nil
	}

	// Convert EncryptedMap to regular map[string]string for JSON marshaling
	plainMap := make(map[string]string)
	for k, v := range encryptedMap {
		plainMap[k] = v
	}

	// Marshal the decrypted data to JSON
	plainTextJSON, err := json.Marshal(plainMap)
	if err != nil {
		return fmt.Errorf("failed to marshal decrypted config to JSON: %w", err)
	}

	// Update the cluster with plain text config
	_, err = impl.dbConnection.Model(&Cluster{}).
		Set("config = ?", string(plainTextJSON)).
		Where("id = ?", cluster.Id).
		Update()
	if err != nil {
		return fmt.Errorf("failed to update cluster config: %w", err)
	}

	return nil
}

// RollbackSpecificCluster rolls back encryption for a specific cluster by ID
func (impl *RollbackServiceImpl) RollbackSpecificCluster(clusterId int) error {
	log.Infof("Starting rollback for specific cluster ID: %d", clusterId)

	// Initialize encryption key
	err := securestore.SetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to set encryption key: %w", err)
	}

	// Get the specific cluster
	var cluster Cluster
	err = impl.dbConnection.Model(&cluster).
		Where("id = ?", clusterId).
		Where("config IS NOT NULL").
		Where("config != ''").
		Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return fmt.Errorf("cluster with ID %d not found or has no config", clusterId)
		}
		return fmt.Errorf("failed to fetch cluster: %w", err)
	}

	err = impl.rollbackSingleCluster(&cluster)
	if err != nil {
		return fmt.Errorf("failed to rollback cluster ID %d: %w", clusterId, err)
	}

	log.Infof("Successfully rolled back cluster ID %d", clusterId)
	return nil
}

// ValidateClusterRollback validates that the rollback was successful by checking if data can be read as plain text
func (impl *RollbackServiceImpl) ValidateRollback() error {
	log.Info("Validating rollback results")

	var clusters []Cluster
	err := impl.dbConnection.Model(&clusters).
		Where("config IS NOT NULL").
		Where("config != ''").
		Select()
	if err != nil {
		return fmt.Errorf("failed to fetch clusters for validation: %w", err)
	}

	validCount := 0
	invalidCount := 0

	for _, cluster := range clusters {
		// Try to parse as plain JSON
		var plainMap map[string]string
		err := json.Unmarshal([]byte(cluster.Config), &plainMap)
		if err != nil {
			log.Errorf("Cluster ID %d: Config is not valid JSON: %v", cluster.Id, err)
			invalidCount++
		} else {
			log.Debugf("Cluster ID %d: Config is valid plain text JSON", cluster.Id)
			validCount++
		}
	}

	log.Infof("Validation completed. Valid: %d, Invalid: %d", validCount, invalidCount)

	if invalidCount > 0 {
		return fmt.Errorf("validation failed: %d clusters have invalid config data", invalidCount)
	}

	return nil
}

// Database connection configuration and utilities
type config struct {
	Addr            string `env:"PG_ADDR" envDefault:"127.0.0.1"`
	Port            string `env:"PG_PORT" envDefault:"5432"`
	User            string `env:"PG_USER" envDefault:"postgres"`
	Password        string `env:"PG_PASSWORD" envDefault:"" secretData:"-"`
	Database        string `env:"PG_DATABASE" envDefault:"orchestrator"`
	ApplicationName string `env:"APP" envDefault:"orchestrator"`
	LocalDev        bool   `env:"RUNTIME_CONFIG_LOCAL_DEV" envDefault:"false"`
}

func getDbConfig() (*config, error) {
	cfg := &config{}
	err := env.Parse(cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, err
}

func newDbConnection() (*pg.DB, error) {
	cfg, err := getDbConfig()
	if err != nil {
		return nil, err
	}
	options := pg.Options{
		Addr:            cfg.Addr + ":" + cfg.Port,
		User:            cfg.User,
		Password:        cfg.Password,
		Database:        cfg.Database,
		ApplicationName: cfg.ApplicationName,
	}
	dbConnection := pg.Connect(&options)
	//check db connection
	var test string
	_, err = dbConnection.QueryOne(&test, `SELECT 1`)

	if err != nil {
		log.Errorf("error in connecting to database %s: %v", cfg.Database, err)
		return nil, err
	} else {
		log.Infof("connected with database %s", cfg.Database)
	}
	return dbConnection, err
}

// ============================================================================
// GITOPS CONFIG TABLE ROLLBACK METHODS
// ============================================================================

// RollbackGitOpsConfigEncryptedFields reads encrypted token data from gitops_config table, decrypts it and stores it back as plain text
func (impl *RollbackServiceImpl) RollbackGitOpsConfigEncryptedFields() error {
	log.Info("Starting rollback of encrypted fields in gitops_config table")

	// Initialize encryption key
	err := securestore.SetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to set encryption key: %w", err)
	}

	var configs []GitOpsConfig
	err = impl.dbConnection.Model(&configs).
		Where("token IS NOT NULL").
		Where("token != ''").
		Select()

	if err != nil {
		return fmt.Errorf("failed to fetch gitops configs: %w", err)
	}

	log.Infof("Found %d gitops configs to process", len(configs))

	successCount := 0
	errorCount := 0

	for _, config := range configs {
		err = impl.rollbackSingleGitOpsConfig(&config)
		if err != nil {
			log.Errorf("Failed to rollback gitops config ID %d: %v", config.Id, err)
			errorCount++
			continue
		}
		successCount++
	}

	log.Infof("Rollback completed. Success: %d, Errors: %d", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("rollback completed with %d errors out of %d total configs", errorCount, len(configs))
	}

	return nil
}

func (impl *RollbackServiceImpl) rollbackSingleGitOpsConfig(config *GitOpsConfig) error {
	// Try to decrypt the token
	var encryptedString securestore.EncryptedString
	err := encryptedString.Scan(config.Token)
	if err != nil {
		// If scanning fails, it might already be plain text
		log.Warnf("GitOps config ID %d: token appears to be already decrypted or invalid: %v", config.Id, err)
		return nil
	}

	// Convert decrypted data to plain text
	plainText := string(encryptedString)

	// Update the database with plain text
	_, err = impl.dbConnection.Model(config).
		Set("token = ?", plainText).
		Where("id = ?", config.Id).
		Update()

	if err != nil {
		return fmt.Errorf("failed to update gitops config ID %d: %w", config.Id, err)
	}

	log.Infof("Successfully rolled back gitops config ID %d", config.Id)
	return nil
}

// RollbackSpecificGitOpsConfig rolls back encryption for a specific gitops config by ID
func (impl *RollbackServiceImpl) RollbackSpecificGitOpsConfig(id int) error {
	log.Infof("Starting rollback for specific gitops config ID: %d", id)

	// Initialize encryption key
	err := securestore.SetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to set encryption key: %w", err)
	}

	var config GitOpsConfig
	err = impl.dbConnection.Model(&config).Where("id = ?", id).Select()
	if err != nil {
		return fmt.Errorf("failed to fetch gitops config ID %d: %w", id, err)
	}

	err = impl.rollbackSingleGitOpsConfig(&config)
	if err != nil {
		return fmt.Errorf("failed to rollback gitops config ID %d: %w", id, err)
	}

	log.Infof("Successfully rolled back gitops config ID %d", id)
	return nil
}

// ValidateGitOpsConfigRollback validates that the rollback was successful
func (impl *RollbackServiceImpl) ValidateGitOpsConfigRollback() error {
	log.Info("Validating gitops config rollback results")

	var configs []GitOpsConfig
	err := impl.dbConnection.Model(&configs).
		Where("token IS NOT NULL").
		Where("token != ''").
		Select()

	if err != nil {
		return fmt.Errorf("failed to fetch gitops configs for validation: %w", err)
	}

	log.Infof("Validating %d gitops configs", len(configs))

	for _, config := range configs {
		// Try to parse as encrypted data - this should fail if rollback was successful
		var encryptedString securestore.EncryptedString
		err := encryptedString.Scan(config.Token)
		if err == nil && string(encryptedString) != config.Token {
			log.Warnf("GitOps config ID %d: token appears to still be encrypted", config.Id)
		} else {
			log.Infof("GitOps config ID %d: token appears to be plain text (rollback successful)", config.Id)
		}
	}

	log.Info("GitOps config validation completed")
	return nil
}

// ============================================================================
// DOCKER ARTIFACT STORE TABLE ROLLBACK METHODS
// ============================================================================

// RollbackDockerArtifactStoreEncryptedFields reads encrypted fields from docker_artifact_store table, decrypts them and stores them back as plain text
func (impl *RollbackServiceImpl) RollbackDockerArtifactStoreEncryptedFields() error {
	log.Info("Starting rollback of encrypted fields in docker_artifact_store table")

	// Initialize encryption key
	err := securestore.SetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to set encryption key: %w", err)
	}

	var stores []DockerArtifactStore
	err = impl.dbConnection.Model(&stores).
		Where("aws_secret_accesskey IS NOT NULL OR password IS NOT NULL").
		Select()

	if err != nil {
		return fmt.Errorf("failed to fetch docker artifact stores: %w", err)
	}

	log.Infof("Found %d docker artifact stores to process", len(stores))

	successCount := 0
	errorCount := 0

	for _, store := range stores {
		err = impl.rollbackSingleDockerArtifactStore(&store)
		if err != nil {
			log.Errorf("Failed to rollback docker artifact store ID %s: %v", store.Id, err)
			errorCount++
			continue
		}
		successCount++
	}

	log.Infof("Rollback completed. Success: %d, Errors: %d", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("rollback completed with %d errors out of %d total stores", errorCount, len(stores))
	}

	return nil
}

func (impl *RollbackServiceImpl) rollbackSingleDockerArtifactStore(store *DockerArtifactStore) error {
	updateFields := make(map[string]interface{})

	// Process aws_secret_accesskey field
	if store.AwsSecretAccessKey != "" {
		var encryptedString securestore.EncryptedString
		err := encryptedString.Scan(store.AwsSecretAccessKey)
		if err != nil {
			log.Warnf("Docker store ID %s: aws_secret_accesskey appears to be already decrypted: %v", store.Id, err)
		} else {
			updateFields["aws_secret_accesskey"] = string(encryptedString)
		}
	}

	// Process password field
	if store.Password != "" {
		var encryptedString securestore.EncryptedString
		err := encryptedString.Scan(store.Password)
		if err != nil {
			log.Warnf("Docker store ID %s: password appears to be already decrypted: %v", store.Id, err)
		} else {
			updateFields["password"] = string(encryptedString)
		}
	}

	// Update the database if we have fields to update
	if len(updateFields) > 0 {
		query := impl.dbConnection.Model(store).Where("id = ?", store.Id)
		for field, value := range updateFields {
			query = query.Set(field+" = ?", value)
		}
		_, err := query.Update()
		if err != nil {
			return fmt.Errorf("failed to update docker artifact store ID %s: %w", store.Id, err)
		}
		log.Infof("Successfully rolled back docker artifact store ID %s", store.Id)
	} else {
		log.Infof("Docker artifact store ID %s: no encrypted fields found to rollback", store.Id)
	}

	return nil
}

// RollbackSpecificDockerArtifactStore rolls back encryption for a specific docker artifact store by ID
func (impl *RollbackServiceImpl) RollbackSpecificDockerArtifactStore(id string) error {
	log.Infof("Starting rollback for specific docker artifact store ID: %s", id)

	// Initialize encryption key
	err := securestore.SetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to set encryption key: %w", err)
	}

	var store DockerArtifactStore
	err = impl.dbConnection.Model(&store).Where("id = ?", id).Select()
	if err != nil {
		return fmt.Errorf("failed to fetch docker artifact store ID %s: %w", id, err)
	}

	err = impl.rollbackSingleDockerArtifactStore(&store)
	if err != nil {
		return fmt.Errorf("failed to rollback docker artifact store ID %s: %w", id, err)
	}

	log.Infof("Successfully rolled back docker artifact store ID %s", id)
	return nil
}

// ValidateDockerArtifactStoreRollback validates that the rollback was successful
func (impl *RollbackServiceImpl) ValidateDockerArtifactStoreRollback() error {
	log.Info("Validating docker artifact store rollback results")

	var stores []DockerArtifactStore
	err := impl.dbConnection.Model(&stores).
		Where("aws_secret_accesskey IS NOT NULL OR password IS NOT NULL").
		Select()

	if err != nil {
		return fmt.Errorf("failed to fetch docker artifact stores for validation: %w", err)
	}

	log.Infof("Validating %d docker artifact stores", len(stores))

	for _, store := range stores {
		// Check aws_secret_accesskey field
		if store.AwsSecretAccessKey != "" {
			var encryptedString securestore.EncryptedString
			err := encryptedString.Scan(store.AwsSecretAccessKey)
			if err == nil && string(encryptedString) != store.AwsSecretAccessKey {
				log.Warnf("Docker store ID %s: aws_secret_accesskey appears to still be encrypted", store.Id)
			}
		}

		// Check password field
		if store.Password != "" {
			var encryptedString securestore.EncryptedString
			err := encryptedString.Scan(store.Password)
			if err == nil && string(encryptedString) != store.Password {
				log.Warnf("Docker store ID %s: password appears to still be encrypted", store.Id)
			}
		}
	}

	log.Info("Docker artifact store validation completed")
	return nil
}

// ============================================================================
// GIT PROVIDER TABLE ROLLBACK METHODS
// ============================================================================

// RollbackGitProviderEncryptedFields reads encrypted fields from git_provider table, decrypts them and stores them back as plain text
func (impl *RollbackServiceImpl) RollbackGitProviderEncryptedFields() error {
	log.Info("Starting rollback of encrypted fields in git_provider table")

	// Initialize encryption key
	err := securestore.SetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to set encryption key: %w", err)
	}

	var providers []GitProvider
	err = impl.dbConnection.Model(&providers).
		Where("password IS NOT NULL OR ssh_private_key IS NOT NULL OR access_token IS NOT NULL").
		Select()

	if err != nil {
		return fmt.Errorf("failed to fetch git providers: %w", err)
	}

	log.Infof("Found %d git providers to process", len(providers))

	successCount := 0
	errorCount := 0

	for _, provider := range providers {
		err = impl.rollbackSingleGitProvider(&provider)
		if err != nil {
			log.Errorf("Failed to rollback git provider ID %d: %v", provider.Id, err)
			errorCount++
			continue
		}
		successCount++
	}

	log.Infof("Rollback completed. Success: %d, Errors: %d", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("rollback completed with %d errors out of %d total providers", errorCount, len(providers))
	}

	return nil
}

func (impl *RollbackServiceImpl) rollbackSingleGitProvider(provider *GitProvider) error {
	updateFields := make(map[string]interface{})

	// Process password field
	if provider.Password != "" {
		var encryptedString securestore.EncryptedString
		err := encryptedString.Scan(provider.Password)
		if err != nil {
			log.Warnf("Git provider ID %d: password appears to be already decrypted: %v", provider.Id, err)
		} else {
			updateFields["password"] = string(encryptedString)
		}
	}

	// Process ssh_private_key field
	if provider.SshPrivateKey != "" {
		var encryptedString securestore.EncryptedString
		err := encryptedString.Scan(provider.SshPrivateKey)
		if err != nil {
			log.Warnf("Git provider ID %d: ssh_private_key appears to be already decrypted: %v", provider.Id, err)
		} else {
			updateFields["ssh_private_key"] = string(encryptedString)
		}
	}

	// Process access_token field
	if provider.AccessToken != "" {
		var encryptedString securestore.EncryptedString
		err := encryptedString.Scan(provider.AccessToken)
		if err != nil {
			log.Warnf("Git provider ID %d: access_token appears to be already decrypted: %v", provider.Id, err)
		} else {
			updateFields["access_token"] = string(encryptedString)
		}
	}

	// Update the database if we have fields to update
	if len(updateFields) > 0 {
		query := impl.dbConnection.Model(provider).Where("id = ?", provider.Id)
		for field, value := range updateFields {
			query = query.Set(field+" = ?", value)
		}
		_, err := query.Update()
		if err != nil {
			return fmt.Errorf("failed to update git provider ID %d: %w", provider.Id, err)
		}
		log.Infof("Successfully rolled back git provider ID %d", provider.Id)
	} else {
		log.Infof("Git provider ID %d: no encrypted fields found to rollback", provider.Id)
	}

	return nil
}

// RollbackSpecificGitProvider rolls back encryption for a specific git provider by ID
func (impl *RollbackServiceImpl) RollbackSpecificGitProvider(id int) error {
	log.Infof("Starting rollback for specific git provider ID: %d", id)

	// Initialize encryption key
	err := securestore.SetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to set encryption key: %w", err)
	}

	var provider GitProvider
	err = impl.dbConnection.Model(&provider).Where("id = ?", id).Select()
	if err != nil {
		return fmt.Errorf("failed to fetch git provider ID %d: %w", id, err)
	}

	err = impl.rollbackSingleGitProvider(&provider)
	if err != nil {
		return fmt.Errorf("failed to rollback git provider ID %d: %w", id, err)
	}

	log.Infof("Successfully rolled back git provider ID %d", id)
	return nil
}

// ValidateGitProviderRollback validates that the rollback was successful
func (impl *RollbackServiceImpl) ValidateGitProviderRollback() error {
	log.Info("Validating git provider rollback results")

	var providers []GitProvider
	err := impl.dbConnection.Model(&providers).
		Where("password IS NOT NULL OR ssh_private_key IS NOT NULL OR access_token IS NOT NULL").
		Select()

	if err != nil {
		return fmt.Errorf("failed to fetch git providers for validation: %w", err)
	}

	log.Infof("Validating %d git providers", len(providers))

	for _, provider := range providers {
		// Check password field
		if provider.Password != "" {
			var encryptedString securestore.EncryptedString
			err := encryptedString.Scan(provider.Password)
			if err == nil && string(encryptedString) != provider.Password {
				log.Warnf("Git provider ID %d: password appears to still be encrypted", provider.Id)
			}
		}

		// Check ssh_private_key field
		if provider.SshPrivateKey != "" {
			var encryptedString securestore.EncryptedString
			err := encryptedString.Scan(provider.SshPrivateKey)
			if err == nil && string(encryptedString) != provider.SshPrivateKey {
				log.Warnf("Git provider ID %d: ssh_private_key appears to still be encrypted", provider.Id)
			}
		}

		// Check access_token field
		if provider.AccessToken != "" {
			var encryptedString securestore.EncryptedString
			err := encryptedString.Scan(provider.AccessToken)
			if err == nil && string(encryptedString) != provider.AccessToken {
				log.Warnf("Git provider ID %d: access_token appears to still be encrypted", provider.Id)
			}
		}
	}

	log.Info("Git provider validation completed")
	return nil
}

// ============================================================================
// REMOTE CONNECTION CONFIG TABLE ROLLBACK METHODS
// ============================================================================

// RollbackRemoteConnectionConfigEncryptedFields reads encrypted fields from remote_connection_config table, decrypts them and stores them back as plain text
func (impl *RollbackServiceImpl) RollbackRemoteConnectionConfigEncryptedFields() error {
	log.Info("Starting rollback of encrypted fields in remote_connection_config table")

	// Initialize encryption key
	err := securestore.SetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to set encryption key: %w", err)
	}

	var configs []RemoteConnectionConfig
	err = impl.dbConnection.Model(&configs).
		Where("ssh_password IS NOT NULL OR ssh_auth_key IS NOT NULL").
		Select()

	if err != nil {
		return fmt.Errorf("failed to fetch remote connection configs: %w", err)
	}

	log.Infof("Found %d remote connection configs to process", len(configs))

	successCount := 0
	errorCount := 0

	for _, config := range configs {
		err = impl.rollbackSingleRemoteConnectionConfig(&config)
		if err != nil {
			log.Errorf("Failed to rollback remote connection config ID %d: %v", config.Id, err)
			errorCount++
			continue
		}
		successCount++
	}

	log.Infof("Rollback completed. Success: %d, Errors: %d", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("rollback completed with %d errors out of %d total configs", errorCount, len(configs))
	}

	return nil
}

func (impl *RollbackServiceImpl) rollbackSingleRemoteConnectionConfig(config *RemoteConnectionConfig) error {
	updateFields := make(map[string]interface{})

	// Process ssh_password field
	if config.SshPassword != "" {
		var encryptedString securestore.EncryptedString
		err := encryptedString.Scan(config.SshPassword)
		if err != nil {
			log.Warnf("Remote connection config ID %d: ssh_password appears to be already decrypted: %v", config.Id, err)
		} else {
			updateFields["ssh_password"] = string(encryptedString)
		}
	}

	// Process ssh_auth_key field
	if config.SshAuthKey != "" {
		var encryptedString securestore.EncryptedString
		err := encryptedString.Scan(config.SshAuthKey)
		if err != nil {
			log.Warnf("Remote connection config ID %d: ssh_auth_key appears to be already decrypted: %v", config.Id, err)
		} else {
			updateFields["ssh_auth_key"] = string(encryptedString)
		}
	}

	// Update the database if we have fields to update
	if len(updateFields) > 0 {
		query := impl.dbConnection.Model(config).Where("id = ?", config.Id)
		for field, value := range updateFields {
			query = query.Set(field+" = ?", value)
		}
		_, err := query.Update()
		if err != nil {
			return fmt.Errorf("failed to update remote connection config ID %d: %w", config.Id, err)
		}
		log.Infof("Successfully rolled back remote connection config ID %d", config.Id)
	} else {
		log.Infof("Remote connection config ID %d: no encrypted fields found to rollback", config.Id)
	}

	return nil
}

// RollbackSpecificRemoteConnectionConfig rolls back encryption for a specific remote connection config by ID
func (impl *RollbackServiceImpl) RollbackSpecificRemoteConnectionConfig(id int) error {
	log.Infof("Starting rollback for specific remote connection config ID: %d", id)

	// Initialize encryption key
	err := securestore.SetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to set encryption key: %w", err)
	}

	var config RemoteConnectionConfig
	err = impl.dbConnection.Model(&config).Where("id = ?", id).Select()
	if err != nil {
		return fmt.Errorf("failed to fetch remote connection config ID %d: %w", id, err)
	}

	err = impl.rollbackSingleRemoteConnectionConfig(&config)
	if err != nil {
		return fmt.Errorf("failed to rollback remote connection config ID %d: %w", id, err)
	}

	log.Infof("Successfully rolled back remote connection config ID %d", id)
	return nil
}

// ValidateRemoteConnectionConfigRollback validates that the rollback was successful
func (impl *RollbackServiceImpl) ValidateRemoteConnectionConfigRollback() error {
	log.Info("Validating remote connection config rollback results")

	var configs []RemoteConnectionConfig
	err := impl.dbConnection.Model(&configs).
		Where("ssh_password IS NOT NULL OR ssh_auth_key IS NOT NULL").
		Select()

	if err != nil {
		return fmt.Errorf("failed to fetch remote connection configs for validation: %w", err)
	}

	log.Infof("Validating %d remote connection configs", len(configs))

	for _, config := range configs {
		// Check ssh_password field
		if config.SshPassword != "" {
			var encryptedString securestore.EncryptedString
			err := encryptedString.Scan(config.SshPassword)
			if err == nil && string(encryptedString) != config.SshPassword {
				log.Warnf("Remote connection config ID %d: ssh_password appears to still be encrypted", config.Id)
			}
		}

		// Check ssh_auth_key field
		if config.SshAuthKey != "" {
			var encryptedString securestore.EncryptedString
			err := encryptedString.Scan(config.SshAuthKey)
			if err == nil && string(encryptedString) != config.SshAuthKey {
				log.Warnf("Remote connection config ID %d: ssh_auth_key appears to still be encrypted", config.Id)
			}
		}
	}

	log.Info("Remote connection config validation completed")
	return nil
}

// ============================================================================
// COMBINED OPERATIONS
// ============================================================================

// RollbackAllTables rolls back encrypted fields in all supported tables
func (impl *RollbackServiceImpl) RollbackAllTables() error {
	log.Info("Starting rollback of all tables")

	tables := []struct {
		name string
		fn   func() error
	}{
		{"cluster", impl.RollbackEncryptedConfig},
		{"gitops_config", impl.RollbackGitOpsConfigEncryptedFields},
		{"docker_artifact_store", impl.RollbackDockerArtifactStoreEncryptedFields},
		{"git_provider", impl.RollbackGitProviderEncryptedFields},
		{"remote_connection_config", impl.RollbackRemoteConnectionConfigEncryptedFields},
	}

	successCount := 0
	errorCount := 0
	var errors []string

	for _, table := range tables {
		log.Infof("Processing table: %s", table.name)
		err := table.fn()
		if err != nil {
			log.Errorf("Failed to rollback table %s: %v", table.name, err)
			errors = append(errors, fmt.Sprintf("%s: %v", table.name, err))
			errorCount++
		} else {
			log.Infof("Successfully rolled back table: %s", table.name)
			successCount++
		}
	}

	log.Infof("Rollback all tables completed. Success: %d, Errors: %d", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("rollback completed with %d errors:\n%s", errorCount, strings.Join(errors, "\n"))
	}

	return nil
}

// ValidateAllRollbacks validates that the rollback was successful for all tables
func (impl *RollbackServiceImpl) ValidateAllRollbacks() error {
	log.Info("Starting validation of all table rollbacks")

	tables := []struct {
		name string
		fn   func() error
	}{
		{"cluster", impl.ValidateRollback},
		{"gitops_config", impl.ValidateGitOpsConfigRollback},
		{"docker_artifact_store", impl.ValidateDockerArtifactStoreRollback},
		{"git_provider", impl.ValidateGitProviderRollback},
		{"remote_connection_config", impl.ValidateRemoteConnectionConfigRollback},
	}

	successCount := 0
	errorCount := 0
	var errors []string

	for _, table := range tables {
		log.Infof("Validating table: %s", table.name)
		err := table.fn()
		if err != nil {
			log.Errorf("Failed to validate table %s: %v", table.name, err)
			errors = append(errors, fmt.Sprintf("%s: %v", table.name, err))
			errorCount++
		} else {
			log.Infof("Successfully validated table: %s", table.name)
			successCount++
		}
	}

	log.Infof("Validation all tables completed. Success: %d, Errors: %d", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("validation completed with %d errors:\n%s", errorCount, strings.Join(errors, "\n"))
	}

	return nil
}
