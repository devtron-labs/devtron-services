/*
 * Copyright (c) 2024. Devtron Inc.
 */

package securestore

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type EncryptionKeyService interface {
	// CreateAndStoreEncryptionKey generates a new AES-256 encryption key and stores it in the attributes repository
	CreateAndStoreEncryptionKey() error

	// RotateEncryptionKey generates a new encryption key and stores it (deactivating the old one)
	RotateEncryptionKey(userId int32) (string, error)

	// GenerateEncryptionKey generates a new AES-256 encryption key (32 bytes)
	GenerateEncryptionKey() (string, error)

	GetEncryptionKey() (string, error)
}

type EncryptionKeyServiceImpl struct {
	logger               *zap.SugaredLogger
	attributesRepository AttributesRepository
}

func NewEncryptionKeyServiceImpl(
	logger *zap.SugaredLogger,
	attributesRepository AttributesRepository) (*EncryptionKeyServiceImpl, error) {
	impl := &EncryptionKeyServiceImpl{
		logger:               logger,
		attributesRepository: attributesRepository,
	}
	err := impl.CreateAndStoreEncryptionKey()
	if err != nil {
		return nil, err
	}
	return impl, nil
}

// GenerateEncryptionKey generates a new AES-256 encryption key (32 bytes = 256 bits)
func (impl *EncryptionKeyServiceImpl) GenerateEncryptionKey() (string, error) {
	// Generate 32 random bytes for AES-256
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		impl.logger.Errorw("error generating random encryption key", "err", err)
		return "", fmt.Errorf("failed to generate encryption key: %w", err)
	}
	// Encode to hex string for storage
	keyHex := hex.EncodeToString(key)
	return keyHex, nil
}

// CreateAndStoreEncryptionKey generates a new AES-256 encryption key and stores it in the attributes repository
func (impl *EncryptionKeyServiceImpl) CreateAndStoreEncryptionKey() error {
	// Check if encryption key already exists
	encryptionKeyModel, err := impl.attributesRepository.FindByKey(ENCRYPTION_KEY)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error checking for existing encryption key", "err", err)
		return err
	}

	if encryptionKeyModel != nil && encryptionKeyModel.Id > 0 && len(encryptionKeyModel.Value) > 0 {
		encryptionKey = []byte(encryptionKeyModel.Value)
		impl.logger.Warnw("encryption key already exists", "keyId", encryptionKeyModel.Id)
	} else {
		// Generate new encryption key
		encryptionKeyNew, err := impl.GenerateEncryptionKey()
		if err != nil {
			return err
		}
		// Store in repository
		err = impl.attributesRepository.SaveEncryptionKeyIfNotExists(encryptionKeyNew)
		if err != nil {
			impl.logger.Errorw("error storing encryption key", "err", err)
			return fmt.Errorf("failed to store encryption key: %w", err)
		}
		encryptionKey = []byte(encryptionKeyNew)
		impl.logger.Infow("Successfully created and stored encryption key")
	}
	return nil
}

// RotateEncryptionKey generates a new encryption key and stores it (deactivating the old one)
func (impl *EncryptionKeyServiceImpl) RotateEncryptionKey(userId int32) (string, error) {
	impl.logger.Infow("Rotating encryption key", "userId", userId)

	// Generate new encryption key
	newEncryptionKey, err := impl.GenerateEncryptionKey()
	if err != nil {
		return "", err
	}

	// Store in repository (this will deactivate the old key)
	err = impl.attributesRepository.SaveEncryptionKeyIfNotExists(newEncryptionKey)
	if err != nil {
		impl.logger.Errorw("error rotating encryption key", "err", err)
		return "", fmt.Errorf("failed to rotate encryption key: %w", err)
	}
	//TODO: also need to rotate encryption's already done
	impl.logger.Infow("Successfully rotated encryption key", "userId", userId)
	return newEncryptionKey, nil
}

// GetEncryptionKey retrieves the active encryption key from the repository
func (impl *EncryptionKeyServiceImpl) GetEncryptionKey() (string, error) {
	key, err := impl.attributesRepository.GetEncryptionKey()
	if err != nil {
		if err == pg.ErrNoRows {
			impl.logger.Errorw("encryption key not found in repository")
			return "", fmt.Errorf("encryption key not found, please create one first")
		}
		impl.logger.Errorw("error retrieving encryption key", "err", err)
		return "", err
	}
	return key, nil
}
