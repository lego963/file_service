package main

import (
	"file_service/internal/file_service"
	"file_service/pkg/grpcserver"
	"log"
)

func main() {
	srv := grpcserver.NewGRPCServer()
	fileService := file_service.NewFileService()
	srv.RegisterFileService(fileService)

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
