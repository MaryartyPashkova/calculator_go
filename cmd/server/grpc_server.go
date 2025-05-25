package main

import (
	"context"

	"log"

	"net"

	"calculator/internal/pb"

	"calculator/internal/service"

	"google.golang.org/grpc"
)

type grpcServer struct {
	pb.UnimplementedCalculatorServiceServer
	calculator *service.CalculatorService
}

func (s *grpcServer) Calculate(ctx context.Context, req *pb.CalculateRequest) (*pb.CalculateResponse, error) {
	instructions := make([]service.Instruction, 0, len(req.Instructions))
	for _, instr := range req.Instructions {
		left := parseValue(instr)
		right := parseRight(instr)
		instructions = append(instructions, service.Instruction{
			Type:  instr.Type,
			Op:    instr.Op,
			Var:   instr.Var,
			Left:  left,
			Right: right,
		})
	}

	results, err := s.calculator.Run(context.Background(), instructions)
	if err != nil {
		log.Printf("Ошибка выполнения: %v", err)
		return nil, err
	}

	items := make([]*pb.ResultItem, 0, len(results))
	for _, item := range results {
		items = append(items, &pb.ResultItem{
			Var:   item.Var,
			Value: item.Value,
		})
	}

	return &pb.CalculateResponse{Items: items}, nil
}

func parseValue(instr *pb.Instruction) interface{} {
	switch v := instr.LeftType.(type) {
	case *pb.Instruction_LeftInt:
		return v.LeftInt
	case *pb.Instruction_LeftVar:
		return v.LeftVar
	}
	return nil
}

func parseRight(instr *pb.Instruction) interface{} {
	switch v := instr.RightType.(type) {
	case *pb.Instruction_RightInt:
		return v.RightInt
	case *pb.Instruction_RightVar:
		return v.RightVar
	}
	return nil
}

func startGRPCServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCalculatorServiceServer(s, &grpcServer{
		calculator: service.NewCalculatorService(),
	})
	log.Printf("gRPC сервер запущен на порту %d\n", 50051)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
