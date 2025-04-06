package telemetry

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

func NewTLSConfig(cfg TlsConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{}

	//1. Load client certificate and private key
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	//2. Load CA certificate
	if cfg.RootCAFile != "" {
		caCert, err := os.ReadFile(cfg.RootCAFile)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, err
		}
		tlsConfig.RootCAs = caCertPool
	}

	//3. Set TLS version
	if cfg.MinVersion != "" {
		switch cfg.MinVersion {
		case "TLS10":
			tlsConfig.MinVersion = tls.VersionTLS10
		case "TLS11":
			tlsConfig.MinVersion = tls.VersionTLS11
		case "TLS12":
			tlsConfig.MinVersion = tls.VersionTLS12
		}
	}

	//4. InsecureSkipVerify
	if cfg.InsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	return tlsConfig, nil
}

func LoadTLSCredentials(cfg TlsConfig) (credentials.TransportCredentials, error) {
	tlsConfig, err := NewTLSConfig(cfg)
	if err != nil {
		fmt.Println("failed to create TLS config:", err)
		return nil, err
	}

	return credentials.NewTLS(tlsConfig), nil
}
