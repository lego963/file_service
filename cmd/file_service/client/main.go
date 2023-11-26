package main

import (
	"context"
	api "file_service/api/file_service"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
)

const (
	serverAddress = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Создаем клиент
	client := api.NewFileServiceClient(conn)

	// Вызываем метод SaveFile
	saveFileResponse, err := client.SaveFile(context.Background(), &api.FileRequest{
		Filename: "example.txt",
		Data:     []byte("Hello, gRPC!"),
	})
	if err != nil {
		log.Fatalf("SaveFile failed: %v", err)
	}
	fmt.Printf("SaveFile Response: %v\n", saveFileResponse)

	// Вызываем метод GetFileList
	fileListResponse, err := client.GetFileList(context.Background(), &api.Empty{})
	if err != nil {
		log.Fatalf("GetFileList failed: %v", err)
	}
	fmt.Println("File List:")
	for _, fileInfo := range fileListResponse.Files {
		fmt.Printf("Filename: %s, Created At: %s, Updated At: %s\n", fileInfo.Filename, fileInfo.CreatedAt, fileInfo.UpdatedAt)
	}

	// Вызываем метод GetFile
	getFileResponse, err := client.GetFile(context.Background(), &api.FileRequest{
		Filename: "example.txt",
	})
	if err != nil {
		log.Fatalf("GetFile failed: %v", err)
	}

	fileResponse, _ := getFileResponse.Recv()
	// Сохраняем полученные данные в файл
	err = os.WriteFile("downloaded_example.txt", fileResponse.Data, 0644)
	if err != nil {
		log.Fatalf("Failed to write downloaded file: %v", err)
	}

	fmt.Println("File downloaded successfully.")
}
