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
 *  * @date 2018-2019
 *
 */

package sysconfigcc

import (
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"sort"
	"strconv"
	"time"
)

type SysConfigChainCode struct {
}

func (s *SysConfigChainCode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	log.Info("*** SysConfigChainCode system contract init ***")
	return shim.Success([]byte("Success"))
}

func (s *SysConfigChainCode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "getAllSysParamsConf":
		log.Info("Start getAllSysParamsConf Invoke")
		resultByte, err := s.getAllSysParamsConf(stub)
		if err != nil {
			jsonResp := "{\"Error\":\"getAllSysParamsConf err: " + err.Error() + "\"}"
			return shim.Error(jsonResp)
		}
		return shim.Success(resultByte)
	case "getSysParamValByKey":
		log.Info("Start getSysParamValByKey Invoke")
		resultByte, err := s.getSysParamValByKey(stub, args)
		if err != nil {
			jsonResp := "{\"Error\":\"getSysParamValByKey err: " + err.Error() + "\"}"
			return shim.Error(jsonResp)
		}
		return shim.Success(resultByte)
	case "updateSysParamWithoutVote":
		log.Info("Start updateSysParamWithoutVote Invoke")
		resultByte, err := s.updateSysParamWithoutVote(stub, args)
		if err != nil {
			jsonResp := "{\"Error\":\"updateSysParamWithoutVote err: " + err.Error() + "\"}"
			return shim.Error(jsonResp)
		}
		return shim.Success(resultByte)
	case "getWithoutVoteResult":
		log.Info("Start getWithoutVoteResult Invoke")
		resultByte, err := stub.GetState(sysParam)
		if err != nil {
			jsonResp := "{\"Error\":\"getWithoutVoteResult err: " + err.Error() + "\"}"
			return shim.Success([]byte(jsonResp))
		}
		return shim.Success(resultByte)
	case "getVotesResult":
		log.Info("Start getVotesResult Invoke")
		resultByte, err := s.getVotesResult(stub, args)
		if err != nil {
			jsonResp := "{\"Error\":\"getVotesResult err: " + err.Error() + "\"}"
			return shim.Success([]byte(jsonResp))
		}
		return shim.Success(resultByte)
	case "createVotesTokens":
		log.Info("Start createVotesTokens Invoke")
		resultByte, err := s.createVotesTokens(stub, args)
		if err != nil {
			jsonResp := "{\"Error\":\"createVotesTokens err: " + err.Error() + "\"}"
			return shim.Success([]byte(jsonResp))
		}
		return shim.Success(resultByte)
	case "nodesVote":
		log.Info("Start nodesVote Invoke")
		resultByte, err := s.nodesVote(stub, args)
		if err != nil {
			jsonResp := "{\"Error\":\"nodesVote err: " + err.Error() + "\"}"
			return shim.Success([]byte(jsonResp))
		}
		return shim.Success(resultByte)
	default:
		log.Error("Invoke funcName err: ", "error", funcName)
		jsonResp := "{\"Error\":\"Invoke funcName err: " + funcName + "\"}"
		return shim.Error(jsonResp)
	}
}

func (s *SysConfigChainCode) getVotesResult(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//params check
	//if len(args) < 1 {
	//	return nil, fmt.Errorf("need 1 args (AssetID String)")
	//}

	//assetIDStr
	//assetIDStr := strings.ToUpper(args[0])
	//check name is exist or not
	tkInfo := getSymbols(stub)
	if tkInfo == nil {
		return nil, fmt.Errorf("Token not exist")
	}

	//get token information
	var topicSupports []SysTopicSupports
	err := json.Unmarshal(tkInfo.VoteContent, &topicSupports)
	if err != nil {
		return nil, fmt.Errorf("Results format invalid, Error!!!")
	}

	//
	isVoteEnd := false
	headerTime, err := stub.GetTxTimestamp(10)
	if err != nil {
		return nil, fmt.Errorf("GetTxTimestamp invalid, Error!!!")
	}
	if headerTime.Seconds > tkInfo.VoteEndTime.Unix() {
		isVoteEnd = true
	}
	//calculate result
	var supportResults []*modules.SysSupportResult
	for i, oneTopicSupport := range topicSupports {
		oneResult := &modules.SysSupportResult{}
		oneResult.TopicIndex = uint64(i) + 1
		oneResult.TopicTitle = oneTopicSupport.TopicTitle
		oneResultSort := sortSupportByCount(oneTopicSupport.VoteResults)
		oneResult.VoteResults = append(oneResult.VoteResults, oneResultSort...)
		//for i := uint64(0); i < oneTopicSupport.SelectMax; i++ {
		//	oneResult.VoteResults = append(oneResult.VoteResults, oneResultSort[i])
		//}
		supportResults = append(supportResults, oneResult)
	}

	//token
	asset := tkInfo.AssetID
	tkID := modules.SysTokenIDInfo{IsVoteEnd: isVoteEnd, CreateAddr: tkInfo.CreateAddr, TotalSupply: tkInfo.TotalSupply,
		SupportResults: supportResults, AssetID: asset.String(), CreateTime: tkInfo.VoteEndTime.UTC(), LeastNum: tkInfo.LeastNum}

	//return json
	tkJson, err := json.Marshal(tkID)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return tkJson, nil //test
}

func (s *SysConfigChainCode) createVotesTokens(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//params check
	if len(args) < 5 {
		return nil, fmt.Errorf("need 5 args (Name,VoteType,TotalSupply,VoteEndTime,VoteContentJson)")
	}
	//get creator
	createAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	//TODO 基金会地址
	//foundationAddress, _ := stub.GetSystemConfig("FoundationAddress")
	//if createAddr != foundationAddress {
	//	jsonResp := "{\"Error\":\"Only foundation can call this function\"}"
	//	return nil, fmt.Errorf(jsonResp)
	//}
	//==== convert params to token information
	var vt modules.VoteToken
	//name symbol
	vt.Name = args[0]
	vt.Symbol = "SVOTE"

	//vote type
	//if args[1] == "0" {
	//	vt.VoteType = byte(0)
	//} else if args[1] == "1" {
	//	vt.VoteType = byte(1)
	//} else if args[1] == "2" {
	//	vt.VoteType = byte(2)
	//} else {
	//	jsonResp := "{\"Error\":\"Only string, 0 or 1 or 2\"}"
	//	return shim.Success([]byte(jsonResp))
	//}
	//total supply
	totalSupply, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to convert total supply\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	if totalSupply == 0 {
		jsonResp := "{\"Error\":\"Can't be zero\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	vt.TotalSupply = totalSupply
	leastNum, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to convert least numbers\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	if leastNum == 0 {
		jsonResp := "{\"Error\":\"Can't be zero\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	//VoteEndTime
	VoteEndTime, err := time.Parse("2006-01-02 15:04:05", args[3])
	if err != nil {
		jsonResp := "{\"Error\":\"No vote end time\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	vt.VoteEndTime = VoteEndTime
	//VoteContent
	var voteTopics []SysVoteTopic
	err = json.Unmarshal([]byte(args[4]), &voteTopics)
	if err != nil {
		jsonResp := "{\"Error\":\"VoteContent format invalid\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	//init support
	var supports []SysTopicSupports
	for _, oneTopic := range voteTopics {
		var oneSupport SysTopicSupports
		oneSupport.TopicTitle = oneTopic.TopicTitle
		for _, oneOption := range oneTopic.SelectOptions {
			oneResult := &modules.SysVoteResult{}
			oneResult.SelectOption = oneOption
			oneSupport.VoteResults = append(oneSupport.VoteResults, oneResult)
		}
		//oneResult.SelectOptionsNum = uint64(len(oneRequest.SelectOptions))
		oneSupport.SelectMax = oneTopic.SelectMax
		supports = append(supports, oneSupport)
	}
	voteContentJson, err := json.Marshal(supports)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate voteContent Json\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	vt.VoteContent = voteContentJson

	txid := stub.GetTxID()
	assetID, _ := modules.NewAssetId(vt.Symbol, modules.AssetType_VoteToken,
		0, common.Hex2Bytes(txid[2:]), modules.UniqueIdType_Null)
	//assetIDStr := assetID.String()
	//check name is only or not
	tkInfo := getSymbols(stub)
	if tkInfo != nil {
		jsonResp := "{\"Error\":\"Repeat AssetID\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	//convert to json
	createJson, err := json.Marshal(vt)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate token Json\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	//last put state
	info := SysTokenInfo{vt.Name, vt.Symbol, createAddr.String(), leastNum, totalSupply,
		VoteEndTime, voteContentJson, assetID}

	err = setSymbols(stub, &info)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to set symbols\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	//set token define
	err = stub.DefineToken(byte(modules.AssetType_VoteToken), createJson, createAddr.String())
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.DefineToken\"}"
		return nil, fmt.Errorf(jsonResp)
	}

	return createJson, nil //test
}

func (s *SysConfigChainCode) nodesVote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//params check
	if len(args) < 1 {
		return nil, fmt.Errorf("need 1 args (SupportRequestJson)")
	}

	//check token
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return nil, fmt.Errorf("GetInvokeTokens failed")
	}
	voteNum := uint64(0)
	assetIDStr := ""
	for i := 0; i < len(invokeTokens); i++ {
		if invokeTokens[i].Asset.AssetId == modules.PTNCOIN {
			continue
		} else if invokeTokens[i].Address == "P1111111111111111111114oLvT2" {
			if assetIDStr == "" {
				assetIDStr = invokeTokens[i].Asset.String()
				voteNum += invokeTokens[i].Amount
			} else if invokeTokens[i].Asset.AssetId.String() == assetIDStr {
				voteNum += invokeTokens[i].Amount
			}
		}
	}
	if voteNum == 0 || assetIDStr == "" { //no vote token
		return nil, fmt.Errorf("Vote token empty")
	}

	//check name is exist or not
	tkInfo := getSymbols(stub)
	if tkInfo == nil {
		return nil, fmt.Errorf("Token not exist")
	}

	//parse support requests
	var supportRequests []SysSupportRequest
	err = json.Unmarshal([]byte(args[0]), &supportRequests)
	if err != nil {
		return nil, fmt.Errorf("SupportRequestJson format invalid")
	}
	//get token information
	var topicSupports []SysTopicSupports
	err = json.Unmarshal(tkInfo.VoteContent, &topicSupports)
	if err != nil {
		return nil, fmt.Errorf("Results format invalid, Error!!!")

	}

	if voteNum < uint64(len(supportRequests)) { //vote token more than request
		return nil, fmt.Errorf("Vote token more than support request")

	}

	//check time
	headerTime, err := stub.GetTxTimestamp(10)
	if err != nil {
		return nil, fmt.Errorf("GetTxTimestamp invalid, Error!!!")

	}
	if headerTime.Seconds > tkInfo.VoteEndTime.Unix() {
		return nil, fmt.Errorf("Vote is over")

	}

	//save support
	indexHistory := make(map[uint64]uint8)
	indexRepeat := false
	for _, oneSupport := range supportRequests {
		topicIndex := oneSupport.TopicIndex - 1
		if _, ok := indexHistory[topicIndex]; ok { //check select repeat
			indexRepeat = true
			break
		}
		indexHistory[topicIndex] = 1
		if topicIndex < uint64(len(topicSupports)) { //1.check index, must not out of total
			if uint64(len(oneSupport.SelectIndexs)) <= topicSupports[topicIndex].SelectMax { //2.check one select's options, must not out of select's max
				lenOfVoteResult := uint64(len(topicSupports[topicIndex].VoteResults))
				selIndexHistory := make(map[uint64]uint8)
				for _, index := range oneSupport.SelectIndexs {
					selectIndex := index - 1
					if _, ok := selIndexHistory[selectIndex]; ok { //check select repeat
						break
					}
					selIndexHistory[selectIndex] = 1
					if selectIndex < lenOfVoteResult { //3.index must be real select options
						topicSupports[topicIndex].VoteResults[selectIndex].Num += 1
					}
				}
			}
		}
	}
	if indexRepeat {
		return nil, fmt.Errorf("Repeat index of select option ")

	}
	voteContentJson, err := json.Marshal(topicSupports)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate voteContent Json")

	}
	tkInfo.VoteContent = voteContentJson

	//save token information
	err = setSymbols(stub, tkInfo)
	if err != nil {
		return nil, fmt.Errorf("Failed to set symbols")

	}
	return []byte("NodesVote success."), nil
}

func (s *SysConfigChainCode) getAllSysParamsConf(stub shim.ChaincodeStubInterface) ([]byte, error) {
	sysVal, err := stub.GetState("sysConf")
	if err != nil {
		return nil, err
	}
	return sysVal, nil
}

func (s *SysConfigChainCode) updateSysParamWithoutVote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//invokeFromAddr, err := stub.GetInvokeAddress()
	//if err != nil {
	//	return nil, err
	//}
	//TODO 基金会地址
	//foundationAddress, _ := stub.GetSystemConfig("FoundationAddress")
	//if invokeFromAddr != foundationAddress {
	//	jsonResp := "{\"Error\":\"Only foundation can call this function\"}"
	//	return nil, fmt.Errorf(jsonResp)
	//}
	//key := args[0]
	//newValue := args[1]
	//oldValue, err := stub.GetState(args[0])
	//if err != nil {
	//	return nil, err
	//}
	//err = stub.PutState(key, []byte(newValue))
	//if err != nil {
	//	return nil, err
	//}
	//sysValByte, err := stub.GetState("sysConf")
	//if err != nil {
	//	return nil, err
	//}
	//sysVal := &core.SystemConfig{}
	//err = json.Unmarshal(sysValByte, sysVal)
	//if err != nil {
	//	return nil, err
	//}
	//switch key {
	//case "DepositAmountForJury":
	//	sysVal.DepositAmountForJury = newValue
	//case "DepositRate":
	//	sysVal.DepositRate = newValue
	//case "FoundationAddress":
	//	sysVal.FoundationAddress = newValue
	//case "DepositAmountForMediator":
	//	sysVal.DepositAmountForMediator = newValue
	//case "DepositAmountForDeveloper":
	//	sysVal.DepositAmountForDeveloper = newValue
	//case "DepositPeriod":
	//	sysVal.DepositPeriod = newValue
	//case "RootCAHolder":
	//	sysVal.RootCAHolder = newValue
	//}
	//sysValByte, err = json.Marshal(sysVal)
	//if err != nil {
	//	return nil, err
	//}
	//err = stub.PutState("sysConf", sysValByte)
	//if err != nil {
	//	return nil, err
	//}
	//jsonResp := "{\"Success\":\"update value from " + string(oldValue) + " to " + newValue + "\"}"
	//return []byte(jsonResp), nil

	//TODO mediator 换届时的相关处理
	modify := &modules.FoundModify{}
	modify.Key = args[0]
	modify.Value = args[1]
	resultBytes, err := stub.GetState(sysParam)
	if err != nil {
		return nil, err
	}
	var modifies []*modules.FoundModify
	if resultBytes == nil {
		modifies = append(modifies, modify)
	} else {
		err := json.Unmarshal(resultBytes, &modifies)
		if err != nil {
			return nil, err
		}
		modifies = append(modifies, modify)
	}
	modifyByte, err := json.Marshal(modifies)
	err = stub.PutState(sysParam, modifyByte)
	if err != nil {
		return nil, err
	}
	return []byte(modifyByte), nil
}

func (s *SysConfigChainCode) getSysParamValByKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		jsonResp := "{\"Error\":\" need 1 args (AssetID String)\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	val, err := stub.GetSystemConfig(args[0])
	//val, err := stub.GetState(args[0])
	if err != nil {
		return nil, err
	}
	jsonResp := "{\"" + args[0] + "\":\"" + string(val) + "\"}"
	return []byte(jsonResp), nil
}

func getSymbols(stub shim.ChaincodeStubInterface) *SysTokenInfo {
	//
	tkInfo := SysTokenInfo{}
	//TODO
	//tkInfoBytes, _ := stub.GetState(symbolsKey + assetID)
	tkInfoBytes, _ := stub.GetState(sysParams)
	if len(tkInfoBytes) == 0 {
		return nil
	}
	//
	err := json.Unmarshal(tkInfoBytes, &tkInfo)
	if err != nil {
		return nil
	}
	return &tkInfo
}
func setSymbols(stub shim.ChaincodeStubInterface, tkInfo *SysTokenInfo) error {
	val, err := json.Marshal(tkInfo)
	if err != nil {
		return err
	}
	err = stub.PutState(sysParams, val)
	return err
}

// A slice of TopicResult that implements sort.Interface to sort by Value.
type VoteResultList []*modules.SysVoteResult

func (p VoteResultList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p VoteResultList) Len() int           { return len(p) }
func (p VoteResultList) Less(i, j int) bool { return p[i].Num > p[j].Num }

// A function to turn a map into a TopicResultList, then sort and return it.
func sortSupportByCount(tpl VoteResultList) VoteResultList {
	sort.Stable(tpl) //sort.Sort(tpl)
	return tpl
}
