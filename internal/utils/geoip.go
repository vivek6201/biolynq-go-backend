package utils

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
)

type GeoIPService struct {
	reader *geoip2.Reader
}

// NewGeoIPService loads the GeoIP database. If it fails, it returns a service with a nil reader
// which will degrade gracefully and fall back to public HTTP API lookup.
func NewGeoIPService(dbPath string) *GeoIPService {
	if dbPath == "" {
		log.Println("GeoIP: Database path is empty. Geographic tracking will use public HTTP API lookup.")
		return &GeoIPService{reader: nil}
	}

	reader, err := geoip2.Open(dbPath)
	if err != nil {
		log.Printf("GeoIP: Local database not loaded (%v). Geographic tracking will use public HTTP API lookup.", err)
		return &GeoIPService{reader: nil}
	}

	log.Printf("GeoIP: Successfully loaded database at %s", dbPath)
	return &GeoIPService{reader: reader}
}

// Close closes the underlying GeoIP database reader
func (g *GeoIPService) Close() {
	if g.reader != nil {
		_ = g.reader.Close()
	}
}

// Lookup takes an IP string and returns (country, city)
func (g *GeoIPService) Lookup(ipStr string) (string, string) {
	ipStr = strings.TrimSpace(ipStr)

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "Unknown", "Unknown"
	}

	// Local & Private range check (e.g. 127.0.0.1, Docker gateways, LAN IPs)
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() {
		return "Local", "Local"
	}

	// 1. Try local MMDB first (extremely fast, offline)
	if g.reader != nil {
			record, err := g.reader.City(ip)
			if err == nil {
				country := record.Country.Names["en"]
				city := record.City.Names["en"]
				if country == "" {
					country = "Unknown"
				}
				if city == "" {
					city = "Unknown"
				}
				return country, city
			}
		}

	// 2. Fall back to free public HTTP API (ip-api.com)
	return lookupIPAPI(ipStr)
}

// lookupIPAPI makes an HTTP request to ip-api.com to resolve the IP address.
func lookupIPAPI(ipStr string) (string, string) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://ip-api.com/json/" + ipStr)
	if err != nil {
		log.Printf("GeoIP Fallback: HTTP request to ip-api.com failed: %v", err)
		return "Unknown", "Unknown"
	}
	defer resp.Body.Close()

	var result struct {
		Status  string `json:"status"`
		Country string `json:"country"`
		City    string `json:"city"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "Unknown", "Unknown"
	}

	if result.Status != "success" {
		return "Unknown", "Unknown"
	}

	country := result.Country
	city := result.City
	if country == "" {
		country = "Unknown"
	}
	if city == "" {
		city = "Unknown"
	}

	return country, city
}
