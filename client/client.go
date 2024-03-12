package main

import (
	"context"
	pb "distributed_file_system/grpc/master_client"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
)

func main() {

	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:2323", grpc.WithInsecure())
	if err != nil {
		fmt.Println("did not connect:", err)
		return
	}
	defer conn.Close()

	// Create a new client
	client := pb.NewClientToMasterServiceClient(conn)

	// Upload request to server
	resp, err := client.UploadFile(context.Background(), &pb.UploadRequest{Filename: "test.mp4"})
	if err != nil {
		fmt.Println("UploadFile failed:", err)
		return
	}
	fmt.Println("UploadFile Response:", resp)
	
	// Open socket connection with the given port to upload the file to server
	err = streamMP4File(int(resp.Port), "files/test.txt")
	if err != nil {
		fmt.Println("Error streaming file:", err)
		return
	}
	fmt.Println("File uploaded successfully")


	// TODO: Download file from server

}


func streamMP4File(port int, filePath string) error {
    // Open the MP4 file
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Start listening on the specified port
    listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
    if err != nil {
        return err
    }
    defer listener.Close()

    // Accept incoming connections
    for {
        conn, err := listener.Accept()
        if err != nil {
            return err
        }

        // Start streaming the file data to the connection
        go func(conn net.Conn) {
            defer conn.Close()
            buffer := make([]byte, 1024)
            for {
                bytesRead, err := file.Read(buffer)
                if err != nil {
                    // End of file
                    if err == io.EOF {
                        break
                    }
                    fmt.Println("Error reading from file:", err)
                    return
                }
                _, err = conn.Write(buffer[:bytesRead])
                if err != nil {
                    fmt.Println("Error writing to connection:", err)
                    return
                }
            }
        }(conn)
    }
}