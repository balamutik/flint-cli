package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

func main() {
	fmt.Println("🔍 Testing Compression Functions Separately")
	fmt.Println("==========================================")

	// Read test file
	testFile := "test_data/file1.txt"
	data, err := os.ReadFile(testFile)
	if err != nil {
		fmt.Printf("❌ Failed to read file: %v\n", err)
		return
	}

	fmt.Printf("📋 Original data: %q (%d bytes)\n", string(data), len(data))

	// Test our compression function
	compressed, err := compressData(data)
	if err != nil {
		fmt.Printf("❌ Compression failed: %v\n", err)
		return
	}

	fmt.Printf("📋 Compressed data: %d bytes\n", len(compressed))
	fmt.Printf("📋 Compressed data (hex): %x\n", compressed[:20]) // First 20 bytes

	// Test our decompression function
	decompressed, err := decompressData(compressed)
	if err != nil {
		fmt.Printf("❌ Decompression failed: %v\n", err)
		return
	}

	fmt.Printf("📋 Decompressed data: %q (%d bytes)\n", string(decompressed), len(decompressed))

	// Check if they match
	if string(data) == string(decompressed) {
		fmt.Println("✅ Compression/decompression cycle PASSED")
	} else {
		fmt.Println("❌ Compression/decompression cycle FAILED")
	}

	// Let's also test if the compressed data has correct gzip header
	fmt.Println("\n🔍 Checking gzip header...")
	if len(compressed) >= 2 {
		if compressed[0] == 0x1f && compressed[1] == 0x8b {
			fmt.Println("✅ Gzip magic header present")
		} else {
			fmt.Printf("❌ Invalid gzip header: %02x %02x\n", compressed[0], compressed[1])
		}
	}
}

// compressData compresses data using gzip compression
func compressData(data []byte) ([]byte, error) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return compressed.Bytes(), nil
}

// decompressData decompresses gzip-compressed data
func decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
