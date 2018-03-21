package main

import (
	"./logger"
	util "./utils"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"math/big"
	"net"
	"sync"
	"time"
)

var cfg *util.Config

func log_init() {
	logger.SetConsole(cfg.Log.Console)
	logger.SetRollingFile(cfg.Log.Dir, cfg.Log.Name, cfg.Log.Num, cfg.Log.Size, logger.KB)
	//ALL，DEBUG，INFO，WARN，ERROR，FATAL，OFF
	logger.SetLevel(logger.ERROR)
	if cfg.Log.Level == "info" {
		logger.SetLevel(logger.INFO)
	} else if cfg.Log.Level == "error" {
		logger.SetLevel(logger.ERROR)
	}
}
func init() {
	cfg = &util.Config{}

	if !util.LoadConfig("seeker.toml", cfg) {
		return
	}
	log_init()
	// initialize()
}

const ua = "manspreading"
const ver = "1.0.0"

// statusData is the network packet for the status message.
type statusData struct {
	ProtocolVersion uint32
	NetworkId       uint64
	TD              *big.Int
	CurrentBlock    common.Hash
	GenesisBlock    common.Hash
}

// newBlockData is the network packet for the block propagation message.
type newBlockData struct {
	Block *types.Block
	TD    *big.Int
}

type conn struct {
	p  *p2p.Peer
	rw p2p.MsgReadWriter
}

type proxy struct {
	lock           sync.RWMutex
	upstreamNode   *discover.Node
	upstreamConn   *conn
	downstreamConn *conn
	upstreamState  statusData
	srv            *p2p.Server
}

var pxy *proxy
var upstreamUrl = "127.0.0.1:30304"
var listenAddr = "127.0.0.1:36666"
var privkey = ""

func test2() {
	var nodekey *ecdsa.PrivateKey
	if *privkey != "" {
		nodekey, _ = crypto.LoadECDSA(*privkey)
		fmt.Println("Node Key loaded from ", *privkey)
	} else {
		nodekey, _ = crypto.GenerateKey()
		crypto.SaveECDSA("./nodekey", nodekey)
		fmt.Println("Node Key generated and saved to ./nodekey")
	}

	node, _ := discover.ParseNode(*upstreamUrl)
	pxy = &proxy{
		upstreamNode: node,
	}

	config := p2p.Config{
		PrivateKey:     nodekey,
		MaxPeers:       2,
		NoDiscovery:    true,
		DiscoveryV5:    false,
		Name:           common.MakeName(fmt.Sprintf("%s/%s", ua, node.ID.String()), ver),
		BootstrapNodes: []*discover.Node{node},
		StaticNodes:    []*discover.Node{node},
		TrustedNodes:   []*discover.Node{node},

		Protocols: []p2p.Protocol{newManspreadingProtocol()},

		ListenAddr: *listenAddr,
		Logger:     log.New(),
	}
	config.Logger.SetHandler(log.StdoutHandler)

	pxy.srv = &p2p.Server{Config: config}

	// Wait forever
	var wg sync.WaitGroup
	wg.Add(2)
	err := pxy.srv.Start()
	wg.Done()
	if err != nil {
		fmt.Println(err)
	}
	wg.Wait()
}
func main() {
	test2()
	c := make(chan int, 1)

	<-c
}
