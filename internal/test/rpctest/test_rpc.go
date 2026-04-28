package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	agentregistryv1 "hypervisor/internal/transport/grpc/agentregistryv1"
)

func main() {
	target := "hypervisor.auroracloud.local:9443"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := agentregistryv1.NewAgentRegistryClient(conn)
	resp, err := client.BootstrapEnrollAgent(ctx, &agentregistryv1.BootstrapEnrollAgentRequest{
		BootstrapToken: "invalid-token",
		RequestedAgentId: "test-agent",
		CsrPem: "test-csr",
		Hostname: "test-host",
	})

	if err != nil {
		fmt.Printf("RPC failed: %v\n", err)
	} else {
		fmt.Printf("RPC succeeded: %v\n", resp)
	}
}
