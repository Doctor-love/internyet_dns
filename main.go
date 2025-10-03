package main

import (
	"os"
	"log"
	"slices"
	"strings"
	"net"
	"net/http"
	"regexp"
)

var configurationDirectory, rootDomain, listenAddress string
var subDomainRegexp, _ = regexp.Compile("^[a-z]([-a-z0-9]*[a-z0-9])?$")

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

	var clientGreatHouses []string
	clientGreatHousesHeader := request.Header.Get("X-Internyet-Client-Great-Houses")
	if clientGreatHousesHeader != "" {
		clientGreatHouses = strings.Split(clientGreatHousesHeader, ",")
	} 

	if request.Header.Get("X-SillyCSRF") != "false" {
		http.Error(response, "Request missing CSRF header", http.StatusBadRequest)
		return
	}

	path := request.URL.Path
	path = strings.TrimPrefix(path, "/api/v2/")
	pathParts := strings.Split(path, "/")

	var greatHouseSubDomain, targetAddress string
	var isForGreatHouse bool
	
	if len(pathParts) == 3 {
		targetAddress = pathParts[2]

	} else if len(pathParts) == 4 {
		isForGreatHouse = true
		greatHouseSubDomain = pathParts[2]
		targetAddress = pathParts[3]
				
	} else {
		http.Error(response, "URL path format is invalid", http.StatusBadRequest)
		return
	}

	recordType := pathParts[0]
	subDomain := pathParts[1]

	if recordType != "A" && recordType != "AAAA" {
		http.Error(response, "Record type is invalid: " + recordType, http.StatusBadRequest)
		return
	}

	if subDomain == "" {
		http.Error(response, "URL part for sub-domain is empty", http.StatusBadRequest)
		return
	}

	if isForGreatHouse && greatHouseSubDomain == "" {
		http.Error(response, "URL part for great house is empty", http.StatusBadRequest)
		return
	}

	if targetAddress == "" {
		http.Error(response, "URL part for target address is empty", http.StatusBadRequest)
		return
	}

	if len(subDomain) > 63 || len(greatHouseSubDomain) > 63 {
		http.Error(response, "URL part for sub-domain/great house is too long", http.StatusBadRequest)
		return
	}
	
	if ! subDomainRegexp.MatchString(subDomain) {
		http.Error(response, "URL part for sub-domain is invalid", http.StatusBadRequest)
		return
	}

	if isForGreatHouse && ! subDomainRegexp.MatchString(greatHouseSubDomain) {
		http.Error(response, "URL part for great house is invalid", http.StatusBadRequest)
		return
	}

	if targetAddress == "this" {
		targetAddress = sourceAddress
	}

	var domain string
	if isForGreatHouse {
		domain = subDomain + "." + greatHouseSubDomain + ".g." + rootDomain

	} else {
		domain = subDomain + "." + clientAlias + ".p." + rootDomain
	}

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

	if isForGreatHouse {
		if ! slices.Contains(clientGreatHouses, greatHouseSubDomain) {
			http.Error(response, "Client isn't member of specified great house", http.StatusBadRequest)
			return
		}
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
	http.HandleFunc("/api/v2/", configurationHandler)
	log.Print("Starting listener on ", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
