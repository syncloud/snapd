package store

import (
	"context"
	"encoding/json"
	"fmt"
	"encoding/base64"
	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/asserts/assertstest"
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
	"time"
)

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
}
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

func (s *SyncloudStore) Assertion(assertType *asserts.AssertionType, primaryKey []string, user *auth.UserState) (asserts.Assertion, error) {
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
