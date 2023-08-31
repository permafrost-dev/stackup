package gateway

import (
	"strings"
)

type GatewayMiddlewareStore struct {
	PreMiddleware  []*GatewayUrlRequestMiddleware
	PostMiddleware []*GatewayUrlResponseMiddleware
}

var GatewayMiddleware = &GatewayMiddlewareStore{
	PreMiddleware:  []*GatewayUrlRequestMiddleware{},
	PostMiddleware: []*GatewayUrlResponseMiddleware{},
}

func (mws *GatewayMiddlewareStore) HasMiddleware(name string) bool {
	for _, mw := range mws.PreMiddleware {
		if strings.EqualFold(mw.Name, name) {
			return true
		}
	}
	for _, mw := range mws.PostMiddleware {
		if strings.EqualFold(mw.Name, name) {
			return true
		}
	}
	return false
}

func (mws *GatewayMiddlewareStore) AddPreMiddleware(mw *GatewayUrlRequestMiddleware) {
	if mws.HasMiddleware(mw.Name) {
		return
	}

	mws.PreMiddleware = append(mws.PreMiddleware, mw)
}

func (mws *GatewayMiddlewareStore) AddPostMiddleware(mw *GatewayUrlResponseMiddleware) {
	if mws.HasMiddleware(mw.Name) {
		return
	}

	mws.PostMiddleware = append(mws.PostMiddleware, mw)
}

func (mws *GatewayMiddlewareStore) GetPreMiddleware(name string) *GatewayUrlRequestMiddleware {
	for _, mw := range mws.PreMiddleware {
		if strings.EqualFold(mw.Name, name) {
			return mw
		}
	}
	return nil
}

func (mws *GatewayMiddlewareStore) GetPostMiddleware(name string) *GatewayUrlResponseMiddleware {
	for _, mw := range mws.PostMiddleware {
		if strings.EqualFold(mw.Name, name) {
			return mw
		}
	}
	return nil
}
