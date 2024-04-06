package main

import (
	"bufio"
	"context"
	client "distributed_file_system/grpc/client"
	master "distributed_file_system/grpc/master"
	"fmt"
	"io"
	"net"
	"os"

	"google.golang.org/grpc"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:8000", grpc.WithInsecure())
	if err != nil {
		fmt.Println("did not connect:", err)
		return
	}
	defer conn.Close()

	// Create a new client
	clientMaster := client.NewClientClient(conn)

	// Create a new master client
	masterClient := master.NewMasterClient(conn)

	for {
		// Ask user if he wants to upload or download a file
		fmt.Println("Do you want to upload or download a file? (upload/download)")
		var action string
		fmt.Scanln(&action)

		if action == "upload" {
			fmt.Println("Enter the filename to upload:")
			var filename string
			fmt.Scanln(&filename)

			// send upload request to server
			resp, err := masterClient.RequestToUpload(context.Background(), &master.UploadRequest{Filename: filename})
			if err != nil {
				fmt.Println("UploadFile failed:", err)
				break
			}
			fmt.Println("UploadFile Response:", resp)

			// Open socket connection with the given IP to upload the file to server in a new goroutine
			go func() {
				err := streamMP4File(resp.Ip + ":" + fmt.Sprint(resp.Port), filename)
				if err != nil {
					fmt.Println("Error streaming file:", err)
					return
				}

				// Send upload success message to server
				clientMaster.UploadSuccess(context.Background(),&client.Success{Success: true})
				fmt.Println("File uploaded successfully")
			}()

		} else if action == "download" {
			fmt.Println("Enter the filename to download:")
			var filename string
			fmt.Scanln(&filename)

			// send download request to server
			resp, err := masterClient.RequestToDonwload(context.Background(), &master.DownloadRequest{Filename: filename})
			if err != nil {
				fmt.Println("DownloadFile failed:", err)
				break
			}
			fmt.Println("DownloadFile Response:", resp)

			// Download the file from the server in a new goroutine
			go func() {

				// convert resp.NodesAddresses to []string
				nodesIps := make([]string, len(resp.NodesAddresses))
				for i := 0; i < len(resp.NodesAddresses); i++ {
					nodesIps[i] = resp.NodesAddresses[i].Ip + ":" + fmt.Sprint(resp.NodesAddresses[i].Port)
				}

				// Download the file
				err = downloadStream(nodesIps, "../files/" + filename + ".mp4", resp.Filesize)
				if err != nil {
					fmt.Println("Error downloading file:", err)
					return
				}
				fmt.Println("File downloaded successfully")
			}()

            
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

// streamMP4File streams an MP4 file to the server on the specified port
func streamMP4File(ip string, filename string) error {
	println(ip)
	// Open the MP4 file
	file, err := os.Open("files/" + filename + ".mp4")
	if err != nil {
		return err
	}
	// defer file.Close()

	// Start listening on the specified port
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		return err
	}
	// defer conn.Close()

	// Send the file name to the server
	_, err = conn.Write([]byte(filename + ".mp4" + "\n"))
	if err != nil {
		return err
	}

	// Start streaming the file data to the connection
	go func(conn net.Conn) {
		defer conn.Close()
		defer file.Close()

		reader := bufio.NewReader(file)
		_, err = reader.WriteTo(conn)
		if err != nil {
			fmt.Println("Error uploading file:", err)
			return
		}

		// buffer := make([]byte, 1024)
		// for {
		// 	bytesRead, err := file.Read(buffer)
		// 	if err != nil {
		// 		// End of file
		// 		if err == io.EOF {
		// 			break
		// 		}
		// 		fmt.Println("Error reading from file:", err)
		// 		return
		// 	}
		// 	_, err = conn.Write(buffer[:bytesRead])
		// 	if err != nil {
		// 		fmt.Println("Error writing to connection:", err)
		// 		return
		// 	}
		// }
	}(conn)

	return nil
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

	// send file name to server
	_, err = conn.Write([]byte(filePath))
	if err != nil {
		done <- err
		return
	}

	// send start and end offset to server
	_, err = conn.Write([]byte(fmt.Sprintf("%d,%d", startOffset, endOffset)))
	if err != nil {
		done <- err
		return
	}
	
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

// downloadStream downloads a file from the server by splitting it into chunks and downloading concurrently from multiple ports
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
