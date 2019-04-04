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

package rwset

import (
	"errors"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag"
)

type RwSetTxMgr struct {
	//rwLock            	sync.RWMutex
	name      string
	baseTxSim map[string]TxSimulator
}

func NewRwSetMgr(name string) (*RwSetTxMgr, error) {
	return &RwSetTxMgr{name, make(map[string]TxSimulator)}, nil
}

// NewTxSimulator implements method in interface `txmgmt.TxMgr`
func (m *RwSetTxMgr) NewTxSimulator(idag dag.IDag, chainid string, txid string) (TxSimulator, error) {
	log.Debugf("constructing new tx simulator")
	hash := common.HexToHash(txid)
	if _, ok := m.baseTxSim[chainid]; ok {
		if m.baseTxSim[chainid].(*RwSetTxSimulator).txid == hash {
			log.Infof("chainid[%s] , txid[%s]already exit", chainid, txid)
			return m.baseTxSim[chainid], nil
		}
	}

	t := NewBasedTxSimulator(idag, hash)
	if t == nil {
		return nil, errors.New("NewBaseTxSimulator is failed.")
	}
	m.baseTxSim[chainid] = t
	log.Infof("creat new rwSetTx")

	return t, nil
}

func init() {

}
