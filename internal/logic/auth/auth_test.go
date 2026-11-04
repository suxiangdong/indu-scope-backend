package auth

import "testing"

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		password string
		wantErr  bool
	}{
		{name: "valid", phone: "13800138000", password: "Password123!"},
		{name: "short phone", phone: "123", password: "Password123!", wantErr: true},
		{name: "non numeric phone", phone: "1380013800a", password: "Password123!", wantErr: true},
		{name: "short password", phone: "13800138000", password: "short", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCredentials(tt.phone, tt.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
