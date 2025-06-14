package utils

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ArchiveSuite struct {
	suite.Suite
}

type zipInMemoryTestCase struct {
	name          string
	fileName      string
	content       string
	expectedError bool
}

func (s *ArchiveSuite) TestZipInMemory() {
	// Read the large random text file
	randomContent, err := os.ReadFile(filepath.Join("__testdata", "large_random.txt"))
	s.NoError(err)

	cases := []zipInMemoryTestCase{
		{
			name:          "test zip in memory",
			fileName:      "index.js",
			content:       "console.log('Hello, World!');",
			expectedError: false,
		},
		{
			name:          "test zip in memory with large random file",
			fileName:      "large_random.txt",
			content:       string(randomContent),
			expectedError: true,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			zipB64Encoded, err := ZipInMemory(tc.fileName, tc.content)
			if tc.expectedError {
				s.Error(err)
			} else {
				s.NoError(err)
				s.NotEmpty(zipB64Encoded)

				zipBytes, err := base64.StdEncoding.DecodeString(zipB64Encoded)
				s.NoError(err)

				// Verify that the output is a valid base64-encoded zip file.
				zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
				s.NoError(err)
				s.Equal(1, len(zipReader.File))
				s.Equal(tc.fileName, zipReader.File[0].Name)
				readCloser, err := zipReader.File[0].Open()
				s.NoError(err)
				content, err := io.ReadAll(readCloser)
				s.NoError(err)
				s.Equal(tc.content, string(content))
			}
		})
	}
}

func TestArchiveSuite(t *testing.T) {
	suite.Run(t, new(ArchiveSuite))
}
