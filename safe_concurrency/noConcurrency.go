package main

//For this purpose we're assuming you have the directory srcDir with files in it
// Also assumes you have the directory destDir to copy the files to

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

func copySrcToDest(srcDir, destDir string, mainLogger *logrus.Logger) {

	//check source and dest directories exist

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		mainLogger.Errorf("Source directory %s does not exist: %v\n", srcDir, err)
		return
	} else if err != nil {
		mainLogger.Errorf("Error checking source directory %s: %v\n", srcDir, err)
		return
	}

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		mainLogger.Errorf("Destination directory %s does not exist: %v\n", destDir, err)
		return
	} else if err != nil {
		mainLogger.Errorf("Error checking destination directory %s: %v\n", destDir, err)
		return
	}

	//walk through src
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			mainLogger.Errorf("Error accessing path %s: %v\n", path, err)
			return err
		}

		// Create corresponding path in destination directory
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			mainLogger.Errorf("Error determining relative path for %s: %v\n", path, err)
			return err
		}
		destPath := filepath.Join(destDir, relPath)

		// Check if dir or file in src
		if info.IsDir() {
			// Create dir if it is a directory
			if err := os.MkdirAll(destPath, info.Mode()); err != nil {
				mainLogger.Errorf("Error creating directory %s: %v\n", destPath, err)
				return err
			}
		} else {
			// Copy file if file
			if err := copy.Copy(path, destPath); err != nil {
				mainLogger.Errorf("Error copying file %s to %s: %v\n", path, destPath, err)
				return err
			}
			mainLogger.Infof("Copied file %s to %s\n", path, destPath)

		}
		return nil
	})

	if err != nil {
		mainLogger.Errorf("Error walking through source directory %s: %v\n", srcDir, err)
	} else {
		mainLogger.Infof("Successfully copied all files from %s to %s\n", srcDir, destDir)
	}

}
