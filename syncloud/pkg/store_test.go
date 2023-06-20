package pkg

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/store/pkg/model"
	"testing"
)

func TestApp(t *testing.T) {

	app := &model.App{
		Name:    "name",
		Summary: "summary",
		Icon:    "url",
	}
	info, err := app.ToInfo("1", 0, "3", "4")
	assert.NoError(t, err)

	bytes, err := json.MarshalIndent(info, "", "  ")
	assert.NoError(t, err)

	var result model.Snap
	err = json.Unmarshal(bytes, &result)
	assert.NoError(t, err)
}
