package utils

import (
	"net"
	"testing"
)

func TestGetLocalIP(t *testing.T) {
	ip := GetLocalIP()

	// Should not be empty
	if ip == "" {
		t.Error("GetLocalIP returned empty string")
	}

	// Should be a valid IP address or "localhost"
	if ip != "localhost" {
		parsed := net.ParseIP(ip)
		if parsed == nil {
			t.Errorf("GetLocalIP returned invalid IP address: %s", ip)
		}

		// Should not be loopback (unless that's all we have)
		if parsed.IsLoopback() && ip != "127.0.0.1" {
			t.Logf("Warning: GetLocalIP returned loopback address: %s", ip)
		}
	}
}

func TestGetLocalIPReturnsConsistentResult(t *testing.T) {
	// Call the function multiple times and ensure it returns the same result
	ip1 := GetLocalIP()
	ip2 := GetLocalIP()
	ip3 := GetLocalIP()

	if ip1 != ip2 || ip2 != ip3 {
		t.Errorf("GetLocalIP returned inconsistent results: %s, %s, %s", ip1, ip2, ip3)
	}
}
