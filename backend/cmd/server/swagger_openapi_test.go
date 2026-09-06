package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSwagger2ToOpenAPI3(t *testing.T) {
	src := []byte(`{
		"swagger":"2.0",
		"info":{"title":"StarByte API","version":"1.0"},
		"host":"localhost:8080",
		"basePath":"/api/v1",
		"paths":{
			"/auth/login":{
				"post":{
					"consumes":["application/json"],
					"produces":["application/json"],
					"parameters":[{"in":"body","name":"body","required":true,"schema":{"$ref":"#/definitions/dto.LoginRequest"}}],
					"responses":{"200":{"description":"ok","schema":{"$ref":"#/definitions/response.Response"}}}
				}
			},
			"/stats/{provider}":{
				"parameters":[{"type":"string","name":"provider","in":"path","required":true}],
				"get":{
					"parameters":[
						{"type":"string","enum":["csv","excel"],"name":"format","in":"query"},
						{"type":"integer","default":1,"name":"page","in":"query"}
					],
					"responses":{"200":{"description":"ok"}}
				}
			}
		},
		"definitions":{
			"dto.LoginRequest":{"type":"object"},
			"response.Response":{"type":"object"}
		},
		"securityDefinitions":{
			"BearerAuth":{"type":"apiKey","name":"Authorization","in":"header"}
		}
	}`)
	out, err := swagger2ToOpenAPI3(src)
	require.NoError(t, err)
	var spec map[string]any
	require.NoError(t, json.Unmarshal(out, &spec))
	require.Equal(t, "3.0.0", spec["openapi"])
	servers, ok := spec["servers"].([]any)
	require.True(t, ok)
	require.Equal(t, "http://localhost:8080/api/v1", servers[0].(map[string]any)["url"])
	paths := spec["paths"].(map[string]any)
	login := paths["/auth/login"].(map[string]any)["post"].(map[string]any)
	rb := login["requestBody"].(map[string]any)
	content := rb["content"].(map[string]any)["application/json"].(map[string]any)
	schema := content["schema"].(map[string]any)
	require.Equal(t, "#/components/schemas/dto.LoginRequest", schema["$ref"])
	stats := paths["/stats/{provider}"].(map[string]any)
	pathParams := stats["parameters"].([]any)
	require.Len(t, pathParams, 1)
	provider := pathParams[0].(map[string]any)
	require.Equal(t, "path", provider["in"])
	require.Equal(t, "provider", provider["name"])
	require.Nil(t, provider["type"])
	require.Equal(t, "string", provider["schema"].(map[string]any)["type"])
	getStats := stats["get"].(map[string]any)
	queryParams := getStats["parameters"].([]any)
	require.Len(t, queryParams, 2)
	format := queryParams[0].(map[string]any)
	require.Nil(t, format["type"])
	require.Nil(t, format["enum"])
	formatSchema := format["schema"].(map[string]any)
	require.Equal(t, "string", formatSchema["type"])
	require.Equal(t, []any{"csv", "excel"}, formatSchema["enum"])
	page := queryParams[1].(map[string]any)
	require.Nil(t, page["type"])
	require.Nil(t, page["default"])
	pageSchema := page["schema"].(map[string]any)
	require.Equal(t, "integer", pageSchema["type"])
	require.Equal(t, float64(1), pageSchema["default"])
	comps := spec["components"].(map[string]any)
	require.Contains(t, comps["schemas"].(map[string]any), "dto.LoginRequest")
	require.Contains(t, comps["securitySchemes"].(map[string]any), "BearerAuth")
}

func TestRegisterSwaggerProdDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("APP_ENV", "prod")
	r := gin.New()
	registerSwagger(r)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil))
	require.Equal(t, http.StatusNotFound, w.Code)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/swagger/openapi.json", nil))
	require.Equal(t, http.StatusNotFound, w2.Code)
}

func TestRegisterSwaggerDevServesOpenAPI3(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("APP_ENV", "dev")
	r := gin.New()
	require.NotPanics(t, func() { registerSwagger(r) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger/openapi.json", nil))
	require.Equal(t, http.StatusOK, w.Code)
	var spec map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &spec))
	require.Equal(t, "3.0.0", spec["openapi"])
	paths, ok := spec["paths"].(map[string]any)
	require.True(t, ok)
	me := paths["/auth/sessions/{user_id}"].(map[string]any)["get"].(map[string]any)
	params := me["parameters"].([]any)
	found := false
	for _, raw := range params {
		p := raw.(map[string]any)
		if p["name"] == "user_id" {
			require.Nil(t, p["type"])
			require.Equal(t, "string", p["schema"].(map[string]any)["type"])
			found = true
		}
	}
	require.True(t, found)
}
