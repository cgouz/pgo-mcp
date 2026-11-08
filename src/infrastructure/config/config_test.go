package config

import (
	"reflect"
	"testing"
)

func TestDatabaseAllowed(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		allowed  []string
		allowAll bool
		wantErr  bool
	}{
		{name: "empty", allowed: []string{}},
		{name: "wildcard", raw: " * ", allowed: []string{}, allowAll: true},
		{name: "list", raw: "analytics, education,analytics", allowed: []string{"analytics", "education"}},
		{name: "empty entry", raw: "analytics,,education", wantErr: true},
		{name: "mixed wildcard first", raw: "*,analytics", wantErr: true},
		{name: "mixed wildcard last", raw: "analytics,*", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			allowed, allowAll, err := databaseAllowed(test.raw)
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(allowed, test.allowed) {
				t.Fatalf("expected %v, got %v", test.allowed, allowed)
			}
			if allowAll != test.allowAll {
				t.Fatalf("expected allowAll=%v, got %v", test.allowAll, allowAll)
			}
		})
	}
}

func TestDiscoveryConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URI", "postgresql://reader:password@localhost/database")
	t.Setenv("MCP_AUTH_TOKEN", "01234567890123456789012345678901")
	t.Setenv("DATABASE_DISCOVERY", "false")
	t.Setenv("DATABASE_ALLOWED", "analytics")

	if _, err := Load(); err == nil {
		t.Fatal("expected contradictory discovery configuration error")
	}
}
