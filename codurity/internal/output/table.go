package output

import (
	"fmt"
	"io"
	"strings"
)

func PrintTable(w io.Writer, headers []string, rows [][]string) {
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	format := ""
	for i, w := range colWidths {
		if i > 0 {
			format += "  "
		}
		format += fmt.Sprintf("%%-%ds", w)
	}
	format += "\n"

	fmt.Fprintf(w, format, toAny(headers)...)
	for _, row := range rows {
		fmt.Fprintf(w, format, toAny(row)...)
	}
}

func PrintSeparator(w io.Writer, widths []int) {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("-", w)
	}
	fmt.Fprintf(w, "%s\n", strings.Join(parts, "  "))
}

func toAny(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}
