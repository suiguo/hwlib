package smtp

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/gomail.v2"
)

var regMail = `<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="utf-8">
    <meta name="theme-color" content="#000000">
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <meta name="browsermode" content="application">
    <meta name="full-screen" content="yes" />
    <meta name="x5-fullscreen" content="true" />
    <meta name="x5-page-mode" content="app" />
    <meta name="360-fullscreen" content="true" />
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <meta http-equiv="x-dns-prefetch-control" content="on" />
    <meta name="viewport"
        content="width=320.1,initial-scale=1,minimum-scale=1,maximum-scale=1,user-scalable=no,minimal-ui" />
    <meta name="apple-mobile-web-app-title" content="yeziyuan" />
    <meta content="telephone=no" name="format-detection" />
    <meta name="fullscreen" content="yes">
    <title>HuiOne</title>
    <style>
        * {
            margin: 0;
            padding: 0;
        }

        #root {
            display: flex;
            flex-direction: column;
            flex: 1;
        }

        .head {
            width: 100%;
            height: 4px;
            background-color: #F5222D;
        }

        .log_div {
            margin-top: 50px;
            margin-left: 50px;
            display: flex;
            flex-direction: row;
        }

        .log_img {
            width: 121px;
            height: 38px;
        }
        .page_title {
            margin-top: 36px;
            margin-left: 50px;
            color: #000000;
            font-size: 30px;
            font-weight: 500;
        }
        .welcome_text {
            margin-top: 16px;
            margin-left: 50px;
            color: #000000;
            font-size: 14px;
            font-weight: 400;
        }
        .verification_code_text {
            margin-top: 36px;
            margin-left: 50px;
            color: #000000;
            font-size: 14px;
            font-weight: 400;
        }
        .verification_code_num {
            margin-top: 8px;
            margin-left: 50px;
            color: #F5222D;
            font-size: 30px;
            font-weight: 500;
        }
        .explain_text {
            display: flex;
            flex-direction: row;
            margin-top: 16px;
            margin-left: 50px;
            color: #000000;
            font-size: 14px;
            font-weight: 400;
        }
        .explain_text .explain_text_red {
            color: #F5222D;
        }
        .statement_text {
            margin-top: 64px;
            margin-left: 50px;
            color: #707A8A;
            font-size: 14px;
            font-weight: 400;
        }
    </style>
</head>

<body>
    <div id="root">
        <div class="head" />
        <div class="log_div">
            <img class="log_img" src='https://hwwallet.s3.ap-northeast-1.amazonaws.com/v1/mail/favicon.png' />
        </div>
        <div class="page_title">账户注册</div>
        <div class="welcome_text">欢迎来到HuiOne机构，使用以下验证码确认注册。</div>
        <div class="verification_code_text">您的验证码</div>
        <div class="verification_code_num">#code</div>
        <div class="explain_text">验证码的有效期为5分钟。请不要与任何人分享此代码。</div>
        <div class="explain_text">如非您本人操作，请立即<div class="explain_text_red">联系客服</div>。</div>
        <div class="statement_text">本邮件由系统自动发送，请勿回复。</div>
    </div>
</body>
</html>`
var clientMap sync.Map

type Option func(*cfg)
type cfg struct {
	Head       map[string][]string
	Body       string
	BodyType   string
	Attach     []string
	AddrHeader []string
}

type SmtpClient struct {
	*gomail.Dialer
}

func newCfg() *cfg {
	tmp := &cfg{
		Head:       make(map[string][]string),
		Body:       "",
		Attach:     make([]string, 0),
		AddrHeader: make([]string, 0),
	}
	tmp.Head["From"] = make([]string, 0)
	tmp.Head["To"] = make([]string, 0)
	tmp.Head["Subject"] = make([]string, 0)
	return tmp
}

/*
m.SetHeader("From", "alex@example.com")
m.SetHeader("To", "bob@example.com", "cora@example.com")
m.SetAddressHeader("Cc", "dan@example.com", "Dan")
m.SetHeader("Subject", "Hello!")
m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
m.Attach("/home/Alex/lolcat.jpg")
*/
//发送人
func WithFrom(from ...string) Option {
	return func(in *cfg) {
		in.Head["From"] = append(in.Head["From"], from...)
	}
}

// 发送到
func WithTo(to ...string) Option {
	return func(in *cfg) {
		in.Head["To"] = append(in.Head["To"], to...)
	}
}

// 抄送
func WithAddrHeader(feild string, addr string, name string) Option {
	return func(in *cfg) {
		in.AddrHeader = []string{feild, addr, name}
	}
}

// 标题
func WithTitle(sub string) Option {
	return func(in *cfg) {
		in.Head["Subject"] = append(in.Head["Subject"], sub)
	}
}

// 邮件内容("text/html","html内容")
func WithBody(body_type string, body string) Option {
	return func(in *cfg) {
		in.Body = body
		in.BodyType = body_type
	}
}
func WithAttachs(path ...string) Option {
	return func(in *cfg) {
		in.Attach = append(in.Attach, path...)
	}
}

// 默认注册code 模板
func WithBodyReg(code string) Option {
	body := strings.ReplaceAll(regMail, "#code", code)
	return WithBody("text/html", body)
}

func GetClient(host string, port int, username string, password string) *SmtpClient {
	instacne_key := fmt.Sprintf("%s_%d_%s_%s", host, port, username, password)
	tmp, ok := clientMap.Load(instacne_key)
	if ok {
		return tmp.(*SmtpClient)
	}
	cli := gomail.NewDialer(host, port, username, password)
	cli.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	tmp = &SmtpClient{Dialer: cli}
	clientMap.Store(instacne_key, tmp)
	return tmp.(*SmtpClient)
}

// 发送邮件
func (cli *SmtpClient) SendMail(opt ...Option) error {
	mail_cfg := newCfg()
	for _, val := range opt {
		val(mail_cfg)
	}
	mail := gomail.NewMessage()
	for key, val := range mail_cfg.Head {
		mail.SetHeader(key, val...)
	}
	if len(mail_cfg.AddrHeader) == 3 {
		mail.SetAddressHeader(mail_cfg.AddrHeader[0], mail_cfg.AddrHeader[1], mail_cfg.AddrHeader[2])
	}
	if mail_cfg.BodyType != "" {
		mail.SetBody(mail_cfg.BodyType, mail_cfg.Body)
	}
	for _, val := range mail_cfg.Attach {
		mail.Attach(val)
	}
	return cli.DialAndSend(mail)
}
