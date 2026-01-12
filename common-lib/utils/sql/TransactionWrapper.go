/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sql

import (
	"github.com/devtron-labs/common-lib/utils"
	"github.com/go-pg/pg"
	"sync"
	"time"
)

type TransactionWrapper interface {
	StartTx() (*pg.Tx, error)
	RollbackTx(tx *pg.Tx) error
	CommitTx(tx *pg.Tx) error
}

type TransactionUtilImpl struct {
	dbConnection *pg.DB
	txStartTimes map[*pg.Tx]time.Time
	mu           sync.Mutex
	serviceName  string
}

func NewTransactionUtilImpl(db *pg.DB, serviceName string) *TransactionUtilImpl {
	return &TransactionUtilImpl{
		dbConnection: db,
		txStartTimes: make(map[*pg.Tx]time.Time),
		serviceName:  serviceName,
	}
}
func (impl *TransactionUtilImpl) RollbackTx(tx *pg.Tx) error {
	impl.observeTxHoldDuration(tx)
	return tx.Rollback()
}
func (impl *TransactionUtilImpl) CommitTx(tx *pg.Tx) error {
	impl.observeTxHoldDuration(tx)
	return tx.Commit()
}
func (impl *TransactionUtilImpl) StartTx() (*pg.Tx, error) {
	tx, err := impl.dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	impl.mu.Lock()
	impl.txStartTimes[tx] = time.Now()
	impl.mu.Unlock()
	return tx, nil
}

func (impl *TransactionUtilImpl) observeTxHoldDuration(tx *pg.Tx) {
	if tx == nil {
		return
	}
	impl.mu.Lock()
	startTime, exists := impl.txStartTimes[tx]
	if exists {
		delete(impl.txStartTimes, tx)
	}
	impl.mu.Unlock()

	if exists {
		utils.ObserveTxHold(startTime, impl.serviceName)
	}
}
