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

func TestValidIDNumber(t *testing.T) {
	if !validIDNumber("11010519491231002X") {
		t.Fatal("expected valid Chinese 18-digit identity number")
	}
	if validIDNumber("110105194912310021") {
		t.Fatal("expected invalid identity checksum")
	}
	if validIDNumber("11010519491331002X") {
		t.Fatal("expected invalid identity date")
	}
}
