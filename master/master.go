package main

import (
	"context"
	// clientPb "distributed_file_system/grpc/client"
	masterPb "distributed_file_system/grpc/master"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"
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
    dataKeeperId uint32
	fileSize int64
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
	//Add File Data To Table -> Add data file entry
	var fileData []FileData
	
	fileData = append(fileData, FileData{
		filePath: req.FilePath,
		dataKeeperId: req.NodeId,
		fileSize: req.FileSize,
	})
	
	fileLookupTable.Set(req.FileName, fileData)

	// notify client of success

	return &emptypb.Empty{}, nil 
}

func(s *masterServer) RegisterNode(ctx context.Context, req *masterPb.RegisterRequest) (*masterPb.RegisterResponse, error) {
	//Here we should add the node to nodes lookup table
	//Add the new node id
	fmt.Println("inside RegisterNode")
	lastNodeId++
	nodeId := lastNodeId
	//start a timer that defines if the node will be killed or not
	timer := time.NewTimer(time.Second)
	//Create a new Node in which we will have all the needed data as tcp download and upload ports , and grpc port , and also its ip
	//Then we define the node's status if its alive or not then the timer which will be used to kill it
	node := Node{
		downloadPort: req.DownloadPort,
		uploadPort: req.UploadPort,
		grpcPort: req.GrpcPort,
		ip: req.Ip,
		isAlive: true,
		timer: timer,
	}
	//Add the new Node
	nodesLookupTable.Set(nodeId, node)
	//Wait For Timer
	go waitForTimer(timer, nodeId)
	//Return response to the node which is it's Id
	res := &masterPb.RegisterResponse{NodeId: nodeId}
	return res, nil
}

func(s *masterServer) RequestToUpload(ctx context.Context, req *masterPb.UploadRequest) (*masterPb.HostAddress, error) {
	//Here we should look for all the nodes available that has that file
	//Firstly as we initialize nodeIds from 0 and increment it, then we want to generate a random number from 0 to lastIdx to select a randomly datakeeper
	rand.Seed(time.Now().UnixNano())
	var node Node

	for {
		randomNumber := rand.Intn(int(lastNodeId))
		node := nodesLookupTable.Get(uint32(randomNumber))
		if(node.isAlive){
			break
		}
	}
	//Reply to the client with node's ip and host
	return &masterPb.HostAddress{
		Ip: node.ip,
		Port:node.uploadPort,
	}, nil
}

func(s *masterServer) RequestToDonwload(ctx context.Context, req *masterPb.DownloadRequest) (*masterPb.DownloadResponse, error) {
	fileName := req.Filename
	fileData := fileLookupTable.Get(fileName)
	var addresses[]*masterPb.HostAddress;
	
	for i:=0; i<len(fileData); i++ {
		nodeId:= fileData[i].dataKeeperId
		node := nodesLookupTable.Get((uint32(nodeId)))
		if(node.isAlive){
			addresses = append(addresses, &masterPb.HostAddress{Ip: node.ip, Port: node.downloadPort})
		}
	}
	return &masterPb.DownloadResponse{NodesAddresses: addresses, Filesize: fileData[0].fileSize}, nil
}

func waitForTimer(timer *time.Timer, nodeId uint32) {
	<-timer.C
	node := nodesLookupTable.Get(nodeId)
	node.isAlive = false
	nodesLookupTable.Set(nodeId, node)
}

func runGrpcServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	masterPb.RegisterMasterServer(grpcServer, &masterServer{})

	fmt.Printf("server running")
	grpcServer.Serve(lis)
}

// func connectClient() clientPb.ClientClient {
// 	clientAddress := fmt.Sprintf("%s:%d", "host", 1247578)
	
// 	conn, err := grpc.Dial(clientAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatalf("failed to connect: %v", err)
// 	}
// 	defer conn.Close()

// 	return clientPb.NewClientClient(conn)
// }

func main() {
	utils.ParseConfig("config/master.json", &config)
	println(config.MasterPort)
	runGrpcServer()
}
