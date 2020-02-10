package store

import (
	"context"

	"github.com/snapcore/snapd/overlord/auth"
	"github.com/snapcore/snapd/snap"
)

// Store represents the ubuntu snap store
type SyncloudStore struct {
	store *Store
}

// New creates a new Store with the given access configuration and for given the store id.
func NewSyncloudStore(cfg *Config, dauthCtx DeviceAndAuthContext) *SyncloudStore {
	return &SyncloudStore{
		store: New(cfg, dauthCtx),
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

	resp, err := s.downloadIndex(channel)
	if err != nil {
		return nil, err
	}

	apps, err := parseIndex(resp, s.cfg.StoreBaseURL)
	if err != nil {
		return nil, err
	}

	version, err := s.downloadVersion(channel, snapSpec.Name)
	if err != nil {
		return nil, ErrSnapNotFound
	}

	info := apps[snapSpec.Name].toInfo(s.cfg.StoreBaseURL, channel, version)

	return info, nil
}
