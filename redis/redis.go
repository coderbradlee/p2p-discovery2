package redis

import (
	// "fmt"
	// "log"
	"math/big"
	// "strconv"
	// "strings"
	// "time"

	//"github.com/ethereum/go-ethereum/common"
	//"github.com/ethereumproject/go-ethereum/common"
	"gopkg.in/redis.v3"

	"../ethhelp"
	"../util"
)

func (b *BlockData) RewardInShannon() int64 {
	reward := new(big.Int).Div(b.Reward, ethhelp.Shannon)
	return reward.Int64()
}

func (b *BlockData) serializeHash() string {
	if len(b.Hash) > 0 {
		return b.Hash
	} else {
		return "0x0"
	}
}

func (b *BlockData) RoundKey() string {
	return join(b.RoundHeight, b.Hash)
}

func (b *BlockData) key() string {
	return join(b.UncleHeight, b.Orphan, b.Nonce, b.serializeHash(), b.Timestamp, b.Difficulty, b.TotalShares, b.Reward, b.NodeName)
}

//func NewRedisClient(cfg *Config, prefix string) *RedisClient {
//	client := redis.NewFailoverClient(&redis.FailoverOptions{
//		MasterName:    cfg.MasterName,
//		SentinelAddrs: cfg.SentinelAddrs,
//		Password:      cfg.Password,
//		DB:            cfg.Database,
//		PoolSize:      cfg.PoolSize,
//	})
//
//	return &RedisClient{client: client, prefix: prefix}
//}
func join(args ...interface{}) string {
	s := make([]string, len(args))
	for i, v := range args {
		switch v.(type) {
		case string:
			s[i] = v.(string)
		case int64:
			s[i] = strconv.FormatInt(v.(int64), 10)
		case uint64:
			s[i] = strconv.FormatUint(v.(uint64), 10)
		case float64:
			s[i] = strconv.FormatFloat(v.(float64), 'f', 0, 64)
		case bool:
			if v.(bool) {
				s[i] = "1"
			} else {
				s[i] = "0"
			}
		case *big.Int:
			n := v.(*big.Int)
			if n != nil {
				s[i] = n.String()
			} else {
				s[i] = "0"
			}
		default:
			panic("Invalid type specified for conversion")
		}
	}
	return strings.Join(s, ":")
}
func NewRedisClient(cfg *Config, prefix string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Endpoint,
		Password: cfg.Password,
		DB:       cfg.Database,
		PoolSize: cfg.PoolSize,
	})
	return &RedisClient{client: client, prefix: prefix}
}

func NewMinerShareRedisClient(cfg *Config, prefix string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Endpoint,
		Password: cfg.Password,
		DB:       cfg.Database,
		PoolSize: cfg.PoolSize,
	})

	return &RedisClient{client: client, prefix: prefix}
}

type Config struct {
	Endpoint string `json:"endpoint"`
	Password string `json:"password"`
	Database int64  `json:"database"`
	PoolSize int    `json:"poolSize"`
}

type RedisClient struct {
	client *redis.Client
	prefix string
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}

func (r *RedisClient) Check() (string, error) {
	return r.client.Ping().Result()
}

func (r *RedisClient) BgSave() (string, error) {
	return r.client.BgSave().Result()
}
func (r *RedisClient) formatKey(args ...interface{}) string {
	return join(r.prefix, join(args...))
}

// Always returns list of addresses. If Redis fails it will return empty list.
// func (r *RedisClient) GetBlacklist() ([]string, error) {
// 	cmd := r.client.SMembers(r.formatKey("blacklist"))
// 	if cmd.Err() != nil {
// 		return []string{}, cmd.Err()
// 	}
// 	return cmd.Val(), nil
// }

// // Always returns list of IPs. If Redis fails it will return empty list.
// func (r *RedisClient) GetWhiteList() ([]string, error) {
// 	cmd := r.client.SMembers(r.formatKey("whitelist"))
// 	if cmd.Err() != nil {
// 		return []string{}, cmd.Err()
// 	}
// 	return cmd.Val(), nil
// }

// func (r *RedisClient) WriteNodeState(id string, height uint64, diff *big.Int) error {
// 	tx := r.client.Multi()
// 	defer tx.Close()

// 	now := util.MakeTimestamp() / 1000

// 	_, err := tx.Exec(func() error {
// 		tx.HSet(r.formatKey("nodes"), join(id, "name"), id)
// 		tx.HSet(r.formatKey("nodes"), join(id, "height"), strconv.FormatUint(height, 10))
// 		tx.HSet(r.formatKey("nodes"), join(id, "difficulty"), diff.String())
// 		tx.HSet(r.formatKey("nodes"), join(id, "lastBeat"), strconv.FormatInt(now, 10))
// 		return nil
// 	})
// 	return err
// }

// func (r *RedisClient) GetAllMinerAccount() (account []string, err error) {
// 	var c int64
// 	for {
// 		now := util.MakeTimestamp() / 1000
// 		c, keys, err := r.client.Scan(c, r.formatKey("miners", "*"), now).Result()

// 		if err != nil {
// 			return account, err
// 		}
// 		for _, key := range keys {
// 			m := strings.Split(key, ":")
// 			if len(m) >= 2 && strings.Index(strings.ToLower(m[2]), "0x") == 0 {
// 				account = append(account, m[2])
// 			}
// 		}
// 		if c == 0 {
// 			break
// 		}
// 	}
// 	return account, nil
// }

// 获取有算力多addr
// coin:hashrate:addr 设置过期时间
// func (r *RedisClient) GetHashrateMinerAccount() (account []string, err error) {
// 	var c int64
// 	for {
// 		now := util.MakeTimestamp() / 1000
// 		c, keys, err := r.client.Scan(c, r.formatKey("hashrate", "*"), now).Result()

// 		if err != nil {
// 			return account, err
// 		}
// 		for _, key := range keys {
// 			m := strings.Split(key, ":")
// 			if len(m) >= 2 && strings.Index(strings.ToLower(m[2]), "0x") == 0 {
// 				account = append(account, m[2])
// 			}
// 		}
// 		if c == 0 {
// 			break
// 		}
// 	}
// 	return account, nil
// }

// func (r *RedisClient) GetMinerCharts(hashNum int64, login string) (stats []*MinerCharts, err error) {
// 	tx := r.client.Multi()
// 	defer tx.Close()
// 	now := util.MakeTimestamp() / 1000
// 	cmds, err := tx.Exec(func() error {
// 		// 172800等于48小时
// 		tx.ZRemRangeByScore(r.formatKey("charts", "miner", login), "-inf", fmt.Sprint("(", now-172800))
// 		tx.ZRevRangeWithScores(r.formatKey("charts", "miner", login), 0, hashNum)
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	stats = convertMinerChartsResults(cmds[1].(*redis.ZSliceCmd))
// 	return stats, nil
// }
// func (r *RedisClient) GetNodeHash(nodeHashLen int64) (stats []NodeHashs, err error) {

// 	tx := r.client.Multi()
// 	defer tx.Close()

// 	now := util.MakeTimestamp() / 1000

// 	cmds, err := tx.Exec(func() error {
// 		// 172800等于48小时
// 		// 259200 =3*24小时
// 		tx.ZRemRangeByScore(r.formatKey("charts", "nodeHash"), "-inf", fmt.Sprint("(", now-259200))
// 		tx.ZRevRangeWithScores(r.formatKey("charts", "nodeHash"), 0, nodeHashLen)
// 		return nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}
// 	stats = convertNodeHashResults(cmds[1].(*redis.ZSliceCmd))
// 	return stats, nil
// }
// func (r *RedisClient) GetPoolCharts(poolHashLen int64) (stats []*PoolCharts, err error) {

// 	tx := r.client.Multi()
// 	defer tx.Close()

// 	now := util.MakeTimestamp() / 1000

// 	cmds, err := tx.Exec(func() error {
// 		// 172800等于48小时
// 		// 259200 =3*24小时
// 		tx.ZRemRangeByScore(r.formatKey("charts", "pool"), "-inf", fmt.Sprint("(", now-259200))
// 		tx.ZRevRangeWithScores(r.formatKey("charts", "pool"), 0, poolHashLen)
// 		return nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	stats = convertPoolChartsResults(cmds[1].(*redis.ZSliceCmd))
// 	//log.Println(stats["poolCharts"])
// 	return stats, nil
// }

// func (r *RedisClient) GetNodeStates() ([]map[string]interface{}, error) {
// 	cmd := r.client.HGetAllMap(r.formatKey("nodes"))
// 	if cmd.Err() != nil {
// 		return nil, cmd.Err()
// 	}
// 	m := make(map[string]map[string]interface{})
// 	for key, value := range cmd.Val() {
// 		parts := strings.Split(key, ":")
// 		if val, ok := m[parts[0]]; ok {
// 			val[parts[1]] = value
// 		} else {
// 			node := make(map[string]interface{})
// 			node[parts[1]] = value
// 			m[parts[0]] = node
// 		}
// 	}
// 	v := make([]map[string]interface{}, len(m), len(m))
// 	i := 0
// 	for _, value := range m {
// 		v[i] = value
// 		i++
// 	}
// 	return v, nil
// }

// func (r *RedisClient) checkPoWExist(height uint64, params []string) (bool, error) {
// 	// Sweep PoW backlog for previous blocks, we have 3 templates back in RAM
// 	r.client.ZRemRangeByScore(r.formatKey("pow"), "-inf", fmt.Sprint("(", height-8))
// 	val, err := r.client.ZAdd(r.formatKey("pow"), redis.Z{Score: float64(height), Member: strings.Join(params, ":")}).Result()
// 	return val == 0, err
// }

// func (r *RedisClient) WriteBlock(login, id, nodeName string, params []string, diff, roundDiff int64, height uint64, window time.Duration, ip string) (bool, error) {
// 	exist, err := r.checkPoWExist(height, params)
// 	if err != nil {
// 		return false, err
// 	}
// 	// Duplicate share, (nonce, powHash, mixDigest) pair exist
// 	if exist {
// 		return true, nil
// 	}
// 	tx := r.client.Multi()
// 	defer tx.Close()

// 	ms := util.MakeTimestamp()
// 	ts := ms / 1000

// 	cmds, err := tx.Exec(func() error {
// 		r.writeShare(tx, ms, ts, login, id, diff, window, ip, nodeName, height, params, "block")
// 		tx.HSet(r.formatKey("stats"), "lastBlockFound", strconv.FormatInt(ts, 10))
// 		tx.HDel(r.formatKey("stats"), "roundShares")
// 		tx.ZIncrBy(r.formatKey("finders"), 1, login)
// 		tx.HIncrBy(r.formatKey("miners", login), "blocksFound", 1)
// 		tx.Rename(r.formatKey("shares", "roundCurrent"), r.formatRound(int64(height), params[0]))
// 		tx.Rename(r.formatKey("shares", "roundNodeCurrent"), r.formatNodeRound(int64(height), params[0]))
// 		tx.HGetAllMap(r.formatRound(int64(height), params[0]))
// 		//rename xvalid,加nonce，防止一条记录中有多个block
// 		//tx.Rename(r.formatKey("xvalid", height),r.formatKey("xvalid", height,params[0]))
// 		return nil
// 	})
// 	if err != nil {
// 		return false, err
// 	} else {
// 		//cmds[12]
// 		sharesMap, _ := cmds[12].(*redis.StringStringMapCmd).Result()
// 		totalShares := int64(0)
// 		for _, v := range sharesMap {
// 			n, _ := strconv.ParseInt(v, 10, 64)
// 			totalShares += n
// 		}
// 		hashHex := strings.Join(params, ":")
// 		s := join(hashHex, ts, roundDiff, totalShares, nodeName)

// 		//写入candidate到miner的reids中
// 		err = r.writeCandidate(float64(height), s)
// 		if err != nil {
// 			log.Println("write candidate to share redis error:", err)
// 			log.Println(float64(height), s)
// 		}
// 		//r.WriteXvalid(login, id, ip, nodeName, ms, diff, window, height, params, "block")

// 		cmd := r.client.ZAdd(r.formatKey("blocks", "candidates"), redis.Z{Score: float64(height), Member: s})

// 		return false, cmd.Err()
// 	}
// }

// func (r *RedisClient) writeShare(tx *redis.Multi, ms, ts int64, login, id string, diff int64, expire time.Duration, ip, nodeName string, height uint64, params []string, shareType string) {
// 	//新增zset类型的分享值，一个高度一段分享值
// 	//tx.ZAdd(r.formatKey("xvalid", height), redis.Z{Score: float64(ms), Member: join(login, ip, id, diff, params[0], params[1], params[2], nodeName, shareType)})
// 	//tx.ZIncrBy(r.formatKey("xvalid", height), float64(diff), join(login, ip, id, ms,  nodeName, shareType))
// 	//tx.Expire(r.formatKey("xvalid", height), expire)

// 	tx.HIncrBy(r.formatKey("shares", "roundCurrent"), login, diff)
// 	tx.HIncrBy(r.formatKey("shares", "roundNodeCurrent"), r.formatKey(login, nodeName), diff)
// 	tx.ZAdd(r.formatKey("hashrate"), redis.Z{Score: float64(ts), Member: join(diff, login, id, ms, nodeName)})
// 	tx.ZAdd(r.formatKey("hashrate", login), redis.Z{Score: float64(ts), Member: join(diff, id, ms, ip)})
// 	tx.Expire(r.formatKey("hashrate", login), expire) // Will delete hashrates for miners that gone
// 	tx.HSet(r.formatKey("miners", login), "lastShare", strconv.FormatInt(ts, 10))
// }

// func (r *RedisClient) WriteShare(login, id string, params []string, diff int64, height uint64, window time.Duration, ip, nodeName string) (bool, error) {
// 	exist, err := r.checkPoWExist(height, params)
// 	if err != nil {
// 		return false, err
// 	}
// 	// Duplicate share, (nonce, powHash, mixDigest) pair exist
// 	if exist {
// 		return true, nil
// 	}
// 	tx := r.client.Multi()
// 	defer tx.Close()

// 	ms := util.MakeTimestamp()
// 	ts := ms / 1000

// 	_, err = tx.Exec(func() error {
// 		r.writeShare(tx, ms, ts, login, id, diff, window, ip, nodeName, height, params, "share")
// 		tx.HIncrBy(r.formatKey("stats"), "roundShares", diff)
// 		return nil
// 	})

// 	//r.WriteXvalid(login, id, ip, nodeName, ms, diff, window, height, params, "share")

// 	return false, err
// }

// func (r *RedisClient) SetMinerChartsExpire(login string) {
// 	r.client.Expire(r.formatKey("charts", "miner", login), 48*time.Hour)
// }

// func (r *RedisClient) WriteMinerCharts(time1 int64, time2, k string, hash, workerOnline int64) error {
// 	s := join(time1, time2, hash)

// 	cmd := r.client.ZAdd(r.formatKey("charts", "miner", k), redis.Z{Score: float64(time1), Member: s})

// 	return cmd.Err()
// }

// func (r *RedisClient) WriteNodeHash(time1 int64, time2 string, nodeHash map[string]int64) error {
// 	var strs string
// 	for k, v := range nodeHash {
// 		strs += join(k, v, "")
// 	}
// 	if len(strs) > 1 {
// 		strs = string(strs[:len(strs)-1])
// 	} else {
// 		strs = "0:0"
// 	}
// 	s := join(time1, time2, strs)
// 	log.Println("node hash:", s)
// 	cmd := r.client.ZAdd(r.formatKey("charts", "nodeHash"), redis.Z{Score: float64(time1), Member: s})
// 	return cmd.Err()
// }

// func (r *RedisClient) WritePoolCharts(time1 int64, time2 string, poolHash string) error {
// 	s := join(time1, time2, poolHash)
// 	cmd := r.client.ZAdd(r.formatKey("charts", "pool"), redis.Z{Score: float64(time1), Member: s})
// 	return cmd.Err()
// }

// func (r *RedisClient) WriteNotWorkRig(login, workId string, worker Worker) error {
// 	s := join(workId)
// 	cmd := r.client.ZAdd(r.formatKey("notwork", login), redis.Z{Score: float64(worker.LastBeat), Member: s})
// 	return cmd.Err()
// }

// //为proxy单独一个xvalid方法
// func (r *RedisClient) WriteXvalid(login, id, ip, nodeName string, ms, diff int64, expire time.Duration, height uint64, params []string, shareType string) error {
// 	cmd := r.Client().ZAdd(r.formatKey("xvalid", height), redis.Z{Score: float64(ms), Member: join(login, ip, id, diff, params[0], params[1], params[2], nodeName, shareType)})
// 	return cmd.Err()
// }

// //单独写一个candidate方法
// func (r *RedisClient) writeCandidate(height float64, s string) error {
// 	cmd := r.client.ZAdd(r.formatKey("xblocks"), redis.Z{Score: height, Member: s})
// 	return cmd.Err()
// }

// func (r *RedisClient) formatKey(args ...interface{}) string {
// 	return join(r.prefix, join(args...))
// }

// func join(args ...interface{}) string {
// 	s := make([]string, len(args))
// 	for i, v := range args {
// 		switch v.(type) {
// 		case string:
// 			s[i] = v.(string)
// 		case int64:
// 			s[i] = strconv.FormatInt(v.(int64), 10)
// 		case uint64:
// 			s[i] = strconv.FormatUint(v.(uint64), 10)
// 		case float64:
// 			s[i] = strconv.FormatFloat(v.(float64), 'f', 0, 64)
// 		case bool:
// 			if v.(bool) {
// 				s[i] = "1"
// 			} else {
// 				s[i] = "0"
// 			}
// 		case *big.Int:
// 			n := v.(*big.Int)
// 			if n != nil {
// 				s[i] = n.String()
// 			} else {
// 				s[i] = "0"
// 			}
// 		default:
// 			panic("Invalid type specified for conversion")
// 		}
// 	}
// 	return strings.Join(s, ":")
// }

// type PendingPayment struct {
// 	Timestamp int64  `json:"timestamp"`
// 	Amount    int64  `json:"amount"`
// 	Address   string `json:"login"`
// }

// type PendingPayment2 struct {
// 	MakeTs   int64 //make timestamp or sendTs
// 	Amount   int64
// 	Login    string
// 	MapLogin string
// 	Nonce    int64
// 	Txid     string
// 	Status   string //txid has check no yes
// }

// func (r *RedisClient) IsMinerExists(login string) (bool, error) {
// 	return r.client.Exists(r.formatKey("miners", login)).Result()
// }

// func (r *RedisClient) GetMinerStats(login string, maxPayments int64) (map[string]interface{}, error) {
// 	stats := make(map[string]interface{})

// 	tx := r.client.Multi()
// 	defer tx.Close()

// 	cmds, err := tx.Exec(func() error {
// 		tx.HGetAllMap(r.formatKey("miners", login))
// 		tx.ZRevRangeWithScores(r.formatKey("payments", login), 0, maxPayments-1)
// 		tx.ZCard(r.formatKey("payments", login))
// 		tx.HGet(r.formatKey("shares", "roundCurrent"), login)
// 		return nil
// 	})

// 	if err != nil && err != redis.Nil {
// 		return nil, err
// 	} else {
// 		result, _ := cmds[0].(*redis.StringStringMapCmd).Result()
// 		stats["stats"] = convertStringMap(result)
// 		payments := convertPaymentsResults(cmds[1].(*redis.ZSliceCmd))
// 		stats["payments"] = payments
// 		stats["paymentsTotal"] = cmds[2].(*redis.IntCmd).Val()
// 		roundShares, _ := cmds[3].(*redis.StringCmd).Int64()
// 		stats["roundShares"] = roundShares
// 	}

// 	return stats, nil
// }

// //todo del
// // WARNING: Must run it periodically to flush out of window hashrate entries
// func (r *RedisClient) FlushStaleStats(window, largeWindow, findWorkerTime time.Duration) (int64, error) {
// 	now := util.MakeTimestamp() / 1000
// 	max := fmt.Sprint("(", now-int64(window/time.Second))
// 	total, err := r.client.ZRemRangeByScore(r.formatKey("hashrate"), "-inf", max).Result()
// 	if err != nil {
// 		return total, err
// 	}

// 	var c int64
// 	miners := make(map[string]struct{})
// 	max = fmt.Sprint("(", now-int64(findWorkerTime/time.Second))

// 	for {
// 		var keys []string
// 		var err error
// 		c, keys, err = r.client.Scan(c, r.formatKey("hashrate", "*"), 100).Result()
// 		if err != nil {
// 			return total, err
// 		}
// 		for _, row := range keys {
// 			login := strings.Split(row, ":")[2]
// 			if _, ok := miners[login]; !ok {
// 				n, err := r.client.ZRemRangeByScore(r.formatKey("hashrate", login), "-inf", max).Result()
// 				if err != nil {
// 					return total, err
// 				}
// 				miners[login] = struct{}{}
// 				total += n
// 			}
// 		}
// 		if c == 0 {
// 			break
// 		}
// 	}
// 	return total, nil
// }

// func (r *RedisClient) CollectStats(smallWindow time.Duration, maxBlocks, maxPayments int64) (map[string]interface{}, error) {
// 	window := int64(smallWindow / time.Second)
// 	stats := make(map[string]interface{})

// 	tx := r.client.Multi()
// 	defer tx.Close()

// 	now := util.MakeTimestamp() / 1000

// 	cmds, err := tx.Exec(func() error {
// 		tx.ZRemRangeByScore(r.formatKey("hashrate"), "-inf", fmt.Sprint("(", now-window))
// 		tx.ZRangeWithScores(r.formatKey("hashrate"), 0, -1)
// 		tx.HGetAllMap(r.formatKey("stats"))
// 		tx.ZRevRangeWithScores(r.formatKey("blocks", "candidates"), 0, -1)
// 		tx.ZRevRangeWithScores(r.formatKey("blocks", "immature"), 0, -1)
// 		tx.ZRevRangeWithScores(r.formatKey("blocks", "matured"), 0, maxBlocks-1)
// 		tx.ZCard(r.formatKey("blocks", "candidates"))
// 		tx.ZCard(r.formatKey("blocks", "immature"))
// 		tx.ZCard(r.formatKey("blocks", "matured")) //todo del
// 		tx.ZCard(r.formatKey("payments", "all"))
// 		tx.ZRevRangeWithScores(r.formatKey("payments", "all"), 0, maxPayments-1)
// 		tx.ZRevRangeWithScores(r.formatKey("blocks", "matured"), 0, 1000)

// 		return nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	result, _ := cmds[2].(*redis.StringStringMapCmd).Result()
// 	stats["stats"] = convertStringMap(result)
// 	candidates := convertCandidateResults(cmds[3].(*redis.ZSliceCmd))
// 	stats["candidates"] = candidates
// 	stats["candidatesTotal"] = cmds[6].(*redis.IntCmd).Val()

// 	immature := convertBlockResults(cmds[4].(*redis.ZSliceCmd))
// 	stats["immature"] = immature
// 	stats["immatureTotal"] = cmds[7].(*redis.IntCmd).Val()

// 	matured := convertBlockResults(cmds[5].(*redis.ZSliceCmd))
// 	stats["matured"] = matured
// 	matured1000 := convertBlockResults(cmds[11].(*redis.ZSliceCmd))
// 	stats["matured1000"] = matured1000

// 	//stats["maturedTotal"] = cmds[8].(*redis.IntCmd).Val()
// 	//出块总数，使用coin:stats记录，不再每次统计
// 	stats["maturedTotal"] = result["matured"]

// 	payments := convertPaymentsResults(cmds[10].(*redis.ZSliceCmd))
// 	stats["payments"] = payments
// 	stats["paymentsTotal"] = cmds[9].(*redis.IntCmd).Val()

// 	totalHashrate, miners, nodeHash := convertMinersStats(window, cmds[1].(*redis.ZSliceCmd))
// 	stats["miners"] = miners
// 	stats["minersTotal"] = len(miners)
// 	stats["hashrate"] = totalHashrate
// 	stats["nodeHash"] = nodeHash

// 	return stats, nil
// }

// /**
// 统计单个旷工的详情(总算力、未工作的机器，工作中的机器和机器提交IP)
// */
// func (r *RedisClient) CollectWorkersStats(sWindow, lWindow, findWorkerTime time.Duration, login string) (map[string]interface{}, error) {
// 	smallWindow := int64(sWindow / time.Second)
// 	largeWindow := int64(lWindow / time.Second)
// 	findWorker := int64(findWorkerTime / time.Second)
// 	stats := make(map[string]interface{})

// 	tx := r.client.Multi()
// 	defer tx.Close()

// 	now := util.MakeTimestamp() / 1000
// 	cmds, err := tx.Exec(func() error {
// 		tx.ZRemRangeByScore(r.formatKey("hashrate", login), "-inf", fmt.Sprint("(", now-largeWindow))
// 		tx.ZRangeWithScores(r.formatKey("hashrate", login), 0, -1)
// 		tx.ZRemRangeByScore(r.formatKey("notwork", login), "-inf", fmt.Sprint("(", now-findWorker))
// 		tx.ZRangeWithScores(r.formatKey("notwork", login), 0, -1)
// 		return nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	totalHashrate := int64(0)
// 	currentHashrate := int64(0)
// 	online := int64(0)
// 	offline := int64(0)
// 	notWorks := convertNotWorkersStats(cmds[3].(*redis.ZSliceCmd))
// 	workers := convertWorkersStats(smallWindow, largeWindow, cmds[1].(*redis.ZSliceCmd))

// 	for id, worker := range workers {
// 		delete(notWorks, id)
// 		timeOnline := now - worker.startedAt
// 		if timeOnline < 600 {
// 			timeOnline = 600
// 		}

// 		boundary := timeOnline
// 		if timeOnline >= smallWindow {
// 			boundary = smallWindow
// 		}
// 		worker.HR = worker.HR / boundary

// 		boundary = timeOnline
// 		if timeOnline >= largeWindow {
// 			boundary = largeWindow
// 		}
// 		worker.TotalHR = worker.TotalHR / boundary

// 		//if worker.LastBeat < (now - smallWindow / 2) {
// 		if worker.LastBeat < (now - smallWindow) {
// 			worker.Offline = true
// 			offline++
// 			r.WriteNotWorkRig(login, id, worker)
// 		} else {
// 			online++
// 		}

// 		currentHashrate += worker.HR
// 		totalHashrate += worker.TotalHR
// 		workers[id] = worker
// 	}
// 	for id, worker := range notWorks {
// 		workers[id] = worker
// 		offline++
// 	}
// 	stats["workers"] = workers
// 	stats["workersTotal"] = (online + offline)
// 	stats["workersOnline"] = online
// 	stats["workersOffline"] = offline
// 	stats["hashrate"] = totalHashrate
// 	stats["currentHashrate"] = currentHashrate
// 	return stats, nil
// }

// func (r *RedisClient) CollectLuckStats(windows []int) (map[string]interface{}, error) {
// 	stats := make(map[string]interface{})

// 	tx := r.client.Multi()
// 	defer tx.Close()

// 	max := int64(windows[len(windows)-1])

// 	cmds, err := tx.Exec(func() error {
// 		tx.ZRevRangeWithScores(r.formatKey("blocks", "immature"), 0, -1)
// 		tx.ZRevRangeWithScores(r.formatKey("blocks", "matured"), 0, max-1)
// 		return nil
// 	})
// 	if err != nil {
// 		return stats, err
// 	}
// 	blocks := convertBlockResults(cmds[0].(*redis.ZSliceCmd), cmds[1].(*redis.ZSliceCmd))

// 	calcLuck := func(max int) (int, float64, float64, float64) {
// 		var total int
// 		var sharesDiff, uncles, orphans float64
// 		for i, block := range blocks {
// 			if i > (max - 1) {
// 				break
// 			}
// 			if block.Uncle {
// 				uncles++
// 			}
// 			if block.Orphan {
// 				orphans++
// 			}
// 			sharesDiff += float64(block.TotalShares) / float64(block.Difficulty)
// 			total++
// 		}
// 		if total > 0 {
// 			sharesDiff /= float64(total)
// 			uncles /= float64(total)
// 			orphans /= float64(total)
// 		}
// 		return total, sharesDiff, uncles, orphans
// 	}
// 	for _, max := range windows {
// 		total, sharesDiff, uncleRate, orphanRate := calcLuck(max)
// 		row := map[string]float64{
// 			"luck": sharesDiff, "uncleRate": uncleRate, "orphanRate": orphanRate,
// 		}
// 		stats[strconv.Itoa(total)] = row
// 		if total < max {
// 			break
// 		}
// 	}
// 	return stats, nil
// }

// //暂时清除xinvalid
// //暂时清除xvalid
// func (r *RedisClient) DelXvalid() {

// 	cmd := r.Client().Keys(r.formatKey("xinvalid") + ":*")
// 	keys := cmd.Val()
// 	for _, key := range keys {
// 		_, err := r.Client().Del(key).Result()
// 		if err != nil {
// 			log.Println("DelXvalid", "xinvalid", err)
// 		}
// 	}

// 	cmd2 := r.Client().Keys(r.formatKey("xvalid") + ":*")
// 	keys2 := cmd2.Val()
// 	for _, key := range keys2 {
// 		_, err := r.Client().Del(key).Result()
// 		if err != nil {
// 			log.Println("DelXvalid", "xvalid", err)
// 		}
// 	}
// }
