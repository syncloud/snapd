package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/snapcore/snapd/arch"
	"github.com/snapcore/snapd/httputil"
	"github.com/snapcore/snapd/jsonutil/safejson"
	"github.com/snapcore/snapd/logger"
	"github.com/snapcore/snapd/overlord/auth"
	"github.com/snapcore/snapd/snap"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var SHA3_384 = "hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR"

// Store represents the ubuntu snap store
type SyncloudStore struct {
	store *Store
	url   string
}

type App struct {
	Name     string          `json:"id"`
	Summary  safejson.String `json:"name"`
	Icon     string          `json:"icon,omitempty"`
	Enabled  bool            `json:"enabled,omitempty"`
	Required bool            `json:"required"`
}

func constructSnapId(name string, version string) string {
	return fmt.Sprintf("%s.%s", name, version)
}

func deconstructSnapId(snapId string) (string, string) {
	parts := strings.Split(snapId, ".")
	return parts[0], parts[1]
}

func (a *App) toInfo(baseUrl string, channel string, version string) *snap.Info {

	appType := snap.TypeApp
	if a.Required {
		appType = snap.TypeBase
	}

	revision, _ := strconv.Atoi(version)
	//	if err != nil {
	//		return nil, fmt.Errorf("Unable to get revision: %s", err)
	//	}
	snapId := constructSnapId(a.Name, version)
	logger.Noticef("snapid: %s", snapId)

	details := snapDetails{
		SnapID:           snapId,
		Name:             a.Name,
		Summary:          a.Summary,
		Version:          version,
		Type:             appType,
		Architectures:    []string{"amd64", "armhf"},
		Revision:         revision,
		Channel:          channel,
		AnonDownloadURL:  fmt.Sprintf("%s/apps/%s_%s_%s.snap", baseUrl, a.Name, version, arch.DpkgArchitecture()),
		DownloadSha3_384: SHA3_384,
	}

	return infoFromRemote(&details)
}

type Index struct {
	Apps []json.RawMessage `json:"apps"`
}

// New creates a new Store with the given access configuration and for given the store id.
func NewSyncloudStore(cfg *Config, dauthCtx DeviceAndAuthContext) *SyncloudStore {
	return &SyncloudStore{
		store: New(cfg, dauthCtx),
		url:   "http://apps.syncloud.org",
	}
}

func (s *SyncloudStore) SetCacheDownloads(fileCount int) {
	s.store.SetCacheDownloads(fileCount)
}

// SnapInfo returns the snap.Info for the store-hosted snap matching the given spec, or an error.
func (s *SyncloudStore) SnapInfo(_ context.Context, snapSpec SnapSpec, user *auth.UserState) (*snap.Info, error) {

	//storeInfo := &storeInfo {}
	//storeInfo.Name = snapSpec.Name
	//storeInfo.ChannelMap =

	//info, err := infoFromStoreInfo(storeInfo)
	//if err != nil {
	//	return nil, err
	//}

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

	version, err := s.downloadVersion(channel, snapSpec.Name, user)
	if err != nil {
		return nil, ErrSnapNotFound
	}

	info := apps[snapSpec.Name].toInfo(s.url, channel, version)

	return info, nil
}

func (s *SyncloudStore) downloadIndex(channel string, user *auth.UserState) (string, error) {
	indexUrl, err := url.Parse(s.url + "/releases/" + channel + "/index-v2")
	if err != nil {
		return "", err
	}
	reqOptions := &requestOptions{
		Method: "GET",
		URL:    indexUrl,
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
		return s.store.doRequest(ctx, s.store.client, reqOptions, user)
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

func decodeStringBody(resp *http.Response) (string, error) {
	ok := resp.StatusCode == 200 || resp.StatusCode == 201
	if !ok {
		return "", fmt.Errorf("store is not responding, code: %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func parseIndex(resp string) (map[string]*App, error) {
	var index Index
	err := json.Unmarshal([]byte(resp), &index)
	if err != nil {
		return nil, err
	}

	apps := make(map[string]*App)

	for i := range index.Apps {
		app := &App{
			Enabled: true,
		}
		err := json.Unmarshal(index.Apps[i], app)
		if err != nil {
			return nil, err
		}
		if !app.Enabled {
			continue
		}
		apps[app.Name] = app

	}

	return apps, nil

}

func (s *SyncloudStore) downloadVersion(channel string, name string, user *auth.UserState) (string, error) {

	versionUrl, err := url.Parse(s.url + "/releases/" + channel + "/" + name + ".version")
	if err != nil {
		return "", err
	}
	reqOptions := &requestOptions{
		Method: "GET",
		URL:    versionUrl,
		Accept: halJsonContentType,
	}

	version, err := s.retryRequestString(context.TODO(), reqOptions, user)
	if err != nil {
		return "", err
	}

	return version, nil

}
