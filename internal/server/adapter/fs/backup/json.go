package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

const (
	tempFileTemplate = "metrics_backup_*.json.tmp"
)

func NewJSON(fileStoragePath string) *jsonFormatter {
	jsonFormatter := &jsonFormatter{
		fileStoragePath: fileStoragePath,
	}
	return jsonFormatter
}

type jsonFormatter struct {
	fileStoragePath string
}

func (jf *jsonFormatter) Write(ctx context.Context, metricses []domain.Metrics) error {
	logger := domain.GetMainLogger()
	if jf.fileStoragePath == "" {
		logger.Errorw("Write", "status", "error", "msg", "fileStoragePath is not specified")
		return os.ErrNotExist
	}

	file, err := os.OpenFile(jf.fileStoragePath, os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		return err
	}

	defer file.Close()

	// Пишем во временный файл
	tmpDir := os.TempDir()
	file, err = os.CreateTemp(tmpDir, tempFileTemplate)

	if err != nil {
		logger.Infow("Write", "status", "ok", "error", "can't create temp file")
		return err
	}

	defer os.Rename(file.Name(), jf.fileStoragePath) // Просто переименовываем временный файл

	err = json.NewEncoder(file).Encode(metricses)
	if err != nil {
		logger.Errorw("Write", "status", "error", "msg", err.Error())
		return err
	}

	logger.Infow("Write", "status", "ok", "msg", fmt.Sprintf("metrics stored into file %v", jf.fileStoragePath))
	return nil
}
func (jf *jsonFormatter) Read(ctx context.Context) ([]domain.Metrics, error) {
	logger := domain.GetMainLogger()
	var result []domain.Metrics

	if jf.fileStoragePath == "" {
		logger.Errorw("Read", "status", "error", "msg", "fileStoragePath is not specified")
		return result, os.ErrNotExist
	}

	file, err := os.Open(jf.fileStoragePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Infow("Read", "status", "ok", "msg", fmt.Sprintf("dump file %v not exists", jf.fileStoragePath))
			return result, os.ErrNotExist
		}
		return nil, err
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&result)
	if err != nil {
		logger.Infow("Read", "status", "error", "msg", fmt.Sprintf("can't restore backup from file %v", jf.fileStoragePath))
		return nil, err
	}

	return result, nil
}
