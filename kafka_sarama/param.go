package kafkasarama

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	"github.com/xdg-go/scram"
)

var (
	sHA256 scram.HashGeneratorFcn = sha256.New
	sHA512 scram.HashGeneratorFcn = sha512.New
)

type xDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (x *xDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}

func (x *xDGSCRAMClient) Step(challenge string) (response string, err error) {
	response, err = x.ClientConversation.Step(challenge)
	return
}

func (x *xDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}

func createTLSConfiguration(certFile string, keyFile string, caFile string, skip bool) (t *tls.Config) {
	t = &tls.Config{
		InsecureSkipVerify: skip,
	}
	if certFile != "" && keyFile != "" && caFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Fatal(err)
		}

		caCert, err := os.ReadFile(caFile)
		if err != nil {
			log.Fatal(err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: skip,
		}
	}
	return t
}
