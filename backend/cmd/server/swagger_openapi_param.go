package main

// swaggerParamSchemaKeys are Swagger 2.0 fields that live on the parameter
// object but must be nested under schema in OpenAPI 3.
var swaggerParamSchemaKeys = []string{
	"type", "format", "items", "default", "enum",
	"maximum", "exclusiveMaximum", "minimum", "exclusiveMinimum",
	"maxLength", "minLength", "pattern",
	"maxItems", "minItems", "uniqueItems", "multipleOf",
}

func convertParameterList(raw any) any {
	params, ok := raw.([]any)
	if !ok {
		return rewriteRefs(raw)
	}
	out := make([]any, 0, len(params))
	for _, p := range params {
		pm, ok := p.(map[string]any)
		if !ok {
			out = append(out, p)
			continue
		}
		out = append(out, convertNonBodyParameter(pm))
	}
	return out
}

func convertNonBodyParameter(pm map[string]any) map[string]any {
	if _, ok := pm["$ref"]; ok {
		return rewriteRefs(pm).(map[string]any)
	}
	copied := copyMap(pm)
	if _, hasSchema := copied["schema"]; !hasSchema {
		schema := map[string]any{}
		for _, key := range swaggerParamSchemaKeys {
			if v, ok := copied[key]; ok {
				schema[key] = v
				delete(copied, key)
			}
		}
		if len(schema) > 0 {
			copied["schema"] = schema
		}
	}
	return rewriteRefs(copied).(map[string]any)
}
