package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/syncloud/store/machine"
	"github.com/syncloud/store/model"
	"github.com/syncloud/store/rest"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

type Index interface {
	Refresh() error
	Read(channel string) (map[string]*model.Snap, bool)
	Find(channel string, query string) *model.SearchResults
	Info(name string) *model.SearchResults
}

type IndexCache struct {
	indexByChannel map[string]map[string]*model.Snap
	lock           sync.RWMutex
	client         rest.Client
	baseUrl        string
	logger         *zap.Logger
}

func New(client rest.Client, baseUrl string, logger *zap.Logger) *IndexCache {
	return &IndexCache{
		client:         client,
		baseUrl:        baseUrl,
		logger:         logger,
		indexByChannel: make(map[string]map[string]*model.Snap),
	}
}

func (i *IndexCache) Find(channel string, query string) *model.SearchResults {
	apps, ok := i.Read(channel)
	if !ok {
		i.logger.Warn("no channel in the index", zap.String("channel", channel))
		return nil
	}
	results := &model.SearchResults{}
	for name, app := range apps {
		if query == "*" || query == "" || query == name {
			result := &model.SearchResult{
				Revision: model.SearchRevision{Channel: channel},
				Snap:     *app,
				Name:     app.Name,
				SnapID:   app.SnapID,
			}
			results.Results = append(results.Results, result)
		}
	}
	return results
}

func (i *IndexCache) Refresh() error {
	fmt.Println("refresh cache")
	channels := []string{"master", "rc", "stable"}
	for _, channel := range channels {
		index, err := i.downloadIndex(channel)
		if err != nil {
			return err
		}
		if index == nil {
			i.logger.Warn("index not found", zap.String("channel", channel))
			continue
		}
		i.WriteIndex(channel, index)
	}
	fmt.Println("refresh cache finished")
	return nil
}

func (i *IndexCache) downloadIndex(channel string) (map[string]*model.Snap, error) {
	resp, code, err := i.client.Get(fmt.Sprintf("%s/releases/%s/index-v2", i.baseUrl, channel))
	if err != nil {
		return nil, err
	}

	if code != 200 {
		return nil, nil
	}

	index, err := i.parseIndex(resp)
	if err != nil {
		return nil, err
	}
	apps := make(map[string]*model.Snap)
	for _, indexApp := range index {
		app, err := i.downloadAppInfo(indexApp, channel)
		if err != nil {
			return nil, err
		}
		if app == nil {
			i.logger.Info("not found", zap.String("app", indexApp.Name), zap.String("channel", channel))
			continue
		}
		apps[indexApp.Name] = app
	}

	return apps, nil
}

func (i *IndexCache) downloadAppInfo(app *model.App, channel string) (*model.Snap, error) {
	versionUrl := fmt.Sprintf("%s/releases/%s/%s.%s.version", i.baseUrl, channel, app.Name, machine.DPKGArch)
	i.logger.Info("version", zap.String("url", versionUrl))
	resp, code, err := i.client.Get(versionUrl)
	if err != nil {
		return nil, err
	}
	if code == 404 {
		return nil, nil
	}
	version := resp
	downloadUrl := fmt.Sprintf("%s/apps/%s_%s_%s.snap", i.baseUrl, app.Name, version, machine.DPKGArch)

	resp, _, err = i.client.Get(fmt.Sprintf("%s/apps/%s_%s_%s.snap.size", i.baseUrl, app.Name, version, machine.DPKGArch))
	if err != nil {
		return nil, err
	}
	size, err := strconv.ParseInt(resp, 10, 0)
	if err != nil {
		if channel == "stable" {
			i.logger.Warn("not valid size", zap.String("app", app.Name), zap.Error(err))
		}
		return nil, nil
	}

	resp, _, err = i.client.Get(fmt.Sprintf("%s/apps/%s_%s_%s.snap.sha384", i.baseUrl, app.Name, version, machine.DPKGArch))
	if err != nil {
		return nil, err
	}
	sha384Encoded := resp
	sha384, err := base64.RawURLEncoding.DecodeString(sha384Encoded)
	if err != nil {
		return nil, err
	}
	return app.ToInfo(version, size, fmt.Sprintf("%x", sha384), downloadUrl)
}

func (i *IndexCache) WriteIndex(channel string, index map[string]*model.Snap) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.indexByChannel[channel] = index
}

func (i *IndexCache) Read(channel string) (map[string]*model.Snap, bool) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	apps, ok := i.indexByChannel[channel]
	return apps, ok
}

func (i *IndexCache) Start() error {
	err := i.Refresh()
	if err != nil {
		i.logger.Error("error", zap.Error(err))
		return err
	}
	go func() {
		for range time.Tick(time.Minute * 60) {
			err := i.Refresh()
			if err != nil {
				i.logger.Error("error", zap.Error(err))
			}
		}
	}()
	return nil
}

func (i *IndexCache) parseIndex(resp string) (map[string]*model.App, error) {
	var index model.Index
	err := json.Unmarshal([]byte(resp), &index)
	if err != nil {
		i.logger.Error("cannot parse index response", zap.Error(err))
		return nil, err
	}

	apps := make(map[string]*model.App)

	for ind := range index.Apps {
		app := &model.App{
			Enabled: true,
		}
		err := json.Unmarshal(index.Apps[ind], app)
		if err != nil {
			return nil, err
		}
		if !app.Enabled {
			continue
		}
		i.logger.Info("index", zap.String("app", app.Name))
		apps[app.Name] = app

	}

	return apps, nil

}
