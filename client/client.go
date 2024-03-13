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
	for {
		// Set up a connection to the server.
		conn, err := grpc.Dial("localhost:2323", grpc.WithInsecure())
		if err != nil {
			fmt.Println("did not connect:", err)
			return
		}
		defer conn.Close()

		// Create a new client
		client := pb.NewClientToMasterServiceClient(conn)

		// Ask user if he wants to upload or download a file
		fmt.Println("Do you want to upload or download a file? (upload/download)")
		var action string
		fmt.Scanln(&action)

		if action == "upload" {
			fmt.Println("Enter the filename to upload:")
			var filename string
			fmt.Scanln(&filename)

			// Upload request to server
			resp, err := client.UploadFile(context.Background(), &pb.UploadRequest{Filename: filename})
			if err != nil {
				fmt.Println("UploadFile failed:", err)
				return
			}
			fmt.Println("UploadFile Response:", resp)

			// Open socket connection with the given port to upload the file to server
			err = streamMP4File(int(resp.Port), "../files/"+filename)
			if err != nil {
				fmt.Println("Error streaming file:", err)
				return
			}
			fmt.Println("File uploaded successfully")

		} else if action == "download" {
			fmt.Println("Enter the filename to download:")
			var filename string
			fmt.Scanln(&filename)

			// Download request to server
			resp, err := client.DownloadFile(context.Background(), &pb.DownloadRequest{Filename: filename})
			if err != nil {
				fmt.Println("DownloadFile failed:", err)
				return
			}
			fmt.Println("DownloadFile Response:", resp)

			// Download the file from the server

			err = downloadStream(resp.IPs, "../files/"+filename, resp.Filesize)
            if err != nil {
                fmt.Println("Error downloading file:", err)
                return
            }
            fmt.Println("File downloaded successfully")

            
		} else {
            fmt.Println("Invalid action")
        }

		// Ask user if he wants to continue
		fmt.Println("Do you want to upload/download another file? (yes/no)")
		var answer string
		fmt.Scanln(&answer)
		if answer != "yes" {
			break
		}
	}

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

// downloadChunk downloads a chunk of the file from the server on the specified port
func downloadChunk(ip string, startOffset, endOffset int64, filePath string, done chan<- error) {
	// Establish connection to the server
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		done <- err
		return
	}
	defer conn.Close()

	// Create or open the file to write
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		done <- err
		return
	}
	defer file.Close()

	// Set the offset for writing to the file
	if _, err = file.Seek(startOffset, 0); err != nil {
		done <- err
		return
	}

	// Read from connection and write to file
	_, err = io.CopyN(file, conn, endOffset-startOffset)
	if err != nil && err != io.EOF {
		done <- err
		return
	}

	done <- nil // Signal success
}

// downloadFile downloads a file from the server by splitting it into chunks and downloading concurrently from multiple ports
func downloadStream(IPs []string, filePath string, chunkSize int64) error {
	// Channel to communicate errors from goroutines
	done := make(chan error)

	// Calculate the number of chunks (number of IPs)
	numChunks := int64(len(IPs))

	// Calculate the chunk size
	chunkSize = (chunkSize + numChunks - 1) / numChunks // Round up

	// Start a goroutine for each chunk
	for i, ip := range IPs {
		startOffset := int64(i) * chunkSize
		endOffset := startOffset + chunkSize
		go downloadChunk(ip, startOffset, endOffset, filePath, done)
	}

	// Wait for all goroutines to finish
	for range IPs {
		if err := <-done; err != nil {
			return err // Return the first error encountered
		}
	}

	return nil // No errors occurred
}
