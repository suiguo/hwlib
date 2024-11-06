package sms

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/nyaruka/phonenumbers"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	// _ "github.com/nyaruka/phonenumbers"
)

type SmsType string

// const Twilio = "twilio"
const TwilioUrl = "https://api.twilio.com/2010-04-01/Accounts/"
const (
	Twilio SmsType = "twilio"
)

type SmsClient interface {
	SendSms(phone string, msg string, opt ...SmsOpt) error
}
type twilioClient struct {
	cfg *smsCfg
	*twilio.RestClient
}
type smsCfg struct {
	Account   string
	MessageId string
	AuthToken string
}
type SmsOpt func(c *smsCfg)

func WithAccount(account string) SmsOpt {
	return func(c *smsCfg) {
		c.Account = account
	}
}
func WithMsgId(msgId string) SmsOpt {
	return func(c *smsCfg) {
		c.MessageId = msgId
	}
}

func WithToken(token string) SmsOpt {
	return func(c *smsCfg) {
		c.AuthToken = token
	}
}

// 是否是合法的手机号 国家号用" "隔开true 合法 false 不合法
func isvaildPhone(phone string) bool {
	tmp := strings.Split(strings.ReplaceAll(phone, "+", ""), " ")
	if len(tmp) != 2 {
		return false
	}
	contry_code, err := strconv.Atoi(tmp[0])
	if err != nil {
		return false
	}
	region := phonenumbers.GetRegionCodeForCountryCode(contry_code)
	number, err := phonenumbers.Parse(tmp[1], region)
	if err != nil || number == nil {
		return false
	}
	return phonenumbers.IsValidNumber(number)
}

func (t *twilioClient) SendSms(phone string, msg string, opt ...SmsOpt) error {
	cfg := &smsCfg{}
	for _, val := range opt {
		val(cfg)
	}
	if t.cfg.Account == "" {
		return fmt.Errorf("twilio need accout")
	}
	if t.cfg.MessageId == "" {
		return fmt.Errorf("twilio need msg id")
	}
	if t.cfg.AuthToken == "" {
		return fmt.Errorf("twilio need auth token")
	}
	if !isvaildPhone(phone) {
		return fmt.Errorf("not vaild phone number")
	}
	if t.RestClient == nil {
		t.RestClient = twilio.NewRestClientWithParams(twilio.ClientParams{
			Username:   t.cfg.Account,   //TWILIO_ACCOUNT_SID
			Password:   t.cfg.AuthToken, //
			AccountSid: t.cfg.Account,   //
		})
	}
	// client
	params := &api.CreateMessageParams{}
	params = params.SetBody(msg).SetMessagingServiceSid(t.cfg.MessageId).SetTo(strings.TrimSpace(phone))
	// params.set
	resp, err := t.Api.CreateMessage(params)
	if err != nil {
		return err
	} else {
		if resp.Sid == nil {
			return fmt.Errorf("sid is nil")
		}
	}
	return nil
}

func GetSmsClient(sms_type SmsType, opt ...SmsOpt) (SmsClient, error) {
	cfg := &smsCfg{}
	for _, val := range opt {
		val(cfg)
	}
	switch sms_type {
	case Twilio:
		return &twilioClient{cfg: cfg}, nil
	}
	return nil, nil
}

func Post(request_url string, data map[string]string, user string, pwd string) ([]byte, error) {
	req_data := url.Values{}
	for key, val := range data {
		req_data.Add(key, val)
	}
	req, err := http.NewRequest("POST", request_url, bytes.NewReader([]byte(req_data.Encode())))
	if err != nil {
		return nil, err
	}
	auth := fmt.Sprintf("%s:[%s]", user, pwd)

	req.Header.Set("Authorization", "Basic "+base64.RawURLEncoding.EncodeToString([]byte(auth)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
