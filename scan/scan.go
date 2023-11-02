package scan

import (
	"sync"
)

type ChainScanCfg struct {
	Chain        ChainType
	ConfirmNum   int
	ContractList []Contract
	Rpc          []string
}
type storeTool struct {
	Working []chan struct{}
	GoNum   int64
	ScanTool
	cfg ChainScanCfg
}

// NewScan gonum 追赶时最多多少个请求
func newScan(gonum int64, cfgs ...ChainScanCfg) *Scan {
	s := &Scan{
		popChan: make(chan []*ContractTokenTran, 2000),
	}
	for _, cfg := range cfgs {
		t := &storeTool{
			ScanTool: newTool(cfg.Chain, cfg.Rpc, cfg.ContractList...),
			cfg:      cfg,
			GoNum:    gonum,
			Working:  make([]chan struct{}, gonum),
		}
		for idx := range t.Working {
			t.Working[idx] = make(chan struct{}, 1)
		}
		s.chain.Store(cfg.Chain, t)
	}
	return s
}

type Scan struct {
	chain   sync.Map
	popChan chan []*ContractTokenTran
}

func (s *Scan) AddContract(chainType ChainType, contracts ...Contract) {
	tool, ok := s.chain.Load(chainType)
	if !ok {
		return
	}
	stool, ok := tool.(*storeTool)
	if !ok {
		return
	}
	stool.AddContract(contracts...)
}
func (s *Scan) Process() {
	s.chain.Range(func(key, value any) (next bool) {
		next = true
		tool, ok := value.(*storeTool)
		if !ok {
			return
		}
		nowBlockNum, err := tool.GetBlockNum()
		if err != nil {
			return
		}
		for i := 0; i < int(tool.GoNum); i++ {
			idx := i
			go s.process(tool, idx, nowBlockNum)
		}
		return
	})
}
func (s *Scan) Result() <-chan []*ContractTokenTran {
	return s.popChan
}
func (s *Scan) process(t *storeTool, idx int, nowBlockNum int64) {
	select {
	case t.Working[idx] <- struct{}{}:
		defer func() {
			<-t.Working[idx]
		}()
		var scanBlock int64
		var err error
		scanBlock, err = GetLastWork(t, idx)
		if err != nil {
			return
		}
		if scanBlock == 0 {
			scanBlock = nowBlockNum
			err = SetLastWork(t, idx, scanBlock-1)
			if err != nil {
				return
			}
		}
		for {
			//
			scanBlock++
			if scanBlock%t.GoNum != int64(idx) {
				continue
			}
			if nowBlockNum-scanBlock < int64(t.cfg.ConfirmNum) {
				return
			}
			Logger.Info("Process", "now_block", nowBlockNum, "scan_block", scanBlock)
			results, err := t.GetLog(scanBlock)
			if err != nil {
				Logger.Info("Process", "idx", idx, "block", scanBlock, "status", err)
				return
			}
			err = SetLastWork(t, idx, scanBlock)
			if err != nil {
				Logger.Info("Process", "idx", idx, "block", scanBlock, "status", err)
				return
			}
			if results != nil {
				s.popChan <- results
				Logger.Info("Process", "idx", idx, "block", scanBlock, "nowblock", nowBlockNum, "status", "success")
			} else {
				Logger.Info("Process", "idx", idx, "block", scanBlock, "nowblock", nowBlockNum, "status", "success", "transfers", 0)
			}

		}
	default:
		return
	}
}
