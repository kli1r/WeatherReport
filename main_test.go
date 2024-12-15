package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseLocationCoords(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		wantErr bool
	}{
		{
			name:    "more args",
			args:    "12.51, 3467.32, 623.123",
			wantErr: true,
		}, {
			name:    "less args",
			args:    "12.51",
			wantErr: true,
		}, {
			name:    "another separator",
			args:    "345.12-210",
			wantErr: true,
		}, {
			name:    "not float",
			args:    "мытипа,цифры",
			wantErr: true,
		}, {
			name:    "not in the range",
			args:    "128.11,-301",
			wantErr: true,
		}, {
			name:    "must be fine",
			args:    "39.01,-154.48",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseLocationCoords(tt.args)
			if tt.wantErr {
				assert.NotNil(t, err, "Должна быть ошибка!")
			} else {
				assert.NoError(t, err, "Ошибки быть не должно! %v", err)
			}
		})
	}
}
