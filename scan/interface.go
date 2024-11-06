package scan

type ChainType string

const (
	Tron     ChainType = "TRON"
	Eth      ChainType = "ETH"
	BSC      ChainType = "BSC"
	Arbitrum ChainType = "Aribitrum"
)
const TransferTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

// const EthJsonRpc = "https://rpc.ankr.com/eth/a6afc2cc81e33de7db377aeded161d882690963c778274a034916d0d40898930"

// const TronJsonRpc = "http://16.162.23.80:8090"
// const TronJsonRpc2 = "http://16.162.23.80:8091"

type Contract struct {
	Addr      string
	TokenName string
	Decimals  uint8
}

type ScanTool interface {
	GetBlockNum() (int64, error)
	GetLog(blockNum int64) ([]*ContractTokenTran, error)
	ChainType() ChainType
	AddContract(...Contract)
}

func newTool(chain ChainType, rpc []string, contracts ...Contract) ScanTool {
	switch chain {
	case Tron:
		//波场处理
		t := &tronTool{url: rpc[0]}
		t.AddContract(contracts...)
		return t
	case Eth, BSC, Arbitrum:
		t := &ethTool{chain_type: chain, url: rpc[0]}
		t.AddContract(contracts...)
		return t
	}
	return nil
}
