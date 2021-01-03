package certCheck

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
	"time"
)

type CertCheckConfig struct {
	Host string `json:"host"`
}

type CertCheckResult struct {
	ValidFrom string `json:"validFrom"`
	ValidTo   string `json:"validTo"`
	DaysLeft  int    `json:"daysLeft"`
	Expired   bool   `json:"expired"`
	Host      string `json:"host"`
}

func Handle(config string) interface{} {
	parsedConfig := CertCheckConfig{}
	err := json.Unmarshal([]byte(config), &parsedConfig)
	if err != nil {
		// TODO: Handle error
		return nil
	}

	fetchedCert, err := getServerCertificate(parsedConfig.Host)
	if err != nil {
		// TODO: Handle error
		return nil
	}

	oneDay := time.Hour * 24
	return CertCheckResult{
		ValidFrom: fetchedCert.NotBefore.Format(time.RFC3339),
		ValidTo:   fetchedCert.NotAfter.Format(time.RFC3339),
		DaysLeft:  int(fetchedCert.NotAfter.Sub(time.Now()).Truncate(oneDay) / oneDay),
		Expired:   fetchedCert.NotAfter.Before(time.Now()),
		Host:      parsedConfig.Host,
	}
}

func getServerCertificate(host string) (*x509.Certificate, error) {
	timeoutSeconds := 5

	d := &net.Dialer{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	conn, err := tls.DialWithDialer(d, "tcp", host+":443", &tls.Config{
		InsecureSkipVerify: true,
		MaxVersion:         0,
	})
	if err != nil {
		return &x509.Certificate{}, err
	}
	defer conn.Close()

	return conn.ConnectionState().PeerCertificates[0], nil
}
