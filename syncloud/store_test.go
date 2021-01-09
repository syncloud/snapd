package syncloud

import (
	"testing"
	. "gopkg.in/check.v1"
	"net/url"
)

func TestStore(t *testing.T) { TestingT(t) }

type configTestSuite struct{}

var _ = Suite(&configTestSuite{})

func (suite *configTestSuite) TestParse(c *C) {

	baseURL, _ := url.Parse("http://apps.syncloud.org")

	snaps, _ := parseIndex(`{
	  "apps" : [
	    {
	      "name" : "app1",
	      "id" : "app1",
	      "required" : false,
	      "ui": true,
	      "icon": "app1-128.png",
	      "description": "desc1",
	      "enabled": true
	    },
	    {
	      "name" : "app2",
	      "id" : "app2",
	      "required" : false,
	      "ui": true,
	      "icon": "app2-128.png",
	      "description": "desc2",
	      "enabled": false
	    },
	    {
	      "name" : "app3",
	      "id" : "app3",
	      "required" : true,
	      "ui": true,
	      "icon": "app3-128.png",
	      "description": "desc3"
	    }
	  ]
	}
	`, baseURL)

	c.Assert(len(snaps), Equals, 2)

	c.Assert(snaps["app1"].Name, Equals, "app1")
	c.Assert(snaps["app1"].Required, Equals, false)

	c.Assert(snaps["app3"].Name, Equals, "app3")
	c.Assert(snaps["app3"].Required, Equals, true)
}

