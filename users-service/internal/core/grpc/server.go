package grpcclient

import (
	"crypto/tls"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewGRPCServer(cfg *config.Config, interceptors ...grpc.UnaryServerInterceptor) *grpc.Server {
	var opts []grpc.ServerOption

	if len(interceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(interceptors...))
	}

	if cfg.TLSEnabled {
		creds, err := loadServerTLSCredentials(cfg)
		if err != nil {
			panic(fmt.Sprintf("failed to load TLS credentials: %v", err))
		}
		opts = append(opts, grpc.Creds(creds))
	}

	return grpc.NewServer(opts...)
}

func loadServerTLSCredentials(cfg *config.Config) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load server cert: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(tlsConfig), nil
}
