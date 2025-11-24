package main

import (
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang/v2"
)

func getIPData(ipAddr *netip.Addr) *geoip2.City {
	db, err := geoip2.Open("/geoip/GeoLite2-City.mmdb")
	if err != nil {
		log.Printf("error occurred when trying to retrieve data from mmdb: %v", err)
		return nil
	}
	defer db.Close()

	record, err := db.City(*ipAddr)
	if err != nil {
		log.Printf("error occurred when trying to retrieve data from mmdb: %v", err)
		return nil
	}

	if !record.HasData() {
		log.Printf("empty data returned when trying to retrieve data from mmdb")
		return nil
	}

	return record
}

func generateFilename(ipStr string) string {
	ipStr = strings.ReplaceAll(ipStr, ".", "")
	ipStr = strings.ReplaceAll(ipStr, ":", "")

	return ipStr
}

func isBot(r *http.Request) bool {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor == "" {
		return true // This should always run behind a proxy
	}

	userAgent := strings.ToLower(r.UserAgent())

	// Discord embed
	if userAgent == "mozilla/5.0 (macintosh; intel mac os x 10.10; rv:38.0) gecko/20100101 firefox/38.0" {
		return true
	}

	// Bots that ignore robots.txt
	if userAgent == "" || strings.Contains(userAgent, "bot") || strings.Contains(userAgent, "embed") ||
		strings.Contains(userAgent, "crawl") || strings.Contains(userAgent, "spider") ||
		strings.Contains(userAgent, "scrape") || strings.Contains(userAgent, "scrape") {
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

func handleVideo(w http.ResponseWriter, r *http.Request) {
	if isBot := isBot(r); isBot {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("418 - I'm a teapot"))
		return
	}

	ipAddr := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
	filePath := fmt.Sprintf("/tmp/%v.mp4", generateFilename(ipAddr))

	// File already exists, grabbing it and sending it instead of regenerating it.
	if _, err := os.Stat(filePath); err == nil {
		handleFileSend(filePath, w, r)
		return
	}

	netIPAddr, err := netip.ParseAddr(ipAddr)
	if err != nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("418 - I'm a teapot"))
		log.Printf("error occurred when trying to retrieve data from IPInfo: %v", err)
		return
	}

	ipData := getIPData(&netIPAddr)
	if ipData == nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("418 - I'm a teapot"))
		log.Printf("error occurred when trying to generate video : %v", err)
		return
	}

	err = generateVideo(filePath, ipData)
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

func handleButton(w http.ResponseWriter, r *http.Request) {
	if isBot := isBot(r); isBot {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("418 - I'm a teapot"))
		return
	}

	userAgent := strings.TrimSpace(strings.Split(r.Header.Get("User-Agent"), ",")[0])
	ipAddr := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
	filePath := fmt.Sprintf("/tmp/%v.gif", generateFilename(ipAddr))

	// File already exists, grabbing it and sending it instead of regenerating it.
	if _, err := os.Stat(filePath); err == nil {
		handleFileSend(filePath, w, r)
		return
	}

	netIPAddr, err := netip.ParseAddr(ipAddr)
	if err != nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("418 - I'm a teapot"))
		log.Printf("error occurred when trying to retrieve data from IPInfo: %v", err)
		return
	}
	ipData := getIPData(&netIPAddr)
	if ipData == nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("418 - I'm a teapot"))
		log.Printf("error occurred when trying to generate video : %v", err)
		return
	}

	err = generateButton(filePath, userAgent, ipData)
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
	http.HandleFunc("/wits.mp4", func(w http.ResponseWriter, r *http.Request) {
		handleVideo(w, r)
	})
	http.HandleFunc("/wits.gif", func(w http.ResponseWriter, r *http.Request) {
		handleButton(w, r)
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
