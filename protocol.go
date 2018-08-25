package main

import (
	// "hash"
	// "golang.org/x/text"
	// "encoding/hex"
	"./logger"
	"fmt"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	// "github.com/ethereum/go-ethereum/rlp"
	"io"
)

func newManspreadingProtocol() p2p.Protocol {
	return p2p.Protocol{
		Name:    eth.ProtocolName,
		Version: eth.ProtocolVersions[0],
		Length:  eth.ProtocolLengths[0],
		Run:     pxy.handle,
		NodeInfo: func() interface{} {
			fmt.Println("Noop: NodeInfo called")
			return nil
		},
		PeerInfo: func(id discover.NodeID) interface{} {
			fmt.Println("Noop: PeerInfo called")
			return nil
		},
	}
}

func (pxy *proxy) handle(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	// logger.Info("peers:", pxy.srv.Peers())
	//先处理dao分叉的问题
	// DAOForkBlock:=big.NewInt(1920000)
	// if daoBlock := DAOForkBlock; daoBlock != nil {
	// 	// Request the peer's DAO fork header for extra-data validation
	// 	if err := p.RequestHeadersByNumber(daoBlock.Uint64(), 1, 0, false); err != nil {
	// 		fmt.Println("RequestHeadersByNumber:",err)
	// 		return err
	// 	}
	// 	// Start a timer to disconnect if the peer doesn't reply in time
	// 	daoChallengeTimeout:=15 * time.Second
	// 	p.forkDrop = time.AfterFunc(daoChallengeTimeout, func() {
	// 		p.Log().Debug("Timed out DAO fork-check, dropping")
	// 		// pm.removePeer(p.id)
	// 	})
	// 	// Make sure it's cleaned up if the peer dies off
	// 	defer func() {
	// 		if p.forkDrop != nil {
	// 			p.forkDrop.Stop()
	// 			p.forkDrop = nil
	// 		}
	// 	}()
	// }
	for {
		msg, err := rw.ReadMsg()
		if err != nil {
			if err == io.EOF {
				pxy.lock.Lock()

				if _, ok := pxy.upstreamConn[p.ID()]; ok {
					delete(pxy.upstreamConn, p.ID())
					logger.Error("delete:", p.ID())
				}
				pxy.lock.Unlock()
			}

			return err
		}
		switch msg.Code {
		case eth.StatusMsg:
			pxy.handleStatus(p, msg, rw)
		case eth.NewBlockMsg:
			pxy.handleNewBlockMsg(p, msg)
		case eth.NewBlockHashesMsg:
			pxy.handleNewBlockHashesMsg(p, msg)
		// case eth.BlockHeadersMsg:
		// 	pxy.handleBlockHeadersMsg(p, msg)
		default:
			break
		}
	}
	return nil
}

// func handle1(p *p2p.Peer, rw p2p.MsgReadWriter) error {
// 	logger.Info("peers:", pxy.srv.Peers())
// 	for {
// 		msg, err := rw.ReadMsg()
// 		if err != nil {
// 			if err == io.EOF {
// 				pxy.lock.Lock()

// 				if _, ok := pxy.upstreamConn[p.ID()]; ok {
// 					delete(pxy.upstreamConn, p.ID())
// 				}
// 				pxy.lock.Unlock()
// 			}

// 			return err
// 		}
// 		switch msg.Code {
// 		case eth.StatusMsg:
// 			var myMessage statusData
// 			err = msg.Decode(&myMessage)
// 			if err != nil {
// 				fmt.Println("decode statusData err: ", err)
// 				return err
// 			}
// 			fmt.Println("genesis:", myMessage.GenesisBlock.Hex())
// 			// fmt.Println("statusData:",myMessage)
// 			pxy.lock.RLock()
// 			_, ok := pxy.upstreamConn[p.ID()]
// 			pxy.lock.RUnlock()
// 			//如果连接不存在
// 			if !ok {
// 				pxy.lock.Lock()
// 				if myMessage.TD.Cmp(pxy.bestState.TD) > 0 {
// 					pxy.upstreamConn[p.ID()] = &conn{p, rw}
// 					pxy.bestState = statusData{
// 						ProtocolVersion: myMessage.ProtocolVersion,
// 						NetworkId:       myMessage.NetworkId,
// 						TD:              myMessage.TD,
// 						CurrentBlock:    myMessage.CurrentBlock,
// 						GenesisBlock:    myMessage.GenesisBlock,
// 					}
// 					// fmt.Println("StatusMsg:", myMessage.TD.Text(16), " from ", p.RemoteAddr().String())
// 					fmt.Println("StatusMsg:", myMessage, " from ", p.RemoteAddr().String())
// 				}
// 				pxy.lock.Unlock()
// 				fmt.Println("NOT EXISTS")
// 			}
// 			// pxy.lock.Unlock()
// 			err = p2p.Send(rw, eth.StatusMsg, &statusData{
// 				ProtocolVersion: myMessage.ProtocolVersion,
// 				NetworkId:       myMessage.NetworkId,
// 				TD:              myMessage.TD,
// 				CurrentBlock:    myMessage.CurrentBlock,
// 				GenesisBlock:    myMessage.GenesisBlock,
// 				// ProtocolVersion: pxy.bestState.ProtocolVersion,
// 				// NetworkId:       pxy.bestState.NetworkId,
// 				// TD:              pxy.bestState.TD,
// 				// CurrentBlock:    pxy.bestState.CurrentBlock,
// 				// GenesisBlock:    pxy.bestState.GenesisBlock,
// 			})

// 			if err != nil {
// 				fmt.Println("handshake err: ", err)
// 				return err
// 			}
// 		case eth.NewBlockMsg:
// 			fmt.Println("NewBlockMsg")
// 			// {
// 			// 	pxy.lock.RLock()
// 			// 	_, ok := pxy.upstreamConn[p.ID()]
// 			// 	pxy.lock.RUnlock()
// 			// 	// pxy.lock.Lock()
// 			// 	// defer pxy.lock.Unlock()
// 			// 	if !ok {
// 			// 		fmt.Println("NewBlockMsg:no id")
// 			// 		return nil
// 			// 	}
// 			// }

// 			// fmt.Println("NewBlockMsg2")
// 			// fmt.Println("msg.Code: ", formateCode(msg.Code))
// 			var myMessage newBlockData
// 			err = msg.Decode(&myMessage)
// 			if err != nil {
// 				fmt.Println("decode newBlockMsg err: ", err)
// 				return err
// 			}
// 			if p.ID() == pxy.upstreamNode.ID {
// 				fmt.Println("newblockmsg:", myMessage.Block.Number().Text(10), " from bootnode", p.RemoteAddr().String())
// 			}
// 			fmt.Println("NewBlockMsg xx:", myMessage.Block.Number(), " from ", p.RemoteAddr().String())

// 			{
// 				pxy.lock.Lock()
// 				// defer pxy.lock.Unlock()
// 				if myMessage.TD.Cmp(pxy.bestState.TD) > 0 {
// 					pxy.bestState = statusData{
// 						ProtocolVersion: gversion,
// 						NetworkId:       gnetworkid,
// 						TD:              myMessage.TD,
// 						CurrentBlock:    myMessage.Block.Hash(),
// 						GenesisBlock:    genesis,
// 					}
// 					pxy.bestHei = myMessage.Block.Number().Uint64()
// 					fmt.Println("NewBlockMsg:", myMessage.Block.Number(), " from ", p.RemoteAddr().String())
// 				}
// 				pxy.lock.Unlock()
// 			}

// 			// need to re-encode msg
// 			size, r, err := rlp.EncodeToReader(myMessage)
// 			if err != nil {
// 				fmt.Println("encoding newBlockMsg err: ", err)
// 			}
// 			relay(p2p.Msg{Code: eth.NewBlockMsg, Size: uint32(size), Payload: r})
// 		case eth.NewBlockHashesMsg:
// 			fmt.Println("NewBlockHashesMsg")
// 			// pxy.lock.Lock()
// 			{
// 				pxy.lock.RLock()
// 				_, ok := pxy.upstreamConn[p.ID()]
// 				pxy.lock.RUnlock()
// 				// defer pxy.lock.Unlock()
// 				if !ok {
// 					fmt.Println("NewBlockHashesMsg:no id")
// 					return nil
// 				}
// 			}
// 			fmt.Println("NewBlockHashesMsg2")
// 			var announces newBlockHashesData
// 			if err := msg.Decode(&announces); err != nil {
// 				fmt.Println("decoding NewBlockHashesMsg err: ", err)
// 				return err
// 			}
// 			// Mark the hashes as present at the remote node

// 			{
// 				pxy.lock.Lock()
// 				for _, block := range announces {
// 					fmt.Println("NewBlockHashesMsg xx:", block.Number, " p:", p.RemoteAddr().String(), " Caps:", p.Caps())
// 					if block.Number > pxy.bestHei {
// 						fmt.Println("NewBlockHashesMsg:", block.Number, " p:", p.RemoteAddr().String(), " Caps:", p.Caps())
// 						pxy.bestHei = block.Number
// 					}
// 				}
// 				pxy.lock.Unlock()
// 			}
// 			// pxy.lock.Unlock()
// 		default:
// 			break
// 		}
// 	}
// 	return nil
// }
// func handle2(p *p2p.Peer, rw p2p.MsgReadWriter) error {
// 	// fmt.Println("Run called")
// 	// fmt.Println("peers:",pxy.srv.Peers())
// 	logger.Info("peers:", pxy.srv.Peers())
// 	for {
// 		// fmt.Println("Waiting for msg...")
// 		msg, err := rw.ReadMsg()

// 		if err != nil {
// 			// fmt.Println("readMsg err: ", err)

// 			if err == io.EOF {
// 				// fmt.Println(fromWhom(p.ID().String()), " has dropped its connection...")
// 				pxy.lock.Lock()
// 				// if p.ID() == pxy.upstreamNode.ID {
// 				// 	pxy.upstreamConn = nil
// 				// } else {
// 				// 	pxy.downstreamConn = nil
// 				// }
// 				if _, ok := pxy.upstreamConn[p.ID()]; ok {
// 					delete(pxy.upstreamConn, p.ID())
// 				}
// 				// if _, ok := pxy.upstreamState[p.ID()]; ok {
// 				// 	delete(pxy.upstreamState, p.ID())
// 				// }
// 				pxy.lock.Unlock()
// 			}

// 			return err
// 		}
// 		if msg.Code == eth.TxMsg {
// 			return err
// 		}
// 		// fmt.Println("Got a msg from: ", fromWhom(p.ID().String()[:8]))
// 		// fmt.Println("msg.Code: ", formateCode(msg.Code))

// 		if msg.Code == eth.StatusMsg { // handshake
// 			var myMessage statusData
// 			err = msg.Decode(&myMessage)
// 			if err != nil {
// 				fmt.Println("decode statusData err: ", err)
// 				return err
// 			}

// 			// fmt.Println("ProtocolVersion: ", myMessage.ProtocolVersion)
// 			// fmt.Println("NetworkId:       ", myMessage.NetworkId)
// 			// fmt.Println("TD:              ", myMessage.TD)
// 			// fmt.Println("CurrentBlock:    ", myMessage.CurrentBlock.Hex())
// 			// fmt.Println("GenesisBlock:    ", myMessage.GenesisBlock.Hex())

// 			pxy.lock.Lock()
// 			// if p.ID() == pxy.upstreamNode.ID {
// 			// 	pxy.upstreamState = myMessage
// 			// 	pxy.upstreamConn = &conn{p, rw}
// 			// } else {
// 			// 	pxy.downstreamConn = &conn{p, rw}
// 			// }
// 			fmt.Println("genesis:", myMessage.GenesisBlock.Hex())
// 			if _, ok := pxy.upstreamConn[p.ID()]; !ok {
// 				// delete(pxy.upstreamConnm, p.ID())
// 				// pxy.upstreamConn=append(pxy.upstreamConn,)

// 				if myMessage.TD.Cmp(pxy.bestState.TD) > 0 && myMessage.GenesisBlock.Hex() == pxy.bestState.GenesisBlock.Hex() {
// 					pxy.upstreamConn[p.ID()] = &conn{p, rw}
// 					// pxy.maxtd = myMessage.TD
// 					// pxy.bestHash = myMessage.CurrentBlock
// 					// pxy.bestState.TD=myMessage.TD
// 					pxy.bestState = statusData{
// 						ProtocolVersion: myMessage.ProtocolVersion,
// 						NetworkId:       myMessage.NetworkId,
// 						TD:              myMessage.TD,
// 						CurrentBlock:    myMessage.CurrentBlock,
// 						GenesisBlock:    myMessage.GenesisBlock,
// 					}
// 					// fmt.Println("StatusMsg:", myMessage.TD.Text(16), " from ", p.RemoteAddr().String())
// 					fmt.Println("StatusMsg:", myMessage, " from ", p.RemoteAddr().String())
// 				}

// 			}
// 			// if val, ok := pxy.upstreamState[p.ID()]; ok {
// 			// 	delete(pxy.upstreamState, p.ID())
// 			// }
// 			// pxy.upstreamState[p.ID()] = myMessage
// 			// temptd := pxy.maxtd
// 			// temphash := pxy.bestHash
// 			err = p2p.Send(rw, eth.StatusMsg, &statusData{
// 				ProtocolVersion: myMessage.ProtocolVersion,
// 				NetworkId:       myMessage.NetworkId,
// 				TD:              myMessage.TD,
// 				CurrentBlock:    myMessage.CurrentBlock,
// 				GenesisBlock:    myMessage.GenesisBlock,
// 			})

// 			if err != nil {
// 				fmt.Println("handshake err: ", err)
// 				return err
// 			}
// 		} else if msg.Code == eth.NewBlockMsg {
// 			pxy.lock.Lock()
// 			// defer pxy.lock.Unlock()
// 			if _, ok := pxy.upstreamConn[p.ID()]; !ok {
// 				pxy.lock.Unlock()
// 				return nil
// 			}
// 			// fmt.Println("msg.Code: ", formateCode(msg.Code))
// 			var myMessage newBlockData
// 			err = msg.Decode(&myMessage)
// 			if err != nil {
// 				fmt.Println("decode newBlockMsg err: ", err)
// 				return err
// 			}
// 			// fmt.Println("newblockmsg:", myMessage.Block.Number(), " coinbase:", myMessage.Block.Coinbase().Hex(), " extra:", string(myMessage.Block.Extra()[:]))
// 			// fmt.Println("newblockmsg:", myMessage.Block.Number().Text(10), " coinbase:", myMessage.Block.Coinbase().Hex())
// 			// if myMessage.Block.Number() > startBlock {
// 			// if myMessage.TD > startTD {
// 			// if myMessage.TD.Cmp(startTD) > 0 {
// 			// 	pxy.upstreamConn[p.ID()] = &conn{p, rw}
// 			// 	startTD = myMessage.TD
// 			// 	startBlock = myMessage.Block.Hash()
// 			// 	fmt.Println("newblockmsg:", myMessage.Block.Number().Text(10), " from ", p.RemoteAddr().String())
// 			// }
// 			// fmt.Println("newblockmsg:", myMessage.Block.Number().Text(10), " from ", p.RemoteAddr().String())
// 			// startBlock = myMessage.Block.Number()
// 			// startTD = myMessage.TD
// 			// }
// 			if p.ID() == pxy.upstreamNode.ID {
// 				// pxy.upstreamState.CurrentBlock = myMessage.Block.Hash()
// 				// pxy.upstreamState.TD = myMessage.TD
// 				fmt.Println("newblockmsg:", myMessage.Block.Number().Text(10), " from bootnode", p.RemoteAddr().String())
// 			}
// 			fmt.Println("NewBlockMsg xx:", myMessage.Block.Number(), " from ", p.RemoteAddr().String())

// 			//TODO: handle newBlock from downstream
// 			// if _, ok := pxy.upstreamConn[p.ID()]; ok {
// 			// delete(pxy.upstreamState, p.ID())

// 			if myMessage.TD.Cmp(pxy.bestState.TD) > 0 {
// 				// pxy.upstreamState[p.ID()].CurrentBlock = myMessage.Block.Hash()
// 				// pxy.upstreamState[p.ID()].TD = myMessage.TD
// 				// old := pxy.upstreamState[p.ID()]
// 				pxy.bestState = statusData{
// 					// ProtocolVersion: val.ProtocolVersion,
// 					// NetworkId:       val.NetworkId,
// 					TD:           myMessage.TD,
// 					CurrentBlock: myMessage.Block.Hash(),
// 					GenesisBlock: genesis,
// 				}
// 				pxy.bestHei = myMessage.Block.Number().Uint64()
// 				fmt.Println("NewBlockMsg:", myMessage.Block.Number(), " from ", p.RemoteAddr().String())
// 			}
// 			// }
// 			pxy.lock.Unlock()

// 			// need to re-encode msg
// 			size, r, err := rlp.EncodeToReader(myMessage)
// 			if err != nil {
// 				fmt.Println("encoding newBlockMsg err: ", err)
// 			}
// 			relay(p2p.Msg{Code: eth.NewBlockMsg, Size: uint32(size), Payload: r})

// 		} else if msg.Code == eth.NewBlockHashesMsg {
// 			pxy.lock.Lock()
// 			// defer pxy.lock.Unlock()
// 			if _, ok := pxy.upstreamConn[p.ID()]; !ok {
// 				pxy.lock.Unlock()
// 				return nil
// 			}
// 			var announces newBlockHashesData
// 			if err := msg.Decode(&announces); err != nil {
// 				fmt.Println("decoding NewBlockHashesMsg err: ", err)
// 				return err
// 			}
// 			// Mark the hashes as present at the remote node

// 			// pxy.lock.Lock()
// 			for _, block := range announces {
// 				fmt.Println("NewBlockHashesMsg xx:", block.Number, " p:", p.RemoteAddr().String(), " Caps:", p.Caps())
// 				if block.Number > pxy.bestHei {
// 					fmt.Println("NewBlockHashesMsg:", block.Number, " p:", p.RemoteAddr().String(), " Caps:", p.Caps())
// 					pxy.bestHei = block.Number
// 				}
// 			}
// 			pxy.lock.Unlock()
// 		}
// 	}
// 	return nil
// }

// func relay2(p *p2p.Peer, msg p2p.Msg) {
// 	var err error
// 	pxy.lock.RLock()
// 	defer pxy.lock.RUnlock()
// 	if p.ID() != pxy.upstreamNode.ID && pxy.upstreamConn != nil {
// 		err = pxy.upstreamConn.rw.WriteMsg(msg)
// 	} else if p.ID() == pxy.upstreamNode.ID && pxy.downstreamConn != nil {
// 		err = pxy.downstreamConn.rw.WriteMsg(msg)
// 	} else {
// 		fmt.Println("One of upstream/downstream isn't alive: ", pxy.srv.Peers())
// 	}

// 	if err != nil {
// 		fmt.Println("relaying err: ", err)
// 	}
// }
func relay(msg p2p.Msg) {
	var err error
	connmap := make(map[discover.NodeID]*conn)
	pxy.lock.RLock()
	for key, value := range pxy.upstreamConn {
		connmap[key] = value
	}
	pxy.lock.RUnlock()
	// if p.ID() != pxy.upstreamNode.ID && pxy.upstreamConn != nil {
	// 	err = pxy.upstreamConn.rw.WriteMsg(msg)
	// } else if p.ID() == pxy.upstreamNode.ID && pxy.downstreamConn != nil {
	// 	err = pxy.downstreamConn.rw.WriteMsg(msg)
	// } else {
	// 	fmt.Println("One of upstream/downstream isn't alive: ", pxy.srv.Peers())
	// }
	for _, v := range connmap {
		// if v.p.c
		err = v.rw.WriteMsg(msg)
		if err != nil {
			logger.Error("relaying err: ", err)
		} else {
			// logger.Info("send:", v.p.RemoteAddr().String())
		}

	}

}

// func (pxy *proxy) upstreamAlive() bool {
// 	for _, peer := range pxy.srv.Peers() {
// 		if peer.ID() == pxy.upstreamNode.ID {
// 			return true
// 		}
// 	}
// 	return false
// }

func fromWhom(nodeId string) string {
	// if nodeId == pxy.upstreamNode.ID.String() {
	// 	return "upstream:" + nodeId
	// } else {
	// 	return "downstream:" + nodeId
	// }
	return nodeId
}
func formateCode(code uint64) (ret string) {
	// StatusMsg          = 0x00
	// 	NewBlockHashesMsg  = 0x01
	// 	TxMsg              = 0x02
	// 	GetBlockHeadersMsg = 0x03
	// 	BlockHeadersMsg    = 0x04
	// 	GetBlockBodiesMsg  = 0x05
	// 	BlockBodiesMsg     = 0x06
	// 	NewBlockMsg        = 0x07

	// 	// Protocol messages belonging to eth/63
	// 	GetNodeDataMsg = 0x0d
	// 	NodeDataMsg    = 0x0e
	// 	GetReceiptsMsg = 0x0f
	// 	ReceiptsMsg    = 0x10
	switch code {
	case eth.StatusMsg:
		ret = "StatusMsg"
	case eth.NewBlockHashesMsg:
		ret = "NewBlockHashesMsg"
	case eth.TxMsg:
		ret = "TxMsg"
	case eth.GetBlockHeadersMsg:
		ret = "GetBlockHeadersMsg"
	case eth.BlockHeadersMsg:
		ret = "BlockHeadersMsg"
	case eth.GetBlockBodiesMsg:
		ret = "GetBlockBodiesMsg"
	case eth.BlockBodiesMsg:
		ret = "BlockBodiesMsg"
	case eth.NewBlockMsg:
		ret = "NewBlockMsg"
	case eth.GetNodeDataMsg:
		ret = "GetNodeDataMsg"
	case eth.NodeDataMsg:
		ret = "NodeDataMsg"
	case eth.GetReceiptsMsg:
		ret = "GetReceiptsMsg"
	case eth.ReceiptsMsg:
		ret = "ReceiptsMsg"
	default:
		ret = "unknown"
	}
	return
}
