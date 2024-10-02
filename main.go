package main

import (
	"os"
	"log"
	"strings"
	"net"
	"net/http"
	"regexp"
)

var configurationDirectory, rootDomain, listenAddress string

// ---
func init() {
	configurationDirectory = os.Getenv("IDNS_CONF_ROOT")
	rootDomain = os.Getenv("IDNS_ROOT_DOMAIN")
	listenAddress = os.Getenv("IDNS_LISTEN_ADDRESS")
}

// ---
func configurationHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(response, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sourceAddress := request.Header.Get("X-Forwarded-For")
	if sourceAddress == "" {
		http.Error(response, "Request missing XFF header", http.StatusBadRequest)
		return
	}

	clientAlias := request.Header.Get("X-Internyet-Client-Alias")
	if clientAlias == "" {
		http.Error(response, "Request missing alias header", http.StatusBadRequest)
		return
	}

	if request.Header.Get("X-SillyCSRF") != "false" {
		http.Error(response, "Request missing CSRF header", http.StatusBadRequest)
		return
	}

	path := request.URL.Path
	path = strings.TrimPrefix(path, "/api/v1/")
	pathParts := strings.Split(path, "/")
	if len(pathParts) != 3 {
		http.Error(response, "URL path format is invalid", http.StatusBadRequest)
		return
	}

	recordType := pathParts[0]
	subDomain := pathParts[1]
	targetAddress := pathParts[2]

	if recordType != "A" && recordType != "AAAA" {
		http.Error(response, "Record type is invalid: " + recordType, http.StatusBadRequest)
		return
	}

	if subDomain == "" {
		http.Error(response, "URL part for sub-domain is empty", http.StatusBadRequest)
		return
	}

	if targetAddress == "" {
		http.Error(response, "URL part for target address is empty", http.StatusBadRequest)
		return
	}

	if ! regexp.MustCompile("^[a-z]+$").MatchString(subDomain) {
		http.Error(response, "URL part for sub-domain is not a-z: " + subDomain, http.StatusBadRequest)
		return
	}

	if targetAddress == "this" {
		targetAddress = sourceAddress
	}

	domain := subDomain + "." + clientAlias + "." + rootDomain
	log.Printf(
		"Participant \"%s@%s\" is trying to register %s \"%s\" to \"%s\"",
		clientAlias, sourceAddress, recordType, domain, targetAddress)

	parsedAddress := net.ParseIP(targetAddress)
	if parsedAddress == nil {
		http.Error(response, "Failed to parse target address: " + targetAddress, http.StatusBadRequest)
		return
	}

	if ! parsedAddress.IsPrivate() {
		http.Error(response, "Target address is not local/private", http.StatusBadRequest)
		return
	}

	if parsedAddress.IsLoopback() || parsedAddress.IsUnspecified() {
		http.Error(response, "Target address is not a valid type", http.StatusBadRequest)
		return
	}

	var addressType string
	if parsedAddress.To4() != nil {
		addressType = "v4"
		
	} else {
		addressType = "v6"
	}

	if recordType == "A" && addressType == "v6" {
		http.Error(response, "IPv6 target address is invalid for A record", http.StatusBadRequest)
		return
	}

	if recordType == "AAAA" && addressType == "v4" {
		http.Error(response, "IPv4 target address is invalid for AAAA record", http.StatusBadRequest)
		return
	}

	filePath := configurationDirectory + "/" + recordType + "_" + domain
	hostsEntry := parsedAddress.String() + " " + domain
	log.Printf("Writing \"%s\" to \"%s\"", hostsEntry, filePath)

	fileHandle, err := os.Create(filePath)
	if err != nil {
		http.Error(response, "Failed to write hosts entry", http.StatusInternalServerError)
		return
	}

	defer fileHandle.Close()

	fileHandle.WriteString(hostsEntry + "\n")
}

// ---
func main() {
	http.HandleFunc("/api/v1/", configurationHandler)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
