package hookstate

import (
	"fmt"
	"github.com/snapcore/snapd/i18n/dumb"
	"github.com/snapcore/snapd/overlord/state"
	"github.com/snapcore/snapd/snap"
)

// HookSetup is a reference to a hook within a specific snap.
type HookSetup struct {
	Snap     string        `json:"snap"`
	Revision snap.Revision `json:"revision"`
	Hook     string        `json:"hook"`
	Optional bool          `json:"optional,omitempty"`
}

// HookTask returns a task that will run the specified hook. Note that the
// initial context must properly marshal and unmarshal with encoding/json.
func HookTask(st *state.State, summary string, setup *HookSetup, contextData map[string]interface{}) *state.Task {
	task := st.NewTask("run-hook", summary)
	task.Set("hook-setup", setup)

	// Initial data for Context.Get/Set.
	if len(contextData) > 0 {
		task.Set("hook-context", contextData)
	}
	return task
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
