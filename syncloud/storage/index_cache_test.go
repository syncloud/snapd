package storage

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/store/log"
	"testing"
)

type Response struct {
	body string
	code int
	err  error
}

func OK(body string) Response {
	return Response{body: body, code: 200}
}

type ClientStub struct {
	response map[string]Response
}

func (c *ClientStub) Get(url string) (string, int, error) {
	fmt.Println(url)
	response, ok := c.response[url]
	if !ok {
		return "", 404, nil
	}
	return response.body, response.code, response.err
}

func TestIndexCache_Refresh_EmptySize(t *testing.T) {

	client := &ClientStub{
		response: map[string]Response{
			"http://localhost/releases/master/index-v2": OK(`
{
  "apps" : [
    {
      "name" : "Platform",
      "id" : "platform",
      "required" : true,
      "ui": false
    }
  ]
}
`),
			"http://localhost/releases/master/platform.amd64.version": OK("123"),
			"http://localhost/apps/platform__amd64.snap.size":         OK(""),
		},
	}

	cache := New(client, "http://localhost", log.Default())
	err := cache.Refresh()
	assert.NoError(t, err)

	cache.Read("test")

}
