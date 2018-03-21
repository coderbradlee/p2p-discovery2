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
	// "net"
	"sync"
	// "time"
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

const (
	ua          = "manspreading"
	ver         = "1.0.0"
	upstreamUrl = "enode://344d2d76587b931a8dccb61f5f3280c9486068ef2758252cf5c6ebc29d4385581137c45e2c218e4ee23a0b14d23ecb6ec12521362e9919380c3b00ff5401bea2@10.81.64.116:30304" //geth2
	// upstreamUrl = "enode://2998c333662a61620126e8a5a44545b8c0b362ec8a89b246a3e2e15a076983525e148ef113152d2836b976fb8de860b03f997012793870d78ae0a56e565d8398@118.31.112.214:30304" //getf1

	listenAddr = "0.0.0.0:36666"
	privkey    = ""
)

// statusData is the network packet for the status message.
type statusData struct {
	ProtocolVersion uint32
	NetworkId       uint64
	TD              *big.Int
	CurrentBlock    common.Hash
	GenesisBlock    common.Hash
}
type newBlockHashesData []struct {
	Hash   common.Hash // Hash of one particular block being announced
	Number uint64      // Number of one particular block being announced
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

func test2() {
	var nodekey *ecdsa.PrivateKey
	if privkey != "" {
		nodekey, _ = crypto.LoadECDSA(privkey)
		fmt.Println("Node Key loaded from ", privkey)
	} else {
		nodekey, _ = crypto.GenerateKey()
		crypto.SaveECDSA("./nodekey", nodekey)
		fmt.Println("Node Key generated and saved to ./nodekey")
	}

	node, err := discover.ParseNode(upstreamUrl)
	if err != nil {
		fmt.Println("discover.ParseNode:", err)
		return
	}
	pxy = &proxy{
		upstreamNode: node,
	}

	config := p2p.Config{
		PrivateKey:     nodekey,
		MaxPeers:       200,
		NoDiscovery:    false,
		DiscoveryV5:    false,
		Name:           common.MakeName(fmt.Sprintf("%s/%s", ua, node.ID.String()), ver),
		BootstrapNodes: []*discover.Node{node},
		StaticNodes:    []*discover.Node{node},
		TrustedNodes:   []*discover.Node{node},

		Protocols: []p2p.Protocol{newManspreadingProtocol()},

		ListenAddr: listenAddr,
		Logger:     log.New(),
	}
	// config.Logger.SetHandler(log.StdoutHandler)

	pxy.srv = &p2p.Server{Config: config}

	//设置初值
	// 5294375 2881436154511909728
	pxy.upstreamState.CurrentBlock = common.StringToHash("0x58f3ea40c3d1ffdea3c88b8d77ede6bdc2ecd6dc88b24aa2479304c359a043e5")
	pxy.upstreamState.TD = big.NewInt(2881436154511909728)
	// Wait forever
	var wg sync.WaitGroup
	wg.Add(2)
	err = pxy.srv.Start()
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
