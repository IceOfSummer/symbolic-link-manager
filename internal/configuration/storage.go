package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/symbolic-link-manager/internal/localizer"
)

var cache *configuration = nil
var lazyLoadConfigPath string

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	if lazyLoadConfigPath != "" {
		return lazyLoadConfigPath
	}
	p := path.Join(AppHome(), "configuration.json")
	lazyLoadConfigPath = p
	return p
}

// readConfig 读取配置文件, 如果有修改了数据, 应该调用 [saveConfig] 进行持久化。
func readConfig() configuration {
	if cache != nil {
		return *cache
	}
	configFilePath := getConfigPath()
	_, err := os.Stat(configFilePath)
	if os.IsNotExist(err) {
		return configuration{
			DeclaredLinkNames: make([]string, 0),
			Links:             make([]*Link, 0),
			Binds:             map[string][]*LinkBindItem{},
		}
	}
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		panic(err)
	}
	var configuration = configuration{
		DeclaredLinkNames: make([]string, 0),
		Links:             make([]*Link, 0),
		Binds:             map[string][]*LinkBindItem{},
	}
	err = json.Unmarshal(content, &configuration)
	if err != nil && len(content) > 0 {
		fmt.Println("Failed to read json config.")
		panic(err)
	}
	cache = &configuration
	return configuration
}

// saveConfig 保存配置
func saveConfig(configuration *configuration) {
	configFilePath := getConfigPath()
	cache = configuration
	content, err := json.Marshal(configuration)
	if err != nil {
		fmt.Println("Failed to save json config.")
		panic(err)
	}
	err = os.WriteFile(configFilePath, content, 0b110_110_100)
	if err != nil {
		panic(err)
	}
}

// AddEnvDeclaration 添加一个环境变量声明
func AddEnvDeclaration(declarationName string) {
	config := readConfig()
	config.DeclaredLinkNames = append(config.DeclaredLinkNames, declarationName)
	saveConfig(&config)
}

// isDeclarationExist 判断链接是否已经声明
func (th *configuration) isDeclarationExist(declarationName string) bool {
	p := th.findDeclaration(declarationName)
	if p == -1 {
		return false
	}
	return true
}

// 搜索变量声明的位置，如果没找到，返回 -1
func (th *configuration) findDeclaration(declarationName string) int {
	for pos, v := range th.DeclaredLinkNames {
		if v == declarationName {
			return pos
		}
	}
	return -1
}

// AddEnvValue 添加一个环境变量的值
func AddEnvValue(env *Link) error {
	config := readConfig()
	if !config.isDeclarationExist(env.Name) {
		return errors.New("对应的环境变量没有声明")
	}
	config.Links = append(config.Links, env)
	saveConfig(&config)
	return nil
}

// AddBind target 绑定到 src
func AddBind(srcName, srcAlias, targetName, targetAlias string) error {
	config := readConfig()

	old, ok := config.Binds[srcName]

	entity := &LinkBindItem{
		CurrentTag: srcAlias,
		TargetName: targetName,
		TargetTag:  targetAlias,
	}
	if ok {
		config.Binds[srcName] = append(old, entity)
	} else {
		config.Binds[srcName] = []*LinkBindItem{entity}
	}
	saveConfig(&config)
	return nil
}

// ListLinkNames
// 列出已经声明的链接
func ListLinkNames() []string {
	return readConfig().DeclaredLinkNames
}

// ListLinkTags 列出所有链接的值。
//
// 当不传 [name] 时，返回所有的值
func ListLinkTags(name string) []*Link {
	config := readConfig()
	if name == "" {
		return config.Links
	}
	result := make([]*Link, 0)
	for _, v := range config.Links {
		if v.Name == name {
			result = append(result, v)
		}
	}
	return result
}

// FindLinkByNameAndTag
// 根据名称和标签搜素链接
func FindLinkByNameAndTag(name, alias string) *Link {
	envs := ListLinkTags(name)
	for _, v := range envs {
		if v.Tag == alias {
			return v
		}
	}
	return nil
}

// ListBinds
// 列出所有的绑定
func ListBinds(linkName, tag string) []*LinkBindItem {
	config := readConfig()

	value, ok := config.Binds[linkName]
	if !ok {
		return make([]*LinkBindItem, 0)
	}
	result := make([]*LinkBindItem, 0)
	for _, item := range value {
		if item.CurrentTag == tag {
			result = append(result, item)
		}
	}
	return result
}

// GetAllBinds
// 获取所有的绑定
func GetAllBinds() BindsData {
	return readConfig().Binds
}

// rebuildDeclaredLinks
// 根据 [Link] 重新创建已经声明的链接
func rebuildDeclaredLinks(links []*Link) []string {
	var names = make([]string, 0)
	set := make(map[string]struct{})
	for _, link := range links {
		_, ok := set[link.Name]
		if !ok {
			set[link.Name] = struct{}{}
			names = append(names, link.Name)
		}
	}
	return names
}

// DeleteLink 删除链接.
// 如果不提供第二个参数, 则删除全部.
// 返回被删除的元素, 如果整个链接被删除，则第二个参数返回 true
func DeleteLink(linkName, alias string) ([]*Link, bool, error) {
	config := readConfig()

	if !config.isDeclarationExist(linkName) {
		return []*Link{}, false, localizer.CreateNoSuchLinkError(linkName)
	}

	var newLinks []*Link
	var deleted []*Link

	for _, link := range config.Links {
		if link.Name != linkName {
			newLinks = append(newLinks, link)
			continue
		}
		if alias == "" || link.Tag == alias {
			deleted = append(deleted, link)
			continue
		}
		newLinks = append(newLinks, link)
	}
	newDeclaredLinkNames := rebuildDeclaredLinks(newLinks)
	config.Links = newLinks
	config.DeclaredLinkNames = newDeclaredLinkNames
	saveConfig(&config)
	return deleted, !config.isDeclarationExist(linkName), nil
}

// DeleteBind 删除对应的绑定.
//
// 返回 [true] 表示删除成功
func DeleteBind(rootLinkName string, linkBindItem *LinkBindItem) bool {
	config := readConfig()

	result, ok := config.Binds[rootLinkName]
	if !ok {
		return false
	}
	for i, item := range result {
		if item.TargetName == linkBindItem.TargetName &&
			item.TargetTag == linkBindItem.TargetTag &&
			item.CurrentTag == linkBindItem.CurrentTag {
			result = append(result[:i], result[i+1:]...)
			config.Binds[rootLinkName] = result
			saveConfig(&config)
			return true
		}
	}
	return false
}

// RenameLinkDeclaration 重命名链接声明
// 如果没有找到旧的声明，将会返回一个错误.
func RenameLinkDeclaration(oldName, newName string) error {
	config := readConfig()
	pos := config.findDeclaration(oldName)
	if pos == -1 {
		return localizer.CreateNoSuchLinkError(oldName)
	}
	np := config.findDeclaration(newName)
	if np != -1 {
		return localizer.CreateLinkNameAlreadyExistError(newName)
	}

	for _, link := range config.Links {
		if link.Name == oldName {
			link.Name = newName
		}
	}
	config.DeclaredLinkNames = rebuildDeclaredLinks(config.Links)

	oldBinds, ok := config.Binds[oldName]
	if ok {
		config.Binds[newName] = oldBinds
		delete(config.Binds, oldName)
	}
	saveConfig(&config)
	return nil
}

// UpdateTag 更新链接值
func UpdateTag(name, tag string, updateEntity Link) error {
	config := readConfig()
	link := FindLinkByNameAndTag(name, tag)
	if link == nil {
		return localizer.CreateNoSuchLinkError(name)
	}
	if updateEntity.Tag != "" {
		link.Tag = updateEntity.Tag
	}
	if updateEntity.Path != "" {
		link.Path = updateEntity.Path
	}
	saveConfig(&config)
	return nil
}

type UpdateBindDTO struct {
	// Required
	SrcName string
	// Required
	SrcTag string
	// Required
	TargetName string
	// Required
	TargetTag string
	// Optional
	NewName string
	// Optional
	NewAlias string
}

// UpdateBind 更新绑定
func UpdateBind(dto UpdateBindDTO) error {
	config := readConfig()
	bind, ok := config.Binds[dto.SrcName]
	if !ok {
		return localizer.CreateNoSuchTagError(dto.SrcName, dto.SrcTag)
	}
	for _, item := range bind {
		if item.TargetName == dto.TargetName && item.TargetTag == dto.TargetTag {
			if dto.NewName != "" {
				item.TargetName = dto.NewName
			}
			if dto.NewAlias != "" {
				item.TargetTag = dto.NewAlias
			}
			saveConfig(&config)
			return nil
		}
	}
	return localizer.CreateNoSuchTagError(dto.TargetName, dto.TargetTag)
}
