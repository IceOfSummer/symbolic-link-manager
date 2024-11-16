package cmd

import (
	"github.com/stretchr/testify/assert"
	"github.com/symbolic-link-manager/internal/configuration"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func readLink(linkName string) (string, error) {
	home := configuration.AppHome()
	lk := path.Join(home, "app", linkName)
	return os.Readlink(lk)
}

func TestUse(t *testing.T) {
	linkName, tag := "TestUse", "TestUse_tag"
	path0, err := prepareTestDirectory("TestUse")
	assert.NoError(t, err)

	ExecuteCommand(t, "add", "link", linkName)
	ExecuteCommand(t, "add", "tag", linkName, tag, path0)
	ExecuteCommand(t, "use", linkName, tag)

	p, err := readLink(linkName)
	assert.NoError(t, err)
	assert.Equal(t, path0, p)
}

func TestBindSwitch(t *testing.T) {
	cur, target := CreateBind(t, "TestBindSwitch", true)
	ExecuteCommand(t, "use", cur.Name, cur.Tag)

	p, err := readLink(cur.Name)
	assert.NoError(t, err)
	ap, err := filepath.Abs(cur.Path)
	assert.NoError(t, err)
	assert.Equal(t, ap, p)

	p1, err := readLink(target.Name)
	assert.NoError(t, err)
	ap1, err := filepath.Abs(target.Path)
	assert.NoError(t, err)
	assert.Equal(t, ap1, p1)
}
