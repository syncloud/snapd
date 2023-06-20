package pkg

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net"
	"os"
)

type Refresher interface {
	RefreshCache() error
}

type Api struct {
	refresher Refresher
	echo      *echo.Echo
}

func NewApi(refresher Refresher) *Api {
	return &Api{
		echo:      echo.New(),
		refresher: refresher,
	}
}

func (a *Api) Start() error {
	a.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	a.echo.Use(middleware.Recover())

	a.echo.POST("/refresh", a.Refresh)

	_ = os.RemoveAll(InternalApi)
	l, err := net.Listen("unix", InternalApi)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}
	a.echo.Listener = l
	return a.echo.Start("")
}

func (a *Api) Refresh(_ echo.Context) error {
	return a.refresher.RefreshCache()
}
