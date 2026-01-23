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

// categoryFromString converts a JSON category string to LanguageCategory
var categoryFromString = map[string]LanguageCategory{
	"primary": CategoryPrimary,
	"web":     CategoryWeb,
	"infra":   CategoryInfra,
	"data":    CategoryData,
	"docs":    CategoryMarkup,
	"other":   CategoryOther,
}

// GetLanguageCategory returns the category for a language
func GetLanguageCategory(language string) LanguageCategory {
	if cfg, ok := GetLanguageConfig(language); ok && cfg.Category != "" {
		if cat, ok := categoryFromString[cfg.Category]; ok {
			return cat
		}
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
