package scan

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/shopspring/decimal"
)

// /
var trxDecimal = decimal.New(1, 6)

type SolidityData struct {
	BlockID      string              `json:"blockID"`
	BlockHeader  SolidityBlockHeader `json:"block_header"`
	Transactions []Transaction       `json:"transactions"`
}

type SolidityBlockHeader struct {
	RawData          BlockHeaderRawData `json:"raw_data"`
	WitnessSignature string             `json:"witness_signature"`
}

type BlockHeaderRawData struct {
	Number         int64  `json:"number"`
	TxTrieRoot     string `json:"txTrieRoot"`
	WitnessAddress string `json:"witness_address"`
	ParentHash     string `json:"parentHash"`
	Version        int64  `json:"version"`
	Timestamp      int64  `json:"timestamp"`
}

type Transaction struct {
	Ret        []Ret              `json:"ret"`
	Signature  []string           `json:"signature"`
	TxID       string             `json:"txID"`
	RawData    TransactionRawData `json:"raw_data"`
	RawDataHex string             `json:"raw_data_hex"`
}

type TransactionRawData struct {
	Contract      []SolidityContract `json:"contract"`
	RefBlockBytes string             `json:"ref_block_bytes"`
	RefBlockHash  string             `json:"ref_block_hash"`
	Expiration    int64              `json:"expiration"`
	Timestamp     int64              `json:"timestamp,omitempty"`
	FeeLimit      int64              `json:"fee_limit,omitempty"`
	Data          string             `json:"data,omitempty"`
}

type SolidityContract struct {
	Parameter    Parameter `json:"parameter"`
	Type         string    `json:"type"`
	PermissionID *int64    `json:"Permission_id,omitempty"`
}

type Parameter struct {
	Value   Value  `json:"value"`
	TypeURL string `json:"type_url"`
}

type Value struct {
	Amount          int64   `json:"amount,omitempty"`
	OwnerAddress    string  `json:"owner_address"`
	ToAddress       string  `json:"to_address,omitempty"`
	Data            *string `json:"data,omitempty"`
	ContractAddress string  `json:"contract_address,omitempty"`
	Resource        *string `json:"resource,omitempty"`
	ReceiverAddress *string `json:"receiver_address,omitempty"`
	AssetName       *string `json:"asset_name,omitempty"`
	Balance         *int64  `json:"balance,omitempty"`
	AccountAddress  *string `json:"account_address,omitempty"`
	Votes           []Vote  `json:"votes"`
	CallValue       *int64  `json:"call_value,omitempty"`
}

type Vote struct {
	VoteAddress string `json:"vote_address"`
	VoteCount   int64  `json:"vote_count"`
}

type Ret struct {
	ContractRet string `json:"contractRet"`
}

const (
	AccountCreateContract      = "AccountCreateContract"
	DelegateResourceContract   = "DelegateResourceContract"
	TransferAssetContract      = "TransferAssetContract"
	TransferContract           = "TransferContract"
	TriggerSmartContract       = "TriggerSmartContract"
	UnDelegateResourceContract = "UnDelegateResourceContract"
	UnfreezeBalanceContract    = "UnfreezeBalanceContract"
	VoteWitnessContract        = "VoteWitnessContract"
	WithdrawBalanceContract    = "WithdrawBalanceContract"
)

const (
	OutOfEnergy = "OUT_OF_ENERGY"
	Success     = "SUCCESS"
)

type Element struct {
	Log             []Log    `json:"log"`
	Fee             int64    `json:"fee,omitempty"`
	BlockNumber     int64    `json:"blockNumber"`
	ContractResult  []string `json:"contractResult"`
	BlockTimeStamp  int64    `json:"blockTimeStamp"`
	Receipt         Receipt  `json:"receipt"`
	ID              string   `json:"id"`
	ContractAddress string   `json:"contract_address,omitempty"`
}

type Log struct {
	Address string   `json:"address"`
	Data    string   `json:"data"`
	Topics  []string `json:"topics"`
}

type Receipt struct {
	Result             string `json:"result,omitempty"`
	EnergyPenaltyTotal int64  `json:"energy_penalty_total,omitempty"`
	EnergyFee          int64  `json:"energy_fee,omitempty"`
	EnergyUsageTotal   int64  `json:"energy_usage_total,omitempty"`
	OriginEnergyUsage  int64  `json:"origin_energy_usage,omitempty"`
	NetUsage           int64  `json:"net_usage,omitempty"`
	EnergyUsage        int64  `json:"energy_usage,omitempty"`
	NetFee             int64  `json:"net_fee,omitempty"`
}

const getNowBlock = "/walletsolidity/getblock"
const getLastBlock = "/wallet/getblock"

const getTranByNum = "/wallet/gettransactioninfobyblocknum"
const getTrxTranByNum = "/walletsolidity/getblockbynum"

type TronBlockInfo struct {
	BlockID     string      `json:"blockID"`
	BlockHeader BlockHeader `json:"block_header"`
}

type BlockHeader struct {
	RawData          RawData `json:"raw_data"`
	WitnessSignature string  `json:"witness_signature"`
}

type RawData struct {
	Number         int64  `json:"number"`
	TxTrieRoot     string `json:"txTrieRoot"`
	WitnessAddress string `json:"witness_address"`
	ParentHash     string `json:"parentHash"`
	Version        int64  `json:"version"`
	Timestamp      int64  `json:"timestamp"`
}
type tronTool struct {
	url string
	// solidUrl   string
	monitorMap sync.Map // map[string]*Contract
}

func (t *tronTool) GetBlockNum() (int64, error) {
	resp, code, err := Request(Post, t.url+getNowBlock, nil, nil)
	if err != nil {
		return 0, err
	}
	if code != 200 {
		return 0, fmt.Errorf("code not 200")
	}
	block := &TronBlockInfo{}
	err = json.Unmarshal(resp, block)
	if err != nil {
		return 0, err
	}
	if block.BlockHeader.RawData.Number > 0 {
		return block.BlockHeader.RawData.Number, nil
	}
	return 0, fmt.Errorf("block numer is zero")
}

func (t *tronTool) getLastBlockNum() (int64, error) {
	resp, code, err := Request(Post, t.url+getLastBlock, nil, nil)
	if err != nil {
		return 0, err
	}
	if code != 200 {
		return 0, fmt.Errorf("code not 200")
	}
	block := &TronBlockInfo{}
	err = json.Unmarshal(resp, block)
	if err != nil {
		return 0, err
	}
	if block.BlockHeader.RawData.Number > 0 {
		return block.BlockHeader.RawData.Number, nil
	}
	return 0, fmt.Errorf("block numer is zero")
}
func (t *tronTool) GetLog(blockNum int64) ([]*ContractTokenTran, error) {
	lastBlockNum, err := t.getLastBlockNum()
	if err != nil {
		return nil, err
	}
	trx, err := t.TrxTransfer(blockNum, lastBlockNum)
	if err != nil {
		return nil, err
	}
	//trc20交易
	out, err := t.Trc20Transfer(blockNum, lastBlockNum, trx)
	if err != nil {
		return nil, err
	}
	return out, err
}

// TrxTransfer 爬块 trx 交易
func (t *tronTool) TrxTransfer(blockNum int64, lastBlock int64) (map[string]*ContractTokenTran, error) {
	param := make(map[string]any)
	param["num"] = blockNum
	param["visible"] = true
	resp, code, err := Request(Post, t.url+getTrxTranByNum, nil, param)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("code is not 200")
	}
	data := &SolidityData{}
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	out := make(map[string]*ContractTokenTran)
	for _, rawTran := range data.Transactions {
		if len(rawTran.Ret) == 0 {
			continue
		}
		if len(rawTran.RawData.Contract) == 0 {
			continue
		}
		tmp := &ContractTokenTran{
			Chain:             string(t.ChainType()),
			BlockNum:          data.BlockHeader.RawData.Number,
			TransferTimestamp: rawTran.RawData.Timestamp,
			Confirmations:     lastBlock - data.BlockHeader.RawData.Number,
			TxId:              rawTran.TxID,
			Success:           rawTran.Ret[0].ContractRet == Success,
			Remark:            rawTran.Ret[0].ContractRet,
			Transfers:         make([]*CallbackTransfer, 0),
		}
		for idx, contract := range rawTran.RawData.Contract {
			value := contract.Parameter.Value
			if contract.Type != TransferContract {
				continue
			}
			amount := decimal.NewFromInt(value.Amount)
			if value.OwnerAddress == "" || value.ToAddress == "" || amount.LessThanOrEqual(decimal.Zero) {
				continue
			}
			tmp.Transfers = append(tmp.Transfers, &CallbackTransfer{
				FromAddress: value.OwnerAddress,
				ToAddress:   value.ToAddress,
				Contract:    "TRX",
				Symbol:      "TRX",
				Amount:      amount.Div(trxDecimal).String(),
				LogIdx:      idx,
			})
		}
		if len(tmp.Transfers) > 0 {
			out[rawTran.TxID] = tmp
		}
	}
	return out, nil
}
func (t *tronTool) Trc20Transfer(blockNum int64, lastBlock int64, trx map[string]*ContractTokenTran) ([]*ContractTokenTran, error) {
	param := make(map[string]any)
	param["num"] = blockNum
	param["visible"] = true
	resp, code, err := Request(Post, t.url+getTranByNum, nil, param)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("code is not 200")
	}
	showData := make([]Element, 0)
	err = json.Unmarshal(resp, &showData)
	if err != nil {
		return nil, err
	}
	out := make([]*ContractTokenTran, 0)

	for _, logs := range showData {
		if logs.Receipt.Result == "" {
			tmp, ok := trx[logs.ID]
			if ok {
				feeCoin := decimal.NewFromInt(logs.Receipt.EnergyFee + logs.Receipt.NetFee)
				feeCoin = feeCoin.Div(trxDecimal)
				// usdt := feeCoin.Mul(price)
				tmp.FeeSymbol = "TRX"
				tmp.FeeAmountCoin = feeCoin.String()
				// tmp.FeeSymbolPrice = price.String()
				// tmp.FeeAmountUsdt = usdt.String()
			}
			continue
		}
		feeCoin := decimal.NewFromInt(logs.Receipt.EnergyFee + logs.Receipt.NetFee)
		feeCoin = feeCoin.Div(trxDecimal)
		// usdt := feeCoin.Mul(price)
		transferData := &ContractTokenTran{
			Chain:             string(t.ChainType()),
			BlockNum:          logs.BlockNumber,
			TransferTimestamp: logs.BlockTimeStamp,
			TxId:              logs.ID,
			FeeSymbol:         "TRX",
			// FeeSymbolPrice:    price.String(),
			FeeAmountCoin: feeCoin.String(),
			// FeeAmountUsdt:     usdt.String(),
			Confirmations: lastBlock - logs.BlockNumber,
			Success:       logs.Receipt.Result == Success,
			Remark:        logs.Receipt.Result,
		}
		for idx, log := range logs.Log {
			if len(log.Topics) != 3 {
				continue
			}
			if len(log.Topics[0]) != 64 || len(log.Topics[1]) != 64 || len(log.Topics[2]) != 64 {
				continue
			}
			contractInfo, ok := t.GetContract(log.Address)
			if !ok {
				continue
			}
			if !strings.HasPrefix(log.Topics[0], "0x") {
				log.Topics[0] = fmt.Sprintf("0x%s", log.Topics[0])
			}
			if log.Topics[0] != TransferTopic {
				continue
			}
			from := log.Topics[1]
			to := log.Topics[2]
			from = from[24:]
			to = to[24:]
			from = "41" + from
			to = "41" + to
			from = address.HexToAddress(from).String()
			to = address.HexToAddress(to).String()
			tranVal, err := hex.DecodeString(log.Data)
			if err != nil {
				continue
			}
			val := new(big.Int).SetBytes(tranVal)
			amount, err := ChainValue(val.String(), contractInfo.Decimals)
			if err != nil {
				continue
			}
			transferData.Transfers = append(transferData.Transfers, &CallbackTransfer{
				FromAddress: from,
				ToAddress:   to,
				Contract:    contractInfo.Addr,
				Symbol:      contractInfo.TokenName,
				Amount:      amount.String(),
				LogIdx:      idx,
			})
			if len(transferData.Transfers) > 0 {
				out = append(out, transferData)
			}
		}
	}
	for _, val := range trx { //这里由map转slice因此结果是无序的
		out = append(out, val)
	}
	return out, nil
}

func (t *tronTool) GetTrc20Decimal(addr string) {
	if strings.HasPrefix(addr, "T") {
		n, _ := address.Base58ToAddress(addr)
		addr = n.Hex()
	}
	resp, _, err := Request(Post, t.url+"/wallet/triggerconstantcontract", nil,
		map[string]string{
			"owner_address":     "410000000000000000000000000000000000000000", //空地址
			"contract_address":  addr,
			"function_selector": "decimals()",
		})
	fmt.Println(string(resp), err)
}

func (t *tronTool) AddContract(c ...Contract) {
	for idx := range c {
		data := c[idx]
		t.monitorMap.Store(data.Addr, &data)
	}
}

func (t *tronTool) GetContract(address string) (*Contract, bool) {
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

func (t *tronTool) ChainType() ChainType {
	return Tron
}
