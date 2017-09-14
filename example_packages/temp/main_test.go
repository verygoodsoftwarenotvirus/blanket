package dairyclient

import (
	"fmt"
	"io/ioutil"
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
