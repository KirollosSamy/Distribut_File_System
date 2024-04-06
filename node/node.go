package main

import (
	"context"
	masterPb "distributed_file_system/grpc/master"
	nodePb "distributed_file_system/grpc/node"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"distributed_file_system/utils"
)

var NodeId uint32
var master masterPb.MasterClient

type nodeServer struct {
	nodePb.UnimplementedNodeServer
}

type Config struct {
	NodeUploadPort uint32 `josn:"NODE_UPLOAD_PORT"`
	NodeDownloadPort uint32 `josn:"NODE_DOWNLOAD_PORT"`
	NodeGrpcPort uint32 `josn:"NODE_GRPC_PORT"`
	MasterHost string `json:"MASTER_HOST"`
	MasterPort uint32 `json:"MASTER_PORT"`
}

var config Config

func (s *nodeServer) Replicate(ctx context.Context, req *nodePb.ReplicateRequest) (*emptypb.Empty, error) {
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
		stream.Send(&masterPb.HeartBeat{NodeId: NodeId})
		time.Sleep(time.Second)
	}
}

func receiveFile(conn net.Conn) {
	fileName, fileSize, err := utils.DownloadFile(conn)

	if err != nil {
		log.Println("Error downloading file from client: ", err)
		return
	}

	master.ConfirmUpload(context.Background(), &masterPb.FileUploadStatus{
		FileName: fileName,
		FilePath: "files/" + fileName,
		FileSize: fileSize,
	})
}

func runUploadServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.NodeUploadPort))
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.NodeDownloadPort))
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.NodeGrpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	nodePb.RegisterNodeServer(grpcServer, &nodeServer{})
	grpcServer.Serve(lis)
}

func connectMaster() masterPb.MasterClient {
	masterAddress := fmt.Sprintf("%s:%d", config.MasterHost, config.MasterPort)
	
	conn, err := grpc.Dial(masterAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	registerNode()

	return masterPb.NewMasterClient(conn)
}

func registerNode() {
	response, err := master.RegisterNode(context.Background(), &masterPb.RegisterRequest{
		Ip: utils.GetMyIp().String(),
		GrpcPort: config.NodeGrpcPort,
		UploadPort: config.NodeUploadPort,
		DownloadPort: config.NodeDownloadPort,
	})
	if err != nil {
		log.Fatalf("can't register node %v", err)
	}

	NodeId = response.NodeId
}

func main() {
	utils.ParseConfig(os.Args[1], &config)

	master = connectMaster()

	go pingMaster()
	go runUploadServer()
	go runDownloadServer()

	runGrpcServer()
}
