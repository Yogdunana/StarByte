package service

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CASPrincipal is the identity returned by CAS serviceValidate.
type CASPrincipal struct {
	User       string
	Attributes map[string]string
}

// TicketValidator validates a Service Ticket against the CAS server.
type TicketValidator interface {
	Validate(ctx context.Context, service, ticket string) (*CASPrincipal, error)
}

type httpTicketValidator struct {
	serverURL string
	client    *http.Client
}

// NewHTTPTicketValidator talks to 金智 / Apereo CAS serviceValidate endpoints.
func NewHTTPTicketValidator(serverURL string, client *http.Client) TicketValidator {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &httpTicketValidator{serverURL: strings.TrimRight(strings.TrimSpace(serverURL), "/"), client: client}
}

func (v *httpTicketValidator) Validate(ctx context.Context, service, ticket string) (*CASPrincipal, error) {
	if v.serverURL == "" {
		return nil, fmt.Errorf("cas server url is empty")
	}
	// ST 一次性。只有 p3 端点不存在/不可达时才回退，避免 200 已耗票后再打 /serviceValidate。
	p, status, err := v.fetch(ctx, v.serverURL+"/p3/serviceValidate", service, ticket, true)
	if casValidateUnreachable(status, err) {
		p, _, err = v.fetch(ctx, v.serverURL+"/serviceValidate", service, ticket, false)
	}
	return p, err
}

func casValidateUnreachable(status int, err error) bool {
	if status == http.StatusNotFound || status == http.StatusMethodNotAllowed || status == http.StatusNotImplemented {
		return true
	}
	return err != nil && status == 0
}

func (v *httpTicketValidator) fetch(ctx context.Context, endpoint, service, ticket string, wantJSON bool) (*CASPrincipal, int, error) {
	q := url.Values{}
	q.Set("service", service)
	q.Set("ticket", ticket)
	if wantJSON {
		q.Set("format", "JSON")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("cas validate http %d", resp.StatusCode)
	}
	trimmed := strings.TrimSpace(string(body))
	if strings.HasPrefix(trimmed, "{") {
		p, perr := parseCASJSON(trimmed)
		return p, resp.StatusCode, perr
	}
	p, perr := parseCASXML(trimmed)
	return p, resp.StatusCode, perr
}

func parseCASJSON(raw string) (*CASPrincipal, error) {
	var envelope struct {
		ServiceResponse struct {
			AuthenticationSuccess *struct {
				User       string          `json:"user"`
				Attributes json.RawMessage `json:"attributes"`
			} `json:"authenticationSuccess"`
			AuthenticationFailure *struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"authenticationFailure"`
		} `json:"serviceResponse"`
	}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		return nil, fmt.Errorf("parse cas json: %w", err)
	}
	if fail := envelope.ServiceResponse.AuthenticationFailure; fail != nil {
		msg := strings.TrimSpace(fail.Description)
		if msg == "" {
			msg = fail.Code
		}
		if msg == "" {
			msg = "ticket validation failed"
		}
		return nil, fmt.Errorf("cas: %s", msg)
	}
	ok := envelope.ServiceResponse.AuthenticationSuccess
	if ok == nil || strings.TrimSpace(ok.User) == "" {
		return nil, fmt.Errorf("cas json missing user")
	}
	return &CASPrincipal{
		User:       strings.TrimSpace(ok.User),
		Attributes: flattenJSONAttributes(ok.Attributes),
	}, nil
}

func flattenJSONAttributes(raw json.RawMessage) map[string]string {
	out := map[string]string{}
	if len(raw) == 0 {
		return out
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return out
	}
	for k, v := range obj {
		if s := stringifyAttr(v); s != "" {
			out[strings.ToLower(strings.TrimSpace(k))] = s
		}
	}
	return out
}

func stringifyAttr(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%.0f", t))
	case json.Number:
		return strings.TrimSpace(t.String())
	case []any:
		if len(t) == 0 {
			return ""
		}
		return stringifyAttr(t[0])
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

type casXMLResponse struct {
	XMLName xml.Name
	Success *casXMLSuccess `xml:"authenticationSuccess"`
	Failure *casXMLFailure `xml:"authenticationFailure"`
}

type casXMLSuccess struct {
	User       string `xml:"user"`
	Attributes []byte `xml:",innerxml"`
}

type casXMLFailure struct {
	Code    string `xml:"code,attr"`
	Message string `xml:",chardata"`
}

func parseCASXML(raw string) (*CASPrincipal, error) {
	var doc casXMLResponse
	if err := xml.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("parse cas xml: %w", err)
	}
	if doc.Failure != nil {
		msg := strings.TrimSpace(doc.Failure.Message)
		if msg == "" {
			msg = doc.Failure.Code
		}
		if msg == "" {
			msg = "ticket validation failed"
		}
		return nil, fmt.Errorf("cas: %s", msg)
	}
	if doc.Success == nil || strings.TrimSpace(doc.Success.User) == "" {
		return nil, fmt.Errorf("cas xml missing user")
	}
	return &CASPrincipal{
		User:       strings.TrimSpace(doc.Success.User),
		Attributes: parseCASXMLAttributes(string(doc.Success.Attributes)),
	}, nil
}

func parseCASXMLAttributes(inner string) map[string]string {
	out := map[string]string{}
	decoder := xml.NewDecoder(strings.NewReader("<attrs>" + inner + "</attrs>"))
	var current string
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			current = strings.ToLower(t.Name.Local)
		case xml.CharData:
			if current == "" || current == "user" || current == "attributes" || current == "attrs" {
				continue
			}
			val := strings.TrimSpace(string(t))
			if val != "" {
				out[current] = val
			}
		case xml.EndElement:
			current = ""
		}
	}
	return out
}
