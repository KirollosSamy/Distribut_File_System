package main

// pb "distributed_file_system/grpc/master_node"

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
}

NodesLookupTable := make(map[uint32]NodesData)





func main() {

}
