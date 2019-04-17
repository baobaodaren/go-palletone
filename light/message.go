package light

import (
	"fmt"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/dag/modules"
)

func (pm *ProtocolManager) StatusMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received status message")
	// Status messages should never arrive after the handshake
	return errResp(ErrExtraStatusMsg, "uncontrolled status message")
}

// Block header query, collect the requested headers and reply
func (pm *ProtocolManager) AnnounceMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received announce message")
	if p.requestAnnounceType == announceTypeNone {
		return errResp(ErrUnexpectedResponse, "")
	}

	var req announceData
	if err := msg.Decode(&req); err != nil {
		return errResp(ErrDecode, "%v: %v", msg, err)
	}

	if p.requestAnnounceType == announceTypeSigned {
		if err := req.checkSignature(p.pubKey); err != nil {
			log.Trace("Invalid announcement signature", "err", err)
			return err
		}
		log.Trace("Valid announcement signature")
	}

	log.Trace("Announce message content", "number", req.Number, "hash", req.Hash, "td", req.Td, "reorg", req.ReorgDepth)
	if pm.fetcher != nil {
		//pm.fetcher.announce(p, &req)
	}
	return nil
}

func (pm *ProtocolManager) GetBlockHeadersMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received block header request")
	return nil
	// Decode the complex header query
	//var req struct {
	//	ReqID uint64
	//	Query getBlockHeadersData
	//}
	//if err := msg.Decode(&req); err != nil {
	//	return errResp(ErrDecode, "%v: %v", msg, err)
	//}
	//
	//query := req.Query
	//if reject(query.Amount, MaxHeaderFetch) {
	//	return errResp(ErrRequestRejected, "")
	//}
	//
	//hashMode := query.Origin.Hash != (common.Hash{})
	//
	//// Gather headers until the fetch or network limits is reached
	//var (
	//	bytes   common.StorageSize
	//	headers []*types.Header
	//	unknown bool
	//)
	//for !unknown && len(headers) < int(query.Amount) && bytes < softResponseLimit {
	//	// Retrieve the next header satisfying the query
	//	var origin *types.Header
	//	if hashMode {
	//		origin = pm.blockchain.GetHeaderByHash(query.Origin.Hash)
	//	} else {
	//		origin = pm.blockchain.GetHeaderByNumber(query.Origin.Number)
	//	}
	//	if origin == nil {
	//		break
	//	}
	//	number := origin.Number.Uint64()
	//	headers = append(headers, origin)
	//	bytes += estHeaderRlpSize
	//
	//	// Advance to the next header of the query
	//	switch {
	//	case query.Origin.Hash != (common.Hash{}) && query.Reverse:
	//		// Hash based traversal towards the genesis block
	//		for i := 0; i < int(query.Skip)+1; i++ {
	//			if header := pm.blockchain.GetHeader(query.Origin.Hash, number); header != nil {
	//				query.Origin.Hash = header.ParentHash
	//				number--
	//			} else {
	//				unknown = true
	//				break
	//			}
	//		}
	//	case query.Origin.Hash != (common.Hash{}) && !query.Reverse:
	//		// Hash based traversal towards the leaf block
	//		if header := pm.blockchain.GetHeaderByNumber(origin.Number.Uint64() + query.Skip + 1); header != nil {
	//			if pm.blockchain.GetBlockHashesFromHash(header.Hash(), query.Skip+1)[query.Skip] == query.Origin.Hash {
	//				query.Origin.Hash = header.Hash()
	//			} else {
	//				unknown = true
	//			}
	//		} else {
	//			unknown = true
	//		}
	//	case query.Reverse:
	//		// Number based traversal towards the genesis block
	//		if query.Origin.Number >= query.Skip+1 {
	//			query.Origin.Number -= query.Skip + 1
	//		} else {
	//			unknown = true
	//		}
	//
	//	case !query.Reverse:
	//		// Number based traversal towards the leaf block
	//		query.Origin.Number += query.Skip + 1
	//	}
	//}
	//
	//bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + query.Amount*costs.reqCost)
	//pm.server.fcCostStats.update(msg.Code, query.Amount, rcost)
	//return p.SendBlockHeaders(req.ReqID, bv, headers)
}

func (pm *ProtocolManager) BlockHeadersMsg(msg p2p.Msg, p *peer) error {
	if pm.downloader == nil {
		return errResp(ErrUnexpectedResponse, "")
	}

	log.Trace("Received block header response message")
	// A batch of headers arrived to one of our previous requests
	var resp struct {
		ReqID, BV uint64
		Headers   []*modules.Header
	}
	if err := msg.Decode(&resp); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	p.fcServer.GotReply(resp.ReqID, resp.BV)
	if pm.fetcher != nil && pm.fetcher.requestedID(resp.ReqID) {
		pm.fetcher.deliverHeaders(p, resp.ReqID, resp.Headers)
	} else {
		err := pm.downloader.DeliverHeaders(p.id, resp.Headers)
		if err != nil {
			log.Debug(fmt.Sprint(err))
		}
	}
	return nil
}

/*
func (pm *ProtocolManager) GetBlockBodiesMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received block bodies request")
	// Decode the retrieval message
	var req struct {
		ReqID  uint64
		Hashes []common.Hash
	}
	if err := msg.Decode(&req); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	// Gather blocks until the fetch or network limits is reached
	var (
		bytes  int
		bodies []rlp.RawValue
	)
	reqCnt := len(req.Hashes)
	if reject(uint64(reqCnt), MaxBodyFetch) {
		return errResp(ErrRequestRejected, "")
	}
	for _, hash := range req.Hashes {
		if bytes >= softResponseLimit {
			break
		}
		// Retrieve the requested block body, stopping if enough was found
		if data := core.GetBodyRLP(pm.chainDb, hash, core.GetBlockNumber(pm.chainDb, hash)); len(data) != 0 {
			bodies = append(bodies, data)
			bytes += len(data)
		}
	}
	bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
	pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)
	return p.SendBlockBodiesRLP(req.ReqID, bv, bodies)

}

func (pm *ProtocolManager) BlockBodiesMsg(msg p2p.Msg, p *peer) error {
	if pm.odr == nil {
		return errResp(ErrUnexpectedResponse, "")
	}

	log.Trace("Received block bodies response")
	// A batch of block bodies arrived to one of our previous requests
	var resp struct {
		ReqID, BV uint64
		Data      []*types.Body
	}
	if err := msg.Decode(&resp); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	p.fcServer.GotReply(resp.ReqID, resp.BV)
	deliverMsg := &Msg{
		MsgType: MsgBlockBodies,
		ReqID:   resp.ReqID,
		Obj:     resp.Data,
	}
	return nil
}

func (pm *ProtocolManager) GetCodeMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received code request")
	// Decode the retrieval message
	var req struct {
		ReqID uint64
		Reqs  []CodeReq
	}
	if err := msg.Decode(&req); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	// Gather state data until the fetch or network limits is reached
	var (
		bytes int
		data  [][]byte
	)
	reqCnt := len(req.Reqs)
	if reject(uint64(reqCnt), MaxCodeFetch) {
		return errResp(ErrRequestRejected, "")
	}
	for _, req := range req.Reqs {
		// Retrieve the requested state entry, stopping if enough was found
		if header := core.GetHeader(pm.chainDb, req.BHash, core.GetBlockNumber(pm.chainDb, req.BHash)); header != nil {
			statedb, err := pm.blockchain.State()
			if err != nil {
				continue
			}
			account, err := pm.getAccount(statedb, header.Root, common.BytesToHash(req.AccKey))
			if err != nil {
				continue
			}
			code, _ := statedb.Database().TrieDB().Node(common.BytesToHash(account.CodeHash))

			data = append(data, code)
			if bytes += len(code); bytes >= softResponseLimit {
				break
			}
		}
	}
	bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
	pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)
	return p.SendCode(req.ReqID, bv, data)
}

func (pm *ProtocolManager) CodeMsg(msg p2p.Msg, p *peer) error {
	if pm.odr == nil {
		return errResp(ErrUnexpectedResponse, "")
	}

	log.Trace("Received code response")
	// A batch of node state data arrived to one of our previous requests
	var resp struct {
		ReqID, BV uint64
		Data      [][]byte
	}
	if err := msg.Decode(&resp); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	p.fcServer.GotReply(resp.ReqID, resp.BV)
	deliverMsg := &Msg{
		MsgType: MsgCode,
		ReqID:   resp.ReqID,
		Obj:     resp.Data,
	}
}

func (pm *ProtocolManager) GetProofsMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received proofs request")
	// Decode the retrieval message
	var req struct {
		ReqID uint64
		Reqs  []ProofReq
	}
	if err := msg.Decode(&req); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	// Gather state data until the fetch or network limits is reached
	var (
		bytes  int
		proofs proofsData
	)
	reqCnt := len(req.Reqs)
	if reject(uint64(reqCnt), MaxProofsFetch) {
		return errResp(ErrRequestRejected, "")
	}
	for _, req := range req.Reqs {
		// Retrieve the requested state entry, stopping if enough was found
		if header := core.GetHeader(pm.chainDb, req.BHash, core.GetBlockNumber(pm.chainDb, req.BHash)); header != nil {
			statedb, err := pm.blockchain.State()
			if err != nil {
				continue
			}
			var trie state.Trie
			if len(req.AccKey) > 0 {
				account, err := pm.getAccount(statedb, header.Root, common.BytesToHash(req.AccKey))
				if err != nil {
					continue
				}
				trie, _ = statedb.Database().OpenStorageTrie(common.BytesToHash(req.AccKey), account.Root)
			} else {
				trie, _ = statedb.Database().OpenTrie(header.Root)
			}
			if trie != nil {
				var proof light.NodeList
				trie.Prove(req.Key, 0, &proof)

				proofs = append(proofs, proof)
				if bytes += proof.DataSize(); bytes >= softResponseLimit {
					break
				}
			}
		}
	}
	bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
	pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)
	return p.SendProofs(req.ReqID, bv, proofs)
}

func (pm *ProtocolManager) ProofsMsg(msg p2p.Msg, p *peer) error {
	if pm.odr == nil {
		return errResp(ErrUnexpectedResponse, "")
	}

	log.Trace("Received proofs response")
	// A batch of merkle proofs arrived to one of our previous requests
	var resp struct {
		ReqID, BV uint64
		Data      []les.NodeList
	}
	if err := msg.Decode(&resp); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	p.fcServer.GotReply(resp.ReqID, resp.BV)
	deliverMsg := &Msg{
		MsgType: MsgProofsV1,
		ReqID:   resp.ReqID,
		Obj:     resp.Data,
	}
}

func (pm *ProtocolManager) GetHeaderProofsMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received headers proof request")
	// Decode the retrieval message
	var req struct {
		ReqID uint64
		Reqs  []ChtReq
	}
	if err := msg.Decode(&req); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	// Gather state data until the fetch or network limits is reached
	var (
		bytes  int
		proofs []ChtResp
	)
	reqCnt := len(req.Reqs)
	if reject(uint64(reqCnt), MaxHelperTrieProofsFetch) {
		return errResp(ErrRequestRejected, "")
	}
	trieDb := trie.NewDatabase(ethdb.NewTable(pm.chainDb, light.ChtTablePrefix))
	for _, req := range req.Reqs {
		if header := pm.blockchain.GetHeaderByNumber(req.BlockNum); header != nil {
			sectionHead := core.GetCanonicalHash(pm.chainDb, req.ChtNum*light.CHTFrequencyServer-1)
			if root := light.GetChtRoot(pm.chainDb, req.ChtNum-1, sectionHead); root != (common.Hash{}) {
				trie, err := trie.New(root, trieDb)
				if err != nil {
					continue
				}
				var encNumber [8]byte
				binary.BigEndian.PutUint64(encNumber[:], req.BlockNum)

				var proof light.NodeList
				trie.Prove(encNumber[:], 0, &proof)

				proofs = append(proofs, ChtResp{Header: header, Proof: proof})
				if bytes += proof.DataSize() + estHeaderRlpSize; bytes >= softResponseLimit {
					break
				}
			}
		}
	}
	bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
	pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)
	return p.SendHeaderProofs(req.ReqID, bv, proofs)
}

func (pm *ProtocolManager) GetHelperTrieProofsMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received helper trie proof request")
	// Decode the retrieval message
	var req struct {
		ReqID uint64
		Reqs  []HelperTrieReq
	}
	if err := msg.Decode(&req); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	// Gather state data until the fetch or network limits is reached
	var (
		auxBytes int
		auxData  [][]byte
	)
	reqCnt := len(req.Reqs)
	if reject(uint64(reqCnt), MaxHelperTrieProofsFetch) {
		return errResp(ErrRequestRejected, "")
	}

	var (
		lastIdx  uint64
		lastType uint
		root     common.Hash
		auxTrie  *trie.Trie
	)
	nodes := light.NewNodeSet()
	for _, req := range req.Reqs {
		if auxTrie == nil || req.Type != lastType || req.TrieIdx != lastIdx {
			auxTrie, lastType, lastIdx = nil, req.Type, req.TrieIdx

			var prefix string
			if root, prefix = pm.getHelperTrie(req.Type, req.TrieIdx); root != (common.Hash{}) {
				auxTrie, _ = trie.New(root, trie.NewDatabase(ethdb.NewTable(pm.chainDb, prefix)))
			}
		}
		if req.AuxReq == auxRoot {
			var data []byte
			if root != (common.Hash{}) {
				data = root[:]
			}
			auxData = append(auxData, data)
			auxBytes += len(data)
		} else {
			if auxTrie != nil {
				auxTrie.Prove(req.Key, req.FromLevel, nodes)
			}
			if req.AuxReq != 0 {
				data := pm.getHelperTrieAuxData(req)
				auxData = append(auxData, data)
				auxBytes += len(data)
			}
		}
		if nodes.DataSize()+auxBytes >= softResponseLimit {
			break
		}
	}
	bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
	pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)
	return p.SendHelperTrieProofs(req.ReqID, bv, HelperTrieResps{Proofs: nodes.NodeList(), AuxData: auxData})

}

func (pm *ProtocolManager) HeaderProofsMsg(msg p2p.Msg, p *peer) error {
	if pm.odr == nil {
		return errResp(ErrUnexpectedResponse, "")
	}

	log.Trace("Received headers proof response")
	var resp struct {
		ReqID, BV uint64
		Data      []ChtResp
	}
	if err := msg.Decode(&resp); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	p.fcServer.GotReply(resp.ReqID, resp.BV)
	deliverMsg := &Msg{
		MsgType: MsgHeaderProofs,
		ReqID:   resp.ReqID,
		Obj:     resp.Data,
	}
}

func (pm *ProtocolManager) HelperTrieProofsMsg(msg p2p.Msg, p *peer) error {
	if pm.odr == nil {
		return errResp(ErrUnexpectedResponse, "")
	}

	log.Trace("Received helper trie proof response")
	var resp struct {
		ReqID, BV uint64
		Data      HelperTrieResps
	}
	if err := msg.Decode(&resp); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	p.fcServer.GotReply(resp.ReqID, resp.BV)
	deliverMsg := &Msg{
		MsgType: MsgHelperTrieProofs,
		ReqID:   resp.ReqID,
		Obj:     resp.Data,
	}

}

func (pm *ProtocolManager) SendTxMsg(msg p2p.Msg, p *peer) error {
	if pm.txpool == nil {
		return errResp(ErrRequestRejected, "")
	}
	// Transactions arrived, parse all of them and deliver to the pool
	var txs []*types.Transaction
	if err := msg.Decode(&txs); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	reqCnt := len(txs)
	if reject(uint64(reqCnt), MaxTxSend) {
		return errResp(ErrRequestRejected, "")
	}
	pm.txpool.AddRemotes(txs)

	_, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
	pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)
}

func (pm *ProtocolManager) GetTxStatusMsg(msg p2p.Msg, p *peer) error {
	if pm.txpool == nil {
		return errResp(ErrUnexpectedResponse, "")
	}
	// Transactions arrived, parse all of them and deliver to the pool
	var req struct {
		ReqID  uint64
		Hashes []common.Hash
	}
	if err := msg.Decode(&req); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	reqCnt := len(req.Hashes)
	if reject(uint64(reqCnt), MaxTxStatus) {
		return errResp(ErrRequestRejected, "")
	}
	bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
	pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)

	return p.SendTxStatus(req.ReqID, bv, pm.txStatus(req.Hashes))
}

func (pm *ProtocolManager) TxStatusMsg(msg p2p.Msg, p *peer) error {
	if pm.odr == nil {
		return errResp(ErrUnexpectedResponse, "")
	}

	log.Trace("Received tx status response")
	var resp struct {
		ReqID, BV uint64
		Status    []txStatus
	}
	if err := msg.Decode(&resp); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	p.fcServer.GotReply(resp.ReqID, resp.BV)
	return nil
}
*/
