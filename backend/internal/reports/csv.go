// CSV export with spreadsheet formula-injection mitigation (T130).
// Text cells beginning with = + - @ or containing control characters are
// prefixed with a single quote so Excel/LibreOffice treat them as literals.
package reports

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"
)

// csvSafe sanitizes a cell for spreadsheet export.
func csvSafe(v any) string {
	var s string
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		s = x
	case time.Time:
		s = x.UTC().Format(time.RFC3339)
	case []byte:
		s = string(x)
	default:
		s = fmt.Sprintf("%v", v)
	}
	if s == "" {
		return s
	}
	// Formula injection: a leading = + - @ (or tab/CR smuggled first) turns
	// the cell into a spreadsheet formula. Prefix with ' to force text.
	c := s[0]
	if c == '=' || c == '+' || c == '-' || c == '@' || c == '\t' || c == '\r' || c == '\n' {
		return "'" + s
	}
	return s
}

// WriteCSV streams the report to w with UTF-8 BOM for Excel compatibility.
func WriteCSV(w io.Writer, r *Report) error {
	if _, err := w.Write([]byte("\xEF\xBB\xBF")); err != nil {
		return err
	}
	cw := csv.NewWriter(w)

	header := make([]string, len(r.Columns))
	for i, c := range r.Columns {
		header[i] = c.Label
	}
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, row := range r.Rows {
		rec := make([]string, len(r.Columns))
		for i, c := range r.Columns {
			rec[i] = csvSafe(row[c.Key])
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// safeFilename builds a download filename from the report + filter range.
func Filename(name string, f Filters) string {
	from, to := f.From, f.To
	if from == "" {
		from = "all"
	}
	if to == "" {
		to = "all"
	}
	return fmt.Sprintf("ojt-%s-%s-to-%s.csv", name, from, to)
}
