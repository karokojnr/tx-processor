package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"
	"tx-processor/config"
	"tx-processor/db"
	"tx-processor/processor"
	"tx-processor/repository"
)

const (
	DefaultWorkers       = 10
	DefaultBatchSize     = 500
	DefaultChannelBuffer = 10000
)

func main() {
	filePath := flag.String("file", "", "Path to the JSON file (required)")
	workerCount := flag.Int("workers", DefaultWorkers, "Number of concurrent workers")
	batchSize := flag.Int("batch", DefaultBatchSize, "Batch size for processing")
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Usage: processor -file=data.json -workers=10 -batch=500")
		os.Exit(1)
	}

	if err := processFile(*filePath, *workerCount, *batchSize); err != nil {
		log.Fatal(err)
	}
}

func processFile(filePath string, workerCount int, batchSize int) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Starting transaction processor",
		"file", filePath,
		"workers", workerCount,
		"batch_size", batchSize)

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	dbConn, err := db.NewPostgresDB(&cfg.DatabaseConfig)
	if err != nil {
		return fmt.Errorf("db connect: %w", err)
	}
	defer dbConn.Close()

	repo := repository.NewAnalyticsRepo(dbConn)
	proc := processor.NewProcessor(cfg, logger, repo)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		defer signal.Stop(sig)
		<-sig
		logger.Warn("interrupt received, shutting down...")
		cancel()
	}()

	lines := make(chan string, DefaultChannelBuffer)
	scanner := bufio.NewScanner(file)
	// Go’s default scanner buffer is 64KB per line, which may fail for large JSON lines.
	buf := make([]byte, 0, 1024*1024) // 1MB buffer
	scanner.Buffer(buf, 10*1024*1024) // Max token size 10MB
	var wg sync.WaitGroup

	start := time.Now()
	totalLines := 0

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if err := proc.ProcessStream(ctx, lines, batchSize); err != nil {
				logger.Error("worker failed", "id", id, "error", err)
				cancel()
			}
		}(i)
	}

	// Feed input
FeedLoop:
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			break FeedLoop
		default:
			lines <- scanner.Text()
			totalLines++
		}
	}
	close(lines)
	wg.Wait()

	if ctx.Err() != nil {
		logger.Warn("Processing interrupted before completion")
		return ctx.Err()
	}

	elapsed := time.Since(start).Seconds()
	throughput := float64(totalLines) / elapsed

	stats := proc.Snapshot()
	logger.Info("Processing complete",
		"transactions", totalLines,
		"elapsed_sec", elapsed,
		"throughput_tps", throughput,
		"unique_users", len(stats))

	return scanner.Err()
}
