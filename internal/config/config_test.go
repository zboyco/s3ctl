package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestS3ConfigItemValidate(t *testing.T) {
	tests := []struct {
		name    string
		item    S3ConfigItem
		wantErr bool
	}{
		{
			name: "valid endpoint without scheme",
			item: S3ConfigItem{
				Endpoint:        "play.min.io",
				AccessKeyID:     "THISISKEYID",
				SecretAccessKey: "THISISSECRETKEY",
			},
			wantErr: false,
		},
		{
			name: "valid endpoint with port and timeout",
			item: S3ConfigItem{
				Endpoint:        "localhost:9000",
				AccessKeyID:     "THISISKEYID",
				SecretAccessKey: "THISISSECRETKEY",
				Timeout:         30,
			},
			wantErr: false,
		},
		{
			name: "endpoint with scheme is invalid",
			item: S3ConfigItem{
				Endpoint:        "https://play.min.io",
				AccessKeyID:     "THISISKEYID",
				SecretAccessKey: "THISISSECRETKEY",
			},
			wantErr: true,
		},
		{
			name: "endpoint with path is invalid",
			item: S3ConfigItem{
				Endpoint:        "play.min.io/path",
				AccessKeyID:     "THISISKEYID",
				SecretAccessKey: "THISISSECRETKEY",
			},
			wantErr: true,
		},
		{
			name: "timeout omitted is allowed",
			item: S3ConfigItem{
				Endpoint:        "s3.example.com",
				AccessKeyID:     "THISISKEYID",
				SecretAccessKey: "THISISSECRETKEY",
				Timeout:         0,
			},
			wantErr: false,
		},
		{
			name: "timeout below range is invalid",
			item: S3ConfigItem{
				Endpoint:        "s3.example.com",
				AccessKeyID:     "THISISKEYID",
				SecretAccessKey: "THISISSECRETKEY",
				Timeout:         -1,
			},
			wantErr: true,
		},
		{
			name: "timeout above range is invalid",
			item: S3ConfigItem{
				Endpoint:        "s3.example.com",
				AccessKeyID:     "THISISKEYID",
				SecretAccessKey: "THISISSECRETKEY",
				Timeout:         301,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
