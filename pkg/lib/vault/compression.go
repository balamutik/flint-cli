package vault

import (
	"bytes"
	"compress/gzip"
	"io"
)

// compressData compresses data using gzip compression.
// This internal function reduces vault file size by compressing the JSON data
// before encryption. Uses standard gzip compression for broad compatibility.
//
// Parameters:
//   - data: Raw data to compress
//
// Returns:
//   - []byte: Compressed data
//   - error: nil on success, or error if compression fails
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

// decompressData decompresses gzip-compressed data.
// This internal function reverses the compression applied by compressData,
// restoring the original JSON data after decryption.
//
// Parameters:
//   - data: Gzip-compressed data to decompress
//
// Returns:
//   - []byte: Decompressed original data
//   - error: nil on success, or error if decompression fails
func decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
