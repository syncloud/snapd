package store
import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/retry.v1"

	"github.com/snapcore/snapd/arch"
	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/client"
	"github.com/snapcore/snapd/httputil"
	"github.com/snapcore/snapd/jsonutil"
	"github.com/snapcore/snapd/logger"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/overlord/auth"
	"github.com/snapcore/snapd/progress"
	"github.com/snapcore/snapd/release"
	"github.com/snapcore/snapd/snap"

	"golang.org/x/net/context/ctxhttp"
	"io/ioutil"
	"github.com/snapcore/snapd/dirs"
)

// Store represents the ubuntu snap store
type SyncloudStore struct {
	store *Store
  url string
}

// New creates a new Store with the given access configuration and for given the store id.
func NewSyncloudStore(cfg *Config, dauthCtx DeviceAndAuthContext) *SyncloudStore {
	return &SyncloudStore{
		store: New(cfg, dauthCtx),
   url: "http://apps.syncloud.org",
	}
}

func (s *SyncloudStore) SetCacheDownloads(fileCount int) {
	s.store.SetCacheDownloads(fileCount)
}

// SnapInfo returns the snap.Info for the store-hosted snap matching the given spec, or an error.
func (s *SyncloudStore) SnapInfo(ctx context.Context, snapSpec SnapSpec, user *auth.UserState) (*snap.Info, error) {

	storeInfo := &storeInfo {}
	storeInfo.Name = snapSpec.Name
	//storeInfo.ChannelMap =

	info, err := infoFromStoreInfo(storeInfo)
	if err != nil {
		return nil, err
	}

	//channel := s.parseChannel(snapSpec.Channel)
  channel := "stable"
	resp, err := s.downloadIndex(channel, user)
	if err != nil {
		return nil, err
	}

	apps, err := parseIndex(resp)
	if err != nil {
		return nil, err
	}

	version, err := s.downloadVersion(channel, snapSpec.Name)
	if err != nil {
		return nil, ErrSnapNotFound
	}

	info := apps[snapSpec.Name].toInfo(s.url, channel, version)

	return info, nil
}

func (s *SyncloudStore) downloadIndex(channel string, user *auth.UserState) (string, error) {
	reqOptions := &requestOptions{
		Method: "GET",
		URL:    url.Parse(s.url + "/releases/" + channel + "/index-v2"),
		Accept: halJsonContentType,
	}

	resp, err := s.retryRequestString(context.TODO(), reqOptions, user)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func (s *SyncloudStore) retryRequestString(ctx context.Context, reqOptions *requestOptions, user *auth.UserState) (string, error) {
	var reply string
	_, err := httputil.RetryRequest(reqOptions.URL.String(), func() (*http.Response, error) {
		return s.store.doRequest(ctx, s.client, reqOptions, user)
	}, func(resp *http.Response) error {
		resp1, err1 := decodeStringBody(resp)
		reply = resp1
		return err1
	}, defaultRetryStrategy)

	if err != nil {
		return "", fmt.Errorf("%v, url: %s", err, reqOptions.URL.String())
	}

	return reply, err
}

func parseIndex(resp string) (map[string]*App, error) {
	var index Index
	err := json.Unmarshal([]byte(resp), &index)
	if err != nil {
		return nil, err
	}

	apps := make(map[string]*App)

	for i, _ := range index.Apps {
		app := &App{
			Enabled: true,
		}
		err := json.Unmarshal([]byte(index.Apps[i]), app)
		if err != nil {
			return nil, err
		}
		if (!app.Enabled) {
			continue
		}
		apps[app.Name] = app

	}

	return apps, nil

}

func (s *SyncloudStore) downloadVersion(channel string, name string, user *auth.UserState) (string, error) {

	reqOptions := &requestOptions{
		Method: "GET",
		URL:    url.Parse(s.url + "/releases/" + channel + "/" + name + ".version"),
		Accept: halJsonContentType,
	}

	version, err := s.retryRequestString(context.TODO(), reqOptions, user)
	if err != nil {
		return "", err
	}

	return version, nil

}
