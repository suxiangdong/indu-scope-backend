package auth

import "testing"

func TestResolveVerificationRole(t *testing.T) {
	tests := []struct {
		name      string
		current   uint8
		requested uint8
		want      uint8
		wantErr   bool
	}{
		{name: "pending user selects buyer", current: 0, requested: 1, want: 1},
		{name: "pending user selects supplier", current: 0, requested: 2, want: 2},
		{name: "pending user must select", current: 0, requested: 0, wantErr: true},
		{name: "buyer keeps role", current: 1, requested: 0, want: 1},
		{name: "supplier keeps role", current: 2, requested: 2, want: 2},
		{name: "buyer cannot switch", current: 1, requested: 2, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveVerificationRole(tt.current, tt.requested)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveVerificationRole() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("resolveVerificationRole() = %d, want %d", got, tt.want)
			}
		})
	}
}
