// Copyright © 2020 Dmitry Stoletov <info@imega.ru>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package txwrapper

import (
	"context"
	"database/sql"
	"fmt"
)

type TxWrapper struct{ DB *sql.DB }

type TxFunc func(context.Context, *sql.Tx) error

func New(db *sql.DB) *TxWrapper {
	return &TxWrapper{DB: db}
}

func (w *TxWrapper) Transaction(
	ctx context.Context,
	opts *sql.TxOptions,
	txfn TxFunc,
) error {
	wtx, err := w.DB.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction, %w", err)
	}

	if err := txfn(ctx, wtx); err != nil {
		if e := wtx.Rollback(); e != nil {
			return fmt.Errorf("failed to execute transaction, %w", err)
		}

		return err
	}

	if err := wtx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction, %w", err)
	}

	return nil
}
