package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Returns false if the path is a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Layouts:
// Date "2006-01-02" Hour "15:04"
func getTimeFromStrings(date string, hourStr string) (time.Time, error) {
	var timeSb strings.Builder
	if _, err := timeSb.WriteString(date); err != nil {
		return time.Time{}, fmt.Errorf("string builder write failed: %w",
			err)
	}
	if _, err := timeSb.
		WriteString(" "); err != nil {
		return time.Time{}, fmt.Errorf("string builder write failed: %w",
			err)
	}
	if _, err := timeSb.
		WriteString(hourStr); err != nil {
		return time.Time{}, fmt.Errorf("string builder write failed: %w",
			err)
	}

	layout := "2006-01-02 15:04"
	t, err := time.ParseInLocation(layout, timeSb.String(), time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date failed: %w", err)
	}

	return t, nil
}

// From https://stackoverflow.com/a/46767098
type backScanner struct {
	r   io.ReaderAt
	pos int
	err error
	buf []byte
}

func NewScanner(r io.ReaderAt, pos int) *backScanner {
	return &backScanner{r: r, pos: pos}
}

func (s *backScanner) readMore() {
	if s.pos == 0 {
		s.err = io.EOF
		return
	}

	size := 1024
	if size > s.pos {
		size = s.pos
	}
	s.pos -= size
	buf2 := make([]byte, size, size+len(s.buf))

	// ReadAt attemps to read full buf!
	_, s.err = s.r.ReadAt(buf2, int64(s.pos))
	if s.err == nil {
		s.buf = append(buf2, s.buf...)
	}
}

func (s *backScanner) Line() (line string, start int, err error) {
	if s.err != nil {
		return "", 0, s.err
	}
	for {
		lineStart := bytes.LastIndexByte(s.buf, '\n')
		if lineStart >= 0 {
			// We have a complete line:
			var line string
			line, s.buf = string(dropCR(s.buf[lineStart+1:])), s.buf[:lineStart]
			return line, s.pos + lineStart + 1, nil
		}
		// Need more data:
		s.readMore()
		if s.err != nil {
			if s.err == io.EOF {
				if len(s.buf) > 0 {
					return string(dropCR(s.buf)), 0, nil
				}
			}
			return "", 0, s.err
		}
	}
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
