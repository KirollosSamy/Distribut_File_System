package main

import (
	"context"
	pb "distributed_file_system/grpc/master_node"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"distributed_file_system/utils"
)

var NodeId uint32
var master pb.NodeToMasterClient

type masterToNodeServer struct {
	pb.UnimplementedMasterToNodeServer
}

func (s *masterToNodeServer) RegisterNode(ctx context.Context, req *pb.RegisterNodeRequest) (*emptypb.Empty, error) {
	NodeId = req.NodeId
	return &emptypb.Empty{}, nil
}

func (s *masterToNodeServer) Replicate(ctx context.Context, req *pb.ReplicateRequest) (*emptypb.Empty, error) {
	address := fmt.Sprintf("%s:%d", req.OtherNodeIp, req.OtherNodePort)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Println("failed to connect:", err)
		return &emptypb.Empty{}, nil
	}
	defer conn.Close()

	utils.UploadFile(conn, req.Filename)

	return &emptypb.Empty{}, nil
}

func pingMaster() {
	stream, err := master.KeepMeAlive(context.Background())
	if err != nil {
		log.Fatalf("can't ping master %v", err)
	}
	defer stream.CloseAndRecv()

	for {
		stream.Send(&pb.HeartBeat{NodeId: NodeId})
		time.Sleep(time.Second)
	}
}

func receiveFile(conn net.Conn) {
	fileName, fileSize, err := utils.DownloadFile(conn)

	if err != nil {
		log.Println("Error downloading file from client: ", err)
		return
	}

	master.ConfirmUpload(context.Background(), &pb.FileUploadStatus{
		FileName: fileName,
		FilePath: "files/" + fileName,
		FileSize: fileSize,
	})
}

func runUploadServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("NODE_UPLOAD_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("failed to listen:", err)
		} else {
			go receiveFile(conn)
		}
	}
}

func runDownloadServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("NODE_DOWNLOAD_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("failed to listen:", err)
		} else {
			go utils.UploadChunk(conn)
		}
	}
}

func runGrpcServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("NODE_GRPC_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMasterToNodeServer(grpcServer, &masterToNodeServer{})
	grpcServer.Serve(lis)
}

func connectMaster() pb.NodeToMasterClient {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", os.Getenv("MASTER_HOST"), os.Getenv("MASTER_PORT")))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	return pb.NewNodeToMasterClient(conn)
}

func main() {
	master = connectMaster()

	go pingMaster()

	go runUploadServer()

	go runDownloadServer()

	runGrpcServer()
}
