// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

// Package store has support to use the Ubuntu Store for querying and downloading of snaps, and the related services.
package syncloud

import (
	"bytes"
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
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/snapcore/snapd/arch"
	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/asserts/assertstest"
	"github.com/snapcore/snapd/httputil"
	"github.com/snapcore/snapd/logger"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/overlord/auth"
	"github.com/snapcore/snapd/progress"
	"github.com/snapcore/snapd/release"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/store"

	"github.com/snapcore/snapd/dirs"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
	"gopkg.in/retry.v1"
	"io/ioutil"
)

// TODO: better/shorter names are probably in order once fewer legacy places are using this

const (
	// halJsonContentType is the default accept value for store requests
	halJsonContentType = "application/hal+json"
	// jsonContentType is for store enpoints that don't support HAL
	jsonContentType = "application/json"
	// UbuntuCoreWireProtocol is the protocol level we support when
	// communicating with the store. History:
	//  - "1": client supports squashfs snaps
	UbuntuCoreWireProtocol = "1"
)

// the LimitTime should be slightly more than 3 times of our http.Client
// Timeout value
var defaultRetryStrategy = retry.LimitCount(5, retry.LimitTime(38*time.Second,
	retry.Exponential{
		Initial: 300 * time.Millisecond,
		Factor:  2.5,
	},
))

func infoFromRemote(d *snapDetails) *snap.Info {
	info := &snap.Info{}
	info.Architectures = d.Architectures
	info.Type = d.Type
	info.Version = d.Version
	info.Epoch = d.Epoch
	info.RealName = d.Name
	info.SnapID = d.SnapID
	info.Revision = snap.R(d.Revision)
	info.EditedTitle = d.Title
	info.EditedSummary = d.Summary
	info.EditedDescription = d.Description
	info.PublisherID = d.DeveloperID
	info.Publisher = d.Developer
	info.Channel = d.Channel
	info.Sha3_384 = d.DownloadSha3_384
	info.Size = d.DownloadSize
	info.IconURL = d.IconURL
	info.AnonDownloadURL = d.AnonDownloadURL
	info.DownloadURL = d.DownloadURL
	info.Prices = d.Prices
	info.Private = d.Private
	info.Paid = len(info.Prices) > 0
	info.Confinement = snap.ConfinementType(d.Confinement)
	info.Contact = d.Contact
	info.License = d.License
	info.Base = d.Base

	deltas := make([]snap.DeltaInfo, len(d.Deltas))
	for i, d := range d.Deltas {
		deltas[i] = snap.DeltaInfo{
			FromRevision:    d.FromRevision,
			ToRevision:      d.ToRevision,
			Format:          d.Format,
			AnonDownloadURL: d.AnonDownloadURL,
			DownloadURL:     d.DownloadURL,
			Size:            d.Size,
			Sha3_384:        d.Sha3_384,
		}
	}
	info.Deltas = deltas

	screenshots := make([]snap.ScreenshotInfo, 0, len(d.ScreenshotURLs))
	for _, url := range d.ScreenshotURLs {
		screenshots = append(screenshots, snap.ScreenshotInfo{
			URL: url,
		})
	}
	info.Screenshots = screenshots
	// FIXME: once the store sends "contact" for everything, remove
	//        the "SupportURL" part of the if
	if info.Contact == "" {
		info.Contact = d.SupportURL
	}

	// fill in the tracks data
	if len(d.ChannelMapList) > 0 {
		info.Channels = make(map[string]*snap.ChannelSnapInfo)
		info.Tracks = make([]string, len(d.ChannelMapList))
		for i, cm := range d.ChannelMapList {
			info.Tracks[i] = cm.Track
			for _, ch := range cm.SnapDetails {
				// nothing in this channel
				if ch.Info == "" {
					continue
				}
				var k string
				if strings.HasPrefix(ch.Channel, cm.Track) {
					k = ch.Channel
				} else {
					k = fmt.Sprintf("%s/%s", cm.Track, ch.Channel)
				}
				info.Channels[k] = &snap.ChannelSnapInfo{
					Revision:    snap.R(ch.Revision),
					Confinement: snap.ConfinementType(ch.Confinement),
					Version:     ch.Version,
					Channel:     ch.Channel,
					Epoch:       ch.Epoch,
					Size:        ch.DownloadSize,
				}
			}
		}
	}

	return info
}

// Config represents the configuration to access the snap store
type Config struct {
	// Store API base URLs. The assertions url is only separate because it can
	// be overridden by its own env var.
	StoreBaseURL      *url.URL
	AssertionsBaseURL *url.URL

	// StoreID is the store id used if we can't get one through the AuthContext.
	StoreID string

	Architecture string
	Series       string

	DetailFields []string
	DeltaFormat  string

	// CacheDownloads is the number of downloads that should be cached
	CacheDownloads int
}

// setBaseURL updates the store API's base URL in the Config. Must not be used
// to change active config.
func (cfg *Config) SetBaseURL(storeBaseURI *url.URL) error {
	cfg.StoreBaseURL = storeBaseURI

	return nil
}

// Store represents the ubuntu snap store
type Store struct {
	cfg *Config

	architecture string
	series       string

	noCDN bool

	fallbackStoreID string

	detailFields []string
	deltaFormat  string
	// reused http client
	client *http.Client

	authContext auth.AuthContext

	mu                sync.Mutex
	suggestedCurrency string

	cacher store.DownloadCache
}

func respToError(resp *http.Response, msg string) error {
	tpl := "cannot %s: got unexpected HTTP status code %d via %s to %q"
	if oops := resp.Header.Get("X-Oops-Id"); oops != "" {
		tpl += " [%s]"
		return fmt.Errorf(tpl, msg, resp.StatusCode, resp.Request.Method, resp.Request.URL, oops)
	}

	return fmt.Errorf(tpl, msg, resp.StatusCode, resp.Request.Method, resp.Request.URL)
}

func getStructFields(s interface{}) []string {
	st := reflect.TypeOf(s)
	num := st.NumField()
	fields := make([]string, 0, num)
	for i := 0; i < num; i++ {
		tag := st.Field(i).Tag.Get("json")
		idx := strings.IndexRune(tag, ',')
		if idx > -1 {
			tag = tag[:idx]
		}
		if tag != "" {
			fields = append(fields, tag)
		}
	}

	return fields
}

// Extend a base URL with additional unescaped paths.  (url.Parse handles
// resolving relative links, which isn't quite what we want: that goes wrong if
// the base URL doesn't end with a slash.)
func urlJoin(base *url.URL, paths ...string) *url.URL {
	if len(paths) == 0 {
		return base
	}
	url := *base
	url.RawQuery = ""
	paths = append([]string{strings.TrimSuffix(url.Path, "/")}, paths...)
	url.Path = strings.Join(paths, "/")
	return &url
}

// endpointURL clones a base URL and updates it with optional path and query.
func endpointURL(base *url.URL, path string, query url.Values) *url.URL {
	u := *base
	if path != "" {
		u.Path = strings.TrimSuffix(u.Path, "/") + "/" + strings.TrimPrefix(path, "/")
		u.RawQuery = ""
	}
	if len(query) != 0 {
		u.RawQuery = query.Encode()
	}
	return &u
}

// storeURL returns the base store URL, derived from either the given API URL
// or an env var override.
func storeURL(api *url.URL) (*url.URL, error) {
	return api, nil
}

var defaultConfig = Config{}
var syncloudAppsBaseURL *url.URL
var privkey asserts.PrivateKey

var syncloudPrivKey = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: GnuPG v1

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
97WP6Yv2
=gJ0v
-----END PGP PRIVATE KEY BLOCK-----
`

var SHA3_384 = "hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR"

func init() {
	privkey, _ = assertstest.ReadPrivKey(syncloudPrivKey)

	syncloudAppsBaseURL, _ = url.Parse("http://apps.syncloud.org")
	//defaultConfig.SearchURI = urlJoin(storeBaseURI, "api/v1/snaps/search")

	err := defaultConfig.SetBaseURL(syncloudAppsBaseURL)
	if err != nil {
		panic(err)
	}
}

type searchResults struct {
	Payload struct {
		Packages []*snapDetails `json:"clickindex:package"`
	} `json:"_embedded"`
}

type sectionResults struct {
	Payload struct {
		Sections []struct{ Name string } `json:"clickindex:sections"`
	} `json:"_embedded"`
}

// The fields we are interested in
var detailFields = getStructFields(snapDetails{})

// The fields we are interested in for snap.ChannelSnapInfos
var channelSnapInfoFields = getStructFields(channelSnapInfoDetails{})

// The default delta format if not configured.
var defaultSupportedDeltaFormat = "xdelta3"

// New creates a new Store with the given access configuration and for given the store id.
func New(cfg *Config, authContext auth.AuthContext) *Store {
	if cfg == nil {
		cfg = &defaultConfig
	}

	fields := cfg.DetailFields
	if fields == nil {
		fields = detailFields
	}

	architecture := arch.UbuntuArchitecture()
	if cfg.Architecture != "" {
		architecture = cfg.Architecture
	}

	series := release.Series
	if cfg.Series != "" {
		series = cfg.Series
	}

	deltaFormat := cfg.DeltaFormat
	if deltaFormat == "" {
		deltaFormat = defaultSupportedDeltaFormat
	}

	store := &Store{
		cfg:             cfg,
		series:          series,
		architecture:    architecture,
		noCDN:           osutil.GetenvBool("SNAPPY_STORE_NO_CDN"),
		fallbackStoreID: cfg.StoreID,
		detailFields:    fields,
		authContext:     authContext,
		deltaFormat:     deltaFormat,

		client: httputil.NewHTTPClient(&httputil.ClientOpts{
			Timeout:    10 * time.Second,
			MayLogBody: true,
		}),
	}
	store.SetCacheDownloads(cfg.CacheDownloads)

	return store
}

// API endpoint paths
const (
	// see https://wiki.ubuntu.com/AppStore/Interfaces/ClickPackageIndex
	// XXX: Repeating "api/" here is cumbersome, but the next generation
	// of store APIs will probably drop that prefix (since it now
	// duplicates the hostname), and we may want to switch to v2 APIs
	// one at a time; so it's better to consider that as part of
	// individual endpoint paths.
	searchEndpPath      = "api/v1/snaps/search"
	detailsEndpPath     = "api/v1/snaps/details"
	bulkEndpPath        = "api/v1/snaps/metadata"
	ordersEndpPath      = "api/v1/snaps/purchases/orders"
	buyEndpPath         = "api/v1/snaps/purchases/buy"
	customersMeEndpPath = "api/v1/snaps/purchases/customers/me"
	sectionsEndpPath    = "api/v1/snaps/sections"
	commandsEndpPath    = "api/v1/snaps/names"

	deviceNonceEndpPath   = "api/v1/snaps/auth/nonces"
	deviceSessionEndpPath = "api/v1/snaps/auth/sessions"

	assertionsPath = "api/v1/snaps/assertions"
)

func (s *Store) baseURL(defaultURL *url.URL) *url.URL {
	u := defaultURL
	if s.authContext != nil {
		var err error
		_, u, err = s.authContext.ProxyStoreParams(defaultURL)
		if err != nil {
			logger.Debugf("cannot get proxy store parameters from state: %v", err)
		}
	}
	if u != nil {
		return u
	}
	return defaultURL
}

func (s *Store) endpointURL(p string, query url.Values) *url.URL {
	return endpointURL(s.baseURL(s.cfg.StoreBaseURL), p, query)
}

// LoginUser logs user in the store and returns the authentication macaroons.
func LoginUser(username, password, otp string) (string, string, error) {
	return "macaroon", "discharge", nil
}

// hasStoreAuth returns true if given user has store macaroons setup
func hasStoreAuth(user *auth.UserState) bool {
	return user != nil && user.StoreMacaroon != ""
}

// requestOptions specifies parameters for store requests.
type requestOptions struct {
	Method       string
	URL          *url.URL
	Accept       string
	ContentType  string
	ExtraHeaders map[string]string
	Data         []byte
}

func (r *requestOptions) addHeader(k, v string) {
	if r.ExtraHeaders == nil {
		r.ExtraHeaders = make(map[string]string)
	}
	r.ExtraHeaders[k] = v
}

func cancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func decodeJSONBody(resp *http.Response, success interface{}, failure interface{}) error {
	ok := (resp.StatusCode == 200 || resp.StatusCode == 201)
	// always decode on success; decode failures only if body is not empty
	if !ok && resp.ContentLength == 0 {
		return nil
	}
	result := success
	if !ok {
		result = failure
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (s *Store) retryRequestDecodeJSON(ctx context.Context, reqOptions *requestOptions, success interface{}, failure interface{}) (resp *http.Response, err error) {
	return httputil.RetryRequest(reqOptions.URL.String(), func() (*http.Response, error) {
		return s.doRequest(ctx, s.client, reqOptions)
	}, func(resp *http.Response) error {
		return decodeJSONBody(resp, success, failure)
	}, defaultRetryStrategy)
}

func decodeStringBody(resp *http.Response) (string, error) {
	ok := (resp.StatusCode == 200 || resp.StatusCode == 201)
	if !ok {
		return "", fmt.Errorf("store is not responding, code: %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func (s *Store) retryRequestString(ctx context.Context, reqOptions *requestOptions) (string, error) {
	logger.Noticef("download: %v", reqOptions.URL.String())

	var reply string
	_, err := httputil.RetryRequest(reqOptions.URL.String(), func() (*http.Response, error) {
		return s.doRequest(ctx, s.client, reqOptions)
	}, func(resp *http.Response) error {
		resp1, err1 := decodeStringBody(resp)
		reply = resp1
		return err1
	}, defaultRetryStrategy)

	if err != nil {
		err := fmt.Errorf("%v, url: %s", err, reqOptions.URL.String())
		logger.Noticef("download failed: %v", err)
		return "", err
	}

	return reply, err
}

func (s *Store) doRequest(ctx context.Context, client *http.Client, reqOptions *requestOptions) (*http.Response, error) {
	req, err := s.newRequest(reqOptions)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	if ctx != nil {
		resp, err = ctxhttp.Do(ctx, client, req)
	} else {
		resp, err = client.Do(req)
	}
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (s *Store) newRequest(reqOptions *requestOptions) (*http.Request, error) {
	var body io.Reader
	if reqOptions.Data != nil {
		body = bytes.NewBuffer(reqOptions.Data)
	}

	req, err := http.NewRequest(reqOptions.Method, reqOptions.URL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", httputil.UserAgent())
	req.Header.Set("Accept", reqOptions.Accept)

	if reqOptions.ContentType != "" {
		req.Header.Set("Content-Type", reqOptions.ContentType)
	}

	for header, value := range reqOptions.ExtraHeaders {
		req.Header.Set(header, value)
	}

	return req, nil
}

func (s *Store) extractSuggestedCurrency(resp *http.Response) {
	suggestedCurrency := resp.Header.Get("X-Suggested-Currency")

	if suggestedCurrency != "" {
		s.mu.Lock()
		s.suggestedCurrency = suggestedCurrency
		s.mu.Unlock()
	}
}

type ordersResult struct {
	Orders []*order `json:"orders"`
}

type order struct {
	SnapID          string `json:"snap_id"`
	Currency        string `json:"currency"`
	Amount          string `json:"amount"`
	State           string `json:"state"`
	RefundableUntil string `json:"refundable_until"`
	PurchaseDate    string `json:"purchase_date"`
}

// decorateOrders sets the MustBuy property of each snap in the given list according to the user's known orders.
func (s *Store) decorateOrders(snaps []*snap.Info, user *auth.UserState) error {
	// Mark every non-free snap as must buy until we know better.
	hasPriced := false
	for _, info := range snaps {
		if info.Paid {
			info.MustBuy = true
			hasPriced = true
		}
	}

	if user == nil {
		return nil
	}

	if !hasPriced {
		return nil
	}

	var err error

	reqOptions := &requestOptions{
		Method: "GET",
		URL:    s.endpointURL(ordersEndpPath, nil),
		Accept: jsonContentType,
	}
	var result ordersResult
	resp, err := s.retryRequestDecodeJSON(context.TODO(), reqOptions, &result, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == 401 {
		// TODO handle token expiry and refresh
		return ErrInvalidCredentials
	}
	if resp.StatusCode != 200 {
		return respToError(resp, "obtain known orders from store")
	}

	// Make a map of the IDs of bought snaps
	bought := make(map[string]bool)
	for _, order := range result.Orders {
		bought[order.SnapID] = true
	}

	for _, info := range snaps {
		info.MustBuy = mustBuy(info.Paid, bought[info.SnapID])
	}

	return nil
}

// mustBuy determines if a snap requires a payment, based on if it is non-free and if the user has already bought it
func mustBuy(paid bool, bought bool) bool {
	if !paid {
		// If the snap is free, then it doesn't need buying
		return false
	}

	return !bought
}

// A SnapSpec describes a single snap wanted from SnapInfo
type SnapSpec struct {
	Name    string
	Channel string
	// AnyChannel can be set to query for any revision independent of channel
	AnyChannel bool
	// Revision can be set to query for an exact revision
	Revision snap.Revision
}

func (s *Store) downloadAppInfo(app *App, channel string, name string) (*snap.Info, error) {
	version, err := s.retryRequestString(context.Background(), &requestOptions{
		Method: "GET",
		URL:    urlJoin(s.cfg.StoreBaseURL, "releases", channel, fmt.Sprintf("%s.%s.version", app.Name, arch.UbuntuArchitecture())),
		Accept: halJsonContentType,
	})
	if err != nil {
		return nil, err
	}
	downloadUrl := fmt.Sprintf("%s/apps/%s_%s_%s.snap", s.cfg.StoreBaseURL, app.Name, version, arch.UbuntuArchitecture())

	sizeString, err := s.retryRequestString(context.Background(), &requestOptions{
		Method: "GET",
		URL:    urlJoin(s.cfg.StoreBaseURL, "apps", fmt.Sprintf("%s_%s_%s.snap.size", app.Name, version, arch.UbuntuArchitecture())),
		Accept: halJsonContentType,
	})
	if err != nil {
		return nil, err
	}
	size, err := strconv.ParseInt(sizeString, 10, 0)
	if err != nil {
		return nil, err
	}

	sha384, err := s.retryRequestString(context.Background(), &requestOptions{
		Method: "GET",
		URL:    urlJoin(s.cfg.StoreBaseURL, "apps", fmt.Sprintf("%s_%s_%s.snap.sha384", app.Name, version, arch.UbuntuArchitecture())),
		Accept: halJsonContentType,
	})
	if err != nil {
		return nil, err
	}

	return app.toInfo(channel, version, size, sha384, downloadUrl)
}

func (s *Store) downloadVersion(channel string, name string) (string, error) {

	version, err := s.retryRequestString(context.Background(), &requestOptions{
		Method: "GET",
		URL:    urlJoin(s.cfg.StoreBaseURL, "releases", channel, name+".version"),
		Accept: halJsonContentType,
	})
	if err != nil {
		return "", err
	}

	return version, nil

}

// SnapInfo returns the snap.Info for the store-hosted snap matching the given spec, or an error.
func (s *Store) SnapInfo(snapSpec store.SnapSpec, user *auth.UserState) (*snap.Info, error) {
	channel := s.parseChannel(snapSpec.Channel)

	resp, err := s.downloadIndex(channel)
	if err != nil {
		return nil, err
	}

	apps, err := parseIndex(resp, s.cfg.StoreBaseURL)
	if err != nil {
		return nil, err
	}

	info, err := s.downloadAppInfo(apps[snapSpec.Name], channel, snapSpec.Name)
	if err != nil {
		return nil, ErrSnapNotFound
	}

	return info, nil
}

// A Search is what you do in order to Find something
type Search struct {
	Query   string
	Section string
	Private bool
	Prefix  bool
}

// Find finds  (installable) snaps from the store, matching the
// given Search.
func (s *Store) Find(search *store.Search, user *auth.UserState) ([]*snap.Info, error) {
	logger.Noticef("search query: %s", search.Query)
	channel := "stable"
	resp, err := s.downloadIndex(channel)
	if err != nil {
		return nil, err
	}
	apps, err := parseIndex(resp, s.cfg.StoreBaseURL)
	if err != nil {
		return nil, err
	}
	var snaps []*snap.Info
	for name, app := range apps {
		if search.Query == "*" || search.Query == "" || search.Query == name {
			//version, err := s.downloadVersion(channel, name)
			info, err := s.downloadAppInfo(app, channel, name)
			if err != nil {
				logger.Noticef("No version on the channel: %s", channel)
			} else {
				//info := app.toInfo(s.cfg.StoreBaseURL, channel, version)
				snaps = append(snaps, info)
			}
		}
	}

	return snaps, nil
}

func (s *Store) parseChannel(channel string) string {
	if channel == "" {
		return "stable"
	}
	parts := strings.Split(channel, "/")
	return parts[0]

}

func (s *Store) downloadIndex(channel string) (string, error) {
	reqOptions := &requestOptions{
		Method: "GET",
		URL:    urlJoin(s.cfg.StoreBaseURL, "releases", channel, "index-v2"),
		Accept: halJsonContentType,
	}

	resp, err := s.retryRequestString(context.TODO(), reqOptions)
	if err != nil {
		return "", err
	}
	return resp, nil
}

type App struct {
	Name     string `json:"id"`
	Summary  string `json:"name"`
	Icon     string `json:"icon,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
	Required bool   `json:"required"`
}

func constructSnapId(name string, version string) string {
	return fmt.Sprintf("%s.%s", name, version)
}

func deconstructSnapId(snapId string) (string, string) {
	parts := strings.Split(snapId, ".")
	return parts[0], parts[1]
}

func (a *App) toInfo(channel string, version string, downloadSize int64, downloadSha384 string, downloadUrl string) (*snap.Info, error) {
	appType := snap.TypeApp
	if a.Required {
		appType = snap.TypeBase
	}

	revision, err := strconv.Atoi(version)
	if err != nil {
		return nil, fmt.Errorf("unable to get revision: %s", err)
	}
	snapId := constructSnapId(a.Name, version)
	logger.Noticef("snapid: %s", snapId)

	details := snapDetails{
		SnapID:           snapId,
		Name:             a.Name,
		Summary:          a.Summary,
		Version:          version,
		Type:             appType,
		Architectures:    []string{"amd64", "armhf", "arm64"},
		Revision:         revision,
		IconURL:          a.Icon,
		Channel:          channel,
		AnonDownloadURL:  downloadUrl,
		DownloadSha3_384: downloadSha384,
		DownloadSize:     downloadSize,
	}

	return infoFromRemote(&details), nil
}

type Index struct {
	Apps []json.RawMessage `json:"apps"`
}

func parseIndex(resp string, baseUrl *url.URL) (map[string]*App, error) {
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
		if !app.Enabled {
			continue
		}
		apps[app.Name] = app

	}

	return apps, nil

}

func (s *Store) Sections(user *auth.UserState) ([]string, error) {
	return []string{"apps"}, nil
}

func (s *Store) WriteCatalogs(names io.Writer, adder store.SnapAdder) error {
	logger.Noticef("not writing catalogs")
	return nil
}

// RefreshCandidate contains information for the store about the currently
// installed snap so that the store can decide what update we should see
type RefreshCandidate struct {
	SnapID   string
	Revision snap.Revision
	Epoch    string
	Block    []snap.Revision

	// the desired channel
	Channel string
}

// the exact bits that we need to send to the store
type currentSnapJSON struct {
	SnapID      string `json:"snap_id"`
	Channel     string `json:"channel"`
	Revision    int    `json:"revision,omitempty"`
	Epoch       string `json:"epoch"`
	Confinement string `json:"confinement"`
}

type metadataWrapper struct {
	Snaps  []*currentSnapJSON `json:"snaps"`
	Fields []string           `json:"fields"`
}

func currentSnap(cs *RefreshCandidate) *currentSnapJSON {
	logger.Noticef("cutrentSnap %+v", cs)
	// the store gets confused if we send snaps without a snapid
	// (like local ones)
	if cs.SnapID == "" {
		if cs.Revision.Store() {
			logger.Noticef("store.currentSnap got given a RefreshCandidate with an empty SnapID but a store revision!")
		}
		return nil
	}
	if !cs.Revision.Store() {
		logger.Noticef("store.currentSnap got given a RefreshCandidate with a non-empty SnapID but a non-store revision!")
		return nil
	}

	channel := cs.Channel
	if channel == "" {
		channel = "stable"
	}

	return &currentSnapJSON{
		SnapID:   cs.SnapID,
		Channel:  channel,
		Epoch:    cs.Epoch,
		Revision: cs.Revision.N,
		// confinement purposely left empty
	}
}

func (s *Store) ListRefresh(installed []*store.RefreshCandidate, user *auth.UserState, flags *store.RefreshOptions) (snaps []*snap.Info, err error) {
	return nil, errors.New("not implemented yet")

}

type HashError struct {
	name           string
	sha3_384       string
	targetSha3_384 string
}

func (e HashError) Error() string {
	return fmt.Sprintf("sha3-384 mismatch for %q: got %s but expected %s", e.name, e.sha3_384, e.targetSha3_384)
}

// Download downloads the snap addressed by download info and returns its
// filename.
// The file is saved in temporary storage, and should be removed
// after use to prevent the disk from running out of space.
func (s *Store) Download(ctx context.Context, name string, targetPath string, downloadInfo *snap.DownloadInfo, pbar progress.Meter, user *auth.UserState) error {
	logger.Noticef("expected download sha: %s", downloadInfo.Sha3_384)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	//if err := s.cacher.Get(downloadInfo.Sha3_384, targetPath); err == nil {
	//	return nil
	//}

	partialPath := targetPath + ".partial"
	w, err := os.OpenFile(partialPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	resume, err := w.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := w.Close(); cerr != nil && err == nil {
			err = cerr
		}
		if err != nil {
			os.Remove(w.Name())
		}
	}()

	url := downloadInfo.AnonDownloadURL

	err = download(ctx, name, downloadInfo.Sha3_384, url, s, w, resume, pbar)
	// If hashsum is incorrect retry once
	if _, ok := err.(HashError); ok {
		logger.Debugf("Hashsum error on download: %v", err.Error())
		err = w.Truncate(0)
		if err != nil {
			return err
		}
		_, err = w.Seek(0, os.SEEK_SET)
		if err != nil {
			return err
		}
		err = download(ctx, name, downloadInfo.Sha3_384, url, s, w, 0, pbar)
	}

	if err != nil {
		return err
	}

	if err := os.Rename(w.Name(), targetPath); err != nil {
		return err
	}

	return w.Sync()
}

// download writes an http.Request showing a progress.Meter
var download = func(ctx context.Context, name, sha3_384, downloadURL string, s *Store, w io.ReadWriteSeeker, resume int64, pbar progress.Meter) error {
	storeURL, err := url.Parse(downloadURL)
	if err != nil {
		return err
	}

	var finalErr error
	startTime := time.Now()
	for attempt := retry.Start(defaultRetryStrategy, nil); attempt.Next(); {
		reqOptions := &requestOptions{
			Method: "GET",
			URL:    storeURL,
		}
		httputil.MaybeLogRetryAttempt(reqOptions.URL.String(), attempt, startTime)

		h := crypto.SHA3_384.New()

		if resume > 0 {
			reqOptions.ExtraHeaders = map[string]string{
				"Range": fmt.Sprintf("bytes=%d-", resume),
			}
			// seed the sha3 with the already local file
			if _, err := w.Seek(0, os.SEEK_SET); err != nil {
				return err
			}
			n, err := io.Copy(h, w)
			if err != nil {
				return err
			}
			if n != resume {
				return fmt.Errorf("resume offset wrong: %d != %d", resume, n)
			}
		}

		if cancelled(ctx) {
			return fmt.Errorf("The download has been cancelled: %s", ctx.Err())
		}
		var resp *http.Response
		resp, finalErr = s.doRequest(ctx, httputil.NewHTTPClient(nil), reqOptions)

		if cancelled(ctx) {
			return fmt.Errorf("The download has been cancelled: %s", ctx.Err())
		}
		if finalErr != nil {
			if httputil.ShouldRetryError(attempt, finalErr) {
				continue
			}
			break
		}

		if httputil.ShouldRetryHttpResponse(attempt, resp) {
			resp.Body.Close()
			continue
		}

		defer resp.Body.Close()

		switch resp.StatusCode {
		case 200, 206: // OK, Partial Content
		case 402: // Payment Required

			return fmt.Errorf("please buy %s before installing it.", name)
		default:
			return &DownloadError{Code: resp.StatusCode, URL: resp.Request.URL}
		}

		if pbar == nil {
			pbar = progress.Null
		}
		pbar.Start(name, float64(resp.ContentLength))
		mw := io.MultiWriter(w, h, pbar)
		_, finalErr = io.Copy(mw, resp.Body)
		pbar.Finished()
		if finalErr != nil {
			if httputil.ShouldRetryError(attempt, finalErr) {
				// error while downloading should resume
				var seekerr error
				resume, seekerr = w.Seek(0, os.SEEK_END)
				if seekerr == nil {
					continue
				}
				// if seek failed, then don't retry end return the original error
			}
			break
		}

		if cancelled(ctx) {
			return fmt.Errorf("The download has been cancelled: %s", ctx.Err())
		}

		actualSha3 := fmt.Sprintf("%x", h.Sum(nil))
		logger.Debugf("actual sha: %s", actualSha3)

		//if sha3_384 != "" && sha3_384 != actualSha3 {
		//	finalErr = HashError{name, actualSha3, sha3_384}
		//	}
		break
	}
	return finalErr
}

func (s *Store) Assertion(assertType *asserts.AssertionType, primaryKey []string, user *auth.UserState) (asserts.Assertion, error) {
	logger.Debugf("assert type: %s, key: %s", assertType.Name, strings.Join(primaryKey, "/"))

	publicKeyEnc, err := asserts.EncodePublicKey(privkey.PublicKey())
	if err != nil {
		return nil, err
	}

	body := ""
	headers := ""
	switch assertType.Name {
	case "account-key":
		body = string(publicKeyEnc)
	case "snap-declaration":
		snapId := primaryKey[1]
		name, _ := deconstructSnapId(snapId)
		headers = "" +
			"series: " + primaryKey[0] + "\n" +
			"snap-id: " + snapId + "\n" +
			"snap-name: " + name + "\n"
	case "snap-revision":
		nameVersion, err := base64.RawURLEncoding.DecodeString(primaryKey[0])
		if err != nil {
			return nil, err
		}
		snapId := string(nameVersion)
		_, revision := deconstructSnapId(snapId)
		headers = "" +
			"snap-revision: " + revision + "\n" +
			"snap-id: " + snapId + "\n" +
			"snap-size: 1\n" +
			"snap-sha3-384: " + primaryKey[0] + "\n"
	}

	publicKeyId := privkey.PublicKey().ID()
	logger.Debugf("public key id: %s", publicKeyId)

	content := "type: " + assertType.Name + "\n" +
		"authority-id: syncloud\n" +
		"primary-key: " + strings.Join(primaryKey, "/") + "\n" +
		"publisher-id: syncloud\n" +
		"developer-id: syncloud\n" +
		"account-id: syncloud\n" +
		// "display-name: syncloud\n" +
		"revision: 1\n" +
		"sign-key-sha3-384: " + SHA3_384 + "\n" +
		"sha3-384: " + SHA3_384 + "\n" +
		"public-key-sha3-384: " + publicKeyId + "\n" +
		"timestamp: " + time.Now().Format(time.RFC3339) + "\n" +
		"since: " + time.Now().Format(time.RFC3339) + "\n" +
		headers +
		"validation: certified\n" +
		"body-length: " + strconv.Itoa(len(body)) + "\n\n" +
		body +
		"\n\n"

	signature, err := asserts.SignContent([]byte(content), privkey)

	if err != nil {
		return nil, err
	}

	assertionText := content + string(signature[:]) + "\n"

	logger.Debugf("assertion response: \n%s", assertionText)

	asrt, e := asserts.Decode([]byte(assertionText))

	return asrt, e

}

// BuyOptions specifies parameters to buy from the store.
type BuyOptions struct {
	SnapID   string  `json:"snap-id"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"` // ISO 4217 code as string
}

// BuyResult holds the state of a buy attempt.
type BuyResult struct {
	State string `json:"state,omitempty"`
}

// orderInstruction holds data sent to the store for orders.
type orderInstruction struct {
	SnapID   string `json:"snap_id"`
	Amount   string `json:"amount,omitempty"`
	Currency string `json:"currency,omitempty"`
}

type storeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s *storeError) Error() string {
	return s.Message
}

type storeErrors struct {
	Errors []*storeError `json:"error_list"`
}

func (s *storeErrors) Code() string {
	if len(s.Errors) == 0 {
		return ""
	}
	return s.Errors[0].Code
}

func (s *storeErrors) Error() string {
	if len(s.Errors) == 0 {
		return "internal error: empty store error used as an actual error"
	}
	return s.Errors[0].Error()
}

func buyOptionError(message string) (*BuyResult, error) {
	return nil, fmt.Errorf("cannot buy snap: %s", message)
}

func (s *Store) LookupRefresh(installed *store.RefreshCandidate, user *auth.UserState) (*snap.Info, error) {
	logger.Noticef("LookupRefresh: %v", installed)
	channel := s.parseChannel(installed.Channel)
	snapName, _ := deconstructSnapId(installed.SnapID)

	resp, err := s.downloadIndex(channel)
	if err != nil {
		return nil, err
	}

	apps, err := parseIndex(resp, s.cfg.StoreBaseURL)
	if err != nil {
		return nil, err
	}

	info, err := s.downloadAppInfo(apps[snapName], channel, snapName)
	if err != nil {
		return nil, ErrSnapNotFound
	}

	return info, nil
}

func (s *Store) SuggestedCurrency() string {
	return "USD"
}

type storeCustomer struct {
	LatestTOSDate     string `json:"latest_tos_date"`
	AcceptedTOSDate   string `json:"accepted_tos_date"`
	LatestTOSAccepted bool   `json:"latest_tos_accepted"`
	HasPaymentMethod  bool   `json:"has_payment_method"`
}

func (s *Store) Buy(options *store.BuyOptions, user *auth.UserState) (*store.BuyResult, error) {
	return nil, errors.New("not implemented yet")
}

func (s *Store) ReadyToBuy(*auth.UserState) error {
	return errors.New("not implemented yet")
}

func (s *Store) CacheDownloads() int {
	return s.cfg.CacheDownloads
}

func (s *Store) SetCacheDownloads(fileCount int) {
	s.cfg.CacheDownloads = fileCount
	if fileCount > 0 {
		s.cacher = store.NewCacheManager(dirs.SnapDownloadCacheDir, fileCount)
	} else {
		s.cacher = &store.NullCache{}
	}
}
