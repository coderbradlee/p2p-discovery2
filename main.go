package main

import (
	ethpeer "./ethpeer"
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
	// "net"
	// "os"
	"sync"
	"time"

	// "github.com/ethereum/go-ethereum/cmd/utils"
	// "github.com/ethereum/go-ethereum/crypto"
	// "github.com/ethereum/go-ethereum/p2p/discv5"
	// "github.com/ethereum/go-ethereum/p2p/nat"
	// "github.com/ethereum/go-ethereum/p2p/netutil"
	"./redis"
	// "./rpcs"
	// "strings"
)

var cfg *util.Config
var red *redis.RedisClient

const prefix = "eth"

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
	red = redis.NewRedisClient(&redis.Config{Endpoint: "127.0.0.1:6379", Password: "etcpool123", Database: 0, PoolSize: 10}, prefix)
}

const (
	ua  = "manspreading"
	ver = "1.0.0"

	listenAddr = "0.0.0.0:36666"
	privkey    = ""
	//设置初值
	// 5294375 2881436154511909728
)

var (
	// startBlock = common.StringToHash("0x58f3ea40c3d1ffdea3c88b8d77ede6bdc2ecd6dc88b24aa2479304c359a043e5")
	// startTD    = big.NewInt(2881436154511909728)
	// 换个低一些的高度10000
	startBlock = common.HexToHash("0xdc2d938e4cd0a149681e9e04352953ef5ab399d59bcd5b0357f6c0797470a524")
	startTD    = big.NewInt(2303762395359969)
	genesis    = common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3")
	gversion   = uint32(63)
	gnetworkid = uint64(888888)
)

// statusData is the network packet for the status message.
type statusData struct {
	ProtocolVersion uint32
	NetworkId       uint64
	TD              *big.Int
	CurrentBlock    common.Hash
	GenesisBlock    common.Hash
}

func (s *statusData) String() string {
	return fmt.Sprintf("%v %v %v %v %v", s.ProtocolVersion, s.NetworkId, s.TD.Text(16), s.CurrentBlock.Hex(), s.GenesisBlock.Hex())
}

type newBlockHashesData []struct {
	Hash   common.Hash   // Hash of one particular block being announced
	Number uint64        // Number of one particular block being announced
	header *types.Header // Header of the block partially reassembled (new protocol)	重新组装的区块头
	time   time.Time     // Timestamp of the announcement

	origin string
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
	lock         sync.RWMutex
	upstreamNode *discover.Node //第一个连接的node
	// upstreamConn *conn
	upstreamConn map[discover.NodeID]*conn //后面自动连接的peer
	// downstreamConn *conn
	// upstreamState map[discover.NodeID]statusData
	allPeer    map[string]bool
	ethpeerset *ethpeer.PeerSet
	// NewPeerSet
	bestState     statusData
	bestStateChan chan statusData
	srv           *p2p.Server
	// maxtd         *big.Int
	// bestHash      common.Hash
	bestHeiChan     chan bestHeiPeer
	bestHeiChan2    chan bestHeiPeer
	bestHeiAndPeer  bestHeiPeer
	bestHeiAndPeer2 bestHeiPeer
	bestHeader      types.Header
	bestHeaderChan  chan []*types.Header
	hackChan        chan bool
}
type bestHeiPeer struct {
	bestHei uint64
	p       *p2p.Peer
}

func (pxy *proxy) Start() {
	// tick := time.Tick(50000 * time.Millisecond)
	tickPullBestBlock := time.Tick(10 * time.Second)
	//
	go func() {
		for {
			select {
			case hei, ok := <-pxy.bestHeiChan:
				if !ok {
					break
				}
				if hei.bestHei > pxy.bestHeiAndPeer.bestHei {
					pxy.bestHeiAndPeer = hei
				}
			case hei, ok := <-pxy.bestHeiChan2:
				if !ok {
					break
				}
				if hei.bestHei > pxy.bestHeiAndPeer2.bestHei {
					pxy.bestHeiAndPeer2 = hei
				}
			case beststate, ok := <-pxy.bestStateChan:
				if !ok {
					break
				}
				if beststate.TD.Cmp(pxy.bestState.TD) > 0 && beststate.GenesisBlock.Hex() == genesis.Hex() {
					pxy.bestState = beststate
				}
			case bestheaders, ok := <-pxy.bestHeaderChan:
				// []*types.Header
				if !ok {
					break
				}
				for _, h := range bestheaders {
					if h.Number.Cmp(pxy.bestHeader.Number) > 0 {
						pxy.bestHeader = *h
					}
				}

			case <-tickPullBestBlock:
				pxy.startHack()
				// case <-pxy.hackChan:
				// 	go pxy.hackGetConnect() //获取一些可以连接的
			}
		}
	}()
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

	node, err := discover.ParseNode(MainnetBootnodes[0])
	if err != nil {
		fmt.Println("discover.ParseNode:", err)
		return
	}
	ps := ethpeer.NewPeerSet()
	pxy = &proxy{
		upstreamNode: node,
		upstreamConn: make(map[discover.NodeID]*conn, 0),
		allPeer:      make(map[string]bool, 0),
		ethpeerset:   ps,
		// upstreamState: make(map[discover.NodeID]statusData, 0),
		bestState: statusData{
			ProtocolVersion: gversion,
			NetworkId:       gnetworkid,
			TD:              startTD,
			CurrentBlock:    startBlock,
			GenesisBlock:    genesis,
		},
		bestStateChan:  make(chan statusData),
		bestHeiChan:    make(chan bestHeiPeer),
		bestHeiChan2:   make(chan bestHeiPeer),
		bestHeaderChan: make(chan []*types.Header),
		hackChan:       make(chan bool),
	}
	bootstrapNodes := make([]*discover.Node, 0)
	for _, boot := range MainnetBootnodes {
		old, err := discover.ParseNode(boot)
		if err != nil {
			fmt.Println("discover.ParseNode2:", err)
			continue
		}
		// pxy.srv.AddPeer(old)
		bootstrapNodes = append(bootstrapNodes, old)
	}
	config := p2p.Config{
		PrivateKey:  nodekey,
		MaxPeers:    300,
		NoDiscovery: false,
		DiscoveryV5: false,
		Name:        common.MakeName(fmt.Sprintf("%s/%s", ua, node.ID.String()), ver),
		// BootstrapNodes: []*discover.Node{node},
		BootstrapNodes: bootstrapNodes,
		StaticNodes:    []*discover.Node{node},
		TrustedNodes:   []*discover.Node{node},

		Protocols: []p2p.Protocol{newManspreadingProtocol()},

		ListenAddr: listenAddr,
		Logger:     log.New(),
	}
	// config.Logger.SetHandler(log.StdoutHandler)

	pxy.srv = &p2p.Server{Config: config}
	go pxy.hackGetConnect()
	go pxy.hackReal()
	// Wait forever
	var wg sync.WaitGroup
	wg.Add(2)
	err = pxy.srv.Start()
	pxy.Start()
	wg.Done()
	if err != nil {
		fmt.Println(err)
	}
	wg.Wait()
}
func main() {
	// test()
	test2() //write node ip to redis
	// pureHack()//redis read ip and hack
	//
	c := make(chan int, 1)

	<-c
}
