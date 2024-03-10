// SPDX-FileCopyrightText: 2024 PK Lab AG <contact@pklab.io>
// SPDX-License-Identifier: MIT

package storage

import (
	"context"

	"github.com/celenium-io/celestia-indexer/internal/storage"
	"github.com/dipdup-net/indexer-sdk/pkg/sync"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func (module *Module) saveValidators(
	ctx context.Context,
	tx storage.Transaction,
	validators []*storage.Validator,
	jailed *sync.Map[string, *storage.Validator],
	jails []storage.Jail,
) (int, error) {
	if jailed.Len() > 0 {
		jailedVals := make([]*storage.Validator, 0)
		err := jailed.Range(func(address string, val *storage.Validator) (error, bool) {
			if id, ok := module.validatorsByConsAddress[address]; ok {
				val.Id = id
				jailedVals = append(jailedVals, val)
				return nil, false
			}

			return errors.Errorf("unknown jailed validator: %s", address), false
		})
		if err != nil {
			return 0, err
		}

		if err := tx.Jail(ctx, jailedVals...); err != nil {
			return 0, err
		}
	}

	if len(jails) > 0 {
		for i := range jails {
			if id, ok := module.validatorsByConsAddress[jails[i].Validator.ConsAddress]; ok {
				jails[i].ValidatorId = id
			}

			fraction := decimal.Zero
			switch jails[i].Reason {
			case "double_sign":
				fraction = module.slashingForDoubleSign.Copy()
			case "missing_signature":
				fraction = module.slashingForDowntime.Copy()
			}
			if !fraction.IsPositive() {
				continue
			}

			balanceUpdates, err := tx.UpdateSlashedDelegations(ctx, jails[i].ValidatorId, fraction)
			if err != nil {
				return 0, err
			}
			if err := tx.SaveBalances(ctx, balanceUpdates...); err != nil {
				return 0, err
			}
		}

		if err := tx.SaveJails(ctx, jails...); err != nil {
			return 0, err
		}
	}

	if len(validators) == 0 {
		return 0, nil
	}

	count, err := tx.SaveValidators(ctx, validators...)
	if err != nil {
		return 0, errors.Wrap(err, "saving validators")
	}

	if count == 0 {
		return 0, nil
	}

	for i := range validators {
		if validators[i].ConsAddress == "" {
			continue
		}
		module.validatorsByConsAddress[validators[i].ConsAddress] = validators[i].Id
		module.validatorsByAddress[validators[i].Address] = validators[i].Id
	}

	return count, nil
}