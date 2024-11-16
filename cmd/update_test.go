package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateLinkName(t *testing.T) {
	linkName, tag, path := "TestUpdateLinkName", "tag", "/foo/bar"
	linkName1, tag1, path1 := "TestUpdateLinkName1", "tag1", "/foo/bar1"
	ExecuteCommand(t, "add", "link", linkName)
	ExecuteCommand(t, "add", "tag", linkName, tag, path)

	ExecuteCommand(t, "add", "link", linkName1)
	ExecuteCommand(t, "add", "tag", linkName1, tag1, path1)

	ExecuteCommand(t, "add", "bind", linkName+":"+tag, linkName1+":"+tag1)

	newName := linkName + "_new"
	ExecuteCommand(t, "update", "link", linkName, "--name="+newName)

	assert.True(t, LinkNameExist(newName))
	assert.False(t, LinkNameExist(linkName))
	assert.True(t, TagExist(newName, tag, path))
	assert.True(t, BindExist(newName, tag, linkName1, tag1))
}

func TestUpdateTag(t *testing.T) {
	linkName, tag, path := "TestUpdateLinkName", "tag", "/foo/bar"

	ExecuteCommand(t, "add", "link", linkName)
	ExecuteCommand(t, "add", "tag", linkName, tag, path)

	newTag, newPath := tag+"_new", path+"/new"
	ExecuteCommand(t, "update", "tag", linkName, tag, "--tag="+newTag, "--path="+newPath)

	assert.True(t, TagExist(linkName, newTag, newPath))
}

func TestUpdateBind(t *testing.T) {
	cur, target := CreateBind(t, "TestUpdateBind", false)

	linkName, tag, path := "TestUpdateLinkName", "tag", "/foo/bar"
	ExecuteCommand(t, "add", "link", linkName)
	ExecuteCommand(t, "add", "tag", linkName, tag, path)

	ExecuteCommand(t, "update", "bind", cur.Name+":"+cur.Tag, target.Name+":"+target.Tag, "--targetName="+linkName, "--targetTag="+tag)

	assert.True(t, BindExist(cur.Name, cur.Tag, linkName, tag))
	assert.False(t, BindExist(cur.Name, cur.Tag, target.Name, target.Tag))
}
