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
	//timer *time.Timer
	heartBeat int
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
	//timer := nodesLookupTable.Get(nodeId).timer
	//timer.Reset(time.Second)
	node = nodesLookupTable.Get(nodeId)
	node.heartBeat+=1

	return &emptypb.Empty{}, nil 
}

func(s *masterServer) ConfirmUpload(ctx context.Context, req *masterPb.FileUploadStatus) (*emptypb.Empty, error) {
	
	return &emptypb.Empty{}, nil 
}

func(s *masterServer) RegisterNode(ctx context.Context, req *masterPb.RegisterRequest) (*masterPb.RegisterResponse, error) {
	//Here we should add the node to nodes lookup table
	//Add the new node id
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

func getTwoRandomNodes(existingDataNode int32)int32{
	nodeIds := make([]int32,2)
	nodeIds[0] := int32(rand.Intn(lastNodeId))
	for(nodeIds[0]==existingDataNode ||!nodesLookupTable.Get(nodeIds[0]) )
		nodeIds[0] = int32(rand.Intn(lastNodeId))

	nodeIds[1] :=  int32(rand.Intn(lastNodeId))
	for(nodeIds[1]==existingDataNode || nodeIds[1]==nodeIds[0]||!nodesLookupTable.Get(nodeIds[0]) )
		nodeIds[1] = int32(rand.Intn(lastNodeId))
	return nodeIds
}

func getRandomNode(existingDateNode int32)int32{
	nodeId := int32(rand.Intn(lastNodeId))
	for(nodeId==existingDataNode ||!nodesLookupTable.Get(nodeId) )
		nodeId = int32(rand.Intn(lastNodeId))
	return nodeId
}

func getNumOfAliveNodes()int{
	count:=0
	for_,nodeId := range lastNodeId{
		if(nodesLookupTable.Get(nodeId).isAlive)
			count++
	}
	return count
}

func selectNodeToReplicate(fileName string, dataNodeId int32){
	aliveCnt := getNumOfAliveNodes()

	if aliveCnt == 1{
		fmt.Println("There is only one alive data node. Cannot replicate the file.")
		return
	}else if aliveCnt == 2{
		fmt.Println("There are only two alive data nodes. Replicating the file to the other node.")
		nodeId := getRandomNode(dataNodeId)
		//Here we should call the data keeper to replicate the file with that one only
	}else{
		nodeIds := getTwoRandomNodes(dataNodeId)
		//Here we should call the data keeper to replicate the file for the two nodes
	}
}

func checkIfAlive(){
	for{
		time.sleep(3*time.Second)
		for i := range lastNodeId{
			node := nodesLookupTable.Get(i)
			if(node.heartBeat == 0){
				node.isAlive = false
			} else{
				node.isAlive = true
			}
			node.heartBeat = 0
		}
	}
}

func(s *masterServer) RequestToUpload(ctx context.Context, req *masterPb.UploadRequest) (*masterPb.HostAddress, error) {
	//Here we should look for all the nodes available that has that file
	//Firstly as we initialize nodeIds from 0 and increment it, then we want to generate a random number from 0 to lastIdx to select a randomly datakeeper
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(lastNodeId) // Generates a random number between 0 and last node id inclusive
	//Keep looping until we found a node that works
	while(nodesLookupTable[randomNumber].isAlive==false){
		randomNumber = rand.Intn(lastNodeId)
	}
	//Get the upload port of that node
    uploadPort = nodesLookupTable[randomNumber].uploadPort
	//Reply to the client with node's ip and host
	return &masterPb.HostAddress{ip:nodesLookupTable[randomNumber].ip,port:nodesLookupTable[randomNumber].}, nil
}

func(s *masterServer) RequestToDonwload(ctx context.Context, req *masterPb.DownloadRequest) (*masterPb.DownloadResponse, error) {
	fileName := req.filename
	fileData := fileLookupTable[fileName]
	var addresses[]string;
	for(int i=0; i<len(dataKeepers); i++) {
		if(nodesLookupTable[fileData[i].dataKeeperId].isAlive ==true){
			addresses = append(addresses, fmt.Sprintf("%s:%s",nodesLookupTable[fileData[i].dataKeeperId].ip,nodesLookupTable[fileData[i].dataKeeper].downloadPort))
		}
	}
	return &masterPb.DownloadResponse{nodes_addresses:addresses}, nil
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


	runGrpcServer()
	go checkIfAlive()
}
