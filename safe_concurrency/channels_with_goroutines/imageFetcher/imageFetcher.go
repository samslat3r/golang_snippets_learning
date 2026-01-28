package main

// Initialize Logger
// Create Directory
// Fetch Image Links
// Log image links
// Download Images
// Collect Results

// The main idea is to use channels to coordinate the download of images, this is a demonstration but can be modified to download other things or walk through multiple pages to download images or other things

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

var (
	mainLogger      *logrus.Logger
	logFile         = "./imageFetcherLog.log"
	destDir         = "./images"
	pageURL         = "https://www.apache.org/"
	imageExtensions = []string{
		".ai", ".apng", ".arw", ".avif", ".bmp", ".cr2", ".cr3", ".crw",
		".dcr", ".dds", ".dib", ".dng", ".eps", ".erf", ".fff", ".gif",
		".heic", ".heif", ".hdp", ".ico", ".iiq", ".j2k", ".jp2", ".jpf",
		".jpm", ".jpx", ".jxr", ".kdc", ".mef", ".mj2", ".mos", ".mrw",
		".nef", ".nrw", ".orf", ".pjpeg", ".pjp", ".png", ".psd", ".ptx",
		".pxn", ".raf", ".raw", ".r3d", ".rw2", ".rwl", ".rwz", ".srf",
		".sr2", ".srw", ".svg", ".svgz", ".tif", ".tiff", ".webp", ".wdp",
		".xbm", ".x3f", ".jpeg", ".jpg", ".tga", ".exr", ".hdr", ".pcx", ".tga", ".tiff",
		".svg", ".svgz", ".ico", ".cur", ".wbmp", ".dds", ".jng",
	}
)

type result struct {
	url  string
	path string
	err  error
}
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return nil, fmt.Errorf("error loading location: %v", err)
	}
	entry.Time = entry.Time.In(loc)
	timestamp := entry.Time.Format("2006-01-02 3:04PM (PST)")

	var level string
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

func fetchImageLinks(pageURL string) ([]string, error) {
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching page: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing/reading HTML: %v", err)
	}

	var imageLinks []string

	doc.Find("img").Each(func(index int, item *goquery.Selection) {
		src, exists := item.Attr("src")
		if exists && hasImageExtension(src) {
			absoluteURL := toAbsoluteURL(pageURL, src)
			imageLinks = append(imageLinks, absoluteURL)
		}
	})

	return imageLinks, nil
}

func hasImageExtension(link string) bool {
	for _, ext := range imageExtensions {
		if strings.HasSuffix(strings.ToLower(link), ext) {
			return true
		}
	}
	return false
}

func toAbsoluteURL(base, link string) string {
	u, err := url.Parse(link)
	if err != nil || u.IsAbs() {
		return link
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return link
	}

	return baseURL.ResolveReference(u).String()
}

// downloadImage can be boiled down to "http.Get(imageURL)", "os.Create(destPath ) "io.Copy(file, resp.Body)"
// the results channel handles the coordination of this whole program , it is important
func downloadImage(imageURL, destDir string, results chan<- result) {
	mainLogger.Infof("Starting download of of image %s\n", imageURL)

	fileName := filepath.Base(imageURL)
	destPath := filepath.Join(destDir, fileName)

	resp, err := http.Get(imageURL)
	if err != nil {
		mainLogger.Errorf("Error writing file to %s with error %v\n", destPath, err)
		results <- result{imageURL, "", err}
		return
	}

	defer resp.Body.Close()

	file, err := os.Create(destPath)
	if err != nil {
		mainLogger.Errorf("Error creating file %s with error %s\n", destPath, err)
		results <- result{imageURL, "", err}
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		mainLogger.Errorf("Error saving image from %s to %s with error %v\n", imageURL, destPath, err)
		results <- result{imageURL, "", err}
	}

	mainLogger.Infof("Successfully downloaded image from %s to %s\n", imageURL, destPath)
	results <- result{imageURL, destPath, nil}

}

func main() {
	mainLogger = setupLogging(logFile)

	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		mainLogger.Fatalf("Failed creating destination %s with error %v\n", destDir, err)
	}

	imageLinks, err := fetchImageLinks(pageURL)
	if err != nil {
		mainLogger.Fatalf("Could not play fetch with this url: %s with error %v\n", pageURL, err)
	}
	mainLogger.Infof("Found %d images on page %s\n", len(imageLinks), pageURL)

	//and now the magic happens

	//buffered channel
	results := make(chan result, len(imageLinks))
	defer close(results)

	for _, imageURL := range imageLinks {
		go downloadImage(imageURL, destDir, results)
	}

	for i := 0; i < len(imageLinks); i++ {

		//results is a channel, so we can use it as a blocking operation
		//this will wait until a value is sent to the channel

		result := <-results
		if result.err != nil {
			mainLogger.Errorf("Couldn't download %s with error %v\n", result.url, result.err)
		} else {
			mainLogger.Infof("Downloaded image %s to %s\n", result.url, result.path)
		}
	}
	mainLogger.Info("All images downloaded")
}
