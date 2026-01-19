package json

import (
	"encoding/json"
	"io"

	"github.com/rgehrsern/aloc/internal/model"
	"github.com/rgehrsern/aloc/internal/renderer"
)

type JSONRenderer struct {
	writer io.Writer
	pretty bool
}

func NewJSONRenderer(opts renderer.Options) *JSONRenderer {
	return &JSONRenderer{
		writer: opts.Writer,
		pretty: opts.Pretty,
	}
}

func (r *JSONRenderer) Render(report *model.Report) error {
	enc := json.NewEncoder(r.writer)
	if r.pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(report)
}
