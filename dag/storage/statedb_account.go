/*
 *
 *     This file is part of go-palletone.
 *     go-palletone is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *     go-palletone is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *     You should have received a copy of the GNU General Public License
 *     along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"bytes"

	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

// only for serialization(storage)
type accountInfo struct {
	*modules.AccountInfoBase
	VotedMediators []common.Address
}

func newAccountInfo() *accountInfo {
	return &accountInfo{
		AccountInfoBase: modules.NewAccountInfoBase(),
		VotedMediators:  make([]common.Address, 0),
	}
}

func (acc *accountInfo) accountToInfo() *modules.AccountInfo {
	ai := modules.NewAccountInfo()
	ai.AccountInfoBase = acc.AccountInfoBase

	for _, med := range acc.VotedMediators {
		ai.VotedMediators[med] = true
	}

	return ai
}

func infoToAccount(ai *modules.AccountInfo) *accountInfo {
	acc := newAccountInfo()
	acc.AccountInfoBase = ai.AccountInfoBase

	for med, _ := range ai.VotedMediators {
		acc.VotedMediators = append(acc.VotedMediators, med)
	}

	return acc
}

func accountKey(address common.Address) []byte {
	key := append(constants.ACCOUNT_INFO_PREFIX, address.Bytes21()...)

	return key
}

func (statedb *StateDb) RetrieveAccountInfo(address common.Address) (*modules.AccountInfo, error) {
	acc := newAccountInfo()

	err := retrieve(statedb.db, accountKey(address), acc)
	if err != nil {
		log.Debugf("Get account[%s] info throw an error:%s", address.String(), err.Error())
		return nil, err
	}

	return acc.accountToInfo(), nil
}

func (statedb *StateDb) StoreAccountInfo(address common.Address, info *modules.AccountInfo) error {
	err := StoreBytes(statedb.db, accountKey(address), infoToAccount(info))
	if err != nil {
		log.Debugf("Save account info throw an error:%s", err)
	}

	return err
}

func (statedb *StateDb) UpdateAccountInfo(account common.Address,
	accountUpdateOp *modules.AccountUpdateOperation) error {
	accountInfo, err := statedb.RetrieveAccountInfo(account)
	if accountInfo == nil || err != nil {
		accountInfo = modules.NewAccountInfo()
	}

	if accountUpdateOp.DesiredMediatorCount != nil {
		mediatorCountSet := *accountUpdateOp.DesiredMediatorCount
		accountInfo.DesiredMediatorCount = mediatorCountSet
		log.Debugf("Try to update DesiredMediatorCount(%v) for account(%v)", mediatorCountSet, account.Str())
	}

	if accountUpdateOp.VotingMediator != nil {
		mediator := *accountUpdateOp.VotingMediator
		accountInfo.VotedMediators[mediator] = true
		log.Debugf("Try to save voted mediator(%v) for account(%v)", mediator.Str(), account.Str())
	}

	return statedb.StoreAccountInfo(account, accountInfo)
}

func (statedb *StateDb) UpdateAccountBalance(address common.Address, addAmount int64) error {
	key := append(constants.ACCOUNT_PTN_BALANCE_PREFIX, address.Bytes21()...)
	balance := uint64(0)
	data, err := statedb.db.Get(key)
	if err != nil {
		// 第一次更新时， 数据库没有该账户的相关数据
		log.Debugf("Account info for [%s] don't exist,create it first", address.String())
	} else {
		balance = BytesToUint64(data)
	}
	log.Debugf("Update Ptn Balance for address:%s, add Amount:%d", address.String(), addAmount)
	balance = uint64(int64(balance) + addAmount)
	return statedb.db.Put(key, Uint64ToBytes(balance))
}

func (statedb *StateDb) GetAccountBalance(address common.Address) uint64 {
	key := append(constants.ACCOUNT_PTN_BALANCE_PREFIX, address.Bytes21()...)
	balance := uint64(0)
	data, err := statedb.db.Get(key)
	if err == nil {

		balance = BytesToUint64(data)
	}
	return balance
}

func (statedb *StateDb) LookupAccount() map[common.Address]*modules.AccountInfo {
	result := make(map[common.Address]*modules.AccountInfo)

	iter := statedb.db.NewIteratorWithPrefix(constants.ACCOUNT_INFO_PREFIX)
	for iter.Next() {
		key := iter.Key()
		if key == nil {
			continue
		}

		addB := bytes.TrimPrefix(key, constants.ACCOUNT_INFO_PREFIX)
		add := common.BytesToAddress(addB)

		value := iter.Value()
		if value == nil {
			continue
		}

		acc := newAccountInfo()
		err := rlp.DecodeBytes(value, acc)
		if err != nil {
			log.Debug(fmt.Sprintln("Error in Decoding Bytes to AccountInfo: ", err,
				"\nkey: ", key, "\naddress: ", add.Str(), "\nvalue: ", value))
			continue
		}

		result[add] = acc.accountToInfo()
	}

	return result
}
