package main

import (
	"context"
	masterPb "distributed_file_system/grpc/master"
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


//We need to define lookup tables
//first one for files
type FileData struct {
    filePath string
    dataKeeperId int
}

fileLookupTable := make(map[string][]FileData)

//Then define lookup table for nodes
type NodesData struct {
    tcpUploadPort string
	tcpDownloadPort string
    grpcPort string
	ip string
	isAlive bool
	lastTimeToBing string
}

NodesLookupTable := make(map[uint32]NodesData)


type MasterService struct {
	masterPb.UnimplementedMasterServer
}

func(s *MasterService) KeepMeAlive(ctx context.Context, req *masterPb.HeartBeat) (*emptypb.Empty, error) {
	//Here we should restart the timer of that node
	//in another words we should set the lastTimeToBing to now
}

func(s *MasterService) ConfirmUpload(ctx context.Context, req *masterPb.FileUploadStatus) (*emptypb.Empty, error) {
	
}

func(s *MasterService) RegisterNode(ctx context.Context, req *masterPb.RegisterRequest) (masterPb.RegisterResponse, error) {
	//Here we should add the node to nodes lookup table
}

func(s *MasterService) RequestToUpload(ctx context.Context, req *masterPb.UploadRequest) (masterPb.HostAddress, error) {
	//Here we should look for all the nodes available that has that file
}

func(s *MasterService) RequestToDonwload(ctx context.Context, req *masterPb.DownloadRequest) (masterPb.DownloadResponse, error) {
	
}



func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("failed to listen:", err)
		return
	}

	s := grpc.NewServer()

	pb.RegisterTextServiceServer(s, &MasterService{})

	fmt.Println("Server started. Listening on port 8080...")
	if err := s.Serve(lis); err != nil {
		fmt.Println("failed to serve:", err)
	}
}
