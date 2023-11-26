package grpcserver

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/semaphore"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	api "file_service/api/file_service"
	service "file_service/internal/file_service"
	"google.golang.org/grpc"
)

const maxConcurrentConnections = 10
const cleanupInterval = 5 * time.Second
const fileStashFolder = "file_stash"

type GRPCServer struct {
	server      *grpc.Server
	sem         *semaphore.Weighted // Используется для ограничения количества одновременных подключений
	activeConns int
	mu          sync.Mutex
	fileService *service.FileService
}

func NewGRPCServer() *GRPCServer {
	return &GRPCServer{
		server:      grpc.NewServer(),
		sem:         semaphore.NewWeighted(int64(maxConcurrentConnections)),
		fileService: service.NewFileService(),
	}
}

func (s *GRPCServer) RegisterFileService(fileService *service.FileService) {
	api.RegisterFileServiceServer(s.server, fileService)
}

func (s *GRPCServer) Run() error {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":8080", nil)
		log.Fatal(err)
	}()

	go s.runCleanup()

	log.Println("gRPC server is running on :50051")
	return s.server.Serve(limitListener{listener, s.sem})
}

// limitListener реализует net.Listener и используется для ограничения количества одновременных подключений.
type limitListener struct {
	net.Listener
	sem *semaphore.Weighted
}

// Accept реализует метод Accept интерфейса net.Listener.
func (l limitListener) Accept() (net.Conn, error) {
	if err := l.sem.Acquire(context.Background(), 1); err != nil {
		return nil, err
	}
	conn, err := l.Listener.Accept()
	if err != nil {
		l.sem.Release(1)
		return nil, err
	}
	return &limitedConn{conn, l.sem}, nil
}

// limitedConn реализует net.Conn и используется для освобождения семафора при закрытии подключения.
type limitedConn struct {
	net.Conn
	sem *semaphore.Weighted
}

// Close реализует метод Close интерфейса net.Conn.
func (lc *limitedConn) Close() error {
	err := lc.Conn.Close()
	lc.sem.Release(1)
	return err
}

// runCleanup запускает механизм удаления старых файлов с указанной периодичностью.
func (s *GRPCServer) runCleanup() {
	for {
		time.Sleep(cleanupInterval)
		if err := s.fileService.RemoveOldFiles(fmt.Sprintf("%s/", fileStashFolder), 10*time.Second); err != nil {
			log.Printf("Failed to remove old files: %v", err)
		}
	}
}
