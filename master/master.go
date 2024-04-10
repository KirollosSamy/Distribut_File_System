package main

import (
	"context"
	"math"
	"strconv"

	"golang.org/x/exp/slices"

	// clientPb "distributed_file_system/grpc/client"
	masterPb "distributed_file_system/grpc/master"
	nodePb "distributed_file_system/grpc/node"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"distributed_file_system/utils"
)

type Config struct {
	MasterPort uint32 `json:"MASTER_PORT"`
	ReplicationWaitTime time.Duration `json:"REPLICATION_WAIT_TIME"`
	MinReplicas int `json:"MIN_REPLICAS"`
}
var config Config

type FileData struct {
    filePath string
    dataKeeperId uint32
	fileSize int64
}

type Node struct {
    uploadPort uint32
	downloadPort uint32
    grpcPort uint32
	ip string
	isAlive bool
	//timer *time.Timer
	heartBeat int
}

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
	node := nodesLookupTable.Get(nodeId)
	node.heartBeat +=1
	nodesLookupTable.Set(nodeId, node)
	//fmt.Printf("Receiving heartbeat from node %d with value %d\n", nodeId,node.heartBeat)

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

	// TODO: notify client of success, handle multiple clients

	return &emptypb.Empty{}, nil 
}

func(s *masterServer) RegisterNode(ctx context.Context, req *masterPb.RegisterRequest) (*masterPb.RegisterResponse, error) {
	//Here we should add the node to nodes lookup table
	//Add the new node id
	lastNodeId++
	nodeId := lastNodeId
	//start a timer that defines if the node will be killed or not
	// timer := time.NewTimer(time.Second)
	//Create a new Node in which we will have all the needed data as tcp download and upload ports , and grpc port , and also its ip
	//Then we define the node's status if its alive or not then the timer which will be used to kill it
	node := Node{
		downloadPort: req.DownloadPort,
		uploadPort: req.UploadPort,
		grpcPort: req.GrpcPort,
		ip: req.Ip,
		isAlive: true,
		heartBeat: 0,
		// timer: timer,
	}
	//Add the new Node
	nodesLookupTable.Set(nodeId, node)

	newNode := nodesLookupTable.Get(nodeId)
	println("A new Node Registered to the master with downloadPort",newNode.downloadPort," ,uploadPort ",newNode.uploadPort," grpcPort ",newNode.grpcPort," ip",newNode.ip)
	//Wait For Timer
	// go waitForTimer(timer, nodeId)
	//Return response to the node which is it's Id
	res := &masterPb.RegisterResponse{NodeId: nodeId}
	return res, nil
}

func selectNodes(n int, excluded []uint32) []uint32{
	var availableNodes []uint32
	for nodeId, node := range nodesLookupTable.GetMap() {
		if node.isAlive && !slices.Contains(excluded, nodeId) {
			availableNodes = append(availableNodes, nodeId)
		}
	}

	numNodesToSelect := int(math.Min(float64(n), float64(len(availableNodes))))

	var selectedNodes []uint32
	for len(selectedNodes) < numNodesToSelect {
		randId := uint32(rand.Intn(len(availableNodes)))
		if !slices.Contains(selectedNodes, availableNodes[randId]) {
			selectedNodes = append(selectedNodes, availableNodes[randId])
		}
	}

	return selectedNodes
}

func checkIfAlive(){
	for{
		time.Sleep(50*time.Second)
		for i := 1; i <= int(lastNodeId); i++{
			node := nodesLookupTable.Get(uint32(i))
			if(node.heartBeat == 0){
				node.isAlive = false
				fmt.Printf("Node "+ strconv.Itoa(i) + " is Dead\n")
			} else{
				node.isAlive = true
				fmt.Printf("Node "+ strconv.Itoa(i) + " is Alive\n")
			}
			node.heartBeat = 0
			nodesLookupTable.Set(uint32(i), node)
		}
		fmt.Printf("----------------------------------------------------------------\n")
	}
}

func(s *masterServer) RequestToUpload(ctx context.Context, req *masterPb.UploadRequest) (*masterPb.HostAddress, error) {
	//Here we should look for all the nodes available that has that file
	//Firstly as we initialize nodeIds from 0 and increment it, then we want to generate a random number from 0 to lastIdx to select a randomly datakeeper
	rand.Seed(time.Now().UnixNano())
	var node Node

	for {
		randomNumber := rand.Intn(int(lastNodeId)) + 1
		node = nodesLookupTable.Get(uint32(randomNumber))
		if(node.isAlive){
			break
		}
	}
	println(node.uploadPort , node.downloadPort , node.grpcPort)
	//Reply to the client with node's ip and host
	return &masterPb.HostAddress{
		Ip: node.ip,
		Port:node.uploadPort,
	}, nil
}

func(s *masterServer) RequestToDonwload(ctx context.Context, req *masterPb.DownloadRequest) (*masterPb.DownloadResponse, error) {
	fileName := req.Filename
	fileData := fileLookupTable.Get(fileName)
	println("fileData: ", fileData)
	var addresses[]*masterPb.HostAddress;
	
	for i:=0; i<len(fileData); i++ {
		nodeId:= fileData[i].dataKeeperId
		node := nodesLookupTable.Get((uint32(nodeId)))
		if(node.isAlive){
			currentAddress := &masterPb.HostAddress{Ip: node.ip, Port: node.downloadPort}
			println(currentAddress)
			addresses = append(addresses,currentAddress)
		}
	}
	return &masterPb.DownloadResponse{NodesAddresses: addresses, Filesize: fileData[0].fileSize}, nil
}

// func waitForTimer(timer *time.Timer, nodeId uint32) {
// 	<-timer.C
// 	node := nodesLookupTable.Get(nodeId)
// 	node.isAlive = false
// 	nodesLookupTable.Set(nodeId, node)
// }

func runGrpcServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", config.MasterPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	masterPb.RegisterMasterServer(grpcServer, &masterServer{})

	fmt.Printf("server running")
	grpcServer.Serve(lis)
}

func handleReplication(){
	for{
		time.Sleep(config.ReplicationWaitTime * time.Second)


		for fileName, fileReplicas := range fileLookupTable.GetMap() {
			var activeReplicas []uint32
			fmt.Printf("length of fileReplicas in file %s is: %d\n", fileName, len(fileReplicas))
			for _, replica := range fileReplicas{
				if(nodesLookupTable.Get(replica.dataKeeperId).isAlive){
					activeReplicas = append(activeReplicas, replica.dataKeeperId)
					fmt.Println("File is Stored on node "+strconv.Itoa((int(replica.dataKeeperId))))
				}
				fmt.Println("-----------------------------------------")
			}

			numRequiredReplicas := config.MinReplicas - len(activeReplicas)

			if(numRequiredReplicas > 0){
				destinations := selectNodes(numRequiredReplicas, activeReplicas)
				if len(destinations) > 0 {
					fmt.Printf("Replicating file %s to %d nodes with ids: \n", fileName, len(destinations))
					for _, id := range destinations{
						println(id)
					}
				replicate(fileName, activeReplicas[0], destinations)
				//Now I will Loop over the lookup table of files to check it's updated successfully
				for _,fileDate := range fileLookupTable.Get(fileName){
					
					fmt.Println("Ana gwa l handle replicate function")
					fmt.Printf("file %s is in %d\n", fileName, fileDate.dataKeeperId)
				}
				
				fmt.Println("-----------------------------------------")
				}else {
					fmt.Printf("No enough nodes to replicate file %s", fileName)
				}
			}
		}
	}
}

func replicate(fileName string, source uint32, destinations []uint32) {	
	sourceNode := nodesLookupTable.Get(source)
	sourceAddress := fmt.Sprintf("%s:%d", sourceNode.ip, sourceNode.grpcPort)

	conn, err := grpc.Dial(sourceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to node: %v", err)
	}
	defer conn.Close()

	nodeClient := nodePb.NewNodeClient(conn)

	fileReplicas := fileLookupTable.Get(fileName)

	for _, destination := range destinations {
		destinationNode := nodesLookupTable.Get(destination)

		_, err = nodeClient.Replicate(context.Background(), &nodePb.ReplicateRequest{
			Filename: fileName,
			OtherNodeIp: destinationNode.ip,
			OtherNodePort: destinationNode.uploadPort,
		})
		if err != nil {
			log.Printf("Failed to replicate file %s to node %d: %v", fileName, destination, err)
		} else{
			fileReplicas = append(fileReplicas, FileData{
				dataKeeperId: destination,
				fileSize: fileReplicas[0].fileSize,
			})

			for _,fileDate := range fileReplicas{
				fmt.Println("Ana gwa l replicate function")
				fmt.Printf("file %s is in %d\n", fileName, fileDate.dataKeeperId)
			}
			
			fmt.Println("-----------------------------------------")
		}
	}

	for _,fileDate := range fileReplicas{
		fmt.Println("Ana khlst l replicate function")
		fmt.Printf("file %s is in %d\n", fileName, fileDate.dataKeeperId)
	}
	
	fmt.Println("-----------------------------------------")

	fileLookupTable.Set(fileName, fileReplicas)
}

func main() {
	utils.ParseConfig("config/master.json", &config)

	fileLookupTable = utils.NewSafeMap[string, []FileData]()
	nodesLookupTable = utils.NewSafeMap[uint32, Node]()

	go checkIfAlive()
	go handleReplication()

	runGrpcServer()
}
