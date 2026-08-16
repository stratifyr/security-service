// versions:
// 	gofr-cli v0.6.0
// 	gofr.dev v1.37.0
// 	source: security-service.proto

package grpc

import (
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/proto"
)

// Register the gRPC service in your app using the following code in your main.go:
//
// grpc.RegisterSecurityServiceServerWithGofr(app, &grpc.NewSecurityServiceGoFrServer())
//
// SecurityServiceGoFrServer defines the gRPC server implementation.
// Customize the struct with required dependencies and fields as needed.

type SecurityServiceGoFrServer struct {
	health *healthServer
}

func (s *SecurityServiceGoFrServer) GetMarketDays(ctx *gofr.Context) (any, error) {
	return &proto.GetMarketDaysResponse{}, nil
}
func (s *SecurityServiceGoFrServer) GetMetrics(ctx *gofr.Context) (any, error) {
	return &proto.GetMetricsResponse{}, nil
}
func (s *SecurityServiceGoFrServer) GetSecurities(ctx *gofr.Context) (any, error) {
	return &proto.GetSecuritiesResponse{}, nil
}
func (s *SecurityServiceGoFrServer) UpdateSecurity(ctx *gofr.Context) (any, error) {
	return &proto.UpdateSecurityResponse{}, nil
}
func (s *SecurityServiceGoFrServer) CreateOrUpdateSecurityStat(ctx *gofr.Context) (any, error) {
	return &proto.CreateOrUpdateSecurityStatResponse{}, nil
}
func (s *SecurityServiceGoFrServer) GetMarketDataJobs(ctx *gofr.Context) (any, error) {
	return &proto.GetMarketDataJobsResponse{}, nil
}
func (s *SecurityServiceGoFrServer) UpdateMarketDataJob(ctx *gofr.Context) (any, error) {
	return &proto.UpdateMarketDataJobResponse{}, nil
}
