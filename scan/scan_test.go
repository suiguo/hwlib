package scan

import (
	"log"
	"testing"

	"github.com/suiguo/hwlib/logger"
	"github.com/suiguo/hwlib/redis"
)

func TestXxx(t *testing.T) {
	r, _ := redis.GetInstance(logger.NewStdLogger(2), &redis.RedisCfg{
		Url: []struct {
			Host string "json:\"host\""
			Port int    "json:\"port\""
		}{
			{Host: "127.0.0.1", Port: 6379},
		},
	})
	SetRedis(r.Cc)
	scan := NewWork(2, ChainScanCfg{
		Chain:      BSC,
		ConfirmNum: 18, //确认区块数
		ContractList: []Contract{
			{
				Addr:      "0x55d398326f99059ff775485246999027b3197955",
				TokenName: "USDT",
				Decimals:  18,
			},
		}, //默认支持本币
		Rpc: []string{"https://rpc.ankr.com/bsc/a6afc2cc81e33de7db377aeded161d882690963c778274a034916d0d40898930"},
	})
	scan.Run()
	for slice := range scan.Result() {
		for _, transfer := range slice {
			log.Println(transfer)
		}
	}
}
