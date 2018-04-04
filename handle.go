package main

import (
	// "hash"
	// "golang.org/x/text"
	// "encoding/hex"
	"./logger"
	"fmt"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/p2p"
	// "github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/rlp"
	// "io"
	// "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	// "github.com/ethereum/go-ethereum/params"
)

func (pxy *proxy) handleStatus(p *p2p.Peer, msg p2p.Msg, rw p2p.MsgReadWriter) (err error) {
	var myMessage statusData
	err = msg.Decode(&myMessage)
	if err != nil {
		logger.Error("decode statusData err: ", err)
		return err
	}
	// fmt.Println("genesis:", myMessage.GenesisBlock.Hex())
	pxy.lock.Lock()
	// if myMessage.TD.Cmp(pxy.bestState.TD) > 0 {
	pxy.upstreamConn[p.ID()] = &conn{p, rw}
	pxy.lock.Unlock()
	logger.Info("add:", p.ID())
	// 	pxy.bestState = statusData{
	// 		ProtocolVersion: myMessage.ProtocolVersion,
	// 		NetworkId:       myMessage.NetworkId,
	// 		TD:              myMessage.TD,
	// 		CurrentBlock:    myMessage.CurrentBlock,
	// 		GenesisBlock:    myMessage.GenesisBlock,
	// 	}
	// 	fmt.Println("StatusMsg:", myMessage, " from ", p.RemoteAddr().String())
	// }
	// pxy.lock.Unlock()
	pxy.bestStateChan <- statusData{
		ProtocolVersion: myMessage.ProtocolVersion,
		NetworkId:       myMessage.NetworkId,
		TD:              myMessage.TD,
		CurrentBlock:    myMessage.CurrentBlock,
		GenesisBlock:    myMessage.GenesisBlock,
	}
	// pxy.bestHei = myMessage.Block.Number().Uint64()
	// pxy.bestHeiChan <- myMessage.Block.Number().Uint64()
	err = p2p.Send(rw, eth.StatusMsg, &statusData{
		ProtocolVersion: myMessage.ProtocolVersion,
		NetworkId:       myMessage.NetworkId,
		TD:              myMessage.TD,
		CurrentBlock:    myMessage.CurrentBlock,
		// GenesisBlock:    myMessage.GenesisBlock,
		GenesisBlock: myMessage.GenesisBlock,
		// ProtocolVersion: pxy.bestState.ProtocolVersion,
		// NetworkId:       pxy.bestState.NetworkId,
		// TD:              pxy.bestState.TD,
		// CurrentBlock:    pxy.bestState.CurrentBlock,
		// GenesisBlock:    pxy.bestState.GenesisBlock,
	})

	if err != nil {
		logger.Error("handshake err: ", err)
		return err
	}

	return nil
}
func (pxy *proxy) handleNewBlockMsg(p *p2p.Peer, msg p2p.Msg) (err error) {
	// fmt.Println("NewBlockMsg")
	{
		pxy.lock.RLock()
		_, ok := pxy.upstreamConn[p.ID()]
		pxy.lock.RUnlock()
		// pxy.lock.Lock()
		// defer pxy.lock.Unlock()
		if !ok {
			fmt.Println("NewBlockMsg:no id")
			return nil
		}
	}

	// // fmt.Println("NewBlockMsg2")
	// // fmt.Println("msg.Code: ", formateCode(msg.Code))
	var myMessage newBlockData
	err = msg.Decode(&myMessage)
	if err != nil {
		logger.Error("decode newBlockMsg err: ", err)
		return err
	}
	if p.ID() == pxy.upstreamNode.ID {
		logger.Info("newblockmsg:", myMessage.Block.Number().Text(10), " from bootnode", p.RemoteAddr().String())
	}
	// fmt.Println("NewBlockMsg xx:", myMessage.Block.Number(), " from ", p.RemoteAddr().String())

	{
		// pxy.lock.Lock()
		// defer pxy.lock.Unlock()
		// if myMessage.TD.Cmp(pxy.bestState.TD) > 0 {
		// 	pxy.bestState = statusData{
		// 		ProtocolVersion: gversion,
		// 		NetworkId:       gnetworkid,
		// 		TD:              myMessage.TD,
		// 		CurrentBlock:    myMessage.Block.Hash(),
		// 		GenesisBlock:    genesis,
		// 	}
		pxy.bestStateChan <- statusData{
			ProtocolVersion: gversion,
			NetworkId:       gnetworkid,
			TD:              myMessage.TD,
			CurrentBlock:    myMessage.Block.Hash(),
			GenesisBlock:    genesis,
		}
		// pxy.bestHei = myMessage.Block.Number().Uint64()
		pxy.bestHeiChan <- bestHeiPeer{myMessage.Block.Number().Uint64(), p}
		// fmt.Println("NewBlockMsg:", myMessage.Block.Number(), " from ", p.RemoteAddr().String())
		// }
		// pxy.lock.Unlock()
	}

	// need to re-encode msg
	size, r, err := rlp.EncodeToReader(myMessage)
	if err != nil {
		fmt.Println("encoding newBlockMsg err: ", err)
		return err
	}
	relay(p2p.Msg{Code: eth.NewBlockMsg, Size: uint32(size), Payload: r})
	return nil
}
func (pxy *proxy) handleBlockHeadersMsg(p *p2p.Peer, msg p2p.Msg) (err error) {
	fmt.Println("have BlockHeadersMsg")
	var headers []*types.Header
	if err := msg.Decode(&headers); err != nil {
		fmt.Println("handleBlockHeadersMsg:", err)
		return err
	}
	// If no headers were received, but we're expending a DAO fork check, maybe it's that
	// if len(headers) == 0 {
	// 	// Possibly an empty reply to the fork header checks, sanity check TDs
	// 	verifyDAO := true

	// 	// If we already have a DAO header, we can check the peer's TD against it. If
	// 	// the peer's ahead of this, it too must have a reply to the DAO check

	// 	daoHeader := core.BlockChain.GetHeaderByNumber(params.ChainConfig.DAOForkBlock.Uint64())
	// 	// if daoHeader := pm.blockchain.GetHeaderByNumber(pm.chainconfig.DAOForkBlock.Uint64()); daoHeader != nil {
	// 	if daoHeader != nil {
	// 		pxy.lock.RLock()
	// 		td := pxy.bestState.TD
	// 		pxy.lock.RUnlock()

	// 		if td.Cmp(core.BlockChain.GetTd(daoHeader.Hash(), daoHeader.Number.Uint64())) >= 0 {
	// 			verifyDAO = false
	// 		}
	// 	}
	// 	// If we're seemingly on the same chain, disable the drop timer
	// 	if verifyDAO {
	// 		fmt.Println("Seems to be on the same side of the DAO fork")
	// 		pxy.bestHeaderChan <- headers
	// 		return nil
	// 	}
	// }
	pxy.bestHeaderChan <- headers
	return nil
}
func (pxy *proxy) handleNewBlockHashesMsg(p *p2p.Peer, msg p2p.Msg) (err error) {
	// fmt.Println("NewBlockHashesMsg")
	// pxy.lock.Lock()
	{
		pxy.lock.RLock()
		_, ok := pxy.upstreamConn[p.ID()]
		pxy.lock.RUnlock()
		// defer pxy.lock.Unlock()
		if !ok {
			fmt.Println("NewBlockHashesMsg:no id")
			return nil
		}
	}
	// fmt.Println("NewBlockHashesMsg2")
	var announces newBlockHashesData
	if err := msg.Decode(&announces); err != nil {
		logger.Error("decoding NewBlockHashesMsg err: ", err)
		return err
	}
	// Mark the hashes as present at the remote node

	{
		// pxy.lock.Lock()
		for _, block := range announces {
			// fmt.Println("NewBlockHashesMsg xx:", block.Number, " p:", p.RemoteAddr().String(), " Caps:", p.Caps())
			// if block.Number > pxy.bestHei {
			// 	fmt.Println("NewBlockHashesMsg:", block.Number, " p:", p.RemoteAddr().String(), " Caps:", p.Caps())
			// 	pxy.bestHei = block.Number
			// }
			pxy.bestHeiChan2 <- bestHeiPeer{block.Number, p}
			// fmt.Println("NewBlockHashesMsg:", block.Number, " from:", p)
		}
		// pxy.lock.Unlock()
	}
	size, r, err := rlp.EncodeToReader(announces)
	if err != nil {
		fmt.Println("encoding NewBlockHashesMsg err: ", err)
		return err
	}
	relay(p2p.Msg{Code: eth.NewBlockHashesMsg, Size: uint32(size), Payload: r})
	return nil
}
