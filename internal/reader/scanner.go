package reader

import (
	"bufio"
	"context"
	"io"
)

type ScannerReader struct {
	scanner *bufio.Scanner
}

func NewScannerReader(scanner *bufio.Scanner) *ScannerReader {
	return &ScannerReader{scanner: scanner}
}

func (r *ScannerReader) ReadLine(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return r.scanner.Text(), nil
}
