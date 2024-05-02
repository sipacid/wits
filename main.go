package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ipinfo/go/v2/ipinfo"
)

func isBot(r *http.Request) bool {
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP == "" {
		return true // This should always run behind a proxy
	}

	userAgent := strings.ToLower(r.UserAgent())

	// Discord embed
	if userAgent == "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.10; rv:38.0) Gecko/20100101 Firefox/38.0" {
		return true
	}

	// Bots that ignore robots.txt
	if userAgent == "" || strings.Contains(userAgent, "bot") || strings.Contains(userAgent, "embed") || strings.Contains(userAgent, "crawl") || strings.Contains(userAgent, "spider") || strings.Contains(userAgent, "scrape") || strings.Contains(userAgent, "scrape") {
		return true
	}

	return false
}

func handleFileSend(filePath string, w http.ResponseWriter, r *http.Request) {
	_, err := os.Stat(filePath)
	if err != nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("418 - I'm a teapot"))
		log.Printf("error occurred when trying to read exising file: %v", err)
		return
	}

	http.ServeFile(w, r, filePath)
}

func handler(c *ipinfo.Client, w http.ResponseWriter, r *http.Request) {
	// Handle bots
	if isBot := isBot(r); isBot {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("418 - I'm a teapot"))
		return
	}

	ipAddr := r.Header.Get("X-Real-IP")
	filePath := fmt.Sprintf("/tmp/%v.mp4", generateFilename(ipAddr))

	// File already exists, grabbing it and sending it instead of regenerating it.
	if _, err := os.Stat(filePath); err == nil {
		handleFileSend(filePath, w, r)
		return
	}

	ipInfo, err := c.GetIPInfo(net.ParseIP(ipAddr))
	if err != nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("418 - I'm a teapot"))
		log.Printf("error occurred when trying to retrieve data from IPInfo: %v", err)
		return
	}

	err = generateVideo(filePath, ipAddr, ipInfo)
	if err != nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("418 - I'm a teapot"))
		log.Printf("error occurred when trying to generate video : %v", err)
		return
	}

	handleFileSend(filePath, w, r)

	// Aight, we've sent the file, let's delete it after a few mins. Don't care if it fails, fuckyou.
	time.AfterFunc(2*time.Minute, func() {
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to delete video file: %v", err)
		}
	})
}

func main() {
	token, present := os.LookupEnv("IPINFO_TOKEN")
	if !present {
		log.Fatal("IPINFO_TOKEN environment variable not set, exiting...")
	}

	ipinfo_client := ipinfo.NewClient(nil, nil, token)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(ipinfo_client, w, r)
	})
	log.Fatal(http.ListenAndServe(":3000", nil))
}
