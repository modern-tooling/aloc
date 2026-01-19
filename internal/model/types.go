package model

// Role represents the semantic role of a file in the codebase
type Role string

const (
	RoleProd       Role = "prod"
	RoleTest       Role = "test"
	RoleInfra      Role = "infra"
	RoleDocs       Role = "docs"
	RoleConfig     Role = "config"
	RoleGenerated  Role = "generated"
	RoleVendor     Role = "vendor"
	RoleScripts    Role = "scripts"
	RoleExamples   Role = "examples"
	RoleDeprecated Role = "deprecated"
)

// AllRoles contains all possible roles for iteration
var AllRoles = []Role{
	RoleProd,
	RoleTest,
	RoleInfra,
	RoleDocs,
	RoleConfig,
	RoleGenerated,
	RoleVendor,
	RoleScripts,
	RoleExamples,
	RoleDeprecated,
}

// TestKind represents the type of test
type TestKind string

const (
	TestUnit        TestKind = "unit"
	TestIntegration TestKind = "integration"
	TestE2E         TestKind = "e2e"
	TestContract    TestKind = "contract"
	TestFixture     TestKind = "fixture"
)

// AllTestKinds contains all possible test kinds
var AllTestKinds = []TestKind{
	TestUnit,
	TestIntegration,
	TestE2E,
	TestContract,
	TestFixture,
}

// Signal represents the source of classification evidence
type Signal string

const (
	SignalPath         Signal = "path"
	SignalFilename     Signal = "filename"
	SignalExtension    Signal = "extension"
	SignalNeighborhood Signal = "neighborhood"
	SignalHeader       Signal = "header"
	SignalOverride     Signal = "override"
)

// AllSignals contains all possible signals
var AllSignals = []Signal{
	SignalPath,
	SignalFilename,
	SignalExtension,
	SignalNeighborhood,
	SignalHeader,
	SignalOverride,
}

// SemanticColor represents a semantic color token for rendering
type SemanticColor string

const (
	ColorPrimary     SemanticColor = "semantic.primary"
	ColorSafety      SemanticColor = "semantic.safety"
	ColorOperational SemanticColor = "semantic.operational"
	ColorKnowledge   SemanticColor = "semantic.knowledge"
	ColorFragility   SemanticColor = "semantic.fragility"
	ColorLowEmphasis SemanticColor = "semantic.low_emphasis"
	ColorExternal    SemanticColor = "semantic.external"
	ColorWarning     SemanticColor = "semantic.warning"
)

// Color returns the semantic color for a role
func (r Role) Color() SemanticColor {
	switch r {
	case RoleProd:
		return ColorPrimary
	case RoleTest:
		return ColorSafety
	case RoleInfra:
		return ColorOperational
	case RoleDocs:
		return ColorKnowledge
	case RoleConfig:
		return ColorFragility
	case RoleGenerated:
		return ColorLowEmphasis
	case RoleVendor:
		return ColorExternal
	case RoleScripts:
		return ColorPrimary
	case RoleExamples:
		return ColorKnowledge
	case RoleDeprecated:
		return ColorWarning
	default:
		return ColorPrimary
	}
}

// String returns the string representation of a role
func (r Role) String() string {
	return string(r)
}

// String returns the string representation of a test kind
func (t TestKind) String() string {
	return string(t)
}

// String returns the string representation of a signal
func (s Signal) String() string {
	return string(s)
}
