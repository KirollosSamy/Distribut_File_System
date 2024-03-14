package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func UploadFile(conn net.Conn, filename string) {
	defer conn.Close()

	file, err := os.Open(filename)
	if err != nil {
		log.Println("Failed to open file:", err)
		return
	}
	defer file.Close()

	_, err = conn.Write([]byte(filename))
	if err != nil {
		fmt.Println("Error sending filename:", err)
		return
	}

	reader := bufio.NewReader(file)
	_, err = reader.WriteTo(conn)
	if err != nil {
		fmt.Println("Error sending file data:", err)
	}
}

func DownloadFile(conn net.Conn) string {
	defer conn.Close()

	// Read the filename from the client
	reader := bufio.NewReader(conn)
	filename, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading filename:", err)
		return ""
	}
	filename = filename[:len(filename)-1]

	file, err := os.Create("files/" + filename)
	if err != nil {
		log.Println("Failed to create file:", err)
		return ""
	}
	defer file.Close()

	_, err = io.Copy(file, conn)
	if err != nil {
		log.Println("Error copying file:", err)
		return ""
	}

	return filename
}
