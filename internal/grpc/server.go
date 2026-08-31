package grpc

import (
	"context"
	"fmt"
	"net"

	"xboard-go/config"
	"xboard-go/internal/repository"
	"xboard-go/internal/service"
	"xboard-go/pkg/logger"

	"google.golang.org/grpc"
)

// NodeBroadcaster is the global broadcaster instance, accessible from HTTP
// handlers to push config/user updates to connected gRPC nodes.
var NodeBroadcaster *Broadcaster

// GRPCServer wraps the gRPC server lifecycle and its dependencies.
type GRPCServer struct {
	server      *grpc.Server
	listener    net.Listener
	port        int
	broadcaster *Broadcaster
}

// NewServer creates and configures a gRPC server.
//
// Parameters:
//   - port: TCP port to listen on (e.g. 50051)
//   - cfg:  application configuration
//
// It wires up repositories, services, the broadcaster, the auth interceptor,
// and the NodeService implementation.
func NewServer(port int, cfg *config.Config) *GRPCServer {
	// --- dependencies ---
	nodeRepo := repository.NewNodeRepository()
	trafficSvc := service.NewTrafficService()
	nodeUserRepo := repository.NewNodeUserRepository()
	userRepo := repository.NewUserRepository()
	trafficLogRepo := repository.NewTrafficLogRepository()
	uniProxySvc := service.NewUniProxyService(nodeRepo, nodeUserRepo, userRepo, trafficLogRepo, trafficSvc)

	// 复用已有的 Broadcaster（可能已在 main.go 中初始化供 REST/WS 使用）
	broadcaster := NodeBroadcaster
	if broadcaster == nil {
		broadcaster = NewBroadcaster()
		NodeBroadcaster = broadcaster // expose globally for HTTP handlers
	}
	nodeSvcServer := NewNodeServiceServer(nodeRepo, trafficSvc, uniProxySvc, broadcaster, cfg)

	// --- interceptors ---
	auth := NewAuthInterceptor()

	// --- gRPC server options ---
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(auth.UnaryInterceptor),
		grpc.StreamInterceptor(auth.StreamInterceptor),
		grpc.ForceServerCodec(jsonCodec{}),
	}

	server := grpc.NewServer(opts...)

	// --- register the NodeService ---
	RegisterNodeServiceServer(server, nodeSvcServer)

	return &GRPCServer{
		server:      server,
		port:        port,
		broadcaster: broadcaster,
	}
}

// Start begins listening for gRPC connections.
// It blocks until the server is stopped; call this in a goroutine.
func (s *GRPCServer) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("gRPC server failed to listen on %s: %w", addr, err)
	}
	s.listener = lis

	logger.Sugar().Infof("gRPC server listening on %s", addr)

	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("gRPC server serve error: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the gRPC server.
func (s *GRPCServer) Stop() {
	logger.Sugar().Info("gRPC server shutting down...")
	s.server.GracefulStop()
}

// Broadcaster exposes the broadcaster instance so that other parts of the
// application (e.g. admin handlers) can push config/user updates to nodes.
func (s *GRPCServer) Broadcaster() *Broadcaster {
	return s.broadcaster
}

// ---------------------------------------------------------------------------
// Service registration (manual, replacing protoc-generated RegisterXxxServer)
// ---------------------------------------------------------------------------

// RegisterNodeServiceServer registers the NodeServiceServer with the gRPC server.
func RegisterNodeServiceServer(s *grpc.Server, srv NodeServiceServer) {
	s.RegisterService(&nodeService_ServiceDesc, srv)
}

// --- Unary method handlers ---

func _NodeService_Handshake_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandshakeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).Handshake(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/node.NodeService/Handshake",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).Handshake(ctx, req.(*HandshakeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/node.NodeService/GetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).GetConfig(ctx, req.(*NodeConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_GetUsers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).GetUsers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/node.NodeService/GetUsers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).GetUsers(ctx, req.(*UserListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// --- Stream handler ---

func _NodeService_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	// Wrap the raw grpc.ServerStream into our typed NodeService_StreamServer.
	wrapped := NewNodeStreamServer(stream)
	return srv.(NodeServiceServer).Stream(wrapped)
}

// --- Service descriptor ---

var nodeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "xboard.node.NodeService",
	HandlerType: (*NodeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Handshake",
			Handler:    _NodeService_Handshake_Handler,
		},
		{
			MethodName: "GetConfig",
			Handler:    _NodeService_GetConfig_Handler,
		},
		{
			MethodName: "GetUsers",
			Handler:    _NodeService_GetUsers_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _NodeService_Stream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "node.proto",
}
