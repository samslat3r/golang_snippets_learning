package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/otiai10/copy"
	"github.com/sirupsen/logrus"
)

var (
	srcDir     = "/home/sslater/exampleFiles"
	destDir    = "/home/users/sslater/golang_snippets_learning/safe_concurrency/filesgohere"
	logdir     = filepath.Join(destDir, "logs")
	logfile    = filepath.Join(logdir, "logfile.txt")
	mainLogger *logrus.Logger
)

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return nil, fmt.Errorf("error loading location: %v", err)
	}
	entry.Time = entry.Time.In(loc)
	timestamp := entry.Time.Format("2006-01-02 - 3:04PM (PST)")

	level := entry.Level.String()
	switch entry.Level {
	case logrus.InfoLevel:
		level = "[INFO]"
	case logrus.WarnLevel:
		level = "[WARN]"
	case logrus.ErrorLevel:
		level = "[ERROR]"
	case logrus.FatalLevel:
		level = "[FATAL]"
	case logrus.PanicLevel:
		level = "[PANIC]"
	case logrus.DebugLevel:
		level = "[DEBUG]"
	default:
		level = "[UNKNOWN]"
	}

	msg := fmt.Sprintf("%s %s %s\n", timestamp, level, entry.Message)
	return []byte(msg), nil
}

func setupLogging(logFile string) *logrus.Logger {
	mainLogger := logrus.New()

	mainLogFile, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening main log file: %v\n", err)
		os.Exit(1)
	}

	mainLogger.SetOutput(mainLogFile)
	mainLogger.SetLevel(logrus.DebugLevel)
	mainLogger.SetFormatter(&CustomFormatter{})

	return mainLogger
}

// Walks through the directory, sends paths of all files and directories to the channel "paths"
// "Paths" is a channel of strings to which we will send the paths of all files and directories
func walkSourceDir(srcDir string, paths chan<- string, mainLogger *logrus.Logger) {
	defer close(paths)
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			mainLogger.Errorf("Error accessing directory %s: %v\n", path, err)
			return err
		}
		paths <- path
		return nil
	})

	if err != nil {
		mainLogger.Errorf("Error walking source directory %s: %v\n", srcDir, err)
	}
}

// Also will run as a goroutine. It will copy the files from the source directory to the destination directory. We aren't running this directly.
func copyWorker(paths <-chan string, destDir string, results chan<- error, mainLogger *logrus.Logger) {
	for path := range paths {
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			mainLogger.Errorf("Error getting relative path: %v\n", err)
			results <- err
			continue
		}
		destPath := filepath.Join(destDir, relPath)

		info, err := os.Stat(path)
		if err != nil {
			mainLogger.Errorf("Error geting file info for %s: %v\n", path, err)
			results <- err
			continue
		}

		if info.IsDir() {
			//Create the directory in destination
			if err := os.MkdirAll(destPath, info.Mode()); err != nil {
				mainLogger.Errorf("Error creating directory %s: %v\n", destPath, err)
				results <- err
			}
		} else {
			// It must be a file then, so copy the file to the destination
			if err := copy.Copy(path, destPath); err != nil {
				mainLogger.Errorf("Error copying file %s to %s: %v\n", path, destPath, err)
				results <- err
			} else {
				mainLogger.Infof("Copied file %s to %s\n", path, destPath)
			}
		}

		results <- nil
	}
}

// Here is where the magic happens.
func copySrcToDest(srcDir, destDir string, mainLogger *logrus.Logger) {

	//start by checking if destination exists..
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		mainLogger.Errorf("Destination directory does not exist: %s\n", destDir)
		return
	} else if err != nil {
		mainLogger.Errorf("Error checking destination directory: %v\n", err)
		return
	}

	paths := make(chan string)
	results := make(chan error)
	const numWorkers = 4 // Number of worker goroutines, this can be adjusted to suit your needs

	//Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		go copyWorker(paths, destDir, results, mainLogger)
	}

	//Start the source directory walk
	go walkSourceDir(srcDir, paths, mainLogger)

	//collect results
	for i := 0; i < numWorkers; i++ {
		for err := range results {
			if err != nil {
				mainLogger.Errorf("Error during copy: %v\n", err)
			}
		}
	}
}

func main() {
	mainLogger = setupLogging(logfile)
	mainLogger.Info("Starting copy operation")

	copySrcToDest(srcDir, destDir, mainLogger)

	mainLogger.Info("Copy operation completed")
}
