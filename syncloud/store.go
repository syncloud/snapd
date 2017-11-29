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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	"encoding/base64"

	"github.com/snapcore/snapd/arch"
	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/httputil"
	"github.com/snapcore/snapd/logger"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/overlord/auth"
	"github.com/snapcore/snapd/progress"
	"github.com/snapcore/snapd/release"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/store"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
	"gopkg.in/retry.v1"
	"io/ioutil"
	"strconv"
	"crypto/rsa"
	"crypto/rand"
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
var defaultRetryStrategy = retry.LimitCount(5, retry.LimitTime(33*time.Second,
	retry.Exponential{
		Initial: 100 * time.Millisecond,
		Factor:  2.5,
	},
))

func infoFromRemote(d *snapDetails) *snap.Info {
	info := &snap.Info{}
	info.Architectures = d.Architectures
	info.Type = d.Type
	info.Version = d.Version
	info.Epoch = "0"
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
	info.Confinement = snap.ConfinementType(d.Confinement)
	info.Contact = d.Contact

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


// Store represents the ubuntu snap store
type Store struct {
	searchURI      *url.URL
	detailsURI     *url.URL
	bulkURI        *url.URL
	assertionsURI  *url.URL
	ordersURI      *url.URL
	buyURI         *url.URL
	customersMeURI *url.URL
	sectionsURI    *url.URL

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

var defaultConfig = store.Config{}
var syncloudAppsBaseURL *url.URL
var privkey asserts.PrivateKey

func init() {

	pkey, err := rsa.GenerateKey(rand.Reader, 752)
	if err != nil {
		panic(fmt.Errorf("failed to create private key: %v", err))
	}
	privkey = asserts.RSAPrivateKey(pkey)
	//return asserts.RSAPrivateKey(priv), priv

	syncloudAppsBaseURL, _  = url.Parse("http://apps.syncloud.org")
	//defaultConfig.SearchURI = urlJoin(storeBaseURI, "api/v1/snaps/search")
	defaultConfig.DetailsURI = urlJoin(syncloudAppsBaseURL, "releases/master/versions")


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
func New(cfg *store.Config, authContext auth.AuthContext) *Store {
	if cfg == nil {
		cfg = &defaultConfig
	}

	fields := cfg.DetailFields
	if fields == nil {
		fields = detailFields
	}

	rawQuery := ""
	if len(fields) > 0 {
		v := url.Values{}
		v.Set("fields", strings.Join(fields, ","))
		rawQuery = v.Encode()
	}

	var searchURI *url.URL
	if cfg.SearchURI != nil {
		uri := *cfg.SearchURI
		uri.RawQuery = rawQuery
		searchURI = &uri
	}

	var detailsURI *url.URL
	if cfg.DetailsURI != nil {
		uri := *cfg.DetailsURI
		uri.RawQuery = rawQuery
		detailsURI = &uri
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

	// see https://wiki.ubuntu.com/AppStore/Interfaces/ClickPackageIndex
	return &Store{
		searchURI:       searchURI,
		detailsURI:      detailsURI,
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
		return "", errors.New("store is not responding")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}


func (s *Store) retryRequestString(ctx context.Context, reqOptions *requestOptions) (string, error) {
	var reply string
	_, err := httputil.RetryRequest(reqOptions.URL.String(), func() (*http.Response, error) {
		return s.doRequest(ctx, s.client, reqOptions)
	}, func(resp *http.Response) error {
		resp1, err1 := decodeStringBody(resp)
		reply = resp1
		return err1
	}, defaultRetryStrategy)

	if err != nil {
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
		if len(info.Prices) != 0 {
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
		URL:    s.ordersURI,
		Accept: jsonContentType,
	}
	var result ordersResult
	resp, err := s.retryRequestDecodeJSON(context.TODO(), reqOptions, user, &result)
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
		info.MustBuy = mustBuy(info.Prices, bought[info.SnapID])
	}

	return nil
}

// mustBuy determines if a snap requires a payment, based on if it is non-free and if the user has already bought it
func mustBuy(prices map[string]float64, bought bool) bool {
	if len(prices) == 0 {
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

// SnapInfo returns the snap.Info for the store-hosted snap matching the given spec, or an error.
func (s *Store) SnapInfo(snapSpec store.SnapSpec, user *auth.UserState) (*snap.Info, error) {
	// get the query before doing Parse, as that overwrites it
	reqOptions := &requestOptions{
		Method: "GET",
		URL:    s.detailsURI,
		Accept: halJsonContentType,
	}

	//var remote *snapDetails
	resp, err := s.retryRequestString(context.TODO(), reqOptions)
	if err != nil {
		return nil, err
	}


	lines := strings.Split(resp, "\n")
	apps := make(map[string]string)
	for i := 0; i < len(lines); i +=2 {
		values := strings.Split(lines[i], "=")
		apps[values[0]] = values[1]
	}

	versionStr, ok := apps[snapSpec.Name]
	if !ok {
		return nil, ErrSnapNotFound
	}

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return nil,  fmt.Errorf("Unable to get version: %s", err)
	}

	details := snapDetails{
		Name: snapSpec.Name,
		Version: versionStr,
		Architectures: []string{"amd64", "armhf"},
		Revision: version,
		AnonDownloadURL: fmt.Sprintf("%s/apps/%s_%d_%s.snap", syncloudAppsBaseURL, snapSpec.Name, version, arch.UbuntuArchitecture()),

	}
	info := infoFromRemote(&details)

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
	return nil, errors.New("Find is not implemented yet")
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
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

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
			pbar = &progress.NullProgress{}
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
		if sha3_384 != "" && sha3_384 != actualSha3 {
			finalErr = HashError{name, actualSha3, sha3_384}
		}
		break
	}
	return finalErr
}

func (s *Store) Assertion(assertType *asserts.AssertionType, primaryKey []string, user *auth.UserState) (asserts.Assertion, error) {

	blobSHA3_384 := "QlqR0uAWEAWF5Nwnzj5kqmmwFslYPu1IL16MKtLKhwhv0kpBv5wKZ_axf_nf_2cL"
	hashDigest, err := base64.RawURLEncoding.DecodeString(blobSHA3_384)
	if err != nil {
		return nil, err
	}

	digest, err := asserts.EncodeDigest(crypto.SHA3_384, hashDigest)
	if err != nil {
		return nil, err
	}

	publicKeyEnc, err := asserts.EncodePublicKey(privkey.PublicKey())
	if err != nil {
		return nil, err
	}

	//publicKey := string(publicKeyEn)
	println(assertType.Name)

	assertion, err := asserts.Assemble(
		map[string]interface{}{
			"snap-name": "syncloud",
			"snap-id": "syncloud",
			"snap-size": "100",
			"snap-revision": "1",
			"authority-id": "syncloud",
			"publisher-id": "syncloud",
			"developer-id": "syncloud",
			"account-id": "syncloud",
			"display-name": "syncloud",
			"type": assertType.Name,
			"sign-key-sha3-384": digest,
			"sha3-384": digest,
			"snap-sha3-384": digest,
			"public-key-sha3-384": privkey.PublicKey().ID(),
			"timestamp": time.Now().Format(time.RFC3339),
			"since": time.Now().Format(time.RFC3339),
			"series": "1",
			"validation": "certified",
			"body-length": "149",
		},
		publicKeyEnc,
		[]byte(""),
		[]byte("signature"),
	)

	return assertion, err

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

func (s *Store) LookupRefresh(*store.RefreshCandidate, *auth.UserState) (*snap.Info, error) {
	return nil, ErrNoUpdateAvailable
}

func (s *Store) ListRefresh([]*store.RefreshCandidate, *auth.UserState) ([]*snap.Info, error) {
	return nil, errors.New("not implemented yet")

}

func (s *Store) Sections(user *auth.UserState) ([]string, error) {
	return nil, errors.New("not implemented yet")
}

func (s *Store) SuggestedCurrency() string {
	return "USD"
}

func (s *Store) Buy(options *store.BuyOptions, user *auth.UserState) (*store.BuyResult, error) {
	return nil, errors.New("not implemented yet")
}

func (s *Store) ReadyToBuy(*auth.UserState) error {
	return errors.New("not implemented yet")
}
