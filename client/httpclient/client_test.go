package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/ekspand/trusty/internal/config"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projFolder = "../../"

func Test_New(t *testing.T) {
	_, err := New(NewConfig(), nil)
	assert.Error(t, err, "Expected error when calling new on empty host list")
}

func Test_DecodeResponse(t *testing.T) {
	res := http.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(bytes.NewBufferString(`{"code":"MY_CODE","message":"doesn't exist"}`))}
	c, err := New(NewConfig(), []string{"https://localhost:5555"})
	assert.NoError(t, err, "Unexpected error.")
	var body map[string]string
	_, sc, err := c.httpClient.DecodeResponse(&res, &body)
	require.Equal(t, res.StatusCode, sc)
	require.Error(t, err)

	ge, ok := err.(*httperror.Error)
	require.True(t, ok, "Expecting decodeResponse to map a valid error to the Error struct, but was %T %v", err, err)
	assert.Equal(t, "MY_CODE", ge.Code)
	assert.Equal(t, "doesn't exist", ge.Message)
	assert.Equal(t, http.StatusNotFound, ge.HTTPStatus)

	// if the body isn't valid json, we should get returned a json parser error, as well as the body
	invalidResponse := `["foo"}`
	res.Body = ioutil.NopCloser(bytes.NewBufferString(invalidResponse))
	_, sc, err = c.httpClient.DecodeResponse(&res, &body)
	require.Error(t, err)
	assert.Equal(t, invalidResponse, err.Error())

	// error body is valid json, but missing the error field
	res.Body = ioutil.NopCloser(bytes.NewBufferString(`{"foo":"bar"}`))
	_, sc, err = c.httpClient.DecodeResponse(&res, &body)
	assert.True(t, err == nil || err.Error() != `HTTP StatusCode 404: Body: {"foo":"bar"}`,
		"decodeResponse should return entire body for an error response with no error field, but got %v", err)

	// statusCode < 300, with a decodeable body
	res.StatusCode = http.StatusOK
	res.Body = ioutil.NopCloser(bytes.NewBufferString(`{"foo":"baz"}`))
	_, sc, err = c.httpClient.DecodeResponse(&res, &body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, sc)
	assert.Equal(t, "baz", body["foo"], "decodeResponse hasn't correctly decoded the payload, got %+v", body)

	// statusCode < 300, with a parsing error
	res.Body = ioutil.NopCloser(bytes.NewBufferString(`[}`))
	_, sc, err = c.httpClient.DecodeResponse(&res, &body)
	assert.Equal(t, http.StatusOK, sc, "decodeResponse returned unexpected statusCode, expecting 200")
	assert.Error(t, err)
	assert.Equal(t, "unable to decode body response to (*map[string]string) type: invalid character '}' looking for beginning of value", err.Error())
}

func Test_GetTriesHostList(t *testing.T) {
	h := func(id string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"sid":%q}`, id)
		})
	}
	s1 := httptest.NewServer(h("s1"))
	s2 := httptest.NewServer(h("s2"))
	s3 := httptest.NewServer(h("s3"))

	defaultConfig := NewConfig()
	client, err := New(defaultConfig, []string{s1.URL, s2.URL, s3.URL})
	assert.NoError(t, err, "Unexpected error.")

	client.httpClient.Policy.Retries[0] = func(r *http.Request, resp *http.Response, err error, retries int) (bool, time.Duration, string) {
		return retries < 2, time.Millisecond, "connection"
	}

	get := func(hosts ...string) string {
		var res map[string]string
		if _, _, err := client.GetFrom(context.Background(), hosts, "/", &res); err != nil {
			t.Fatalf("GetFrom returned error %v", err)
		}
		return res["sid"]
	}
	post := func(hosts ...string) string {
		var res map[string]string
		_, _, err := client.PostRequestTo(context.Background(), hosts, "/", nil, &res)
		require.NoError(t, err)
		return res["sid"]
	}
	verify := func(exp string, hosts ...string) {
		actual := get(hosts...)
		require.Equal(t, exp, actual)
		actual = post(hosts...)
		require.Equal(t, exp, actual)
	}
	verify("s1", s1.URL, s2.URL, s3.URL)
	verify("s2", s2.URL, s1.URL, s3.URL)
	verify("s3", s3.URL, s2.URL, s1.URL)
	s1.Close()
	verify("s2", s1.URL, s2.URL, s3.URL)
	s3.Close()
	verify("s2", s1.URL, s3.URL, s2.URL)
	s2.Close()
	_, _, err = client.Get(context.Background(), "/", nil)
	require.Error(t, err, "Get with all hosts down should return an error")
}

func Test_Retriable(t *testing.T) {
	retryClient := retriable.New()
	c := Client{
		httpClient: retryClient,
	}

	assert.Equal(t, retryClient, c.Retriable())
}

func Test_SetRetryPolicy(t *testing.T) {
	retryPolicy := &retriable.Policy{
		TotalRetryLimit: 0,
	}
	retryClient := retriable.New()
	assert.NotNil(t, retryClient)
	c := &Client{
		httpClient: retryClient,
	}
	c.WithPolicy(retryPolicy)

	assert.Equal(t, retryPolicy, c.Retriable().Policy)
}

func makeTestHandler(t *testing.T, expURI, responseBody string) http.Handler {
	h := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, expURI, r.RequestURI, "received wrong URI")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, responseBody)
	}
	return http.HandlerFunc(h)
}

// setContentsEqual returns true if the 2 slices contain the same values regardless of order (e.g. pretend the 2 slices are sets)
func setContentsEqual(a, b []string) bool {
	mkSet := func(s []string) map[string]bool {
		res := make(map[string]bool, len(s))
		for _, x := range s {
			res[x] = true
		}
		return res
	}
	sa := mkSet(a)
	sb := mkSet(b)
	return reflect.DeepEqual(sa, sb)
}

func loadConfig(t *testing.T) *config.Configuration {
	cfgFile, err := config.GetConfigAbsFilename("etc/dev/"+config.ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	c, err := config.LoadConfig(cfgFile)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
	return c
}
