package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/settings"
	"github.com/stackup-app/stackup/lib/types"
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

type GatewayHttpResponse struct {
	Url      string `json:"url"`
	Contents string `json:"contents"`
	Code     int    `json:"code"`
}

type Gateway struct {
	Enabled             bool
	AllowedDomains      []string
	DeniedDomains       []string
	AllowedFileExts     []string
	BlockedFileExts     []string
	EnabledMiddleware   []string
	Middleware          []*GatewayUrlRequestMiddleware
	PostMiddleware      []*GatewayUrlResponseMiddleware
	DomainHeaders       *sync.Map
	DomainContentTypes  *sync.Map
	BlockedContentTypes *sync.Map
	JsEngine            types.JavaScriptEngineContract
	HttpClient          *http.Client
	Settings            *settings.Settings
	Cache               *cache.Cache
}

// New initializes the gateway with deny/allow lists
func New(cache *cache.Cache) *Gateway {
	result := Gateway{
		Enabled:             true,
		DeniedDomains:       []string{},
		AllowedDomains:      []string{},
		BlockedFileExts:     []string{},
		AllowedFileExts:     []string{},
		Middleware:          []*GatewayUrlRequestMiddleware{},
		PostMiddleware:      []*GatewayUrlResponseMiddleware{},
		DomainHeaders:       &sync.Map{},
		DomainContentTypes:  &sync.Map{},
		BlockedContentTypes: &sync.Map{},
		HttpClient:          http.DefaultClient,
		Cache:               cache,
	}

	result.Enable()

	return &result
}

func (g *Gateway) setup() {
	// we need the `validateUrl` middleware no matter what, so prepend it to the list of enabled middleware.
	// this middleware is the core functionality of the http gateway's allow/block lists.
	// temp := []string{"validateUrl"}
	g.EnabledMiddleware = []string{"validateUrl"}
	g.EnabledMiddleware = append(g.EnabledMiddleware, "validateUrl")
	GatewayMiddleware.AddPreMiddleware(&ValidateUrlMiddleware)

	// fmt.Printf("enabled middleware: %v\n", g.EnabledMiddleware)

	for _, name := range g.EnabledMiddleware {
		if !GatewayMiddleware.HasMiddleware(name) {
			continue
		}

		g.AddPreMiddleware(GatewayMiddleware.GetPreMiddleware(name))
		g.AddPostMiddleware(GatewayMiddleware.GetPostMiddleware(name))
	}

	g.Enable()
}

// Initializes the gateway using the specified settings, JavascriptEngine, and http client.  If `httpClient` is nil,
// the `http.DefaultClient` is be used.
func (g *Gateway) Initialize(s *settings.Settings, jsEngine types.JavaScriptEngineContract, httpClient *http.Client) {
	if g == nil {
		return
	}

	g.AllowedDomains = []string{"*"}

	g.DomainContentTypes = &sync.Map{}
	g.BlockedContentTypes = &sync.Map{}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	g.Settings = s

	g.Middleware = []*GatewayUrlRequestMiddleware{
		&ValidateUrlMiddleware,
	}

	if s.Gateway.FileExtensions == nil {
		s.Gateway.FileExtensions = &settings.WorkflowSettingsGatewayFileExtensions{
			Allow: []string{"*"},
			Block: []string{"*"},
		}
	}

	g.HttpClient = httpClient
	g.JsEngine = jsEngine
	g.SetAllowedDomains(s.Domains.Allowed)
	g.SetDeniedDomains([]string{"*"})
	g.SetAllowedFileExts(s.Gateway.FileExtensions.Allow)
	g.SetBlockedFileExts(s.Gateway.FileExtensions.Block)
	g.EnabledMiddleware = s.Gateway.Middleware
	g.DeniedDomains = g.normalizeDomainArray(g.DeniedDomains)
	g.AllowedDomains = g.normalizeDomainArray(g.AllowedDomains)

	//g.SetDefaults()

	// `setup` must be called after the settings have been applied above
	g.setup()
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
	g.SetDomainContentTypes("*", g.Settings.Gateway.ContentTypes.Allowed)
	g.SetBlockedContentTypes("*", g.Settings.Gateway.ContentTypes.Blocked)
}

func (g *Gateway) SetAllowedDomains(domains []string) {
	//g.AllowedDomains = g.normalizeDomainArray(domains)
}

func (g *Gateway) SetDeniedDomains(domains []string) {
	g.DeniedDomains = g.normalizeDomainArray(domains)
}

func (g *Gateway) SetDomainHeaders(domain string, headers []string) {
	//g.DomainHeaders.Store(domain, headers)
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
	// fmt.Printf("setting content types for domain %s: %v\n", domain, contentTypes)
	// fmt.Printf("DomainContentTypes: %v\n", g)

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

func (g *Gateway) AddPreMiddleware(mw *GatewayUrlRequestMiddleware) {
	if mw == nil {
		return
	}

	for _, existingMw := range g.Middleware {
		if strings.EqualFold(existingMw.Name, mw.Name) {
			return
		}
	}

	g.Middleware = append(g.Middleware, mw)
}

func (g *Gateway) AddPostMiddleware(mw *GatewayUrlResponseMiddleware) {
	if mw == nil {
		return
	}

	for _, existingMw := range g.PostMiddleware {
		if strings.EqualFold(existingMw.Name, mw.Name) {
			return
		}
	}

	g.PostMiddleware = append(g.PostMiddleware, mw)
}

// The `runUrlRequestPipeline` method runs the middleware pipeline for a URL request, returning
// an error if the request is not allowed AND the gateway is enabled, or nil if the request is
// allowed/the gateway is disabled.
func (g *Gateway) runUrlRequestPipeline(link string) error {
	if g == nil || !g.Enabled {
		return nil
	}

	if !strings.Contains(link, "://") {
		link = "https://" + link
	}

	for _, mw := range g.Middleware {
		if err := (*mw).Handler(g, link); err != nil {
			fmt.Printf("error on mw: %v; %v\n", mw.Name, err)
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

func (g *Gateway) HasCache() bool {
	return g.Cache != nil
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

func (g *Gateway) processHeaders(headers []string) []string {
	result := []string{}

	for _, header := range headers {
		if strings.TrimSpace(header) == "" {
			continue
		}

		if g.JsEngine.IsEvaluatableScriptString(header) {
			header = g.JsEngine.GetEvaluatableScriptString(header)
		}

		result = append(result, os.ExpandEnv(header))
	}

	return result
}

// GetUrl returns the contents of a URL as a string, assuming it
// is allowed by the gateway, otherwise it returns an error.
func (g *Gateway) GetUrl(urlStr string, headers ...string) (string, error) {
	if err := g.runUrlRequestPipeline(urlStr); err != nil {
		fmt.Printf("error: %v\n", err)
		return "", err
	}

	expireTtl := utils.Min(5, g.Cache.DefaultTtl)

	var response *GatewayHttpResponse = &GatewayHttpResponse{
		Url:      urlStr,
		Code:     1,
		Contents: "",
	}

	if g.HasCache() {
		// fmt.Printf(" [gateway] checking cache for response to %s\n", urlStr)
		entry, _ := g.Cache.Get(g.CacheKeyFor(urlStr))
		if entry != nil {
			err := json.Unmarshal([]byte(entry.Value), response)
			// fmt.Printf(" [gateway] found cached response for %v\n", response)

			if response != nil && response.Code > 1 {
				if response.Code != 200 {
					err = fmt.Errorf("HTTP request failed: error %d (url: %s)", response.Code, urlStr)
				}

				return response.Contents, err
			}
		}
	}

	fmt.Printf("gateway.get url == %s\n", urlStr)

	var allHeaders []string = []string{"User-Agent: stackup/1.0"}

	if g.DomainHeaders == nil {
		g.DomainHeaders = &sync.Map{}
	}

	g.DomainHeaders.Range(func(key, value any) bool {
		parsed, _ := url.Parse(urlStr)
		if utils.DomainGlobMatch(key.(string), parsed.Hostname()) {
			allHeaders = append(allHeaders, g.processHeaders(value.([]string))...)
		}
		return true
	})

	allHeaders = append(allHeaders, g.processHeaders(headers)...)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}

	for _, header := range allHeaders {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	resp, err := g.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	response.Code = resp.StatusCode

	if g.HasCache() {
		g.Cache.Set(g.CacheKeyFor(urlStr), cache.NewCacheEntry(response, expireTtl), expireTtl)
	}

	if err = g.runResponsePipeline(resp); err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP request failed: error %d (url: %s)", resp.StatusCode, urlStr)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if g.HasCache() {
		response.Contents = string(body)
		g.Cache.Set(g.CacheKeyFor(urlStr), cache.NewCacheEntry(response, expireTtl), expireTtl)
	}

	return string(body), nil
}

func (g *Gateway) CacheKeyFor(url string) string {
	return "gateway:" + url
}
