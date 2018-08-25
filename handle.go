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
	// "github.com/ethereum/go-ethereum/core/types"
	// "github.com/ethereum/go-ethereum/params"
	ethpeer "./ethpeer"
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
	// pxy.allPeer[p.ID()] = p
	// NewPeer(version int, p *p2p.Peer, rw p2p.MsgReadWriter)
	pp := ethpeer.NewPeer(myMessage.ProtocolVersion, p, rw)
	err = pxy.ethpeerset.Register(pp)
	if err != nil {
		// fmt.Println("pxy.ethpeerset.Register(pp):",err)
	}
	ethpeer := pxy.ethpeerset.Peer(p.ID().String())
	if ethpeer != nil {
		ethpeer.SetHead(myMessage.CurrentBlock, myMessage.TD)
		ethpeer.SetGenesis(myMessage.GenesisBlock)
	}

	pxy.lock.Unlock()
	logger.Info("add:", p.ID())

	pxy.bestStateChan <- statusData{
		ProtocolVersion: myMessage.ProtocolVersion,
		NetworkId:       myMessage.NetworkId,
		TD:              myMessage.TD,
		CurrentBlock:    myMessage.CurrentBlock,
		GenesisBlock:    myMessage.GenesisBlock,
	}
	err = p2p.Send(rw, eth.StatusMsg, &statusData{
		ProtocolVersion: myMessage.ProtocolVersion,
		NetworkId:       myMessage.NetworkId,
		TD:              myMessage.TD,
		CurrentBlock:    myMessage.CurrentBlock,
		// GenesisBlock:    myMessage.GenesisBlock,
		GenesisBlock: myMessage.GenesisBlock,
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

	var myMessage newBlockData
	err = msg.Decode(&myMessage)
	if err != nil {
		logger.Error("decode newBlockMsg err: ", err)
		return err
	}
	if p.ID() == pxy.upstreamNode.ID {
		logger.Info("newblockmsg:", myMessage.Block.Number().Text(10), " from bootnode", p.RemoteAddr().String())
	}
	{
		pxy.bestStateChan <- statusData{
			ProtocolVersion: gversion,
			NetworkId:       gnetworkid,
			TD:              myMessage.TD,
			CurrentBlock:    myMessage.Block.Hash(),
			GenesisBlock:    genesis,
		}
		pxy.bestHeiChan <- bestHeiPeer{bestHei: myMessage.Block.NumberU64(), p: p}
	}
	// myMessage.Block=pxy.bestHeiAndPeer.bestHei
	myMessage.TD = pxy.bestState.TD
	// need to re-encode msg
	size, r, err := rlp.EncodeToReader(myMessage)
	if err != nil {
		fmt.Println("encoding newBlockMsg err: ", err)
		return err
	}
	relay(p2p.Msg{Code: eth.NewBlockMsg, Size: uint32(size), Payload: r})
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
	// announces.Block=pxy.bestHeiAndPeer.bestHei
	// announces.TD=pxy.bestState.TD

	size, r, err := rlp.EncodeToReader(announces)
	if err != nil {
		fmt.Println("encoding NewBlockHashesMsg err: ", err)
		return err
	}
	msgs := p2p.Msg{Code: eth.NewBlockHashesMsg, Size: uint32(size), Payload: r}
	err = pxy.upstreamConn[p.ID()].rw.WriteMsg(msgs)
	if err != nil {
		logger.Error("relaying err: ", err)
	} else {
		logger.Info("send:", p.RemoteAddr().String())
	}
	relay(msgs)

	return nil
}
