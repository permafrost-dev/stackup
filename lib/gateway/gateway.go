package gateway

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	glob "github.com/ryanuber/go-glob"
)

type GatewayUrlRequestMiddleware struct {
	Name    string
	Handler func(g *Gateway, link string) error
}

type GatewayUrlResponseMiddleware struct {
	Name    string
	Handler func(g *Gateway, resp *http.Response) error
}
type Gateway struct {
	Enabled             bool
	AllowedDomains      []string
	DeniedDomains       []string
	Middleware          []*GatewayUrlRequestMiddleware
	PostMiddleware      []*GatewayUrlResponseMiddleware
	DomainHeaders       *sync.Map
	DomainContentTypes  *sync.Map
	BlockedContentTypes *sync.Map
}

// New initializes the gateway with deny/allow lists
func New(deniedDomains, allowedDomains []string) *Gateway {
	result := Gateway{
		Enabled:             true,
		DeniedDomains:       deniedDomains,
		AllowedDomains:      allowedDomains,
		Middleware:          []*GatewayUrlRequestMiddleware{},
		PostMiddleware:      []*GatewayUrlResponseMiddleware{},
		DomainHeaders:       &sync.Map{},
		DomainContentTypes:  &sync.Map{},
		BlockedContentTypes: &sync.Map{},
	}

	result.Initialize()

	return &result
}

func (g *Gateway) Initialize() {
	g.normalizeDataArray(&g.DeniedDomains)
	g.normalizeDataArray(&g.AllowedDomains)

	g.AddMiddleware(&ValidateUrlMiddleware)
	g.AddMiddleware(&VerifyFileTypeMiddleware)
	g.AddPostMiddleware(&VerifyContentTypeMIddleware)

	g.Enable()
}

func (g *Gateway) SetAllowedDomains(domains []string) {
	g.AllowedDomains = domains
	g.normalizeDataArray(&g.AllowedDomains)
}

func (g *Gateway) SetDeniedDomains(domains []string) {
	g.DeniedDomains = domains
	g.normalizeDataArray(&g.DeniedDomains)
}

func (g *Gateway) SetDomainHeaders(domain string, headers []string) {
	g.DomainHeaders.Store(domain, headers)
}

func (g *Gateway) GetDomainHeaders(domain string) []string {
	result := []string{}

	g.DomainHeaders.Range(func(key, value any) bool {
		if glob.Glob(key.(string), domain) {
			result = append(result, value.([]string)...)
		}
		return true
	})

	return result
}

func (g *Gateway) SetBlockedContentTypes(domain string, contentTypes []string) {
	g.BlockedContentTypes.Store(domain, contentTypes)
}

func (g *Gateway) GetBlockedContentTypes(domain string) []string {
	result := []string{}

	g.BlockedContentTypes.Range(func(key, value any) bool {
		if glob.Glob(key.(string), domain) {
			result = append(result, value.([]string)...)
		}
		return true
	})

	return result
}

func (g *Gateway) SetDomainContentTypes(domain string, contentTypes []string) {
	if len(contentTypes) == 0 {
		g.DomainContentTypes.Delete(domain)
		return
	}

	g.DomainContentTypes.Store(domain, contentTypes)
}

func (g *Gateway) GetDomainContentTypes(domain string) []string {
	result := []string{}

	g.DomainContentTypes.Range(func(key, value any) bool {
		if glob.Glob(key.(string), domain) {
			result = append(result, value.([]string)...)
		}
		return true
	})

	return result
}

func (g *Gateway) AddMiddleware(mw *GatewayUrlRequestMiddleware) {
	g.Middleware = append(g.Middleware, mw)
}

func (g *Gateway) AddPostMiddleware(mw *GatewayUrlResponseMiddleware) {
	g.PostMiddleware = append(g.PostMiddleware, mw)
}

// The `runUrlRequestPipeline` function is a method of the `Gateway` struct. It iterates over the
// `Middleware` slice of the `Gateway` struct and executes each middleware function in order. Each
// middleware function takes a `Gateway` instance and a URL `link` as parameters and returns an error.
// If any middleware function returns an error, the `runUrlRequestPipeline` function immediately
// returns that error. If all middleware functions are executed successfully, the function returns
// `nil`.
func (g *Gateway) runUrlRequestPipeline(link string) error {
	for _, mw := range g.Middleware {
		err := (*mw).Handler(g, link)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Gateway) runResponsePipeline(resp *http.Response) error {
	for _, mw := range g.PostMiddleware {
		err := (*mw).Handler(g, resp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Gateway) Allowed(link string) bool {
	return g.runUrlRequestPipeline(link) == nil
}

// processes an array of domains and remove any empty strings and extract hostnames from URLs if
// they are present, then copy the result back to the original array so we have an array of only
// hostnames with or without wildcard characters.
func (g *Gateway) normalizeDataArray(arr *[]string) {
	tempDomains := []string{}

	for _, domain := range *arr {
		if len(strings.TrimSpace(domain)) == 0 {
			continue
		}
		if strings.Contains(domain, "://") {
			parsedUrl, _ := url.Parse(domain)
			domain = parsedUrl.Host
		}

		tempDomains = append(tempDomains, domain)
	}

	copy(*arr, tempDomains)
}

func (g *Gateway) Enable() {
	g.Enabled = true
}

func (g *Gateway) Disable() {
	g.Enabled = false
}

func (g *Gateway) checkArrayForDomainMatch(arr *[]string, s string) bool {
	for _, domain := range *arr {
		if strings.EqualFold(s, domain) {
			return true
		}
		if strings.Contains(domain, "*") && glob.Glob(domain, s) {
			return true
		}
	}

	return false
}

// GetUrl returns the contents of a URL as a string, assuming it
// is allowed by the gateway, otherwise it returns an error.
func (g *Gateway) GetUrl(urlStr string, headers ...string) (string, error) {
	err := g.runUrlRequestPipeline(urlStr)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return "", err
	}

	var tempHeaders []string = []string{"User-Agent: stackup/1.0"}

	g.DomainHeaders.Range(func(key, value any) bool {
		parsed, _ := url.Parse(urlStr)
		if glob.Glob(key.(string), parsed.Hostname()) {
			for _, header := range value.([]string) {
				header := os.ExpandEnv(header)
				tempHeaders = append(tempHeaders, header)
			}
		}
		return true
	})

	for _, header := range headers {
		if strings.TrimSpace(header) != "" {
			header = os.ExpandEnv(header)
			tempHeaders = append(tempHeaders, header)
		}
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}

	for _, header := range tempHeaders {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	err = g.runResponsePipeline(resp)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Read the response body into a byte slice
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
