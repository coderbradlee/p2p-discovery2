package redis

import (
	"math/big"
)

// BlockData
type BlockData struct {
	Height         int64    `json:"height"`
	Timestamp      int64    `json:"timestamp"`
	Difficulty     int64    `json:"difficulty"`
	TotalShares    int64    `json:"shares"`
	Uncle          bool     `json:"uncle"`
	UncleHeight    int64    `json:"uncleHeight"`
	Orphan         bool     `json:"orphan"`
	Hash           string   `json:"hash"`
	Nonce          string   `json:"-"`
	PowHash        string   `json:"-"`
	MixDigest      string   `json:"-"`
	Reward         *big.Int `json:"-"`
	ExtraReward    *big.Int `json:"-"`
	ImmatureReward string   `json:"-"`
	RewardString   string   `json:"reward"`
	RoundHeight    int64    `json:"-"`
	candidateKey   string
	immatureKey    string
	NodeName       string `json:"-"`
}

type PoolCharts struct {
	Timestamp  int64  `json:"timestamp"`
	TimeFormat string `json:"timeFormat"`
	PoolHash   int64  `json:"poolHash"`
}

type MinerCharts struct {
	Timestamp  int64  `json:"timestamp"`
	TimeFormat string `json:"timeFormat"`
	MinerHash  int64  `json:"minerHash"`
}

type NodeHashs struct {
	Timestamp  int64      `json:"timestamp"`
	TimeFormat string     `json:"timeFormat"`
	TotalHash  int64      `json:"totalHash"`
	NodeHash   []NodeHash `json:"nodeHash"`
}

type NodeHash struct {
	NodeName string `json:"name"`
	NodeHash int64  `json:"hash"`
}

//新增xvalid类型
type XValid struct {
	Height      int64 //块高度
	Login       string
	Ip          string
	Id          string
	Share       int64 //分享值
	NonceHex    string
	HashNoNonce string
	MixDigest   string
	TimeStamp   int64 //提交时间
	NodeName    string
	ShareType   string //类型，share和block
}

//新增xcast
type XCast struct {
	Height      int64
	TimeStamp   int64
	UncleHeight int64
	BlockReward int64
	Row         map[string]int64
}

type MinerRedis struct {
	Endpoint string `json:"endpoint"`
	Password string `json:"password"`
	Database int64  `json:"database"`
	PoolSize int    `json:"poolSize"`
}

type Miner struct {
	LastBeat  int64 `json:"lastBeat"`
	HR        int64 `json:"hr"`
	Offline   bool  `json:"offline"`
	startedAt int64
	ip        []string
}

type Worker struct {
	Miner
	TotalHR int64 `json:"hr2"`
}
