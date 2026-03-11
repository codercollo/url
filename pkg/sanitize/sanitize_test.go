package sanitize_test

import (
	"testing"
	"url/pkg/sanitize"
)

func TestURL_ValidHTTPS(t *testing.T) {
	out, err := sanitize.URL("https://example.com/path?q=1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "https://example.com/path?q=1" {
		t.Errorf("unexpected output: %s", out)
	}

}

// Test that URLs without a scheme default to HTTPS
func TestURL_AddsHTTPSSheme(t *testing.T) {
	out, err := sanitize.URL("example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "https://example.com" {
		t.Errorf("unexpected output: %s", out)
	}
}

// Test that empty input returns an error
func TestURL_Empty_ReturnsError(t *testing.T) {
	_, err := sanitize.URL("")
	if err == nil {
		t.Fatalf("expected error for empty URL")
	}
}

// Test that javascript: schema is rejected for security
func TestURL_JavascriptScheme_Rejected(t *testing.T) {
	_, err := sanitize.URL("javascript:alert(1)")
	if err == nil {
		t.Fatal("expected error for javascript: sheme")
	}
}

// Test that file: scheme is rejected
func TestURL_FileScheme_Rejected(t *testing.T) {
	_, err := sanitize.URL("file:///etc/passwd")
	if err == nil {
		t.Fatal("expected error for file: scheme")
	}
}

// Test that localhost URLs are blocked
func TestURL_Localhost_Rejected(t *testing.T) {
	_, err := sanitize.URL("http://localhost:8080/admin")
	if err == nil {
		t.Fatal("expected error for localhost")
	}
}

// Test that provate network IPs are rejected
func TestURL_PrivateIP_Rejected(t *testing.T) {
	_, err := sanitize.URL("http://192.168.1.1/admin")
	if err == nil {
		t.Fatal("expected error for private IP")
	}
}

// Test that credentials in the URL are stripped
func TestURL_StripsUserInfo(t *testing.T) {
	out, err := sanitize.URL("https://user:pass@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "https://example.com" {
		t.Errorf("expected user info stripped, got %s", out)
	}
}

// Test that URL fragraments are removed
func TestURL_StripFragment(t *testing.T) {
	out, err := sanitize.URL("https://example.com/page#section")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "https://example.com/page" {
		t.Errorf("expected fragment stripped, got %s", out)
	}
}

// Test that scheme and host are normalized to lowercase
func TestURL_NormalisesSchemeToLower(t *testing.T) {
	out, err := sanitize.URL("HTTPS://Example.COM/Path")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "https://example.com/Path" {
		t.Errorf("unexpected output: %s", out)
	}
}
