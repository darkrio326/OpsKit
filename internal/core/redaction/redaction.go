package redaction

import (
	"regexp"
	"slices"
	"strings"
)

const Mask = "******"

var defaultKeys = []string{"password", "token", "secret"}

func DefaultKeys() []string {
	return append([]string{}, defaultKeys...)
}

func RedactText(input string, extraKeys ...string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}
	keys := normalizeKeys(extraKeys)
	if len(keys) == 0 {
		return input
	}
	group := buildKeyGroup(keys)

	out := input
	jsonRe := regexp.MustCompile(`(?i)("(?:` + group + `)"\s*:\s*)"(.*?)"`)
	out = jsonRe.ReplaceAllString(out, `${1}"`+Mask+`"`)

	argEqRe := regexp.MustCompile(`(?i)(--?(?:` + group + `))=([^\s]+)`)
	out = argEqRe.ReplaceAllString(out, `${1}=`+Mask)

	argSpaceRe := regexp.MustCompile(`(?i)(--?(?:` + group + `))\s+([^\s]+)`)
	out = argSpaceRe.ReplaceAllString(out, `${1} `+Mask)

	keyValueRe := regexp.MustCompile(`(?i)\b(` + group + `)\b\s*([:=])\s*([^\s,;]+)`)
	out = keyValueRe.ReplaceAllString(out, `${1}${2}`+Mask)
	return out
}

func RedactArgs(args []string, extraKeys ...string) []string {
	keys := normalizeKeys(extraKeys)
	if len(keys) == 0 {
		return append([]string{}, args...)
	}
	out := append([]string{}, args...)
	for i := 0; i < len(out); i++ {
		arg := out[i]
		if eq := strings.Index(arg, "="); eq > 0 {
			name := arg[:eq]
			if isSensitiveName(name, keys) {
				out[i] = name + "=" + Mask
			}
			continue
		}
		if isSensitiveName(arg, keys) && i+1 < len(out) {
			out[i+1] = Mask
		}
	}
	return out
}

func normalizeKeys(extraKeys []string) []string {
	keys := append([]string{}, defaultKeys...)
	for _, k := range extraKeys {
		k = strings.ToLower(strings.TrimSpace(k))
		if k != "" {
			keys = append(keys, k)
		}
	}
	keys = uniq(keys)
	slices.Sort(keys)
	return keys
}

func uniq(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func buildKeyGroup(keys []string) string {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, regexp.QuoteMeta(k))
	}
	return strings.Join(parts, "|")
}

func isSensitiveName(name string, keys []string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	n = strings.TrimLeft(n, "-")
	if n == "" {
		return false
	}
	if slices.Contains(keys, n) {
		return true
	}
	parts := strings.FieldsFunc(n, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	for _, p := range parts {
		if slices.Contains(keys, p) {
			return true
		}
	}
	return false
}
