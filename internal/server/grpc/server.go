package grpc

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pvz_service/internal/app/pvz"
	pvzV1 "pvz_service/protos/gen/go"
)

type GRPCServer struct {
	pvzV1.UnimplementedPVZServiceServer
	pvzService pvz.ServiceInterface
}

func NewGRPCServer(pvzService pvz.ServiceInterface) *GRPCServer {
	return &GRPCServer{
		pvzService: pvzService,
	}
}

func (s *GRPCServer) GetPVZList(ctx context.Context, req *pvzV1.GetPVZListRequest) (*pvzV1.GetPVZListResponse, error) {
	pvzList, err := s.pvzService.ServiceGetPVZs(ctx, "", "", "", "")
	if err != nil {
		return nil, err
	}

	var result []*pvzV1.PVZ
	for _, item := range pvzList {
		regDate := timestamppb.New(item.PVZ.RegistrationDate)

		result = append(result, &pvzV1.PVZ{
			Id:               item.PVZ.ID.String(),
			City:             item.PVZ.City,
			RegistrationDate: regDate,
		})
	}

	return &pvzV1.GetPVZListResponse{
		Pvzs: result,
	}, nil
}
