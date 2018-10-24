package jury

import (
	"github.com/palletone/go-palletone/contracts"
	"time"
)

type ContractResp struct {
	Err  error
	Resp interface{}
}

type ContractReqInf interface {
	do(v contracts.ContractInf) ContractResp
	//getContractInfo() (error)
}

//////
type ContractInstallReq struct {
	chainID   string
	ccName    string
	ccPath    string
	ccVersion string
}

func (req ContractInstallReq) do(v contracts.ContractInf) ContractResp {
	var resp ContractResp

	payload, err := v.Install(req.chainID, req.ccName, req.ccPath, req.ccVersion)
	resp = ContractResp{err, payload}
	return resp
}

type ContractDeployReq struct {
	chainID    string
	templateId []byte
	txid       string
	args       [][]byte
	timeout    time.Duration
}

func (req ContractDeployReq) do(v contracts.ContractInf) ContractResp {
	var resp ContractResp

	_, payload, err := v.Deploy(req.chainID, req.templateId, req.txid, req.args, req.timeout)
	resp = ContractResp{err, payload}
	return resp
}

type ContractInvokeReq struct {
	chainID  string
	deployId []byte
	txid     string
	args     [][]byte
	timeout  time.Duration
}

func (req ContractInvokeReq) do(v contracts.ContractInf) ContractResp {
	var resp ContractResp

	payload, err := v.Invoke(req.chainID, req.deployId, req.txid, req.args, req.timeout)
	resp = ContractResp{err, payload}
	return resp
}

type ContractStopReq struct {
	chainID     string
	deployId    []byte
	txid        string
	deleteImage bool
}

func (req ContractStopReq) do(v contracts.ContractInf) ContractResp {
	var resp ContractResp

	err := v.Stop(req.chainID, req.deployId, req.txid, req.deleteImage)
	resp = ContractResp{err, nil}
	return resp
}

func ContractProcess(req ContractReqInf) (interface{}, error) {
	c := make(chan struct{})

	//todo tmp
	v := &contracts.Contract{}

	var resp interface{}

	go func() {
		defer close(c)
		resp = req.do(v)
	}()

	select {
	case <-c:
		return resp, nil
	}

	return nil, nil
}





