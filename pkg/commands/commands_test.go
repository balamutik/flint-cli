package commands

import "testing"

// TestFormatSize tests the formatSize helper function
func TestFormatSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, test := range tests {
		result := formatSize(test.size)
		if result != test.expected {
			t.Fatalf("formatSize(%d) = %s, expected %s", test.size, result, test.expected)
		}
	}
}

// TestFormatNumber tests the formatNumber helper function
func TestFormatNumber(t *testing.T) {
	tests := []struct {
		num      int64
		expected string
	}{
		{123, "123"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
		{100000, "100,000"},
	}

	for _, test := range tests {
		result := formatNumber(test.num)
		if result != test.expected {
			t.Fatalf("formatNumber(%d) = %s, expected %s", test.num, result, test.expected)
		}
	}
}

// TestFormatFileSize tests the formatFileSize helper function
func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1048576, "1.0MB"},
		{1073741824, "1.0GB"},
		{1099511627776, "1.0TB"},
	}

	for _, test := range tests {
		result := formatFileSize(test.size)
		if result != test.expected {
			t.Fatalf("formatFileSize(%d) = %s, expected %s", test.size, result, test.expected)
		}
	}
}
