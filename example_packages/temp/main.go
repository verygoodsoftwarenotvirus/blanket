package dairyclient

import (
	"net/http"
	"net/url"
	"strings"
)

type DairyclientV1 interface {
	ProductExists(sku string) (bool, error)
	DeleteProduct(sku string) error
	DeleteProductRoot(rootID uint64) error
	DeleteProductOption(optionID uint64) error
	DeleteProductOptionValueForOption(optionID uint64) error
}

type V1Client struct {
	*http.Client
	URL        *url.URL
	AuthCookie *http.Cookie
}

func (dc *V1Client) executeRequest(req *http.Request) (*http.Response, error) {
	req.AddCookie(dc.AuthCookie)
	return dc.Do(req)
}

func (dc *V1Client) buildURL(queryParams map[string][]string, parts ...string) string {
	parts = append([]string{"v1"}, parts...)
	u, _ := url.Parse(strings.Join(parts, "/"))
	queryString := url.Values(queryParams)
	u.RawQuery = queryString.Encode()
	return dc.URL.ResolveReference(u).String()
}

func (dc *V1Client) get(uri string, obj interface{}) error {
	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	res, err := dc.executeRequest(req)
	if err != nil {
		return err
	}
	return nil
}
