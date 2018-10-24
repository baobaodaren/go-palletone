package jury

import (
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/errors"
	"fmt"
	"github.com/palletone/go-palletone/common/log"
)

type PeerType int

const (
	_         PeerType = iota
	TUnknow
	TJury
	TMediator
)

type Juror struct {
	name    string
	address common.Address

	//InitPartSec kyber.Scalar
	InitPartPub kyber.Point
}

type Processor struct {
	name    string
	dag     dag.IDag
	ptype   PeerType
	address common.Address
	quit    chan struct{}

	jurors map[common.Address]Juror //记录所有执行合约的节点信息

	contractExecFeed  event.Feed
	contractExecScope event.SubscriptionScope

	contractBroadcastFeed  event.Feed
	contractBroadcastScope event.SubscriptionScope
}

func (p *Processor) ContractProcess() error {
	return nil
}

func (p *Processor) Consensus() error {
	return nil
}

func (p *Processor) Start() error {
	//资源初始化

	//启动消息接收处理线程

	//合约执行线程

	return nil
}

func (p *Processor) Stop() error {
	return nil
}

func (p *Processor) SubscribeContractEvent(ch chan<- ContractExeEvent) event.Subscription {
	return p.contractExecScope.Track(p.contractExecFeed.Subscribe(ch))
}

type ContractCMD byte

//const (
//	CONTRACT_INSTALL ContractCMD = iota
//	CONTRACT_DEPLOY
//	CONTRACT_INVOKE
//	CONTRACT_STOP
//)

func (p *Processor) ProcessContractEvent(event *ContractExeEvent) error {
	//Processor 需要记录每次合约的执行，在共识通过后清除
	//检查event
	if event == nil {
		return errors.New("param is nil")
	}
	runContractCmd := func(msg *modules.Message) (error) {
		switch msg.App {
		case modules.APP_CONTRACT_TPL:
			{
			}
		case modules.APP_CONTRACT_DEPLOY:
			{
			}
		case modules.APP_CONTRACT_INVOKE:
			{
				req := ContractInvokeReq{
					chainID:  "palletone",
					deployId: msg.Payload.(modules.ContractInvokeRequestPayload).ContractId,
					args:     msg.Payload.(modules.ContractInvokeRequestPayload).Args,
				}

				//todo tmp
				//req := ContractInvokeReq{
				//	chainID:  "palletone",
				//	deployId: []byte("9527"),
				//	txid:     "1234567",
				//	args:     util.ToChaincodeArgs("add", "3", "2"),
				//	timeout:  time.Second * time.Duration(3),
				//}
				payload, err := ContractProcess(req)
				if err != nil {
					log.Error("contract exec fail:%s", err)
					return errors.New("ContractProcess fail")
				}

				log.Info("", payload)
			}
		case modules.APP_CONTRACT_STOP:
			{
			}
		default:
			return errors.New(fmt.Sprintf("event conversion fail,th"))
		}
		return nil
	}

	//var ctr *contractInfo
	if len(event.Tx.TxMessages) > 0 {
		for _, pTx := range event.Tx.TxMessages {
			go runContractCmd(pTx)
		}
	}

	//执行合约命令:install、deploy、invoke、stop
	//event.Tx.TxMessages[0].Payload

	//异常处理，记录标识
	//获取到执行结果，记录到本地内存，用于与接收到其他节点合约执行的结果进行对比

	//将执行结果hash、签名、广播
	//

	return nil
}

func (p *Processor) ProcessContractSigEvent(event *ContractSigEvent) error {
	//检查事件

	//签名检查

	//添加到数据对比队列，与本地合约执行结果进行payload对比
	//这里注意的是，如果本地合约没有执行完成，需等待

	//签名，检查收集2/3以上:
	//活跃mediator---群签
	//jury---检查是否自己是最小签名，即确定自己是否为leader。
	//如果是leader，将签名后的数据以tx形式发给Mediator,后续咋处理？

	//

	return nil
}

func (p *Processor) SubscribeContractBroadcastEvent(ch chan<- ContractSigEvent) event.Subscription {
	return p.contractBroadcastScope.Track(p.contractBroadcastFeed.Subscribe(ch))
}

func (p *Processor) ProcessContractBroadcastEvent(event *ContractSigEvent) error {

	return nil
}
