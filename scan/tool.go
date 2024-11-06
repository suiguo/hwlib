package scan

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/suiguo/hwlib/logger"
)

var Redis redis.Cmdable
var Logger logger.Logger

var restyClient *resty.Client

type Method string

const (
	Post Method = "POST"
	Get  Method = "GET"
)

func init() {
	tsp := &http.Transport{
		DisableKeepAlives:   true,
		MaxIdleConnsPerHost: 500,
		MaxIdleConns:        500,
		TLSHandshakeTimeout: 2 * time.Second,
	}
	cli := &http.Client{
		Transport: tsp,
	}
	restyClient = resty.NewWithClient(cli).SetRetryCount(5).SetTimeout(time.Second * 3)
	Logger = logger.NewStdLogger(2)
}
func SetRedis(r redis.Cmdable) {
	Redis = r
}

// / Request 请求url获得返回结果，queryParams 格式化在参数后面的?a=b&c=d格式，jsonData 是 body json 的数据格式
func Request(method Method, url string, queryParams map[string]string, jsonData any) ([]byte, int, error) {
	r := restyClient.R()
	if queryParams != nil {
		r = r.SetQueryParams(queryParams)
	}
	if jsonData != nil {
		r = r.SetBody(jsonData)
	}
	switch method {
	case Post:
		respon, err := r.Post(url)
		if err != nil {
			return nil, 0, err
		}
		if respon != nil {
			return respon.Body(), respon.StatusCode(), nil
		}
		return nil, 0, errors.Errorf("nil respon")
	case Get:
		respon, err := r.Get(url)
		if err != nil {
			return nil, 0, err
		}
		if respon != nil {
			return respon.Body(), respon.StatusCode(), nil
		}
		return nil, 0, errors.Errorf("nil respon")
	}
	return nil, 0, errors.Errorf("unknow error")
}

func ChainValue(input string, decimals uint8) (decimal.Decimal, error) {
	out, err := decimal.NewFromString(input)
	if err != nil {
		return out, err
	}
	return out.Div(decimal.New(1, int32(decimals))), nil
}
