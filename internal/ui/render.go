package ui

import (
	"fmt"
	"html"
	"strings"
)

// AutoRender turns data into HTML automatically.
// Lists → tables, maps → cards, strings → text blocks.
func AutoRender(data interface{}) string {
	var b strings.Builder
	b.WriteString(pageHeader(""))
	b.WriteString(`<div class="container">`)

	switch v := data.(type) {
	case []interface{}:
		renderTable(&b, v)
	case map[string]interface{}:
		renderCard(&b, v)
	case string:
		b.WriteString(fmt.Sprintf(`<div class="text-block">%s</div>`, html.EscapeString(v)))
	default:
		b.WriteString(fmt.Sprintf(`<div class="text-block">%v</div>`, v))
	}

	b.WriteString(`</div>`)
	b.WriteString(pageFooter())
	return b.String()
}

func renderTable(b *strings.Builder, items []interface{}) {
	if len(items) == 0 {
		b.WriteString(`<div class="text-block" style="color:#555">No items yet.</div>`)
		return
	}

	// Collect all keys from all items
	keys := collectKeys(items)
	if len(keys) == 0 {
		// Not a list of maps — render as simple list
		b.WriteString(`<table><tbody>`)
		for _, item := range items {
			b.WriteString(fmt.Sprintf(`<tr><td>%s</td></tr>`, html.EscapeString(stringify(item))))
		}
		b.WriteString(`</tbody></table>`)
		return
	}

	b.WriteString(`<table><thead><tr>`)
	for _, k := range keys {
		b.WriteString(fmt.Sprintf(`<th>%s</th>`, html.EscapeString(k)))
	}
	b.WriteString(`</tr></thead><tbody>`)

	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		b.WriteString(`<tr>`)
		for _, k := range keys {
			b.WriteString(fmt.Sprintf(`<td>%s</td>`, html.EscapeString(stringify(m[k]))))
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table>`)
}

func renderCard(b *strings.Builder, m map[string]interface{}) {
	b.WriteString(`<div class="card">`)
	for k, v := range m {
		b.WriteString(fmt.Sprintf(
			`<div class="card-row"><span class="card-key">%s</span><span class="card-value">%s</span></div>`,
			html.EscapeString(k),
			html.EscapeString(stringify(v)),
		))
	}
	b.WriteString(`</div>`)
}

func collectKeys(items []interface{}) []string {
	seen := make(map[string]bool)
	var keys []string
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			return nil
		}
		for k := range m {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	return keys
}

func stringify(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func pageHeader(title string) string {
	t := "Drop"
	if title != "" {
		t = html.EscapeString(title)
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<style>%s</style>
</head>
<body>
`, t, DefaultCSS)
}

func pageFooter() string {
	return `
</body>
</html>`
}

// RenderPage builds HTML from a page title and pre-rendered body content.
func RenderPage(title string, body string) string {
	var b strings.Builder
	b.WriteString(pageHeader(title))
	b.WriteString(`<div class="container">`)
	if title != "" {
		b.WriteString(fmt.Sprintf(`<h1>%s</h1>`, html.EscapeString(title)))
	}
	b.WriteString(body)
	b.WriteString(`</div>`)
	b.WriteString(pageFooter())
	return b.String()
}
