package scan

///采用新的bsc client
import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type BlockByNumberResp struct {
	Jsonrpc string              `json:"jsonrpc"`
	ID      int64               `json:"id"`
	Error   Error               `json:"error"`
	Result  BlockByNumberResult `json:"result"`
}

type BlockByNumberResult struct {
	Difficulty       string                     `json:"difficulty"`
	ExtraData        string                     `json:"extraData"`
	GasLimit         string                     `json:"gasLimit"`
	GasUsed          string                     `json:"gasUsed"`
	Hash             string                     `json:"hash"`
	LogsBloom        string                     `json:"logsBloom"`
	Miner            string                     `json:"miner"`
	MixHash          string                     `json:"mixHash"`
	Nonce            string                     `json:"nonce"`
	Number           string                     `json:"number"`
	ParentHash       string                     `json:"parentHash"`
	ReceiptsRoot     string                     `json:"receiptsRoot"`
	Sha3Uncles       string                     `json:"sha3Uncles"`
	Size             string                     `json:"size"`
	StateRoot        string                     `json:"stateRoot"`
	Timestamp        string                     `json:"timestamp"`
	TotalDifficulty  string                     `json:"totalDifficulty"`
	Transactions     []BlockByNumberTransaction `json:"transactions"`
	TransactionsRoot string                     `json:"transactionsRoot"`
	Uncles           []interface{}              `json:"uncles"`
}

type BlockByNumberTransaction struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	Value            string `json:"value"`
	Type             string `json:"type"`
	ChainID          string `json:"chainId"`
}

///

type BlockNumber struct {
	Jsonrpc string `json:"jsonrpc"`
	Error   Error  `json:"error"`
	ID      int64  `json:"id"`
	Result  string `json:"result"`
}

type Error struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// //
type ReceiptRespon struct {
	Jsonrpc string   `json:"jsonrpc"`
	ID      int64    `json:"id"`
	Error   Error    `json:"error"`
	Result  []Result `json:"result"`
}

type Result struct {
	BlockHash         string       `json:"blockHash"`
	BlockNumber       string       `json:"blockNumber"`
	ContractAddress   interface{}  `json:"contractAddress"`
	CumulativeGasUsed string       `json:"cumulativeGasUsed"`
	EffectiveGasPrice string       `json:"effectiveGasPrice"`
	From              string       `json:"from"`
	GasUsed           string       `json:"gasUsed"`
	Logs              []ReceiptLog `json:"logs"`
	LogsBloom         string       `json:"logsBloom"`
	Status            Status       `json:"status"`
	To                string       `json:"to"`
	TransactionHash   string       `json:"transactionHash"`
	TransactionIndex  string       `json:"transactionIndex"`
	Type              Status       `json:"type"`
}

type ReceiptLog struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

type Status string

const (
	FailStatus    Status = "0x0"
	SuccessStatus Status = "0x1"
)

type JsonRpcParam struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int64         `json:"id"`
}

// ///////
type ContractTokenTran struct {
	BlockNum       int64  `json:"blockNum"`
	Chain          string `json:"chain"`
	Confirmations  int64  `json:"confirmations"`
	FeeAmountCoin  string `json:"feeAmountCoin"`
	FeeAmountUsdt  string `json:"feeAmountUsdt"`
	FeeSymbol      string `json:"feeSymbol"`
	FeeSymbolPrice string `json:"feeSymbolPrice"`
	Remark         string `json:"remark"`
	RequestId      string `json:"requestId"`
	Success        bool   `json:"result"` //是否是成功交易
	// Success bo
	Timestamp         int64               `json:"timestamp"`
	TransferTimestamp int64               `json:"transferTimestamp"`
	Transfers         []*CallbackTransfer `json:"transfers"`
	TxId              string              `json:"txid"`
}

type CallbackTransfer struct {
	Amount      string `json:"amount"`
	Contract    string `json:"contract"`
	FromAddress string `json:"fromAddress"`
	LogIdx      int    `json:"logIdx"`
	Symbol      string `json:"symbol"`
	ToAddress   string `json:"toAddress"`
}

type ethTool struct {
	// client     *ethclient.Client
	url        string
	monitorMap sync.Map // map[string]*Contract
	requestId  atomic.Int64
	chain_type ChainType
}

func (t *ethTool) AddContract(c ...Contract) {
	for idx := range c {
		data := c[idx]
		data.Addr = strings.ToLower(data.Addr)
		t.monitorMap.Store(data.Addr, &data)
	}
}
func (t *ethTool) GetBlockNum() (int64, error) {
	for i := 0; i < 3; i++ {
		idx := t.requestId.Add(1)
		out, code, err := Request(Post, t.url, nil, &JsonRpcParam{
			Jsonrpc: "2.0",
			Method:  "eth_blockNumber",
			ID:      idx,
		})
		if err != nil {
			return 0, err
		}
		if code != 200 {
			return 0, fmt.Errorf("not 200 code")
		}
		resp := &BlockNumber{}
		err = json.Unmarshal(out, resp)
		if err != nil {
			return 0, err
		}
		if resp.Error.Code != 0 {
			return 0, fmt.Errorf(resp.Error.Message)
		}
		r, err := strconv.ParseInt(resp.Result, 0, 64)
		if err == nil {
			return r, err
		} else {
			time.Sleep(time.Second * 2)
		}
	}
	return 0, fmt.Errorf("GetBlockNum error")

}

func (t *ethTool) GetContract(address string) (*Contract, bool) {
	address = strings.ToLower(address)
	info, ok := t.monitorMap.Load(address)
	if !ok {
		return nil, false
	}
	contractInfo := info.(*Contract)
	if contractInfo.Addr == "" {
		return nil, false
	}
	return contractInfo, true
}
func (t *ethTool) getconractTransfer(blockNum int64, nowblock int64) ([]*ContractTokenTran, map[string]bool, error) {
	idx := t.requestId.Add(1)
	req := &JsonRpcParam{
		Jsonrpc: "2.0",
		Method:  "eth_getBlockReceipts",
		Params:  []any{fmt.Sprintf("0x%x", blockNum)},
		ID:      idx,
	}
	resp := &ReceiptRespon{}
	for i := 0; i < 3; i++ {
		r, code, err := Request(Post, t.url, nil, req)
		if err != nil {
			return nil, nil, err
		}
		if code != 200 {
			return nil, nil, fmt.Errorf("code not 200")
		}
		err = json.Unmarshal(r, resp)
		if err != nil {
			return nil, nil, err
		}
		if resp.Error.Code != 0 {
			return nil, nil, fmt.Errorf(resp.Error.Message)
		}
		if len(resp.Result) == 0 {
			Logger.Info("CheckStatus", "result", "0", "status", "retry")
		} else {
			continue
		}
	}
	success := make(map[string]bool)
	out := make([]*ContractTokenTran, 0)
	for _, log := range resp.Result {
		transfertmp := &ContractTokenTran{
			BlockNum:      blockNum,
			Chain:         string(t.ChainType()),
			Confirmations: nowblock - blockNum,
			FeeAmountCoin: string(t.ChainType()),
			TxId:          log.TransactionHash,
			Success:       log.Status == SuccessStatus,
			Transfers:     make([]*CallbackTransfer, 0),
		}
		success[log.TransactionHash] = log.Status == SuccessStatus
		if len(log.Logs) != 1 { //扫描直接交易的信息
			continue
		}
		for _, logdata := range log.Logs {
			if len(logdata.Topics) != 3 {
				continue
			}
			if logdata.Removed {
				continue
			}
			contractInfo, ok := t.GetContract(logdata.Address)
			if !ok || contractInfo == nil {
				continue
			}
			from := logdata.Topics[1]
			to := logdata.Topics[2]
			data := logdata.Data
			data = strings.TrimPrefix(data, "0x")
			from = from[26:]
			from = fmt.Sprintf("0x%s", from)
			to = to[26:]
			to = fmt.Sprintf("0x%s", to)
			val := new(big.Int)
			tranVal, err := hex.DecodeString(data)
			if err != nil {
				continue
			}
			val = val.SetBytes(tranVal)
			tmp, err := ChainValue(val.String(), contractInfo.Decimals)
			if err != nil {
				continue
			}
			if err != nil {
				// fmt.Println("err", err, "log", val)
				continue
			}
			idx, _ := strconv.ParseInt(logdata.LogIndex, 0, 32)
			transfertmp.Transfers = append(transfertmp.Transfers, &CallbackTransfer{
				FromAddress: from,
				ToAddress:   to,
				Contract:    logdata.Address,
				Symbol:      contractInfo.TokenName,
				Amount:      tmp.String(),
				LogIdx:      int(idx),
			})
		}
		if len(transfertmp.Transfers) > 0 {
			out = append(out, transfertmp)
		}
	}
	return out, success, nil
}
func (t *ethTool) GetLog(blockNum int64) ([]*ContractTokenTran, error) {
	nowblock, err := t.GetBlockNum()
	if err != nil {
		return nil, err
	}
	has, err := t.hasTransfer(blockNum)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	var contract []*ContractTokenTran
	var success map[string]bool
	for i := 0; i < 5; i++ {
		contract, success, err = t.getconractTransfer(blockNum, nowblock)
		if err == nil && len(success) > 0 {
			break
		}
		time.Sleep(time.Second * 1)
	}
	if err != nil {
		return nil, err
	}
	bnbtransfer, err := t.getBlockByNum(blockNum, success, nowblock)
	if err != nil {
		return nil, err
	}
	contract = append(contract, bnbtransfer...)
	return contract, err
}
func (t *ethTool) getBlockByNum(blockNum int64, success map[string]bool, nowblock int64) ([]*ContractTokenTran, error) {
	idx := t.requestId.Add(1)
	out, code, err := Request(Post, t.url, nil, &JsonRpcParam{
		Jsonrpc: "2.0",
		Method:  "eth_getBlockByNumber",
		ID:      idx,
		Params:  []any{fmt.Sprintf("0x%x", blockNum), true},
	})
	// os.WriteFile("out2.json", out, os.ModePerm)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("code error")
	}
	// return nil, nil
	info := &BlockByNumberResp{}
	err = json.Unmarshal(out, info)
	if err != nil {
		return nil, err
	}
	if info.Error.Code != 0 {
		return nil, fmt.Errorf(info.Error.Message)
	}
	outtransfer := make([]*ContractTokenTran, 0)
	//bnb本币
	for _, val := range info.Result.Transactions {
		transfertmp := &ContractTokenTran{
			Chain:         string(t.ChainType()),
			Confirmations: nowblock - blockNum,
			FeeAmountCoin: string(t.ChainType()),
			BlockNum:      blockNum,
			TxId:          val.Hash,
			Success:       success[val.Hash],
			Transfers:     make([]*CallbackTransfer, 0),
		}
		if val.Input != "0x" {
			continue
		}
		amout, ok := new(big.Int).SetString(strings.ReplaceAll(val.Value, "0x", ""), 16)
		if !ok {
			continue
		}
		realamount, err := ChainValue(amout.String(), 18)
		if err != nil {
			continue
		}
		transfertmp.Transfers = append(transfertmp.Transfers,
			&CallbackTransfer{
				FromAddress: val.From,
				Contract:    "",
				Amount:      realamount.String(),
				ToAddress:   val.To,
				Symbol:      string(t.ChainType()),
			})
		outtransfer = append(outtransfer, transfertmp)
		// outtransfer = append(outtransfer, &)
	}
	return outtransfer, nil
}

func (t *ethTool) ChainType() ChainType {
	return t.chain_type
}

func (t *ethTool) hasTransfer(block int64) (bool, error) {
	idx := t.requestId.Add(1)
	for i := 0; i < 3; i++ {
		out, code, err := Request(Post, t.url, nil, &JsonRpcParam{
			Jsonrpc: "2.0",
			Method:  "eth_getBlockTransactionCountByNumber",
			ID:      idx,
			Params:  []any{fmt.Sprintf("0x%x", block)},
		})
		if err != nil {
			return false, err
		}
		if code != 200 {
			return false, fmt.Errorf("not 200 code")
		}
		resp := &BlockNumber{}
		err = json.Unmarshal(out, resp)
		if err != nil {
			return false, err
		}
		if resp.Error.Message != "" {
			return false, fmt.Errorf(resp.Error.Message)
		}
		num, err := strconv.ParseInt(resp.Result, 0, 64)
		if err == nil {
			return num > 0, nil
		} else {
			Logger.Info("hasTransfer", "resp", resp)
			time.Sleep(time.Second * 1)
		}
	}
	return false, fmt.Errorf("hasTransfer error")
}

// 波场解析
//
//	func (t *ethTool) GetBlockByNumber(block int64) ([]*ContractTokenTran, error) {
//		idx := t.requestId.Add(1)
//		out, err := Request(Post, t.url, t.apiKey, &JsonRpcParam{
//			Jsonrpc: "2.0",
//			Method:  "eth_getBlockByNumber",
//			ID:      idx,
//			Params:  []any{fmt.Sprintf("0x%x", block), true},
//		})
//		if err != nil {
//			fmt.Println(err)
//		}
//		info := &BlockInfo{}
//		err = json.Unmarshal(out, info)
//		if err != nil {
//			return nil, err
//		}
//		out_contract := make([]*ContractTokenTran, 0)
//		now := time.Now()
//		for _, tran := range info.Result.Transactions {
//			if len(tran.Input) != 64*2+len(TransferFix) {
//				continue
//			}
//			if !strings.HasPrefix(tran.Input, TransferFix) {
//				continue
//			}
//			contract, ok := t.monitorMap.Load(tran.To)
//			if !ok {
//				continue
//			}
//			contractInfo := contract.(*Contract)
//			toAddr := tran.Input[len(TransferFix) : len(TransferFix)+64]
//			value := tran.Input[len(TransferFix)+64:]
//			toAddr = toAddr[24:]
//			toAddr = "41" + toAddr
//			val := new(big.Int)
//			tran_val, err := hex.DecodeString(value)
//			if err != nil {
//				return nil, err
//			}
//			val = val.SetBytes(tran_val)
//			tmp, err := ChainValue(val.String(), contractInfo.Decimals)
//			if err != nil {
//				return nil, err
//			}
//			from_str := tran.From
//			if !strings.HasPrefix(from_str, "0x41") && !strings.HasPrefix(from_str, "41") {
//				from_str = fmt.Sprintf("0x41%s", strings.TrimPrefix(tran.From, "0x"))
//			}
//			out_contract = append(out_contract, &ContractTokenTran{
//				From:     address.HexToAddress(from_str).String(),
//				To:       address.HexToAddress(toAddr).String(),
//				Contract: contractInfo.Addr,
//				Token:    contractInfo.TokenName,
//				TxHash:   tran.Hash,
//				Amount:   tmp.String(),
//			})
//		}
//		fmt.Println("cost ", time.Since(now).Milliseconds())
//		return out_contract, nil
//	}
//
// func (t *ethTool)
// eth解析
//
//	func (t *ethTool) GetBlockByNumber(block int64) ([]*ContractTokenTran, error) {
//		idx := t.requestId.Add(1)
//		out, err := Request(Post, t.url, t.apiKey, &JsonRpcParam{
//			Jsonrpc: "2.0",
//			Method:  "eth_getBlockByNumber",
//			ID:      idx,
//			Params:  []any{fmt.Sprintf("0x%x", block), true},
//		})
//		if err != nil {
//			fmt.Println(err)
//		}
//		info := &BlockInfo{}
//		err = json.Unmarshal(out, info)
//		if err != nil {
//			return nil, err
//		}
//		out_contract := make([]*ContractTokenTran, 0)
//		now := time.Now()
//		for _, tran := range info.Result.Transactions {
//			if len(tran.Input) != 64*2+len(TransferFix) {
//				continue
//			}
//			if !strings.HasPrefix(tran.Input, TransferFix) {
//				continue
//			}
//			contract, ok := t.monitorMap.Load(tran.To)
//			if !ok {
//				continue
//			}
//			contractInfo := contract.(*Contract)
//			toAddr := tran.Input[len(TransferFix) : len(TransferFix)+64]
//			fmt.Println(len(toAddr))
//			toAddr = fmt.Sprintf("0x%s", strings.TrimLeft(toAddr, "0"))
//			value := tran.Input[len(TransferFix)+64:]
//			fmt.Println(len(value))
//			// value = fmt.Sprintf("0x%s", strings.TrimLeft(value, "0"))
//			val := new(big.Int)
//			tran_val, err := hex.DecodeString(value)
//			if err != nil {
//				return nil, err
//			}
//			val = val.SetBytes(tran_val)
//			tmp, err := ChainValue(val.String(), contractInfo.Decimals)
//			if err != nil {
//				return nil, err
//			}
//			out_contract = append(out_contract, &ContractTokenTran{
//				From:     tran.From,
//				To:       toAddr,
//				Contract: contractInfo.Addr,
//				Token:    contractInfo.TokenName,
//				TxHash:   tran.Hash,
//				Amount:   tmp.String(),
//			})
//		}
//		fmt.Println("cost ", time.Since(now).Milliseconds())
//		return out_contract, nil
//	}
// 0x000000000000000000000000cfc0f98f30742b6d880f90155d4ebb885e55ab33
