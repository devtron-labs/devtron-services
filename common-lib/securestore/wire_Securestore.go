package securestore

import "github.com/google/wire"

var SecurestoreWireSet = wire.NewSet(
	NewAttributesRepositoryImpl,
	wire.Bind(new(AttributesRepository), new(*AttributesRepositoryImpl)),
	NewEncryptionKeyServiceImpl,
	wire.Bind(new(EncryptionKeyService), new(*EncryptionKeyServiceImpl)))

var SecurestoreWireSetNonOrchDbServices = wire.NewSet(
	NewAttributesRepositoryImplForOrchestrator,
	wire.Bind(new(AttributesRepository), new(*AttributesRepositoryImpl)),
	NewEncryptionKeyServiceImpl,
	wire.Bind(new(EncryptionKeyService), new(*EncryptionKeyServiceImpl)))
