package sanitize

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

// Predefined validation errors returned by the URL sanitizer
var (
	ErrEmptyURL     = errors.New("URL cannot be empty")
	ErrInvalidURL   = errors.New("URL is not valid")
	ErrUnsafeScheme = errors.New("URL scheme not allowed - only http and https are permitted")
	ErrMissingHost  = errors.New("URL must include a host")
	ErrPrivateHost  = errors.New("URL points to a private or reserved address")
	ErrLoopbackHost = errors.New("URL points to loopback address")
)

// Allowed URL schemes for safety
var allowedSchemes = map[string]bool{
	"http":  true,
	"https": true,
}

// privateRanges holds private and reserved IP CIDR blocks
var privateRanges []*net.IPNet

// init parses CIDR ranges and stores them for private IP checks
func init() {
	cidrs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16", // IPv4 link-local
		"100.64.0.0/10",  // shared address space
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 unique local
		"fe80::/10",      // IPv6 link-local
	}

	for _, cidr := range cidrs {
		_, block, _ := net.ParseCIDR(cidr)
		privateRanges = append(privateRanges, block)
	}
}

// URL validates and normalizes a raw URL string
func URL(raw string) (string, error) {
	//Remove surrounding whitespace
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrEmptyURL
	}

	//Add default scheme if missing
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	//Parse the URL
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", ErrInvalidURL
	}

	//Normalize scheme
	parsed.Scheme = strings.ToLower(parsed.Scheme)

	//Allow only safe schemes
	if !allowedSchemes[parsed.Scheme] {
		return "", ErrUnsafeScheme
	}

	//Normalize host
	parsed.Host = strings.ToLower(parsed.Host)
	host := parsed.Hostname()

	//Enusre host exists
	if host == "" {
		return "", ErrMissingHost
	}

	//Block localhost explicitly
	if host == "localhost" {
		return "", ErrLoopbackHost
	}

	//Resolve host to IP addresses
	ips, err := net.LookupHost(host)
	if err != nil {
		ips = []string{host}
	}

	//Validate each resolved IP
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}

		//Block loopback addresses
		if ip.IsLoopback() {
			return "", ErrLoopbackHost
		}

		//Block private/reserved networks
		if isPrivate(ip) {
			return "", ErrPrivateHost
		}
	}

	//Remove credentials
	parsed.User = nil

	//Remove fragment
	parsed.Fragment = ""

	return parsed.String(), nil
}

// isPrivate checks if an IP belongs to a private or reserved range
func isPrivate(ip net.IP) bool {
	for _, block := range privateRanges {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
