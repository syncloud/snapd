package storage

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/store/log"
	"github.com/syncloud/store/model"
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

func TestIndexCache_Find(t *testing.T) {

	cache := &IndexCache{
		indexByChannel: map[string]map[string]*model.Snap{
			"channel1": {
				"app1": &model.Snap{
					Name: "app1",
				},
			},
			"channel2": {
				"app2": &model.Snap{
					Name: "app2",
				},
			},
		},
		logger: log.Default(),
	}
	results := cache.Find("channel1", "")
	assert.Equal(t, 1, len(results.Results))
	assert.Equal(t, "app1", results.Results[0].Name)
}

func TestIndexCache_Find_PopulateChannel(t *testing.T) {

	cache := &IndexCache{
		indexByChannel: map[string]map[string]*model.Snap{
			"channel": {
				"": &model.Snap{
					Name: "app",
				},
			},
		},
		logger: log.Default(),
	}
	results := cache.Find("channel", "")
	assert.Equal(t, "channel", results.Results[0].Revision.Channel)
}
