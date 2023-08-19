package gateway

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/stackup-app/stackup/lib/settings"
	"github.com/stackup-app/stackup/lib/utils"
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
	AllowedFileExts     []string
	BlockedFileExts     []string
	Middleware          []*GatewayUrlRequestMiddleware
	PostMiddleware      []*GatewayUrlResponseMiddleware
	DomainHeaders       *sync.Map
	DomainContentTypes  *sync.Map
	BlockedContentTypes *sync.Map
}

// New initializes the gateway with deny/allow lists
func New(deniedDomains, allowedDomains, blockedFileExts, allowedFileExts []string) *Gateway {
	result := Gateway{
		Enabled:             true,
		DeniedDomains:       deniedDomains,
		AllowedDomains:      allowedDomains,
		BlockedFileExts:     blockedFileExts,
		AllowedFileExts:     allowedFileExts,
		Middleware:          []*GatewayUrlRequestMiddleware{},
		PostMiddleware:      []*GatewayUrlResponseMiddleware{},
		DomainHeaders:       &sync.Map{},
		DomainContentTypes:  &sync.Map{},
		BlockedContentTypes: &sync.Map{},
	}

	result.setup()

	return &result
}

func (g *Gateway) setup() {
	g.AddMiddleware(&ValidateUrlMiddleware)
	g.AddMiddleware(&VerifyFileTypeMiddleware)
	g.AddPostMiddleware(&VerifyContentTypeMIddleware)

	g.Enable()
}

func (g *Gateway) Initialize(s *settings.Settings) {
	g.SetAllowedDomains(s.Domains.Allowed)
	g.SetDeniedDomains([]string{"*"})
	g.SetAllowedFileExts(s.Gateway.FileExtensions.Allow)
	g.SetBlockedFileExts(s.Gateway.FileExtensions.Block)

	g.DeniedDomains = g.normalizeDomainArray(g.DeniedDomains)
	g.AllowedDomains = g.normalizeDomainArray(g.AllowedDomains)
}

func (g *Gateway) SetAllowedFileExts(exts []string) {
	g.AllowedFileExts = exts
}

func (g *Gateway) SetBlockedFileExts(exts []string) {
	g.BlockedFileExts = exts
}

func (g *Gateway) SetDefaults() {
	if len(g.AllowedFileExts) == 0 {
		g.AllowedFileExts = []string{"*"}
	}
	if len(g.BlockedFileExts) == 0 {
		g.BlockedFileExts = []string{"*"}
	}
}

func (g *Gateway) SetAllowedDomains(domains []string) {
	g.AllowedDomains = g.normalizeDomainArray(domains)
}

func (g *Gateway) SetDeniedDomains(domains []string) {
	g.DeniedDomains = g.normalizeDomainArray(domains)
}

func (g *Gateway) SetDomainHeaders(domain string, headers []string) {
	g.DomainHeaders.Store(domain, headers)
}

func (g *Gateway) GetDomainHeaders(domain string) []string {
	result := []string{}

	g.DomainHeaders.Range(func(key, value any) bool {
		if utils.DomainGlobMatch(key.(string), domain) {
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
		if utils.DomainGlobMatch(key.(string), domain) {
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
		if utils.DomainGlobMatch(key.(string), domain) {
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

// The `runUrlRequestPipeline` method runs the middleware pipeline for a URL request, returning
// an error if the request is not allowed AND the gateway is enabled, or nil if the request is
// allowed/the gateway is disabled.
func (g *Gateway) runUrlRequestPipeline(link string) error {
	if !g.Enabled {
		return nil
	}

	if !strings.Contains(link, "://") {
		link = "https://" + link
	}

	for _, mw := range g.Middleware {
		if err := (*mw).Handler(g, link); err != nil {
			return err
		}
	}

	return nil
}

func (g *Gateway) runResponsePipeline(resp *http.Response) error {
	if !g.Enabled {
		return nil
	}

	for _, mw := range g.PostMiddleware {
		if err := (*mw).Handler(g, resp); err != nil {
			return err
		}
	}

	return nil
}

func (g *Gateway) Allowed(link string) bool {
	return g.runUrlRequestPipeline(link) == nil
}

// processes an array of domains and remove any empty strings and extracts hostnames from URLs if
// they are present, then returns a new array without the removed items
func (g *Gateway) normalizeDomainArray(arr []string) []string {
	result := []string{}

	for _, domain := range arr {
		if len(strings.TrimSpace(domain)) == 0 {
			continue
		}
		if strings.Contains(domain, "://") {
			parsedUrl, err := url.Parse(domain)
			if err == nil {
				domain = parsedUrl.Host
			}
		}

		result = append(result, domain)
	}

	return result
}

func (g *Gateway) Enable() {
	g.Enabled = true
}

func (g *Gateway) Disable() {
	g.Enabled = false
}

func (g *Gateway) checkArrayForDomainMatch(arr *[]string, domain string) bool {
	for _, domainPattern := range *arr {
		if strings.EqualFold(domain, domainPattern) || strings.EqualFold(strings.TrimPrefix(domainPattern, "*."), domain) {
			return true
		}
		if utils.DomainGlobMatch(domainPattern, domain) {
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
		if utils.DomainGlobMatch(key.(string), parsed.Hostname()) {
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
