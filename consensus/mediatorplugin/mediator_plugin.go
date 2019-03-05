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
 */

package mediatorplugin

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
)

var (
	errTerminated = errors.New("terminated")
)

func newChainBanner(dag iDag) {
	fmt.Printf("\n" +
		"*   ------- NEW CHAIN -------   *\n" +
		"*   - Welcome to PalletOne! -   *\n" +
		"*   -------------------------   *\n" +
		"\n")

	if dag.GetSlotAtTime(time.Now()) > 200 {
		fmt.Printf("Your genesis seems to have an old timestamp\n" +
			"Please consider using the --genesistime option to give your genesis a recent timestamp\n" +
			"\n")
	}
}

func (mp *MediatorPlugin) SubscribeNewProducedUnitEvent(ch chan<- NewProducedUnitEvent) event.Subscription {
	return mp.newProducedUnitScope.Track(mp.newProducedUnitFeed.Subscribe(ch))
}

func (mp *MediatorPlugin) scheduleProductionLoop() {
	// 1. 计算下一秒的滴答时刻，如果少于50毫秒，则多等一秒开始
	now := time.Now()
	timeToNextSecond := time.Second - time.Duration(now.Nanosecond())
	if timeToNextSecond < 50*time.Millisecond {
		timeToNextSecond += time.Second
	}

	// 2. 安排unit生产循环
	// Start to production unit for expiration
	timeout := time.NewTimer(timeToNextSecond)
	defer timeout.Stop()

	// production unit until termination is requested
	select {
	case <-mp.quit:
		return
	case <-mp.stopProduce:
		return
	case <-timeout.C:
		go mp.unitProductionLoop()
	}
}

// unit生产的状态类型
type ProductionCondition uint8

// unit生产的状态枚举
const (
	Produced ProductionCondition = iota // 正常生产unit
	NotSynced
	NotMyTurn
	NotTimeYet
	//NoPrivateKey
	LowParticipation
	Lag
	Consecutive
	ExceptionProducing
)

func (mp *MediatorPlugin) unitProductionLoop() ProductionCondition {
	// 1. 尝试生产unit
	result, detail := mp.maybeProduceUnit()

	// 2. 打印尝试结果
	switch result {
	case Produced:
		log.Info("Generated unit(" + detail["Hash"] + ") #" + detail["Num"] + " parent(" + detail["ParentHash"] +
			") @" + detail["Timestamp"] + " signed by " + detail["Mediator"])
	case NotSynced:
		log.Info("Not producing unit because production is disabled until we receive a recent unit." +
			" Disable this check with --staleProduce option.")
	case NotTimeYet:
		log.Debug("Not producing unit because next slot time is " + detail["NextTime"] +
			" , but now is " + detail["Now"])
	case NotMyTurn:
		log.Debug("Not producing unit because current scheduled mediator is " +
			detail["ScheduledMediator"])
	case Lag:
		log.Info("Not producing unit because node didn't wake up within 500ms of the slot time." +
			" Scheduled Time is: " + detail["ScheduledTime"] + ", but now is " + detail["Now"])
	//case NoPrivateKey:
	//	log.Info("Not producing unit because I don't have the private key for " +
	//		detail["ScheduledKey"])
	case Consecutive:
		log.Info("Not producing unit because the last unit was generated " +
			"by the same mediator(" + detail["Mediator"] + ")." +
			" This node is probably disconnected from the network so unit production has been disabled." +
			" Disable this check with --allowConsecutive option.")
	case LowParticipation:
		log.Infof("Not producing unit because node appears to be on a minority fork with only %v "+
			"mediator participation.", detail["ParticipationRate"])
	case ExceptionProducing:
		log.Infof("Exception producing unit: %v", detail["Msg"])
	default:
		log.Infof("Unknown condition while producing unit!")
	}

	// 3. 继续循环生产计划
	go mp.scheduleProductionLoop()

	return result
}

func (mp *MediatorPlugin) maybeProduceUnit() (ProductionCondition, map[string]string) {
	//defer func(start time.Time) {
	//	log.Debug("maybeProduceUnit unit elapsed", "elapsed", time.Since(start))
	//}(time.Now())

	detail := make(map[string]string)
	dag := mp.dag

	// 整秒调整，四舍五入
	nowFine := time.Now()
	now := time.Unix(nowFine.Add(500*time.Millisecond).Unix(), 0)

	// 1. 判断是否满足生产的各个条件
	nextSlotTime := dag.GetSlotTime(1)
	// If the next Unit production opportunity is in the present or future, we're synced.
	if !mp.productionEnabled {
		if nextSlotTime.After(now) || nextSlotTime.Equal(now) {
			mp.productionEnabled = true
		} else {
			return NotSynced, detail
		}
	}

	slot := dag.GetSlotAtTime(now)
	// is anyone scheduled to produce now or one second in the future?
	if slot == 0 {
		detail["NextTime"] = nextSlotTime.Format("2006-01-02 15:04:05.000")
		detail["Now"] = now.Format("2006-01-02 15:04:05.000")
		return NotTimeYet, detail
	}

	// this conditional judgment should fail, because now <= HeadUnitTime
	// should have resulted in slot == 0.
	//
	// if this assert triggers, there is a serious bug in dag.GetSlotAtTime()
	// which would result in allowing a later unit to have a timestamp
	// less than or equal to the previous unit
	if !(now.Unix() > dag.HeadUnitTime()) {
		//panic("The later unit have a timestamp less than or equal to the previous!")
		detail["Msg"] = "The property database is being updated because the new unit is received synchronously."
		return ExceptionProducing, detail
	}

	scheduledMediator := dag.GetScheduledMediator(slot)
	if scheduledMediator.Equal(common.Address{}) {
		detail["Msg"] = "The current shuffled mediators is nil!"
		return ExceptionProducing, detail
	}

	// we must control the Mediator scheduled to produce the next Unit.
	//med, ok := mp.mediators[scheduledMediator]
	_, ok := mp.mediators[scheduledMediator]
	if !ok {
		detail["ScheduledMediator"] = scheduledMediator.Str()
		return NotMyTurn, detail
	}

	// 此处应该判断scheduledMediator的签名公钥对应的私钥在本节点是否存在
	ks := mp.ptn.GetKeyStore()
	//err := ks.Unlock(accounts.Account{Address: scheduledMediator}, med.Password)
	//if err != nil {
	//	detail["ScheduledKey"] = scheduledMediator.Str()
	//	return NoPrivateKey, detail
	//}

	if !mp.consecutiveProduceEnabled && dag.IsConsecutiveMediator(scheduledMediator) {
		detail["Mediator"] = scheduledMediator.Str()
		return Consecutive, detail
	}

	pRate := dag.MediatorParticipationRate()
	if pRate < mp.requiredParticipation {
		detail["ParticipationRate"] = fmt.Sprint(pRate / core.PalletOne1Percent)
		return LowParticipation, detail
	}

	// todo 由于当前代码更新数据库没有加锁，可能如下情况：
	// 生产单元的协程满足了前面的判断，此时新收到一个unit正在更新数据库，后面的判断有不能通过
	// todo 调试的代码，暂时注释掉该判断，release版本的代码必须使用该判断
	scheduledTime := dag.GetSlotTime(slot)
	diff := scheduledTime.Sub(now)
	if diff > 500*time.Millisecond || diff < -500*time.Millisecond {
		detail["ScheduledTime"] = scheduledTime.Format("2006-01-02 15:04:05")
		detail["Now"] = now.Format("2006-01-02 15:04:05")
		return Lag, detail
	}

	// 2. 生产单元
	//execute contract
	// todo 待优化
	if err := mp.ptn.ContractProcessor().AddContractLoop(mp.ptn.TxPool(), scheduledMediator, ks); err != nil {
		log.Debug("MaybeProduceUnit", "RunContractLoop err:", err.Error())
	}

	groupPubKey := mp.LocalMediatorPubKey(scheduledMediator)
	newUnit := dag.GenerateUnit(scheduledTime, scheduledMediator, groupPubKey, ks, mp.ptn.TxPool())
	if newUnit == nil || newUnit.IsEmpty() {
		detail["Msg"] = "The newly produced unit is empty!"
		return ExceptionProducing, detail
	}

	unitHash := newUnit.UnitHash
	detail["Num"] = strconv.FormatUint(newUnit.NumberU64(), 10)
	time := time.Unix(newUnit.Timestamp(), 0)
	detail["Timestamp"] = time.Format("2006-01-02 15:04:05")
	detail["Mediator"] = scheduledMediator.Str()
	detail["Hash"] = unitHash.TerminalString()
	detail["ParentHash"] = newUnit.ParentHash()[0].TerminalString()

	// 3. 对 unit 进行群签名和广播
	if mp.groupSigningEnabled {
		go mp.groupSignUnit(scheduledMediator, unitHash)
	}

	// 4. 异步向区块链网络广播新unit
	go mp.newProducedUnitFeed.Send(NewProducedUnitEvent{Unit: newUnit})

	return Produced, detail
}

func (mp *MediatorPlugin) groupSignUnit(localMed common.Address, unitHash common.Hash) {
	// 1. 初始化签名unit相关的签名分片的buf
	mp.recoverBufLock.Lock()
	aSize := mp.dag.ActiveMediatorsCount()
	if _, ok := mp.toTBLSRecoverBuf[localMed]; !ok {
		mp.toTBLSRecoverBuf[localMed] = make(map[common.Hash]*sigShareSet)
	}
	mp.toTBLSRecoverBuf[localMed][unitHash] = newSigShareSet(aSize)
	mp.recoverBufLock.Unlock()

	// 2. 过了 unit 确认时间后，及时删除群签名分片的相关数据，防止内存溢出
	go func() {
		expiration := mp.dag.UnitIrreversibleTime()
		deleteBuf := time.NewTimer(expiration)

		select {
		case <-mp.quit:
			return
		case <-deleteBuf.C:
			mp.recoverBufLock.Lock()
			if _, ok := mp.toTBLSRecoverBuf[localMed][unitHash]; ok {
				log.Debugf("the unit(%v) has expired confirmation time, no longer need the mediator(%v) "+
					"to recover group-sign", unitHash.TerminalString(), localMed.Str())
				delete(mp.toTBLSRecoverBuf[localMed], unitHash)
			}
			mp.recoverBufLock.Unlock()
		}
	}()
}
