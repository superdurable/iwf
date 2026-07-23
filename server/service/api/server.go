// Copyright (c) 2022-2026 Super Durable, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package api

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/superdurable/iwf/config"
	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service/common/log"
	"github.com/superdurable/iwf/service/common/log/tag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const pendingPhaseMsg = "RPC body pending Phase 2–4 UnifiedClient / interpreter rewrite"

// backendTypeCadence matches service.BackendTypeCadence without importing service (iwfidl).
const backendTypeCadence = "cadence"

// BackendClient is the Phase 1 shell dependency; Phase 2 widens this to UnifiedClient.
type BackendClient interface {
	GetBackendType() string
}

// BackendTypeFunc adapts a string-returning backend type getter.
type BackendTypeFunc func() string

func (f BackendTypeFunc) GetBackendType() string { return f() }

// BlobStore is the Phase 1 shell store surface (avoids importing blobstore helpers that still reference iwfidl).
type BlobStore interface{}

// Server hosts FlowService, InternalService, and grpc.health on one plaintext port.
type Server struct {
	cfg        *config.ApiConfig
	grpcServer *grpc.Server
	healthSrv  *health.Server
	flow       *serviceImpl
	listener   net.Listener
	serving    atomic.Bool
	logger     log.Logger
	readyCheck func(context.Context) error
}

// NewServer constructs the gRPC API server. apiCfg and client must be non-nil.
func NewServer(
	apiCfg *config.ApiConfig,
	extStore *config.ExternalStorageConfig,
	client BackendClient,
	logger log.Logger,
	store BlobStore,
	readyCheck func(context.Context) error,
) *Server {
	if apiCfg == nil {
		panic("apiCfg must not be nil")
	}
	if client == nil {
		panic("client must not be nil")
	}
	if logger == nil {
		panic("logger must not be nil")
	}
	if store == nil {
		panic("store must not be nil")
	}
	if extStore == nil {
		panic("extStore must not be nil")
	}
	if readyCheck == nil {
		readyCheck = func(context.Context) error { return nil }
	}

	flow := newServiceImpl(apiCfg, extStore, client, logger, store)
	maxMsg := apiCfg.EffectiveGrpcMaxMessageBytes()
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	healthSrv.SetServingStatus(iwfpb.FlowService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_NOT_SERVING)
	healthSrv.SetServingStatus(iwfpb.InternalService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_NOT_SERVING)

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsg),
		grpc.MaxSendMsgSize(maxMsg),
		grpc.ChainUnaryInterceptor(unaryRecover(logger), unaryLog(logger)),
	)
	iwfpb.RegisterFlowServiceServer(grpcServer, flow)
	iwfpb.RegisterInternalServiceServer(grpcServer, flow)
	healthpb.RegisterHealthServer(grpcServer, healthSrv)

	return &Server{
		cfg:        apiCfg,
		grpcServer: grpcServer,
		healthSrv:  healthSrv,
		flow:       flow,
		logger:     logger,
		readyCheck: readyCheck,
	}
}

// Run listens on :Port and serves until the process exits or GracefulStop.
func (s *Server) Run() error {
	port := s.cfg.Port
	if port == 0 {
		port = config.DefaultApiPort
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s.listener = lis
	s.logger.Info("FlowService gRPC listening", tag.Value(lis.Addr().String()))

	if err := s.refreshReadiness(context.Background()); err != nil {
		s.logger.Error("initial readiness check failed", tag.Error(err))
	} else {
		s.setServing(true)
	}

	return s.grpcServer.Serve(lis)
}

// GracefulStop marks not-serving, then GracefulStop with a bounded fallback to Stop.
func (s *Server) GracefulStop(timeout time.Duration) {
	s.setServing(false)
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		s.grpcServer.Stop()
	}
}

func (s *Server) setServing(serving bool) {
	s.serving.Store(serving)
	st := healthpb.HealthCheckResponse_NOT_SERVING
	if serving {
		st = healthpb.HealthCheckResponse_SERVING
	}
	s.healthSrv.SetServingStatus("", st)
	s.healthSrv.SetServingStatus(iwfpb.FlowService_ServiceDesc.ServiceName, st)
	s.healthSrv.SetServingStatus(iwfpb.InternalService_ServiceDesc.ServiceName, st)
}

func (s *Server) refreshReadiness(ctx context.Context) error {
	return s.readyCheck(ctx)
}

func unaryRecover(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("gRPC handler panic", tag.Value(fmt.Sprintf("%v", r)), tag.Value(info.FullMethod))
				err = status.Errorf(codes.Internal, "internal panic")
			}
		}()
		return handler(ctx, req)
	}
}

func unaryLog(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		logger.Debug("gRPC call",
			tag.Value(info.FullMethod),
			tag.Value(time.Since(start).String()),
			tag.Error(err),
		)
		return resp, err
	}
}

type serviceImpl struct {
	iwfpb.UnimplementedFlowServiceServer
	iwfpb.UnimplementedInternalServiceServer

	apiCfg   *config.ApiConfig
	extStore *config.ExternalStorageConfig
	client   BackendClient
	store    BlobStore
	logger   log.Logger
	started  time.Time
	hostname string
}

func newServiceImpl(
	apiCfg *config.ApiConfig,
	extStore *config.ExternalStorageConfig,
	client BackendClient,
	logger log.Logger,
	store BlobStore,
) *serviceImpl {
	hostname, _ := os.Hostname()
	return &serviceImpl{
		apiCfg:   apiCfg,
		extStore: extStore,
		client:   client,
		store:    store,
		logger:   logger,
		started:  time.Now(),
		hostname: hostname,
	}
}

func (s *serviceImpl) HealthCheck(ctx context.Context, _ *emptypb.Empty) (*iwfpb.HealthInfo, error) {
	return &iwfpb.HealthInfo{
		Condition: "OK",
		Hostname:  s.hostname,
		Duration:  int32(time.Since(s.started).Seconds()),
	}, nil
}

func (s *serviceImpl) StartFlow(context.Context, *iwfpb.StartFlowRequest) (*iwfpb.StartFlowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) PublishToChannel(context.Context, *iwfpb.PublishToChannelRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) StopFlow(context.Context, *iwfpb.StopFlowRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) GetAttributes(context.Context, *iwfpb.GetAttributesRequest) (*iwfpb.GetAttributesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) SetAttributes(context.Context, *iwfpb.SetAttributesRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) LoadBlobs(ctx context.Context, req *iwfpb.LoadBlobsRequest) (*iwfpb.LoadBlobsResponse, error) {
	if req == nil || len(req.GetBlobIds()) == 0 {
		return &iwfpb.LoadBlobsResponse{Values: map[string]*iwfpb.Value{}}, nil
	}
	// Full hydrate path lands with the Phase 2 Value model + blobstore helpers.
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) WaitForFlow(context.Context, *iwfpb.WaitForFlowRequest) (*iwfpb.WaitForFlowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) SearchFlows(context.Context, *iwfpb.SearchFlowsRequest) (*iwfpb.SearchFlowsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) ResetFlow(context.Context, *iwfpb.ResetFlowRequest) (*iwfpb.ResetFlowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) InvokeRPC(context.Context, *iwfpb.InvokeRPCRequest) (*iwfpb.InvokeRPCResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) SkipTimer(context.Context, *iwfpb.SkipTimerRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) UpdateFlowConfig(context.Context, *iwfpb.UpdateFlowConfigRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) WaitForStepCompletion(ctx context.Context, req *iwfpb.WaitForStepCompletionRequest) (*iwfpb.WaitForStepCompletionResponse, error) {
	if s.client.GetBackendType() == backendTypeCadence {
		return nil, status.Errorf(codes.Unimplemented, "WaitForStepCompletion requires Temporal synchronous update")
	}
	_ = ctx
	_ = req
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) WaitForAttribute(ctx context.Context, req *iwfpb.WaitForAttributeRequest) (*emptypb.Empty, error) {
	if s.client.GetBackendType() == backendTypeCadence {
		return nil, status.Errorf(codes.Unimplemented, "WaitForAttribute requires Temporal synchronous update")
	}
	_ = ctx
	_ = req
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) TriggerContinueAsNew(context.Context, *iwfpb.TriggerContinueAsNewRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}

func (s *serviceImpl) DumpFlowForContinueAsNew(
	context.Context, *iwfpb.ContinueAsNewDumpRequest,
) (*iwfpb.ContinueAsNewDumpResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, pendingPhaseMsg)
}
