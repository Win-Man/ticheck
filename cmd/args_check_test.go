package cmd

import (
	"testing"

	"github.com/Win-Man/ticheck/config"
)

// TestValueEqual covers the config/variable comparison used by args-check.
// The happy paths mirror the notes in checkConfigbyComponent:
// off = OFF, 24h0m0s = 24h, 30MiB = 30MB, 40GiB = 40GB.
func TestValueEqual(t *testing.T) {
	tests := []struct {
		name  string
		act   string
		exp   string
		equal bool
	}{
		// plain strings
		{"identical strings", "true", "true", true},
		{"case insensitive", "TRUE", "true", true},
		{"on vs off", "ON", "OFF", false},
		{"off vs OFF", "off", "OFF", true},
		{"empty strings", "", "", true},
		{"different values", "1", "2", false},

		// durations
		{"duration same value different format", "24h0m0s", "24h", true},
		{"duration identical", "3s", "3s", true},
		{"duration different", "3s", "4s", false},
		{"duration different length", "24h15m", "24h0m0s", false},
		{"zero duration", "0", "0", true},

		// byte sizes: "i" and trailing "b" are ignored by areEquivalentUnits
		{"MiB equals MB", "30MiB", "30MB", true},
		{"GiB equals GB", "40GiB", "40GB", true},
		{"KiB equals KB", "128KiB", "128KB", true},
		{"same unit same value", "3867MB", "3867MB", true},
		{"same value different magnitude", "30MB", "30GB", false},
		{"different value same unit", "3867MB", "4000MB", false},

		// known limitation: units are compared textually, not converted
		{"1536MB vs 1.5GB not converted", "1536MB", "1.5GB", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valueEqual(tt.act, tt.exp); got != tt.equal {
				t.Errorf("valueEqual(%q, %q) = %v, want %v", tt.act, tt.exp, got, tt.equal)
			}
			// comparison must be symmetric
			if got := valueEqual(tt.exp, tt.act); got != tt.equal {
				t.Errorf("valueEqual(%q, %q) = %v, want %v (symmetry)", tt.exp, tt.act, got, tt.equal)
			}
		})
	}
}

func TestHasUnit(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"10G", true},
		{"30MiB", true},
		{"10", false},
		{"", false},
		{"off", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := hasUnit(tt.value); got != tt.want {
				t.Errorf("hasUnit(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestParseValueAndUnit(t *testing.T) {
	tests := []struct {
		value    string
		wantNum  float64
		wantUnit string
	}{
		{"30MiB", 30, "MiB"},
		{"1.5GB", 1.5, "GB"},
		{"3867MB", 3867, "MB"},
		{"10", 10, ""},
		{"  7GiB  ", 7, "GiB"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			num, unit := parseValueAndUnit(tt.value)
			if num != tt.wantNum || unit != tt.wantUnit {
				t.Errorf("parseValueAndUnit(%q) = (%v, %q), want (%v, %q)",
					tt.value, num, unit, tt.wantNum, tt.wantUnit)
			}
		})
	}
}

func TestAreEquivalentUnits(t *testing.T) {
	tests := []struct {
		unit1 string
		unit2 string
		want  bool
	}{
		{"GiB", "GB", true},
		{"MiB", "MB", true},
		{"KB", "KiB", true},
		{"GB", "MB", false},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.unit1+"_"+tt.unit2, func(t *testing.T) {
			if got := areEquivalentUnits(tt.unit1, tt.unit2); got != tt.want {
				t.Errorf("areEquivalentUnits(%q, %q) = %v, want %v", tt.unit1, tt.unit2, got, tt.want)
			}
		})
	}
}

func TestParseInstanceAddress(t *testing.T) {
	tests := []struct {
		name     string
		instance string
		wantHost string
		wantPort int
		wantErr  bool
	}{
		{"ipv4", "10.0.0.1:4000", "10.0.0.1", 4000, false},
		{"hostname", "tidb-0.tidb-peer:4000", "tidb-0.tidb-peer", 4000, false},
		{"ipv6", "[::1]:4000", "::1", 4000, false},
		{"surrounding spaces", " 10.0.0.1:4000 ", "10.0.0.1", 4000, false},
		{"missing port", "10.0.0.1", "", 0, true},
		{"non numeric port", "10.0.0.1:abc", "", 0, true},
		{"empty", "", "", 0, true},
		{"port out of range", "10.0.0.1:70000", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, err := parseInstanceAddress(tt.instance)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseInstanceAddress(%q) expected error, got host=%q port=%d", tt.instance, host, port)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseInstanceAddress(%q) unexpected error: %v", tt.instance, err)
			}
			if host != tt.wantHost || port != tt.wantPort {
				t.Errorf("parseInstanceAddress(%q) = (%q, %d), want (%q, %d)",
					tt.instance, host, port, tt.wantHost, tt.wantPort)
			}
		})
	}
}

// TestOpenTiDBInstancesSkipsInvalidRecords makes sure a malformed cluster_info
// row is skipped instead of crashing the whole args-check run.
func TestOpenTiDBInstancesSkipsInvalidRecords(t *testing.T) {
	records := [][]string{
		{},
		{"not-a-host-port"},
		{"10.0.0.1:not-a-port"},
	}
	if got := openTiDBInstances(records, config.DBConfig{}); len(got) != 0 {
		t.Fatalf("openTiDBInstances() returned %d connections, want 0", len(got))
	}
}
