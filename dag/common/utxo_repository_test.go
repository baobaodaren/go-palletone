/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package common

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"

	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

func mockUtxoRepository() *UtxoRepository {
	db, _ := ptndb.NewMemDatabase()
	utxodb := storage.NewUtxoDb(db)
	idxdb := storage.NewIndexDb(db)
	statedb := storage.NewStateDb(db)
	return NewUtxoRepository(utxodb, idxdb, statedb)
}

func TestUpdateUtxo(t *testing.T) {
	rep := mockUtxoRepository()
	rep.UpdateUtxo(common.Hash{}, &modules.PaymentPayload{}, uint32(0))
}

func TestReadUtxos(t *testing.T) {
	rep := mockUtxoRepository()
	utxos, totalAmount := rep.ReadUtxos(common.Address{}, modules.Asset{})
	log.Println(utxos, totalAmount)
}

func TestGetUxto(t *testing.T) {
	dagconfig.DagConfig.DbPath = getTempDir(t)
	log.Println(modules.Input{})
}

func getTempDir(t *testing.T) string {
	d, err := ioutil.TempDir("", "leveldb-test")
	if err != nil {
		t.Fatal(err)
	}
	return d
}

//func TestSaveAssetInfo(t *testing.T) {
//	assetid := modules.PTNCOIN
//	asset := modules.Asset{
//		AssetId:  assetid,
//		UniqueId: assetid,
//	}
//	assetInfo := modules.AssetInfo{
//		GasToken:        "Test",
//		AssetID:      &asset,
//		InitialTotal: 1000000000,
//		Decimal:      100000000,
//	}
//	assetInfo.OriginalHolder.SetString("Mytest")
//}

//func TestWalletBalance(t *testing.T) {
//	rep := mockUtxoRepository()
//	addr := common.Address{}
//	addr.SetString("P1CXn936dYuPKGyweKPZRycGNcwmTnqeDaA")
//	balance := rep.WalletBalance(addr, modules.Asset{})
//	log.Println("Address total =", balance)
//}

//
//func TestGetAccountTokens(t *testing.T) {
//	rep := mockUtxoRepository()
//	addr := common.Address{}
//	addr.SetString("P12EA8oRMJbAtKHbaXGy8MGgzM8AMPYxkNr")
//	tokens, err := rep.GetAccountTokens(addr)
//	if err != nil {
//		log.Println("Get account error:", err.Error())
//	} else if len(tokens) == 0 {
//		log.Println("Get none account")
//	} else {
//		for _, token := range tokens {
//			log.Printf("Token (%s, %v) = %v\n",
//				token.GasToken, token.AssetID.AssetId, token.Balance)
//			// test WalletBalance method
//			log.Println(rep.WalletBalance(addr, *token.AssetID))
//			// test ReadUtxos method
//			utxos, amount := rep.ReadUtxos(addr, *token.AssetID)
//			log.Printf("Addr(%s) balance=%v\n", addr.String(), amount)
//			for outpoint, utxo := range utxos {
//				log.Println(">>> UTXO txhash =", outpoint.TxHash.String())
//				log.Println("    UTXO msg index =", outpoint.MessageIndex)
//				log.Println("    UTXO out index =", outpoint.OutIndex)
//				log.Println("    UTXO amount =", utxo.Amount)
//			}
//		}
//	}
//
//}
