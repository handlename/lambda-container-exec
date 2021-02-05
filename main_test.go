package main

import (
	"fmt"
	"testing"
)

func TestParseS3Path(t *testing.T) {
	tests := []struct {
		inPath    string
		outErr    error
		outBucket string
		outKey    string
	}{
		{
			inPath:    "s3://some-bucket/path/to/obj",
			outErr:    nil,
			outBucket: "some-bucket",
			outKey:    "path/to/obj",
		},
		{
			inPath:    "https://example.com",
			outErr:    fmt.Errorf("not a S3 path"),
			outBucket: "",
			outKey:    "",
		},
		{
			inPath:    "foo bar",
			outErr:    fmt.Errorf("not a S3 path"),
			outBucket: "",
			outKey:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.inPath, func(t *testing.T) {
			bucket, key, err := parseS3Path(test.inPath)

			if err == nil && test.outErr != nil {
				t.Errorf("error: unexpected error '%s'", err)
			}

			if err != nil && err.Error() != test.outErr.Error() {
				t.Errorf("error: got '%s', want '%s'", err, test.outErr)
			}

			if bucket != test.outBucket {
				t.Errorf("Bucket: got '%s', want '%s'", bucket, test.outBucket)
			}

			if key != test.outKey {
				t.Errorf("Key: got '%s', want '%s'", key, test.outKey)
			}
		})
	}
}
