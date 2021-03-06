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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 *
 */

package dag

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	csort "github.com/palletone/go-palletone/core/sort"
	"github.com/palletone/go-palletone/dag/modules"
)

// 投票统计辅助结构体
type voteTally struct {
	candidate  common.Address
	votedCount uint64
}

func newVoteTally(candidate common.Address) *voteTally {
	return &voteTally{
		candidate:  candidate,
		votedCount: 0,
	}
}

type voteTallys []*voteTally

func (vts voteTallys) Len() int {
	return len(vts)
}

func (vts voteTallys) Less(i, j int) bool {
	mVoteI := vts[i].votedCount
	mVoteJ := vts[j].votedCount

	if mVoteI != mVoteJ {
		return mVoteI > mVoteJ
	}

	return vts[i].candidate.Less(vts[j].candidate)
}

func (vts voteTallys) Swap(i, j int) {
	vts[i], vts[j] = vts[j], vts[i]
}

// 获取账户相关投票数据的直方图
func (dag *Dag) performAccountMaintenance() {
	log.Debugf("Tally account voting mediators and setting mediators' count")
	// 1. 初始化数据
	dag.totalVotingStake = 0

	mediators := dag.GetMediators()
	mediatorCount := len(mediators)
	dag.mediatorVoteTally = make([]*voteTally, mediatorCount, mediatorCount)
	mediatorIndex := make(map[common.Address]int, mediatorCount)

	maxMediatorCount := dag.GetChainParameters().MaximumMediatorCount
	dag.mediatorCountHistogram = make([]uint64, maxMediatorCount/2+1)

	// 2. 遍历所有账户
	allAccount := dag.LookupAccount()
	mediatorVoteCount := make(map[common.Address]uint64)
	for _, info := range allAccount {
		//votingStake := dag.unstableStateRep.GetAccountBalance(addr)
		// 累加投票数量
		mediatorVoteCount[info.VotedMediator] += info.Balance
		dag.totalVotingStake += info.Balance
	}

	index := 0
	for mediator, _ := range mediators {
		// 建立 mediator 地址和index 的映射关系
		mediatorIndex[mediator] = index

		// 初始化 mediator 的投票数据
		voteTally := newVoteTally(mediator)
		count := mediatorVoteCount[mediator]
		voteTally.votedCount = count
		dag.mediatorVoteTally[index] = voteTally

		index++
	}

	//minFn := func(x, y int) int {
	//	if x < y {
	//		return x
	//	}
	//	return y
	//}

	// 2. 遍历所有账户
	//allAccount := dag.LookupAccount()
	//mediatorVoteCount := make(map[common.Address]uint64)
	//for _, info := range allAccount {
	//votingStake := dag.unstableStateRep.GetAccountBalance(addr)
	// 累加投票数量
	//dag.mediatorVoteTally[index].votedCount += info.Balance
	//mediatorVoteCount[info.VotedMediator] += info.Balance
	//// 遍历该账户投票的mediator
	//for med, _ := range info.VotedMediators {
	//	index, ok := mediatorIndex[med]
	//
	//	// if they somehow managed to specify an illegal mediator index, ignore it.
	//	if !ok {
	//		continue
	//	}
	//
	//	// 累加投票数量
	//	dag.mediatorVoteTally[index].votedCount += votingStake
	//}
	//
	//// 统计该账户设置的活跃mediator数量
	//desiredMediatorCount := info.DesiredMediatorCount
	//if desiredMediatorCount <= maxMediatorCount {
	//	offset := minFn(int(desiredMediatorCount/2), len(dag.mediatorCountHistogram)-1)
	//
	//	// votes for a number greater than MaximumMediatorCount
	//	// are turned into votes for MaximumMediatorCount.
	//	//
	//	// in particular, this takes care of the case where a
	//	// member was voting for a high number, then the
	//	// parameter was lowered.
	//	dag.mediatorCountHistogram[offset] += votingStake
	//}

	//	dag.totalVotingStake += info.Balance
	//}
}

func (dag *Dag) updateActiveMediators() bool {
	// 1. 统计出活跃mediator数量n
	stakeTarget := (dag.totalVotingStake - dag.mediatorCountHistogram[0]) / 2

	// accounts that vote for 0 or 1 mediator do not get to express an opinion on
	// the number of mediators to have (they abstain and are non-voting accounts)

	mediatorCountIndex := 0
	if stakeTarget > 0 {
		var stakeTally uint64 = 0
		upperLimit := len(dag.mediatorCountHistogram) - 1
		for mediatorCountIndex < upperLimit {
			mediatorCountIndex++
			stakeTally += dag.mediatorCountHistogram[mediatorCountIndex]

			if stakeTally > stakeTarget {
				break
			}
		}
	}

	maxFn := func(x, y int) int {
		if x > y {
			return x
		}
		return y
	}

	gp := dag.GetGlobalProp()
	minMediatorCount := gp.ImmutableParameters.MinimumMediatorCount
	mediatorCount := maxFn(mediatorCountIndex*2+1, int(minMediatorCount))

	mediatorLen := dag.mediatorVoteTally.Len()
	if mediatorLen < mediatorCount {
		log.Debugf("the desired mediator count is %v, the actual mediator count is %v,"+
			" the minimum mediator count is %v", mediatorCount, mediatorLen, minMediatorCount)
		// 保证活跃mediator的总数为奇数
		mediatorCount = (mediatorLen-1)/2*2 + 1
	}
	log.Debugf("In this round, The active mediator's count is %v", mediatorCount)

	// 2. 根据每个mediator的得票数，排序出前n个 active mediator
	csort.PartialSort(dag.mediatorVoteTally, mediatorCount)

	// 3. 更新每个mediator的得票数
	for _, voteTally := range dag.mediatorVoteTally {
		med := dag.GetMediator(voteTally.candidate)
		med.TotalVotes = voteTally.votedCount
		dag.SaveMediator(med, false)
	}

	// 4. 更新 global property 中的 active mediator 和 Preceding Mediators
	gp.PrecedingMediators = gp.ActiveMediators
	gp.ActiveMediators = make(map[common.Address]bool, mediatorCount)
	for index := 0; index < mediatorCount; index++ {
		voteTally := dag.mediatorVoteTally[index]
		gp.ActiveMediators[voteTally.candidate] = true
	}
	dag.SaveGlobalProp(gp, false)

	return isActiveMediatorsChanged(gp)
}

// 判断新一届mediator和上一届mediator是否有变化
func isActiveMediatorsChanged(gp *modules.GlobalProperty) bool {
	precedingMediators := gp.PrecedingMediators
	activeMediators := gp.ActiveMediators

	// 首先考虑活跃mediator个数是否改变
	if len(precedingMediators) != len(activeMediators) {
		return true
	}

	for am, _ := range activeMediators {
		if !precedingMediators[am] {
			return true
		}
	}

	return false
}

func (dag *Dag) updateChainParameters() {
	log.Debugf("update chain parameters")

	dag.UpdateSysParams()

	return
}
