/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package storage

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/util"
)

// value will serialize to rlp encoding bytes
func Store(db ptndb.Database, key string, value interface{}) error {
	return StoreBytes(db, []byte(key), value)
}

//func BatchStore(batch ptndb.Batch, key []byte, value interface{}) error {
//	val, err := rlp.EncodeToBytes(value)
//	if err != nil {
//		return err
//	}
//	err = batch.Put(key, val)
//	if err != nil {
//		log.Error("batch put error", "key:", string(key), "err:", err)
//	}
//	return err
//}
func StoreBytes(db ptndb.Putter, key []byte, value interface{}) error {
	val, err := rlp.EncodeToBytes(value)
	if err != nil {
		return err
	}
	err = db.Put(key, val)
	if err != nil {
		log.Error("StoreBytes", "key:", string(key), "err:", err)
	}
	return err
	/*
		_, err = db.Get(key)
		if err != nil {
			if err.Error() == errors.ErrNotFound.Error() {
				//	if err == errors.New("not found") {
				if err := db.Put(key, val); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if err = db.Delete(key); err != nil {
				return err
			}
			if err := db.Put(key, val); err != nil {
				return err
			}
		}
		return nil
	*/
}

func GetBytes(db ptndb.Database, key []byte) ([]byte, error) {
	val, err := db.Get(key)
	if err != nil {
		return []byte{}, err
	}
	log.Debug("storage GetBytes", "key:", string(key), "value:", string(val))
	return val, nil
}



func StoreString(db ptndb.Putter, key, value string) error {
	return db.Put(util.ToByte(key), util.ToByte(value))
}
func GetString(db ptndb.Database, key string) (string, error) {
	return getString(db, util.ToByte(key))
}

func BatchErrorHandler(err error, errorList *[]error) {
	if err != nil {
		*errorList = append(*errorList, err)
	}
}
