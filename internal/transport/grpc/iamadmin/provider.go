package iamadmin

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"hypervisor/internal/config"
	"hypervisor/internal/errorx"
	"hypervisor/internal/transport/grpc/iamv1"
	"hypervisor/internal/transport/http/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type Provider struct {
	conn   *grpc.ClientConn
	client iamv1.AdminSessionServiceClient
	cfg    config.IAMCfg
}

func NewProvider(cfg config.IAMCfg) (*Provider, error) {
	cfg = normalizeIAMConfig(cfg)
	tlsConfig, err := buildClientTLSConfig(cfg)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(cfg.GRPCTarget, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		return nil, fmt.Errorf("iam grpc: dial: %w", err)
	}

	return &Provider{conn: conn, client: iamv1.NewAdminSessionServiceClient(conn), cfg: cfg}, nil
}

func (p *Provider) Close() error {
	if p == nil || p.conn == nil {
		return nil
	}
	return p.conn.Close()
}

func (p *Provider) AuthorizeAdminSession(ctx context.Context, input middleware.AdminSessionAuthInput) (*middleware.AdminSessionContext, error) {
	if p == nil || p.conn == nil || p.client == nil {
		return nil, errorx.ErrUnauthorized
	}
	if ctx == nil {
		ctx = context.Background()
	}

	timeout := p.cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := p.client.AuthorizeSession(callCtx, &iamv1.AuthorizeAdminSessionRequest{
		SessionToken: input.SessionToken,
		DeviceId:     input.DeviceID,
		DeviceSecret: input.DeviceSecret,
		ClientIp:     input.ClientIP,
		UserAgent:    input.UserAgent,
	})
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return nil, errorx.ErrUnauthorized
		}
		return nil, fmt.Errorf("iam grpc: authorize admin session: %w", err)
	}

	return &middleware.AdminSessionContext{
		AdminUserID:  resp.GetAdminUserId(),
		DisplayName:  resp.GetDisplayName(),
		CredentialID: resp.GetCredentialId(),
		DeviceID:     resp.GetDeviceId(),
		SessionID:    resp.GetSessionId(),
	}, nil
}

func normalizeIAMConfig(cfg config.IAMCfg) config.IAMCfg {
	cfg.GRPCTarget = strings.TrimSpace(cfg.GRPCTarget)
	cfg.GRPCTLSCACertPath = strings.TrimSpace(cfg.GRPCTLSCACertPath)
	cfg.GRPCTLSCertPath = strings.TrimSpace(cfg.GRPCTLSCertPath)
	cfg.GRPCTLSKeyPath = strings.TrimSpace(cfg.GRPCTLSKeyPath)
	cfg.GRPCTLSServerName = strings.TrimSpace(cfg.GRPCTLSServerName)
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 3 * time.Second
	}
	return cfg
}

func buildClientTLSConfig(cfg config.IAMCfg) (*tls.Config, error) {
	if cfg.GRPCTarget == "" {
		return nil, errors.New("iam grpc: target is required")
	}
	if cfg.GRPCTLSCACertPath == "" || cfg.GRPCTLSCertPath == "" || cfg.GRPCTLSKeyPath == "" {
		return nil, errors.New("iam grpc: mtls ca, cert and key are required")
	}

	caPEM, err := os.ReadFile(cfg.GRPCTLSCACertPath)
	if err != nil {
		return nil, fmt.Errorf("iam grpc: read ca: %w", err)
	}
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(caPEM); !ok {
		return nil, errors.New("iam grpc: append ca")
	}

	cert, err := tls.LoadX509KeyPair(cfg.GRPCTLSCertPath, cfg.GRPCTLSKeyPath)
	if err != nil {
		return nil, fmt.Errorf("iam grpc: load client cert: %w", err)
	}

	return &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{cert},
		ServerName:   cfg.GRPCTLSServerName,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
