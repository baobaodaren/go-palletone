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
package manger

import (
	"time"
	"net"
	"os"
	"sync"

	"encoding/hex"
	"google.golang.org/grpc"
	"github.com/spf13/viper"
	"github.com/golang/protobuf/proto"

	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/contracts/accesscontrol"
	"github.com/palletone/go-palletone/contracts/scc"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/common"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/crypto"
	"github.com/pkg/errors"
	"fmt"
)

type CCInfo struct {
	Id      []byte
	Name    string
	Path    string
	Version string

	SysCC  bool
	Enable bool
}

type chain struct {
	version int
	cclist  map[string]*CCInfo
}

var chains = struct {
	sync.RWMutex
	clist map[string]*chain
}{clist: make(map[string]*chain)}

type oldSysCCInfo struct {
	origSystemCC       []*scc.SystemChaincode
	origSysCCWhitelist map[string]string
}

func (osyscc *oldSysCCInfo) reset() {
	scc.MockResetSysCCs(osyscc.origSystemCC)
	viper.Set("chaincode.system", osyscc.origSysCCWhitelist)
}

func chainsInit() {
	chains.clist = nil
	chains.clist = make(map[string]*chain)
}

func addChainCodeInfo(c *chain, cc *CCInfo) error {
	if c == nil || cc == nil {
		return errors.New("chain or ccinfo is nil")
	}

	for k, v := range c.cclist {
		if k == cc.Name && v.Version == cc.Version{
			logger.Errorf("chaincode [%s] , version[%d] already exit, %v", cc.Name, cc.Version, v)
			return errors.New("already exit chaincode")
		}
	}
	c.cclist[cc.Name] = cc

	return nil
}

func setChaincode(cid string, version int, chaincode *CCInfo) error {
	chains.Lock()
	defer chains.Unlock()

	for k, v := range chains.clist {
		if k == cid {
			logger.Errorf("chainId[%s] already exit, %v", cid, v)

			return addChainCodeInfo(v, chaincode)
		}
	}
	cNew := chain{
		version:version,
		cclist:make(map[string]*CCInfo),
	}
	chains.clist[cid] = &cNew

	return addChainCodeInfo(&cNew, chaincode)
}

func getChaincodeList(cid string) (*chain, error) {
	if cid == "" {
		return nil, errors.New("param is nil")
	}

	if chains.clist[cid] != nil {
		return chains.clist[cid], nil
	}
	errmsg := fmt.Sprintf("not find chainId[%s] in chains", cid)

	return nil, errors.New(errmsg)
}

func delChaincode(cid string, ccName string, version int) (error) {
	if cid == "" || ccName == "" {
		return  errors.New("param is nil")
	}

	if chains.clist[cid] != nil {
		for k, _ := range chains.clist[cid].cclist {
			if k == ccName {
				chains.clist[cid].cclist[k] = nil
				logger.Infof("del chaincode[%s]", ccName)
				return nil
			}
		}
	}
	logger.Infof("not find chaincode[%s]", ccName)

	return nil
}

func marshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}

// CreateChaincodeProposalWithTxIDNonceAndTransient creates a proposal from given input
func createChaincodeProposalWithTxIDNonceAndTransient(txid string, typ common.HeaderType, chainID string, cis *peer.ChaincodeInvocationSpec, nonce, creator []byte, transientMap map[string][]byte) (*peer.Proposal, string, error) {
	// get a more appropriate mechanism to handle it in.
	var epoch uint64 = 0

	ccHdrExt := &peer.ChaincodeHeaderExtension{ChaincodeId: cis.ChaincodeSpec.ChaincodeId}
	ccHdrExtBytes, err := proto.Marshal(ccHdrExt)
	if err != nil {
		return nil, "", err
	}

	cisBytes, err := proto.Marshal(cis)
	if err != nil {
		return nil, "", err
	}

	ccPropPayload := &peer.ChaincodeProposalPayload{Input: cisBytes, TransientMap: transientMap}
	ccPropPayloadBytes, err := proto.Marshal(ccPropPayload)
	if err != nil {
		return nil, "", err
	}

	timestamp := util.CreateUtcTimestamp()
	hdr := &common.Header{ChannelHeader: marshalOrPanic(&common.ChannelHeader{
		Type:      int32(typ),
		TxId:      txid,
		Timestamp: timestamp,
		ChannelId: chainID,
		Extension: ccHdrExtBytes,
		Epoch:     epoch}),
		SignatureHeader: marshalOrPanic(&common.SignatureHeader{Nonce: nonce, Creator: creator})}

	hdrBytes, err := proto.Marshal(hdr)
	if err != nil {
		return nil, "", err
	}

	return &peer.Proposal{Header: hdrBytes, Payload: ccPropPayloadBytes}, txid, nil
}

func computeProposalTxID(nonce, creator []byte) (string, error) {
	opdata := append(nonce, creator...)
	digest := util.ComputeSHA256(opdata)

	return hex.EncodeToString(digest), nil
}

func createChaincodeProposalWithTransient(typ common.HeaderType, chainID string, txid string, cis *peer.ChaincodeInvocationSpec, creator []byte, transientMap map[string][]byte) (*peer.Proposal, string, error) {
	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, "", err
	}
	// compute txid
	//txid, err := computeProposalTxID(nonce, creator)
	//if err != nil {
	//	return nil, "", err
	//}

	return createChaincodeProposalWithTxIDNonceAndTransient(txid, typ, chainID, cis, nonce, creator, transientMap)
}

func createChaincodeProposal(typ common.HeaderType, chainID string, txid string, cis *peer.ChaincodeInvocationSpec, creator []byte) (*peer.Proposal, string, error) {
	return createChaincodeProposalWithTransient(typ, chainID, txid, cis, creator, nil)
}

func GetBytesProposal(prop *peer.Proposal) ([]byte, error) {
	propBytes, err := proto.Marshal(prop)
	return propBytes, err
}

func signedEndorserProposa(chainID string, txid string, cs *peer.ChaincodeSpec, creator, signature []byte) (*peer.SignedProposal, *peer.Proposal, error) {
	prop, _, err := createChaincodeProposal(
		common.HeaderType_ENDORSER_TRANSACTION,
		chainID,
		txid,
		&peer.ChaincodeInvocationSpec{ChaincodeSpec: cs},
		creator)
	if err != nil {
		return nil, nil, err
	}

	propBytes, err := GetBytesProposal(prop)
	if err != nil {
		return nil, nil, err
	}

	return &peer.SignedProposal{ProposalBytes: propBytes, Signature: signature}, prop, nil
}

func peerCreateChain(cid string) error {
	chains.Lock()
	defer chains.Unlock()

	//chains.list[cid] = &chain{
	//	//cs: &chainSupport{
	//	//},
	//}

	return nil
}

func peerServerInit() error {
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	peerAddress := viper.GetString("peer.address")
	if peerAddress == "" {
		peerAddress = "0.0.0.0:21726"
	}

	lis, err := net.Listen("tcp", peerAddress)
	if err != nil {
		return err
	}
	ccStartupTimeout := time.Duration(30) * time.Second
	ca, _ := accesscontrol.NewCA()
	pb.RegisterChaincodeSupportServer(grpcServer, core.NewChaincodeSupport(peerAddress, false, ccStartupTimeout, ca))
	go grpcServer.Serve(lis)

	return nil
}

func peerServerDeInit() error {
	defer os.RemoveAll("/home/glh/tmp/chaincodes")
	return nil
}

func systemContractInit() error {
	chainID := util.GetTestChainID()
	peerCreateChain(chainID)
	scc.RegisterSysCCs()
	scc.DeploySysCCs(chainID)
	return nil
}

func systemContractDeInit() error {
	chainID := util.GetTestChainID()
	scc.DeDeploySysCCs(chainID)
	return nil
}

func systemContractStop() error {

	return nil
}
