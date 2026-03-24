package output

// Formatter formats data for output.
type Formatter interface {
	Format(data any) (string, error)
}

// Column defines a table column.
type Column struct {
	Header string
	Width  int
}

// NewFormatter creates the appropriate formatter based on flags.
func NewFormatter(jsonMode bool, format string, noColor bool) Formatter {
	if jsonMode {
		return &JSONFormatter{}
	}
	if format != "" {
		return &TemplateFormatter{Template: format}
	}
	return nil
}
