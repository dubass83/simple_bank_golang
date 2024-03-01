package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	GatewayUserAgentHeader = "grpcgateway-user-agent"
	GatewayClientIP        = "x-forwarded-for"
	UserAgentHeader        = "user-agent"
)

type Metadata struct {
	ClientIP  string
	UserAgent string
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}

	if md, ok := metadata.FromIncomingContext(ctx); ok {

		if userAgent := md.Get(GatewayUserAgentHeader); len(userAgent) > 0 {
			mtdt.UserAgent = userAgent[0]
		}
		if userAgent := md.Get(UserAgentHeader); len(userAgent) > 0 {
			mtdt.UserAgent = userAgent[0]
		}
		if clientIP := md.Get(GatewayClientIP); len(clientIP) > 0 {
			mtdt.ClientIP = clientIP[0]
		}
	}
	if p, ok := peer.FromContext(ctx); ok {
		mtdt.ClientIP = p.Addr.String()
	}

	return mtdt
}
