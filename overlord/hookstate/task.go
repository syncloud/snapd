package hookstate

import (
	"fmt"
	"github.com/snapcore/snapd/i18n/dumb"
	"github.com/snapcore/snapd/overlord/state"
	"github.com/snapcore/snapd/snap"
	"time"
)

// HookSetup is a reference to a hook within a specific snap.
type HookSetup struct {
	Snap     string        `json:"snap"`
	Revision snap.Revision `json:"revision"`
	Hook     string        `json:"hook"`
	Optional bool          `json:"optional,omitempty"`

	Timeout     time.Duration `json:"timeout,omitempty"`
	IgnoreError bool          `json:"ignore-error,omitempty"`
	TrackError  bool          `json:"track-error,omitempty"`

}

func PostInstall(s *state.State, snapName string) *state.Task {
	var summary = fmt.Sprintf(i18n.G("Run post-install hook of %q snap if present"), snapName)
	hooksup := &HookSetup{
		Snap:     snapName,
		Hook:     "post-install",
		Optional: true,
	}
	return HookTask(s, summary, hooksup, nil)
}
