package httpapi

import "github.com/microcosm-cc/bluemonday"

var richTextPolicy = bluemonday.UGCPolicy()

// Sanitize on reads as well as new saves so previously stored HTML is safe in
// both the public pages and the admin contentEditable editors.
func sanitizeContent(value any) {
	switch value := value.(type) {
	case map[string]any:
		_, hasSlug := value["slug"]
		_, hasHeading := value["h1"]
		for key, child := range value {
			rich := key == "body" || key == "intro" || key == "content" || key == "devAbout" || (key == "description" && (hasSlug || hasHeading))
			if text, ok := child.(string); ok && rich {
				value[key] = richTextPolicy.Sanitize(text)
			} else {
				sanitizeContent(child)
			}
		}
	case []any:
		for _, child := range value {
			sanitizeContent(child)
		}
	case []map[string]any:
		for _, child := range value {
			sanitizeContent(child)
		}
	}
}

// A scheduled article becomes visible on its date without requiring a worker
// to change its persisted editorial status. Undated scheduled drafts stay hidden.
func publicBlogSQL(alias string) string {
	if alias != "" {
		alias += "."
	}
	return "((" + alias + "status = 'Published' and (" + alias + "publish_date is null or " + alias + "publish_date <= current_date)) or (" + alias + "status = 'Scheduled' and " + alias + "publish_date <= current_date))"
}
