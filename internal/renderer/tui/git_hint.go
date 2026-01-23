package tui

import (
	"fmt"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RenderGitHint renders the marginal hint when --git is not specified
func RenderGitHint(hint *model.GitHint, theme *renderer.Theme) string {
	if hint == nil || !hint.HasGit {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(theme.Dim.Render("Git history detected · run `aloc --git` for churn & stability signals"))
	sb.WriteString("\n")

	// optional safe signal (repo age and last commit)
	if hint.RepoAge != "" || hint.LastCommit != "" {
		parts := []string{}
		if hint.RepoAge != "" {
			parts = append(parts, fmt.Sprintf("Repo age: %s", hint.RepoAge))
		}
		if hint.LastCommit != "" {
			parts = append(parts, fmt.Sprintf("last commit: %s", hint.LastCommit))
		}
		sb.WriteString(theme.Dim.Render(strings.Join(parts, " · ")))
		sb.WriteString("\n")
	}

	return sb.String()
}
