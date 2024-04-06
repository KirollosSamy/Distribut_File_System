package main

import (
	"context"
	masterPb "distributed_file_system/grpc/master"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"distributed_file_system/utils"
)

type Config struct {
	MasterPort uint32 `josn:"MASTER_PORT"`
}
var config Config

//We need to define lookup tables
//first one for files
type FileData struct {
    filePath string
    dataKeeperId int
}

//Then define lookup table for nodes
type Node struct {
    uploadPort uint32
	downloadPort uint32
    grpcPort uint32
	ip string
	isAlive bool
	timer *time.Timer
	// lastTimeToBing string
}

// fileLookupTable := make(map[string][]FileData)
// NodesLookupTable := make(map[uint32]NodesData)
var fileLookupTable *utils.SafeMap[string, []FileData]
var nodesLookupTable *utils.SafeMap[uint32, Node]

var lastNodeId uint32 = 0

type masterServer struct {
	masterPb.UnimplementedMasterServer
}

func(s *masterServer) KeepMeAlive(ctx context.Context, req *masterPb.HeartBeat) (*emptypb.Empty, error) {
	// restart the timer of that node
	nodeId := req.NodeId
	timer := nodesLookupTable.Get(nodeId).timer
	timer.Reset(time.Second)

	return &emptypb.Empty{}, nil 
}

func(s *masterServer) ConfirmUpload(ctx context.Context, req *masterPb.FileUploadStatus) (*emptypb.Empty, error) {
	
	return &emptypb.Empty{}, nil 
}

func(s *masterServer) RegisterNode(ctx context.Context, req *masterPb.RegisterRequest) (*masterPb.RegisterResponse, error) {
	//Here we should add the node to nodes lookup table
	lastNodeId++
	nodeId := lastNodeId

	timer := time.NewTimer(time.Second)

	node := Node{
		downloadPort: req.DownloadPort,
		uploadPort: req.UploadPort,
		grpcPort: req.GrpcPort,
		ip: req.Ip,
		isAlive: true,
		timer: timer,
	}
	nodesLookupTable.Set(nodeId, node)

	go waitForTimer(timer, nodeId)

	res := &masterPb.RegisterResponse{NodeId: nodeId}
	return res, nil
}

func(s *masterServer) RequestToUpload(ctx context.Context, req *masterPb.UploadRequest) (*masterPb.HostAddress, error) {
	//Here we should look for all the nodes available that has that file

	return &masterPb.HostAddress{}, nil
}

func(s *masterServer) RequestToDonwload(ctx context.Context, req *masterPb.DownloadRequest) (*masterPb.DownloadResponse, error) {
	
	return &masterPb.DownloadResponse{}, nil
}

func waitForTimer(timer *time.Timer, nodeId uint32) {
	<-timer.C
	node := nodesLookupTable.Get(nodeId)
	node.isAlive = false
	nodesLookupTable.Set(nodeId, node)
}

func runGrpcServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.MasterPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	masterPb.RegisterMasterServer(grpcServer, &masterServer{})
	grpcServer.Serve(lis)
}


func main() {
	utils.ParseConfig("config/master.json", &config)

	fileLookupTable = utils.NewSafeMap[string, []FileData]()
	nodesLookupTable = utils.NewSafeMap[uint32, Node]()

	runGrpcServer()
}
