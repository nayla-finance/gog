package iolimit

import (
	"fmt"
	"io"
)

// ReadAll reads all data from the provided io.Reader up to a specified maximum size limit.
// This function is designed to prevent potential memory exhaustion attacks by limiting
// the amount of data that can be read from an untrusted source.
//
// The function uses io.LimitReader internally to cap the reading at maxSize bytes.
// If the actual data size equals or exceeds the maxSize limit, an error is returned
// to indicate that the response exceeded the acceptable size threshold.
//
// Parameters:
//   - r: The io.Reader to read data from (e.g., HTTP response body, file, etc.)
//   - maxSize: The maximum number of bytes allowed to be read. Must be positive.
//
// Returns:
//   - []byte: The complete data read from the reader (only if within size limit)
//   - error: An error if reading fails or if the data size exceeds maxSize
//
// Error conditions:
//   - Returns the underlying io.ReadAll error if reading fails
//   - Returns a size limit error if data equals or exceeds maxSize bytes
//   - The size limit error indicates potential oversized content
//
// Example usage:
//
//	// Reading HTTP response with 1MB limit
//	data, err := ReadAll(response.Body, 1024*1024)
//	if err != nil {
//	    // Handle error (could be read failure or size limit exceeded)
//	    return err
//	}
//	// Process data safely knowing it's within size limits
func ReadAll(r io.Reader, maxSize int64) ([]byte, error) {
	l := io.LimitReader(r, maxSize+1) // Read one extra byte
	readBytes, err := io.ReadAll(l)
	if err != nil {
		return nil, err
	}

	if int64(len(readBytes)) > maxSize {
		return nil, fmt.Errorf("reader exceeded size limit of %d bytes", maxSize)
	}

	return readBytes, nil
}
