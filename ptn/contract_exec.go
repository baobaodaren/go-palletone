package ptn

import (
	"github.com/palletone/go-palletone/consensus/jury"
	"github.com/palletone/go-palletone/common/event"
)

type contractInf interface {
	SubscribeContractEvent(ch chan<- jury.ContractExeEvent) event.Subscription
	ProcessContractEvent(event *jury.ContractExeEvent) error

	ProcessContractSigEvent(event *jury.ContractSigEvent) error

	SubscribeContractBroadcastEvent(ch chan<- jury.ContractSigEvent) event.Subscription
	ProcessContractBroadcastEvent(event *jury.ContractSigEvent) error
}

func (self *ProtocolManager) contractDealRecvLoop() {
	for {
		select {
		case event := <-self.contractExecCh:
			self.contractProc.ProcessContractEvent(&event)

		case <-self.contractExecSub.Err():
			return
		}
	}
}