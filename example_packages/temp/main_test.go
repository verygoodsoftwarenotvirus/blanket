package dairyclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	exampleID       = 666
	exampleURL      = `http://www.dairycart.com`
	exampleUsername = `username`
	examplePassword = `password` // lol not really
	exampleSKU      = `sku`
	exampleBadJSON  = `{"invalid lol}`
)

type subtest struct {
	Message string
	Test    func(t *testing.T)
}

func handlerGenerator(handlers map[string]func(res http.ResponseWriter, req *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for path, handlerFunc := range handlers {
			if r.URL.Path == path {
				handlerFunc(w, r)
				return
			}
		}
	})
}

func runSubtestSuite(t *testing.T, tests []subtest) {
	testPassed := true
	for _, test := range tests {
		if !testPassed {
			t.FailNow()
		}
		testPassed = t.Run(test.Message, test.Test)
	}
}

func createInternalClient(t *testing.T, ts *httptest.Server) *V1Client {
	u, err := url.Parse(ts.URL)
	assert.Nil(t, err, "no error should be returned when parsing a test server's URL")

	c := &V1Client{
		Client: ts.Client(),
		AuthCookie: &http.Cookie{
			Name: "dairycart",
		},
		URL: u,
	}
	return c
}

func TestUnexportedBuildURL(t *testing.T) {
	ts := httptest.NewTLSServer(http.NotFoundHandler())
	defer ts.Close()
	c := createInternalClient(t, ts)

	testCases := []struct {
		query    map[string][]string
		parts    []string
		expected string
	}{
		{
			query:    nil,
			parts:    []string{""},
			expected: fmt.Sprintf("%s/v1/", ts.URL),
		},
		{
			query:    nil,
			parts:    []string{"things", "and", "stuff"},
			expected: fmt.Sprintf("%s/v1/things/and/stuff", ts.URL),
		},
		{
			query:    map[string][]string{"param": {"value"}},
			parts:    []string{"example"},
			expected: fmt.Sprintf("%s/v1/example?param=value", ts.URL),
		},
	}

	for _, tc := range testCases {
		actual := c.buildURL(tc.query, tc.parts...)
		assert.Equal(t, tc.expected, actual, "expected and actual built URLs don't match")
	}
}

func TestExecuteRequestAddsCookieToRequests(t *testing.T) {
	t.Parallel()
	var endpointCalled bool
	exampleEndpoint := "/v1/whatever"

	handlers := map[string]func(res http.ResponseWriter, req *http.Request){
		exampleEndpoint: func(res http.ResponseWriter, req *http.Request) {
			endpointCalled = true
			cookies := req.Cookies()
			if len(cookies) == 0 {
				assert.FailNow(t, "no cookies attached to the request")
			}

			cookieFound := false
			for _, c := range cookies {
				if c.Name == "dairycart" {
					cookieFound = true
				}
			}
			assert.True(t, cookieFound)
		},
	}

	ts := httptest.NewTLSServer(handlerGenerator(handlers))
	defer ts.Close()
	c := createInternalClient(t, ts)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", ts.URL, exampleEndpoint), nil)
	assert.Nil(t, err, "no error should be returned when creating a new request")

	c.executeRequest(req)
	assert.True(t, endpointCalled, "endpoint should have been called")
}

func TestGet(t *testing.T) {
	t.Parallel()
	var normalEndpointCalled bool
	handlers := map[string]func(res http.ResponseWriter, req *http.Request){
		"/v1/normal": func(res http.ResponseWriter, req *http.Request) {
			normalEndpointCalled = true
			assert.Equal(t, req.Method, http.MethodGet, "get should be making GET requests")
			exampleResponse := `
				{
					"things": "stuff"
				}
			`
			fmt.Fprintf(res, exampleResponse)
		},
	}

	ts := httptest.NewTLSServer(handlerGenerator(handlers))
	defer ts.Close()
	c := createInternalClient(t, ts)

	normalUse := func(t *testing.T) {
		expected := struct {
			Things string `json:"things"`
		}{
			Things: "stuff",
		}

		actual := struct {
			Things string `json:"things"`
		}{}

		err := c.get(c.buildURL(nil, "normal"), &actual)
		assert.Nil(t, err)
		assert.Equal(t, expected, actual, "actual struct should equal expected struct")
		assert.True(t, normalEndpointCalled, "endpoint should have been called")
	}

	nilInput := func(t *testing.T) {
		nilErr := c.get(c.buildURL(nil, "whatever"), nil)
		assert.NotNil(t, nilErr)
	}

	nonPointerInput := func(t *testing.T) {
		actual := struct {
			Things string `json:"things"`
		}{}

		ptrErr := c.get(c.buildURL(nil, "whatever"), actual)
		assert.NotNil(t, ptrErr)
	}

	subtests := []subtest{
		{
			Message: "normal use",
			Test:    normalUse,
		},
		{
			Message: "nil input",
			Test:    nilInput,
		},
		{
			Message: "non-pointer input",
			Test:    nonPointerInput,
		},
	}
	runSubtestSuite(t, subtests)
}
