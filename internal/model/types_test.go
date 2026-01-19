package model

import (
	"encoding/json"
	"testing"
)

func TestRoleString(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		{RoleProd, "prod"},
		{RoleTest, "test"},
		{RoleInfra, "infra"},
		{RoleDocs, "docs"},
		{RoleConfig, "config"},
		{RoleGenerated, "generated"},
		{RoleVendor, "vendor"},
		{RoleScripts, "scripts"},
		{RoleExamples, "examples"},
		{RoleDeprecated, "deprecated"},
	}

	for _, tt := range tests {
		if got := tt.role.String(); got != tt.want {
			t.Errorf("Role.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestRoleColor(t *testing.T) {
	tests := []struct {
		role Role
		want SemanticColor
	}{
		{RoleProd, ColorPrimary},
		{RoleTest, ColorSafety},
		{RoleInfra, ColorOperational},
		{RoleDocs, ColorKnowledge},
		{RoleConfig, ColorFragility},
		{RoleGenerated, ColorLowEmphasis},
		{RoleVendor, ColorExternal},
		{RoleScripts, ColorPrimary},
		{RoleExamples, ColorKnowledge},
		{RoleDeprecated, ColorWarning},
	}

	for _, tt := range tests {
		if got := tt.role.Color(); got != tt.want {
			t.Errorf("Role(%s).Color() = %v, want %v", tt.role, got, tt.want)
		}
	}
}

func TestRoleColorDefaultCase(t *testing.T) {
	// test unknown role returns default color
	unknownRole := Role("unknown")
	if got := unknownRole.Color(); got != ColorPrimary {
		t.Errorf("unknown Role.Color() = %v, want %v", got, ColorPrimary)
	}
}

func TestRoleJSONSerialization(t *testing.T) {
	role := RoleTest
	data, err := json.Marshal(role)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	if string(data) != `"test"` {
		t.Errorf("json.Marshal = %s, want %s", data, `"test"`)
	}

	var decoded Role
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded != role {
		t.Errorf("decoded = %v, want %v", decoded, role)
	}
}

func TestAllRolesComplete(t *testing.T) {
	if len(AllRoles) != 10 {
		t.Errorf("AllRoles has %d roles, want 10", len(AllRoles))
	}
}

func TestAllSignalsComplete(t *testing.T) {
	if len(AllSignals) != 6 {
		t.Errorf("AllSignals has %d signals, want 6", len(AllSignals))
	}
}

func TestAllTestKindsComplete(t *testing.T) {
	if len(AllTestKinds) != 5 {
		t.Errorf("AllTestKinds has %d kinds, want 5", len(AllTestKinds))
	}
}

func TestTestKindString(t *testing.T) {
	tests := []struct {
		kind TestKind
		want string
	}{
		{TestUnit, "unit"},
		{TestIntegration, "integration"},
		{TestE2E, "e2e"},
		{TestContract, "contract"},
		{TestFixture, "fixture"},
	}

	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("TestKind.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestSignalString(t *testing.T) {
	tests := []struct {
		signal Signal
		want   string
	}{
		{SignalPath, "path"},
		{SignalFilename, "filename"},
		{SignalExtension, "extension"},
		{SignalNeighborhood, "neighborhood"},
		{SignalHeader, "header"},
		{SignalOverride, "override"},
	}

	for _, tt := range tests {
		if got := tt.signal.String(); got != tt.want {
			t.Errorf("Signal.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestSemanticColorValues(t *testing.T) {
	colors := []SemanticColor{
		ColorPrimary,
		ColorSafety,
		ColorOperational,
		ColorKnowledge,
		ColorFragility,
		ColorLowEmphasis,
		ColorExternal,
		ColorWarning,
	}

	// verify all colors have expected prefix
	for _, c := range colors {
		if len(c) == 0 {
			t.Error("SemanticColor should not be empty")
		}
	}

	// verify unique values
	seen := make(map[SemanticColor]bool)
	for _, c := range colors {
		if seen[c] {
			t.Errorf("duplicate SemanticColor: %v", c)
		}
		seen[c] = true
	}
}
