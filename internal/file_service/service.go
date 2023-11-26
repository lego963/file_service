package file_service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"file_service/api/file_service"
)

const fileStashFolder = "file_stash"

type FileService struct {
	file_service.UnsafeFileServiceServer
}

func NewFileService() *FileService {
	return &FileService{}
}

func (s *FileService) SaveFile(_ context.Context, req *file_service.FileRequest) (*file_service.FileResponse, error) {
	filename := req.Filename
	data := req.Data

	err := os.WriteFile(fmt.Sprintf("%s/%s", fileStashFolder, filename), data, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	filesCounter.WithLabelValues().Inc()

	return &file_service.FileResponse{}, nil
}

func (s *FileService) GetFileList(_ context.Context, _ *file_service.Empty) (*file_service.FileListResponse, error) {
	files, err := os.ReadDir(fmt.Sprintf("%s", fileStashFolder))
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	var fileList []*file_service.FileInfo
	for _, file := range files {
		fileInfo, _ := file.Info()
		fileList = append(fileList, &file_service.FileInfo{
			Filename:  file.Name(),
			CreatedAt: fileInfo.ModTime().Format(time.RFC3339),
			UpdatedAt: fileInfo.ModTime().Format(time.RFC3339),
		})
	}

	requestsCounter.WithLabelValues("GetFileList").Inc()

	return &file_service.FileListResponse{Files: fileList}, nil
}

func (s *FileService) GetFile(req *file_service.FileRequest, stream file_service.FileService_GetFileServer) error {
	filename := req.Filename

	data, err := os.ReadFile(fmt.Sprintf("%s/%s", fileStashFolder, filename))
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	err = stream.Send(&file_service.FileResponse{Data: data})
	if err != nil {
		return fmt.Errorf("failed to send file: %v", err)
	}

	requestsCounter.WithLabelValues("GetFile").Inc()

	return nil
}

func (s *FileService) RemoveOldFiles(directory string, maxAge time.Duration) error {
	files, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, file := range files {
		fileInfo, _ := file.Info()
		if time.Since(fileInfo.ModTime()) > maxAge {
			filePath := filepath.Join(directory, file.Name())
			if err := os.Remove(filePath); err != nil {
				return err
			}
		}
	}

	return nil
}
