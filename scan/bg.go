package scan

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type NodeChainType int
type AddrInfo struct {
	Addr      string `json:"addr"`
	ApiKey    string `json:"api_key"`
	AccountId string `json:"account_id"`
}

const (
	UNKNOW     NodeChainType = iota
	CHAIN_BSC  NodeChainType = iota + 1000
	CHAIN_ARBI NodeChainType = iota + 1000
	ETH        NodeChainType = iota + 1000
	ZKSync     NodeChainType = iota + 1000
	ZKEVM      NodeChainType = iota + 1000
	StarkNet   NodeChainType = iota + 1000
	TronNet    NodeChainType = iota + 1000
)

func (c NodeChainType) IsValid() bool { //暂不支持这种链
	return c.Name() != "Unknow"
}

func (c NodeChainType) Name() string {
	switch c {
	case CHAIN_BSC:
		return "BSC"
	case CHAIN_ARBI:
		return "Arbitrum"
	case TronNet:
		return "Tron"
	case ETH:
		return "ETH"
	case ZKSync:
		return "ZKSync"
	case ZKEVM:
		return "ZKEVM"
	case StarkNet:
		return "StarkNet"
	}
	return "Unknow"
}

func GetLastWork(chain ScanTool, idx int) (int64, error) {
	if Redis == nil {
		return 0, fmt.Errorf("redis nil")
	}
	return Redis.HIncrBy(context.Background(), fmt.Sprintf("scan:node:%d", idx), string(chain.ChainType()), 0).Result()
}

func SetLastWork(chain ScanTool, idx int, block int64) error {
	if Redis == nil {
		return fmt.Errorf("redis nil")
	}
	_, err := Redis.HSet(context.Background(), fmt.Sprintf("scan:node:%d", idx), string(chain.ChainType()), block).Result()
	return err
}

type WorkHandler struct {
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
	scan   *Scan
}

func (w *WorkHandler) Run() {
	w.once.Do(func() {
		go w.work()
	})
}
func (w *WorkHandler) work() {
	idx := 0
	for {
		idx++
		select {
		case <-w.ctx.Done():
			return
		default:
			w.scan.Process()
			time.Sleep(time.Second * 2)
		}
		if idx%5 == 0 {
			idx = 0
			Logger.Error("Workd", "status", "running")
		}
	}
}
func (w *WorkHandler) Stop() {
	w.cancel()
}
func (w *WorkHandler) Result() <-chan []*ContractTokenTran {
	return w.scan.Result()
}

// NewWork maxGoNum 最大执行分组 cfg 链的配置
func NewWork(maxGoNum int, cfgs ...ChainScanCfg) *WorkHandler {
	scan := newScan(int64(maxGoNum), cfgs...)
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkHandler{
		ctx:    ctx,
		cancel: cancel,
		scan:   scan,
	}
}
