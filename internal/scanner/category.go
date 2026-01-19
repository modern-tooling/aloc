package scanner

// LanguageCategory represents a logical grouping of programming languages
type LanguageCategory string

const (
	CategoryPrimary LanguageCategory = "Primary"       // general-purpose programming languages
	CategoryWeb     LanguageCategory = "Web"           // web frontend languages
	CategoryData    LanguageCategory = "Data & Config" // data serialization and config
	CategoryMarkup  LanguageCategory = "Documentation" // markup and documentation
	CategoryInfra   LanguageCategory = "DevOps & IaC"  // infrastructure and build tools
	CategoryOther   LanguageCategory = "Other"         // uncategorized
)

// CategoryOrder defines display order for categories
var CategoryOrder = []LanguageCategory{
	CategoryPrimary,
	CategoryWeb,
	CategoryInfra,
	CategoryData,
	CategoryMarkup,
	CategoryOther,
}

// langToCategory maps language names to their categories
var langToCategory = map[string]LanguageCategory{
	// Primary - general purpose programming languages
	"Go":               CategoryPrimary,
	"TypeScript":       CategoryPrimary,
	"JavaScript":       CategoryPrimary,
	"Python":           CategoryPrimary,
	"Rust":             CategoryPrimary,
	"Java":             CategoryPrimary,
	"Kotlin":           CategoryPrimary,
	"Swift":            CategoryPrimary,
	"C":                CategoryPrimary,
	"C++":              CategoryPrimary,
	"C#":               CategoryPrimary,
	"Ruby":             CategoryPrimary,
	"PHP":              CategoryPrimary,
	"Scala":            CategoryPrimary,
	"Haskell":          CategoryPrimary,
	"Elixir":           CategoryPrimary,
	"Erlang":           CategoryPrimary,
	"Clojure":          CategoryPrimary,
	"Dart":             CategoryPrimary,
	"Groovy":           CategoryPrimary,
	"Lua":              CategoryPrimary,
	"Perl":             CategoryPrimary,
	"R":                CategoryPrimary,
	"Shell":            CategoryPrimary,

	// Web - frontend and styling
	"HTML":             CategoryWeb,
	"CSS":              CategoryWeb,
	"SCSS":             CategoryWeb,
	"SASS":             CategoryWeb,
	"LESS":             CategoryWeb,
	"Vue":              CategoryWeb,
	"Svelte":           CategoryWeb,
	"MDX":              CategoryWeb,

	// Infrastructure - DevOps and IaC
	"Terraform":        CategoryInfra,
	"Makefile":         CategoryInfra,
	"Just":             CategoryInfra,
	"HCL":              CategoryInfra,
	"Dockerfile":       CategoryInfra,
	"Protocol Buffers": CategoryInfra,

	// Data & Config - serialization and configuration
	"JSON":             CategoryData,
	"YAML":             CategoryData,
	"TOML":             CategoryData,
	"XML":              CategoryData,
	"SQL":              CategoryData,

	// Documentation - markup languages
	"Markdown":         CategoryMarkup,
	"Plain Text":       CategoryMarkup,
}

// GetLanguageCategory returns the category for a language
func GetLanguageCategory(language string) LanguageCategory {
	if cat, ok := langToCategory[language]; ok {
		return cat
	}
	return CategoryOther
}

// GetCategoryDisplayOrder returns the display order index for a category
func GetCategoryDisplayOrder(cat LanguageCategory) int {
	for i, c := range CategoryOrder {
		if c == cat {
			return i
		}
	}
	return len(CategoryOrder) // put unknown at end
}
