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

package modules

import "github.com/palletone/go-palletone/common"

//一个账户（地址）的状态信息
//Include:
// personal account P1*
//P2SH account P3*
//Contract account PC*
type AccountInfoBase struct {
	//AccountName string

	//当前账户的PTN余额，Devin:因为操作太频繁，转移到AccountBalance上了
	//PtnBalance uint64

	// 本账户期望的活跃mediator数量
	DesiredMediatorCount uint8
}

func NewAccountInfoBase() *AccountInfoBase {
	return &AccountInfoBase{
		//PtnBalance:         0,
		DesiredMediatorCount: 0,
	}
}

type AccountInfo struct {
	*AccountInfoBase
	//当前账户投票的Mediator
	VotedMediators map[common.Address]bool
}

func NewAccountInfo() *AccountInfo {
	return &AccountInfo{
		AccountInfoBase: NewAccountInfoBase(),
		VotedMediators:  make(map[common.Address]bool),
	}
}

type AccountUpdateOperation struct {
	DesiredMediatorCount *uint8          `json:"desiredMediatorCount"`
	VotingMediator       *common.Address `json:"votingMediator"`
}
