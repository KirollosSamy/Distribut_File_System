package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func UploadFile(conn net.Conn, filename string, dir string) {
	defer conn.Close()

	file, err := os.Open(dir + "/" + filename)
	if err != nil {
		log.Println("Failed to open file:", err)
		return
	}
	defer file.Close()

	_, err = conn.Write([]byte(filename + "\n"))
	if err != nil {
		fmt.Println("Error sending filename:", err)
		return
	}

	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Error sending file data:", err)
	}
}

func UploadChunk(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	filename, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading filename:", err)
		return
	}
	filename = filename[:len(filename)-1]

	offset, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading offset:", err)
		return
	}
	parts := strings.Split(strings.TrimSpace(offset), ":")
    startOffset, _ := strconv.ParseInt(parts[0], 10, 64)
    endOffset, _ := strconv.ParseInt(parts[1], 10, 64)

	println(startOffset)
	println(endOffset)

	file, err := os.Open("files_1/"+filename)
	if err != nil {
		log.Println("Failed to open file:", err)
		return
	}
	defer file.Close()

	_, err = file.Seek(startOffset, 0)
	if err != nil {
		log.Println("Failed to seek in file:", err)
		return
	}

	chunkReader := io.NewSectionReader(file, startOffset, endOffset-startOffset)
	_, err = io.Copy(conn, chunkReader)
	if err != nil {
		log.Println("Error sending file data:", err)
	}

	println("chunk sent")
}

func DownloadFile(conn net.Conn, dir string) (string, int64, error) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	fileName, err := reader.ReadString('\n')
	if err != nil {
		return "", 0, err
	}
	fileName = fileName[:len(fileName)-1]

	os.Mkdir(dir, 0777)
	file, err := os.Create(dir + "/" + fileName)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	fileSize, err := io.Copy(file, conn)
	if err != nil {
		return "", 0, err
	}

	return fileName, fileSize, nil
}

func ParseConfig(filename string, config interface{}) {
	configFile, err := os.Open(filename)
    if err != nil {
        log.Fatal("Error opening config file:", err)
    }
    defer configFile.Close()

    decoder := json.NewDecoder(configFile)
    if err := decoder.Decode(&config); err != nil {
        log.Fatal("Error decoding config JSON:", err)
    }
}

func GetMyIp() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatalf("Error getting IP: %v", err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP
}