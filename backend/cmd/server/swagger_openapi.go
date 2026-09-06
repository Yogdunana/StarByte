package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func swagger2ToOpenAPI3(raw []byte) ([]byte, error) {
	var spec map[string]any
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("parse swagger 2.0: %w", err)
	}
	out := map[string]any{
		"openapi": "3.0.0",
		"info":    spec["info"],
		"paths":   convertPaths(spec["paths"]),
	}
	if tags, ok := spec["tags"]; ok {
		out["tags"] = tags
	}
	host, _ := spec["host"].(string)
	base, _ := spec["basePath"].(string)
	scheme := "http"
	if schemes, ok := spec["schemes"].([]any); ok && len(schemes) > 0 {
		if s, ok := schemes[0].(string); ok && s != "" {
			scheme = s
		}
	}
	if host != "" {
		url := scheme + "://" + host
		if base != "" {
			url += base
		}
		out["servers"] = []any{map[string]any{"url": url}}
	}
	components := map[string]any{}
	if defs, ok := spec["definitions"]; ok {
		components["schemas"] = rewriteRefs(defs)
	}
	if sec, ok := spec["securityDefinitions"]; ok {
		components["securitySchemes"] = convertSecuritySchemes(sec)
	}
	if len(components) > 0 {
		out["components"] = rewriteRefs(components)
	}
	if sec, ok := spec["security"]; ok {
		out["security"] = sec
	}
	return json.MarshalIndent(out, "", "  ")
}

func convertSecuritySchemes(raw any) any {
	obj, ok := raw.(map[string]any)
	if !ok {
		return raw
	}
	out := map[string]any{}
	for name, v := range obj {
		item, ok := v.(map[string]any)
		if !ok {
			out[name] = v
			continue
		}
		copied := copyMap(item)
		if t, _ := copied["type"].(string); t == "apiKey" {
			out[name] = copied
			continue
		}
		out[name] = copied
	}
	return out
}

func convertPaths(raw any) any {
	paths, ok := raw.(map[string]any)
	if !ok {
		return raw
	}
	out := map[string]any{}
	for path, item := range paths {
		ops, ok := item.(map[string]any)
		if !ok {
			out[path] = item
			continue
		}
		converted := map[string]any{}
		for method, op := range ops {
			if method == "parameters" {
				converted[method] = rewriteRefs(op)
				continue
			}
			converted[method] = convertOperation(op)
		}
		out[path] = converted
	}
	return out
}

func convertOperation(raw any) any {
	op, ok := raw.(map[string]any)
	if !ok {
		return raw
	}
	out := copyMap(op)
	consumes := stringSlice(out["consumes"])
	produces := stringSlice(out["produces"])
	delete(out, "consumes")
	delete(out, "produces")
	if len(produces) == 0 {
		produces = []string{"application/json"}
	}
	params, _ := out["parameters"].([]any)
	kept := make([]any, 0, len(params))
	var bodyParams []map[string]any
	var formParams []map[string]any
	for _, p := range params {
		pm, ok := p.(map[string]any)
		if !ok {
			kept = append(kept, p)
			continue
		}
		in, _ := pm["in"].(string)
		switch in {
		case "body":
			bodyParams = append(bodyParams, pm)
		case "formData":
			formParams = append(formParams, pm)
		default:
			kept = append(kept, rewriteRefs(pm))
		}
	}
	if len(kept) > 0 {
		out["parameters"] = kept
	} else {
		delete(out, "parameters")
	}
	if rb := buildRequestBody(bodyParams, formParams, consumes); rb != nil {
		out["requestBody"] = rb
	}
	if resp, ok := out["responses"].(map[string]any); ok {
		out["responses"] = convertResponses(resp, produces)
	}
	return rewriteRefs(out)
}

func buildRequestBody(body, form []map[string]any, consumes []string) map[string]any {
	if len(form) > 0 {
		props := map[string]any{}
		required := []any{}
		for _, p := range form {
			name, _ := p["name"].(string)
			if name == "" {
				continue
			}
			schema := map[string]any{"type": p["type"]}
			if p["type"] == "file" {
				schema["type"] = "string"
				schema["format"] = "binary"
			}
			if desc, ok := p["description"]; ok {
				schema["description"] = desc
			}
			props[name] = schema
			if req, _ := p["required"].(bool); req {
				required = append(required, name)
			}
		}
		ct := "multipart/form-data"
		if len(consumes) > 0 {
			ct = consumes[0]
		}
		schema := map[string]any{"type": "object", "properties": props}
		if len(required) > 0 {
			schema["required"] = required
		}
		return map[string]any{
			"required": true,
			"content": map[string]any{
				ct: map[string]any{"schema": schema},
			},
		}
	}
	if len(body) == 0 {
		return nil
	}
	p := body[0]
	ct := "application/json"
	if len(consumes) > 0 {
		ct = consumes[0]
	}
	schema := p["schema"]
	if schema == nil {
		schema = map[string]any{"type": "object"}
	}
	rb := map[string]any{
		"content": map[string]any{
			ct: map[string]any{"schema": rewriteRefs(schema)},
		},
	}
	if req, _ := p["required"].(bool); req {
		rb["required"] = true
	}
	if desc, ok := p["description"]; ok {
		rb["description"] = desc
	}
	return rb
}

func convertResponses(resp map[string]any, produces []string) map[string]any {
	out := map[string]any{}
	for code, raw := range resp {
		item, ok := raw.(map[string]any)
		if !ok {
			out[code] = raw
			continue
		}
		copied := copyMap(item)
		schema := copied["schema"]
		delete(copied, "schema")
		if schema != nil {
			content := map[string]any{}
			for _, ct := range produces {
				content[ct] = map[string]any{"schema": rewriteRefs(schema)}
			}
			copied["content"] = content
		}
		out[code] = copied
	}
	return out
}

func rewriteRefs(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			if k == "$ref" {
				if s, ok := val.(string); ok {
					out[k] = strings.Replace(s, "#/definitions/", "#/components/schemas/", 1)
					continue
				}
			}
			out[k] = rewriteRefs(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = rewriteRefs(item)
		}
		return out
	default:
		return v
	}
}

func copyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func stringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
