package vault

import (
	"bytes"
	"testing"
)

// TestCompressDecompressData tests the compression and decompression functions
func TestCompressDecompressData(t *testing.T) {
	testData := []byte("This is test data for compression testing. " +
		"It should be compressed and then decompressed back to original form.")

	// Test compression
	compressed, err := compressData(testData)
	if err != nil {
		t.Fatalf("compressData failed: %v", err)
	}

	// Verify compressed data is different from original
	if bytes.Equal(compressed, testData) {
		t.Fatal("Compressed data should be different from original")
	}

	// Test decompression
	decompressed, err := decompressData(compressed)
	if err != nil {
		t.Fatalf("decompressData failed: %v", err)
	}

	// Verify decompressed data matches original
	if !bytes.Equal(decompressed, testData) {
		t.Fatalf("Decompressed data doesn't match original. Expected: %s, got: %s",
			string(testData), string(decompressed))
	}
}

// TestCompressEmptyData tests compression of empty data
func TestCompressEmptyData(t *testing.T) {
	testData := []byte("")

	compressed, err := compressData(testData)
	if err != nil {
		t.Fatalf("compressData failed on empty data: %v", err)
	}

	decompressed, err := decompressData(compressed)
	if err != nil {
		t.Fatalf("decompressData failed on empty compressed data: %v", err)
	}

	if !bytes.Equal(decompressed, testData) {
		t.Fatal("Decompressed empty data doesn't match original empty data")
	}
}

// TestCompressHighlyCompressibleData tests compression of highly repetitive data
func TestCompressHighlyCompressibleData(t *testing.T) {
	// Create highly repetitive data that should compress very well
	testData := bytes.Repeat([]byte("AAAA"), 1000) // 4000 bytes of repeated "AAAA"

	compressed, err := compressData(testData)
	if err != nil {
		t.Fatalf("compressData failed: %v", err)
	}

	// Verify significant compression was achieved
	compressionRatio := float64(len(compressed)) / float64(len(testData))
	if compressionRatio > 0.1 { // Should achieve better than 90% compression
		t.Fatalf("Expected compression ratio < 0.1 for highly repetitive data, got %f", compressionRatio)
	}

	// Verify decompression works
	decompressed, err := decompressData(compressed)
	if err != nil {
		t.Fatalf("decompressData failed: %v", err)
	}

	if !bytes.Equal(decompressed, testData) {
		t.Fatal("Decompressed data doesn't match original highly compressible data")
	}
}

// TestCompressRandomData tests compression of random-like data
func TestCompressRandomData(t *testing.T) {
	// Create data that doesn't compress well (pseudo-random)
	testData := make([]byte, 1000)
	for i := range testData {
		testData[i] = byte(i*37 + 127) // Pseudo-random pattern
	}

	compressed, err := compressData(testData)
	if err != nil {
		t.Fatalf("compressData failed: %v", err)
	}

	// For random data, compression might not achieve much reduction
	// but it should still work
	if len(compressed) == 0 {
		t.Fatal("Compressed data should not be empty")
	}

	// Verify decompression works
	decompressed, err := decompressData(compressed)
	if err != nil {
		t.Fatalf("decompressData failed: %v", err)
	}

	if !bytes.Equal(decompressed, testData) {
		t.Fatal("Decompressed random data doesn't match original")
	}
}

// TestDecompressInvalidData tests decompression of invalid data
func TestDecompressInvalidData(t *testing.T) {
	invalidData := []byte("This is not valid gzip data")

	_, err := decompressData(invalidData)
	if err == nil {
		t.Fatal("Expected error when decompressing invalid data, got nil")
	}
}

// BenchmarkCompressData benchmarks the compression function
func BenchmarkCompressData(b *testing.B) {
	testData := bytes.Repeat([]byte("Test data for compression benchmarking. "), 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compressData(testData)
		if err != nil {
			b.Fatalf("compressData failed: %v", err)
		}
	}
}

// BenchmarkDecompressData benchmarks the decompression function
func BenchmarkDecompressData(b *testing.B) {
	testData := bytes.Repeat([]byte("Test data for decompression benchmarking. "), 100)

	// Pre-compress the data
	compressed, err := compressData(testData)
	if err != nil {
		b.Fatalf("Failed to prepare compressed data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decompressData(compressed)
		if err != nil {
			b.Fatalf("decompressData failed: %v", err)
		}
	}
}
