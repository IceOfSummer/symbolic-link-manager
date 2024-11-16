// Helper functions
package cmd

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/symbolic-link-manager/internal/configuration"
)

func SetUpTestEnvironment() {
	_, filename, _, _ := runtime.Caller(0)
	root := path.Dir(path.Join(filename, ".."))

	err := os.Setenv(configuration.AppHomeEnvKey, path.Join(root, "tmp"))
	if err != nil {
		panic(err)
	}
	_ = os.Mkdir(configuration.AppHome(), 0b111_101_000)
	CleanUp()
}

func CleanUp() {
	target := path.Join(configuration.AppHome(), "configuration.json")
	stat, err := os.Stat(target)
	if err != nil {
		return
	}
	if stat.Size() == 0 {
		return
	}
	_ = os.WriteFile(target, []byte(""), 0b111_101_000)
}

func ExecuteCommand(t *testing.T, args ...string) {
	rootCmd.SetArgs(args)
	assert.Nil(t, rootCmd.Execute())
}

func TestMain(m *testing.M) {
	SetUpTestEnvironment()
	code := m.Run()
	CleanUp()
	os.Exit(code)
}

func Exist[E any](slice []E, searchFn func(ele E) bool) bool {
	for _, v := range slice {
		if searchFn(v) {
			return true
		}
	}
	return false
}

func LinkNameExist(linkName string) bool {
	return Exist(configuration.ListLinkNames(), func(name string) bool {
		return name == linkName
	})
}

// TagExist
// 判断标签是否存在, 如果 [path] 参数为空，则不检查路径
func TagExist(linkName, tag, path string) bool {
	return Exist(configuration.ListLinkTags(linkName), func(link *configuration.Link) bool {
		return link.Tag == tag && (path == link.Path || path == "")
	})
}

func BindExist(linkName, tag, targetLinkName, targetTag string) bool {
	return Exist(configuration.ListBinds(linkName, tag), func(bind *configuration.LinkBindItem) bool {
		return bind.CurrentTag == tag && bind.TargetName == targetLinkName && bind.TargetTag == targetTag
	})
}

// 准备测试使用文件夹
func prepareTestDirectory(linkName string) (string, error) {
	home := configuration.AppHome()

	testRoot := path.Join(home, "test")
	_ = os.Mkdir(testRoot, 0b111_111_101)

	target := path.Join(testRoot, linkName)
	stat, err := os.Stat(target)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
	} else if stat.IsDir() {
		return filepath.FromSlash(target), nil
	}
	err = os.Mkdir(path.Join(testRoot, linkName), 0b111_111_101)
	if err != nil {
		return "", err
	}
	return filepath.FromSlash(target), nil
}

// 创建绑定
func CreateBind(t *testing.T, baseName string, useRealDirectory bool) (*configuration.Link, *configuration.Link) {
	name, tag := baseName, baseName+"_tag"
	name1, tag1 := baseName+"1", baseName+"_tag1"
	var path0, path1 string

	if useRealDirectory {
		p, err := prepareTestDirectory(name)
		assert.NoError(t, err)
		path0 = p
		p, err = prepareTestDirectory(name)
		assert.NoError(t, err)
		path1 = p
	} else {
		path0 = "/fake/" + name
		path1 = "/fake/" + name1
	}
	ExecuteCommand(t, "add", "link", name)
	ExecuteCommand(t, "add", "tag", name, tag, path0)

	ExecuteCommand(t, "add", "link", name1)
	ExecuteCommand(t, "add", "tag", name1, tag1, path1)

	ExecuteCommand(t, "add", "bind", name+":"+tag, name1+":"+tag1)
	return &configuration.Link{Name: name, Tag: tag, Path: path0},
		&configuration.Link{Name: name1, Tag: tag1, Path: path1}
}
