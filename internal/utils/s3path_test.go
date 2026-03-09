package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseS3Path(t *testing.T) {
	tests := []struct {
		name           string
		s3Path         string
		expectedBucket string
		expectedObject string
		wantErr        bool
	}{
		{
			name:           "valid s3 path with object",
			s3Path:         "s3://mybucket/path/to/file.txt",
			expectedBucket: "mybucket",
			expectedObject: "path/to/file.txt",
			wantErr:        false,
		},
		{
			name:           "valid s3 path bucket only",
			s3Path:         "s3://mybucket",
			expectedBucket: "mybucket",
			expectedObject: "",
			wantErr:        false,
		},
		{
			name:           "valid s3 path with trailing slash",
			s3Path:         "s3://mybucket/",
			expectedBucket: "mybucket",
			expectedObject: "",
			wantErr:        false,
		},
		{
			name:           "invalid path without s3 prefix",
			s3Path:         "mybucket/file.txt",
			expectedBucket: "",
			expectedObject: "",
			wantErr:        true,
		},
		{
			name:           "invalid path empty bucket",
			s3Path:         "s3:///file.txt",
			expectedBucket: "",
			expectedObject: "",
			wantErr:        true,
		},
		{
			name:           "invalid path bucket name too short",
			s3Path:         "s3://ab",
			expectedBucket: "",
			expectedObject: "",
			wantErr:        true,
		},
		{
			name:           "invalid path bucket name too long",
			s3Path:         "s3://" + string(make([]byte, 64)),
			expectedBucket: "",
			expectedObject: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, object, err := ParseS3Path(tt.s3Path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBucket, bucket)
				assert.Equal(t, tt.expectedObject, object)
			}
		})
	}
}

func TestParseS3BucketPath(t *testing.T) {
	tests := []struct {
		name           string
		s3Path         string
		expectedBucket string
		wantErr        bool
	}{
		{
			name:           "valid bucket path",
			s3Path:         "s3://mybucket",
			expectedBucket: "mybucket",
			wantErr:        false,
		},
		{
			name:           "valid bucket path with trailing slash",
			s3Path:         "s3://mybucket/",
			expectedBucket: "mybucket",
			wantErr:        false,
		},
		{
			name:           "invalid path with object",
			s3Path:         "s3://mybucket/object",
			expectedBucket: "",
			wantErr:        true,
		},
		{
			name:           "invalid path without s3 prefix",
			s3Path:         "mybucket",
			expectedBucket: "",
			wantErr:        true,
		},
		{
			name:           "invalid path empty bucket",
			s3Path:         "s3://",
			expectedBucket: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, err := ParseS3BucketPath(tt.s3Path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBucket, bucket)
			}
		})
	}
}

func TestValidateBucketName(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		wantErr    bool
	}{
		{
			name:       "valid bucket name",
			bucketName: "mybucket",
			wantErr:    false,
		},
		{
			name:       "valid bucket name with numbers",
			bucketName: "mybucket123",
			wantErr:    false,
		},
		{
			name:       "bucket name too short",
			bucketName: "ab",
			wantErr:    true,
		},
		{
			name:       "bucket name too long",
			bucketName: string(make([]byte, 64)),
			wantErr:    true,
		},
		{
			name:       "bucket name starts with dash",
			bucketName: "-mybucket",
			wantErr:    true,
		},
		{
			name:       "bucket name ends with dash",
			bucketName: "mybucket-",
			wantErr:    true,
		},
		{
			name:       "valid bucket name with dash in middle",
			bucketName: "my-bucket",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBucketName(tt.bucketName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
