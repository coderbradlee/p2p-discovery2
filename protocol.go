package main

import (
	// "golang.org/x/text"
	// "encoding/hex"
	"./logger"
	"fmt"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/rlp"
	"io"
)

func newManspreadingProtocol() p2p.Protocol {
	return p2p.Protocol{
		Name:    eth.ProtocolName,
		Version: eth.ProtocolVersions[0],
		Length:  eth.ProtocolLengths[0],
		Run:     handle,
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

func handle(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	// fmt.Println("Run called")
	// fmt.Println("peers:",pxy.srv.Peers())
	logger.Info("peers:", pxy.srv.Peers())
	for {
		// fmt.Println("Waiting for msg...")
		msg, err := rw.ReadMsg()

		if err != nil {
			// fmt.Println("readMsg err: ", err)

			if err == io.EOF {
				// fmt.Println(fromWhom(p.ID().String()), " has dropped its connection...")
				pxy.lock.Lock()
				// if p.ID() == pxy.upstreamNode.ID {
				// 	pxy.upstreamConn = nil
				// } else {
				// 	pxy.downstreamConn = nil
				// }
				if _, ok := pxy.upstreamConn[p.ID()]; ok {
					delete(pxy.upstreamConn, p.ID())
				}
				if _, ok := pxy.upstreamState[p.ID()]; ok {
					delete(pxy.upstreamState, p.ID())
				}
				pxy.lock.Unlock()
			}

			return err
		}
		if msg.Code == eth.TxMsg {
			return err
		}
		// fmt.Println("Got a msg from: ", fromWhom(p.ID().String()[:8]))
		// fmt.Println("msg.Code: ", formateCode(msg.Code))

		if msg.Code == eth.StatusMsg { // handshake
			var myMessage statusData
			err = msg.Decode(&myMessage)
			if err != nil {
				fmt.Println("decode statusData err: ", err)
				return err
			}

			// fmt.Println("ProtocolVersion: ", myMessage.ProtocolVersion)
			// fmt.Println("NetworkId:       ", myMessage.NetworkId)
			// fmt.Println("TD:              ", myMessage.TD)
			// fmt.Println("CurrentBlock:    ", myMessage.CurrentBlock.Hex())
			// fmt.Println("GenesisBlock:    ", myMessage.GenesisBlock.Hex())

			pxy.lock.Lock()
			// if p.ID() == pxy.upstreamNode.ID {
			// 	pxy.upstreamState = myMessage
			// 	pxy.upstreamConn = &conn{p, rw}
			// } else {
			// 	pxy.downstreamConn = &conn{p, rw}
			// }
			if _, ok := pxy.upstreamConn[p.ID()]; !ok {
				// delete(pxy.upstreamConnm, p.ID())
				// pxy.upstreamConn=append(pxy.upstreamConn,)
				pxy.upstreamConn[p.ID()] = &conn{p, rw}
			}
			// if val, ok := pxy.upstreamState[p.ID()]; ok {
			// 	delete(pxy.upstreamState, p.ID())
			// }
			pxy.upstreamState[p.ID()] = myMessage
			pxy.lock.Unlock()

			err = p2p.Send(rw, eth.StatusMsg, &statusData{
				ProtocolVersion: myMessage.ProtocolVersion,
				NetworkId:       myMessage.NetworkId,
				TD:              startTD,
				CurrentBlock:    startBlock,
				GenesisBlock:    myMessage.GenesisBlock,
			})

			if err != nil {
				fmt.Println("handshake err: ", err)
				return err
			}
		} else if msg.Code == eth.NewBlockMsg {
			fmt.Println("msg.Code: ", formateCode(msg.Code))
			var myMessage newBlockData
			err = msg.Decode(&myMessage)
			if err != nil {
				fmt.Println("decode newBlockMsg err: ", err)
				return err
			}
			// fmt.Println("newblockmsg:", myMessage.Block.Number(), " coinbase:", myMessage.Block.Coinbase().Hex(), " extra:", string(myMessage.Block.Extra()[:]))
			// fmt.Println("newblockmsg:", myMessage.Block.Number().Text(10), " coinbase:", myMessage.Block.Coinbase().Hex())
			fmt.Println("newblockmsg:", myMessage.Block.Number().Text(10), " from ", p.RemoteAddr().String())
			if p.ID() == pxy.upstreamNode.ID {
				// pxy.upstreamState.CurrentBlock = myMessage.Block.Hash()
				// pxy.upstreamState.TD = myMessage.TD
				fmt.Println("newblockmsg:", myMessage.Block.Number().Text(10), " from bootnode", p.RemoteAddr().String())
			}
			pxy.lock.Lock()
			//TODO: handle newBlock from downstream
			if val, ok := pxy.upstreamState[p.ID()]; ok {
				// delete(pxy.upstreamState, p.ID())

				if myMessage.TD.Cmp(pxy.upstreamState[p.ID()].TD) > 0 {
					// pxy.upstreamState[p.ID()].CurrentBlock = myMessage.Block.Hash()
					// pxy.upstreamState[p.ID()].TD = myMessage.TD
					// old := pxy.upstreamState[p.ID()]
					pxy.upstreamState[p.ID()] = statusData{
						ProtocolVersion: val.ProtocolVersion,
						NetworkId:       val.NetworkId,
						TD:              myMessage.TD,
						CurrentBlock:    myMessage.Block.Hash(),
						GenesisBlock:    val.GenesisBlock,
					}
				}
			}
			pxy.lock.Unlock()

			// need to re-encode msg
			size, r, err := rlp.EncodeToReader(myMessage)
			if err != nil {
				fmt.Println("encoding newBlockMsg err: ", err)
			}
			relay(p2p.Msg{Code: eth.NewBlockMsg, Size: uint32(size), Payload: r})

		} else if msg.Code == eth.NewBlockHashesMsg {
			var announces newBlockHashesData
			if err := msg.Decode(&announces); err != nil {
				fmt.Println("decoding NewBlockHashesMsg err: ", err)
				return err
			}
			// Mark the hashes as present at the remote node
			for _, block := range announces {
				// p.MarkBlock(block.Hash)
				fmt.Println("NewBlockHashesMsg:", block.Number," from ",block.origin, " p:", p.RemoteAddr().String())
			}

		}

	}
	return nil
}

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
	pxy.lock.RLock()
	defer pxy.lock.RUnlock()
	// if p.ID() != pxy.upstreamNode.ID && pxy.upstreamConn != nil {
	// 	err = pxy.upstreamConn.rw.WriteMsg(msg)
	// } else if p.ID() == pxy.upstreamNode.ID && pxy.downstreamConn != nil {
	// 	err = pxy.downstreamConn.rw.WriteMsg(msg)
	// } else {
	// 	fmt.Println("One of upstream/downstream isn't alive: ", pxy.srv.Peers())
	// }
	for _, v := range pxy.upstreamConn {
		// if v.p.c
		err = v.rw.WriteMsg(msg)
		if err != nil {
			logger.Error("relaying err: ", err)
		} else {
			logger.Info("send:", v.p.RemoteAddr().String())
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
