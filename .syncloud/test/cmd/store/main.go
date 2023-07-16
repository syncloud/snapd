package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"net/http"
)

func main() {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.Recover())
	e.GET("/api/v1/snaps/sections", Sections)
	e.GET("/api/v1/snaps/names", Names)
	e.POST("/v2/snaps/refresh", Refresh)
	e.GET("/v2/assertions/snap-revision/:key", SnapRevision)
	e.GET("/v2/assertions/snap-declaration/:series/:snap-id", SnapDeclaration)
	e.GET("/v2/assertions/account-key/:key", AccountKey)
	e.GET("/v2/snaps/find", Find)
	e.GET("/v2/snaps/info/:name", Info)

	err := e.Start(":80")
	if err != nil {
		panic(err)
	}
}

func Sections(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/hal+json")
	return c.String(http.StatusOK, `{
  "_embedded": {
    "clickindex:sections": [
      {
        "name": "apps"
      }
    ]
  }
}
`)
}

func Names(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/hal+json")
	return c.String(http.StatusOK, `
{
  "_embedded": {
    "clickindex:package": [
      
    ]
  }
}`)
}

func Refresh(c echo.Context) error {
	req, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	arch := c.Request().Header.Get("Syncloud-Architecture")
	fmt.Printf("refresh arch: %s\n", arch)
	fmt.Printf("request: %s\n", string(req))

	//TODO
	return c.String(http.StatusOK, `
{
  "results": [
    {
      "result": "fetch-assertions",
      "instance-key": "",
      "key": "AAA",
      "assertion-stream-urls": null
    }
  ]
}

`)

}

func SnapRevision(c echo.Context) error {
	key := c.Param("key")
	fmt.Printf("snap revision key %s", key)

	return c.String(http.StatusOK, `

`)
}
func SnapDeclaration(c echo.Context) error {
	key := c.Param("key")
	fmt.Printf("snap revision key %s", key)

	return c.String(http.StatusOK, `

`)
}

func Find(c echo.Context) error {
	channel := c.QueryParam("channel")
	fmt.Printf("channel %s", channel)
	query := c.QueryParam("q")
	fmt.Printf("query %s", query)
	architecture := c.QueryParam("architecture")
	fmt.Printf("architecture %s", architecture)

	return c.String(http.StatusOK, `

{
  "results": [
    {
      "revision": {
        "channel": "stable"
      },
      "snap": {
        "snap-id": "users.272",
        "name": "users",
        "summary": "Users",
        "version": "272",
        "type": "app",
        "architectures": [
          "amd64"
        ],
        "revision": 272,
        "download": {
          "sha3-384": "a8a204614e4504bc8cb539f456d75875f88077fe76baaa19fcf476a9518a25f51cf575f7e92bf04caec2aceee21c07cb",
          "size": 197459968,
          "url": "http://apps.syncloud.org/apps/users_272_amd64.snap"
        },
        "media": [
          {
            "type": "icon",
            "url": "users-128.png",
            "width": 0,
            "height": 0
          }
        ]
      },
      "name": "users",
      "snap-id": "users.272"
    }
  ],
  "error-list": null
}

`)
}

func AccountKey(c echo.Context) error {
	key := c.Param("key")
	fmt.Printf("snap revision key %s", key)

	return c.String(http.StatusOK, `

`)
}

func Info(c echo.Context) error {
	channel := c.QueryParam("channel")
	fmt.Printf("channel %s", channel)
	query := c.QueryParam("q")
	fmt.Printf("query %s", query)
	architecture := c.QueryParam("architecture")
	fmt.Printf("architecture %s", architecture)

	return c.String(http.StatusOK, `

{
  "channel-map": [
    {
      "snap-id": "users.272",
      "name": "users",
      "summary": "Users",
      "version": "272",
      "type": "app",
      "architectures": [
        "amd64"
      ],
      "revision": 272,
      "download": {
        "sha3-384": "a8a204614e4504bc8cb539f456d75875f88077fe76baaa19fcf476a9518a25f51cf575f7e92bf04caec2aceee21c07cb",
        "size": 197459968,
        "url": "http://apps.syncloud.org/apps/users_272_amd64.snap"
      },
      "media": [
        {
          "type": "icon",
          "url": "users-128.png",
          "width": 0,
          "height": 0
        }
      ],
      "channel": {
        "architecture": "amd64",
        "name": "stable",
        "risk": "stable",
        "track": "",
        "released-at": "0001-01-01T00:00:00Z"
      }
    },
    {
      "snap-id": "users.270",
      "name": "users",
      "summary": "Users",
      "version": "270",
      "type": "app",
      "architectures": [
        "amd64"
      ],
      "revision": 270,
      "download": {
        "sha3-384": "bca0a10bd12a30ef259c5d1013bbcff8067e3b0eabdf55170e2685f7fda1c759696662fe9d5df18b29af93923345f013",
        "size": 197459968,
        "url": "http://apps.syncloud.org/apps/users_270_amd64.snap"
      },
      "media": [
        {
          "type": "icon",
          "url": "users-128.png",
          "width": 0,
          "height": 0
        }
      ],
      "channel": {
        "architecture": "amd64",
        "name": "master",
        "risk": "master",
        "track": "",
        "released-at": "0001-01-01T00:00:00Z"
      }
    },
    {
      "snap-id": "users.272",
      "name": "users",
      "summary": "Users",
      "version": "272",
      "type": "app",
      "architectures": [
        "amd64"
      ],
      "revision": 272,
      "download": {
        "sha3-384": "a8a204614e4504bc8cb539f456d75875f88077fe76baaa19fcf476a9518a25f51cf575f7e92bf04caec2aceee21c07cb",
        "size": 197459968,
        "url": "http://apps.syncloud.org/apps/users_272_amd64.snap"
      },
      "media": [
        {
          "type": "icon",
          "url": "users-128.png",
          "width": 0,
          "height": 0
        }
      ],
      "channel": {
        "architecture": "amd64",
        "name": "rc",
        "risk": "rc",
        "track": "",
        "released-at": "0001-01-01T00:00:00Z"
      }
    }
  ],
  "snap": {
    "snap-id": "users.272",
    "name": "users",
    "summary": "Users",
    "version": "272",
    "type": "app",
    "architectures": [
      "amd64"
    ],
    "revision": 272,
    "download": {
      "sha3-384": "a8a204614e4504bc8cb539f456d75875f88077fe76baaa19fcf476a9518a25f51cf575f7e92bf04caec2aceee21c07cb",
      "size": 197459968,
      "url": "http://apps.syncloud.org/apps/users_272_amd64.snap"
    },
    "media": [
      {
        "type": "icon",
        "url": "users-128.png",
        "width": 0,
        "height": 0
      }
    ]
  },
  "name": "users",
  "snap-id": "users.272"
}

`)
}
