package main

import (
	"context"
	masterPb "distributed_file_system/grpc/master"
	nodePb "distributed_file_system/grpc/node"
	"fmt"
	"log"
	"net"
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
	NodeUploadPort uint32 `json:"NODE_UPLOAD_PORT"`
	NodeDownloadPort uint32 `json:"NODE_DOWNLOAD_PORT"`
	NodeGrpcPort uint32 `json:"NODE_GRPC_PORT"`
	MasterHost string `json:"MASTER_HOST"`
	MasterPort uint32 `json:"MASTER_PORT"`
	Directory string `json:"DIRECTORY"`
}

var config Config

func (s *nodeServer) Replicate(ctx context.Context, req *nodePb.ReplicateRequest) (*emptypb.Empty, error) {
	address := fmt.Sprintf("%s:%d", req.OtherNodeIp, req.OtherNodePort)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Println("failed to connect:", err)
		return &emptypb.Empty{}, nil
	}

	dir := fmt.Sprintf("%s_%d", config.Directory, NodeId)
	utils.UploadFile(conn, req.Filename, dir)

	return &emptypb.Empty{}, nil
}

func pingMaster() {
	heartBeat := &masterPb.HeartBeat{NodeId: NodeId}
	for {
		master.KeepMeAlive(context.Background(), heartBeat)
		time.Sleep(time.Second)
	}
}

func receiveFile(conn net.Conn) {
	dir := fmt.Sprintf("%s_%d", config.Directory, NodeId)

	fileName, fileSize, err := utils.DownloadFile(conn, dir)

	if err != nil {
		log.Println("Error downloading file from client: ", err)
		return
	}

	master.ConfirmUpload(context.Background(), &masterPb.FileUploadStatus{
		FileName: fileName,
		FilePath: fmt.Sprintf("%s_%d", config.Directory, NodeId) + fileName,
		FileSize: fileSize,
		NodeId: NodeId,
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

func connectMaster() *grpc.ClientConn {
	masterAddress := fmt.Sprintf("%s:%d", config.MasterHost, config.MasterPort)
	
	conn, err := grpc.Dial(masterAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	master = masterPb.NewMasterClient(conn)

	return conn
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
	utils.ParseConfig("config/node.json", &config)

	conn := connectMaster()
	defer conn.Close()

	registerNode()

	go pingMaster()
	go runUploadServer()
	go runDownloadServer()

	runGrpcServer()
}
