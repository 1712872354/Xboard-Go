package grpc

import (
	"context"
	"strconv"

	"xboard-go/config"
	"xboard-go/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// metadata keys expected by the gRPC node protocol.
const (
	metaAuthToken = "authorization"
	metaNodeID    = "node_id"
)

// AuthInterceptor returns a pair of interceptors (unary + stream) that
// validate incoming node connections.
//
// Authentication rules:
//  1. The "authorization" metadata value is compared against config.Get().App.NodeAPIKey.
//  2. The "node_id" metadata value is parsed as uint32 and injected into the context.
//
// This mirrors the HTTP NodeAPIKeyAuth middleware but for gRPC metadata.
type AuthInterceptor struct{}

// NewAuthInterceptor creates a new AuthInterceptor.
func NewAuthInterceptor() *AuthInterceptor {
	return &AuthInterceptor{}
}

// UnaryInterceptor validates the token for every unary RPC.
func (a *AuthInterceptor) UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	nodeID, err := a.authenticate(ctx)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, ctxKeyNodeID{}, nodeID)
	return handler(ctx, req)
}

// StreamInterceptor validates the token when a bidirectional stream is opened.
func (a *AuthInterceptor) StreamInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	nodeID, err := a.authenticate(stream.Context())
	if err != nil {
		return err
	}
	wrapped := &wrappedStream{
		ServerStream: stream,
		ctx:          context.WithValue(stream.Context(), ctxKeyNodeID{}, nodeID),
	}
	return handler(srv, wrapped)
}

// authenticate extracts and validates the token from gRPC metadata.
// Returns the parsed nodeID on success.
func (a *AuthInterceptor) authenticate(ctx context.Context) (uint32, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// --- validate token ---
	tokens := md.Get(metaAuthToken)
	if len(tokens) == 0 || tokens[0] == "" {
		return 0, status.Error(codes.Unauthenticated, "missing authorization token")
	}
	token := tokens[0]

	expectedKey := config.Get().App.NodeAPIKey
	if token != expectedKey {
		// Also try JWT-based auth (future extension).
		// For now, only the static NodeAPIKey is accepted.
		logger.Sugar().Warnf("gRPC auth: invalid token received")
		return 0, status.Error(codes.Unauthenticated, "invalid authorization token")
	}

	// --- extract node_id ---
	nodeIDs := md.Get(metaNodeID)
	if len(nodeIDs) == 0 || nodeIDs[0] == "" {
		return 0, status.Error(codes.InvalidArgument, "missing node_id in metadata")
	}
	id64, err := strconv.ParseUint(nodeIDs[0], 10, 32)
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "invalid node_id: %s", nodeIDs[0])
	}

	return uint32(id64), nil
}

// ---------------------------------------------------------------------------
// Context helpers
// ---------------------------------------------------------------------------

// ctxKeyNodeID is a private context key type for storing the authenticated node ID.
type ctxKeyNodeID struct{}

// NodeIDFromContext extracts the node ID that was stored by the auth interceptor.
func NodeIDFromContext(ctx context.Context) (uint32, bool) {
	id, ok := ctx.Value(ctxKeyNodeID{}).(uint32)
	return id, ok
}

// ---------------------------------------------------------------------------
// wrappedStream — injects the enriched context into the stream.
// ---------------------------------------------------------------------------

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
