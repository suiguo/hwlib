package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	es7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type RangOpt string

const (
	Less         RangOpt = "<"
	Greater      RangOpt = ">"
	LessEqual    RangOpt = "<="
	GreaterEqual RangOpt = ">="
	IsNot        RangOpt = "-"
)

type SortOpt string

const (
	Asc  SortOpt = "asc"
	Desc SortOpt = "desc"
)

type commonResp struct {
	ScrollID string `json:"_scroll_id"`
	Hits     Hits   `json:"hits"`
}

type Hits struct {
	Hits []Hit `json:"hits"`
}

type Hit struct {
	ID string `json:"_id"`
}
type rang struct {
	opt RangOpt
	val any
}
type param struct {
	termArgs  map[string]any
	rangeArgs map[string][]*rang
	sortArgs  []string
	andStr    string
	orStr     string
}

func (p *param) packageOut() string {
	param_str := ""
	for k, v := range p.termArgs {
		if param_str == "" {
			param_str = fmt.Sprintf(`%s:"%v"`, k, v)
		} else {
			param_str = fmt.Sprintf(`%s AND %s:"%v"`, param_str, k, v)
		}
	}
	for k, r := range p.rangeArgs {
		for _, v := range r {
			if param_str == "" {
				if v.opt == IsNot {
					param_str = fmt.Sprintf(`%s%s:"%v"`, v.opt, k, v.val)
				} else {
					param_str = fmt.Sprintf(`%s:%s%v`, k, v.opt, v.val)
				}
			} else {
				if v.opt == IsNot {
					param_str = fmt.Sprintf(`%s AND %s%s:"%v"`, param_str, v.opt, k, v.val)
				} else {
					param_str = fmt.Sprintf(`%s AND %s:%s%v`, param_str, k, v.opt, v.val)
				}
			}
		}
	}
	if p.andStr != "" {
		if param_str != "" {
			param_str = fmt.Sprintf("(%s) AND (%s)", param_str, p.andStr)
		} else {
			param_str = p.andStr
		}
	}
	if p.orStr != "" {
		if param_str != "" {
			param_str = fmt.Sprintf("(%s) OR (%s)", param_str, p.orStr)
		} else {
			param_str = p.orStr
		}
	}
	return param_str
}

type SearchParam func(p *param)

func GroupAND(and ...SearchParam) SearchParam {
	return func(p *param) {
		tmp := &param{}
		for _, val := range and {
			val(tmp)
		}
		p.andStr = tmp.packageOut()
	}
}

func GroupOR(or ...SearchParam) SearchParam {
	return func(p *param) {
		tmp := &param{}
		for _, val := range or {
			val(tmp)
		}
		p.orStr = tmp.packageOut()
	}
}

func WithEqual(key string, val any) SearchParam {
	return func(p *param) {
		if p.termArgs == nil {
			p.termArgs = make(map[string]any)
		}
		p.termArgs[key] = val
	}
}
func WithRange(key string, opt RangOpt, val any) SearchParam {
	return func(p *param) {
		if p.rangeArgs == nil {
			p.rangeArgs = make(map[string][]*rang)
		}
		_, ok := p.rangeArgs[key]
		if !ok {
			p.rangeArgs[key] = make([]*rang, 0)
		}
		p.rangeArgs[key] = append(p.rangeArgs[key], &rang{opt: opt, val: val})
	}
}
func WithSort(key string, opt SortOpt) SearchParam {
	return func(p *param) {
		if p.sortArgs == nil {
			p.sortArgs = make([]string, 0)
		}
		p.sortArgs = append(p.sortArgs, fmt.Sprintf("%s:%s", key, opt))
	}
}

var instanceMap = make(map[string]*ElasticClient)
var lock sync.RWMutex

type ElasticCfg struct {
	Host     []string `json:"host"`
	UserName string   `json:"username"`
	Pwd      string   `json:"pwd"`
}

func GetInstanceElastic(cfg *ElasticCfg, debug bool) (*ElasticClient, error) {
	if key, err := json.Marshal(cfg); err == nil {
		lock.RLock()
		r, ok := instanceMap[string(key)]
		lock.RUnlock()
		if r == nil || !ok {
			elcfg := es7.Config{
				Addresses: cfg.Host,
			}
			if cfg.UserName != "" {
				elcfg.Username = cfg.UserName
			}
			if cfg.Pwd != "" {
				elcfg.Password = cfg.Pwd
			}
			elcfg.EnableDebugLogger = debug
			cli, err := es7.NewClient(elcfg)
			if err != nil || cli == nil {
				return nil, err
			}

			_, err = cli.Ping()
			if err != nil {
				return nil, err
			}
			// fmt.Println(resp)
			lock.Lock()
			r = &ElasticClient{Client: cli}
			instanceMap[string(key)] = r
			lock.Unlock()
		}
		return r, nil
	} else {
		return nil, err
	}
}

type ElasticClient struct {
	Client *es7.Client
}

func (e *ElasticClient) InsertNewRcord(index string, id int, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	tmp := make([]func(*esapi.IndexRequest), 0)
	tmp = append(tmp, e.Client.Index.WithRefresh("true"))
	if id > -1 {
		tmp = append(tmp, e.Client.Index.WithDocumentID(strconv.Itoa(id)))
	}
	tmp = append(tmp, e.Client.Index.WithRouting("user_order"))
	res, err := e.Client.Index(index, bytes.NewBuffer(body), tmp...)
	if err != nil {
		// fmt.Println(err)
		return err
	}
	defer res.Body.Close()
	// fmt.Println(res)
	return nil
}

func (e *ElasticClient) Del(index string, id int) error {
	_, err := e.Client.Delete(index, strconv.Itoa(id), e.Client.Delete.WithRefresh("true"))
	return err
}
func (e *ElasticClient) Search(index []string, parm ...SearchParam) ([][]byte, error) {
	search_param := &param{}
	for _, tmp := range parm {
		tmp(search_param)
	}
	param_str := search_param.packageOut()
	opts := make([]func(*esapi.SearchRequest), 0)
	opts = append(opts, e.Client.Search.WithIndex(index...))
	if len(search_param.sortArgs) > 0 {
		opts = append(opts, e.Client.Search.WithSort(search_param.sortArgs...))
	}
	opts = append(opts, e.Client.Search.WithScroll(time.Second*60))
	if param_str != "" {
		opts = append(opts,
			e.Client.Search.WithQuery(param_str),
		)
	}
	resp_data := make([][]byte, 0)
	response, err := e.Client.Search(opts...)
	if err != nil {
		return nil, err
	}
	fmt.Println(response)
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("statsu code err[%d]", response.StatusCode)
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	scroll := &commonResp{}
	err = json.Unmarshal(data, scroll)
	if err != nil {
		return nil, err
	}
	resp_data = append(resp_data, data)
	scr_id := scroll.ScrollID
	for {
		if scr_id == "" {
			return resp_data, nil
		}
		d, s, e := e.Scroll(scr_id)
		if e != nil {
			return nil, e
		}
		if s == "done" || s == "" {
			return resp_data, nil
		}
		resp_data = append(resp_data, d)
		scr_id = s
	}
}

func (e *ElasticClient) Scroll(srcroll_id string) ([]byte, string, error) {
	response, err := e.Client.Scroll(
		e.Client.Scroll.WithScroll(time.Second*60),
		e.Client.Scroll.WithScrollID(srcroll_id),
	)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, "", fmt.Errorf("statsu code err")
	}
	// fmt.Println(response)
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}
	scroll := &commonResp{}
	err = json.Unmarshal(data, scroll)
	if err != nil {
		return nil, "", err
	}
	if len(scroll.Hits.Hits) == 0 {
		return nil, "done", nil
	}
	return data, scroll.ScrollID, nil
}
