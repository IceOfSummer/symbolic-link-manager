package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Link struct {
	// 链接名称
	Name string
	// 链接别名
	Alias string
	// 链接路径
	Path string
}

type configuration struct {
	DeclariedLinkNames []string
	Envs               []Link
	// Env.Name:Env.Alias -> Env
	Binds map[string]Link
}

var cache *configuration = nil

// 读取配置文件
func readConfig() configuration {
	if cache != nil {
		return *cache
	}
	_, err := os.Stat(configFilePath)
	if os.IsNotExist(err) {
		return configuration{
			DeclariedLinkNames: make([]string, 0),
			Envs:               make([]Link, 0),
			Binds:              map[string]Link{},
		}
	}
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		panic(err)
	}
	var configuration configuration
	err = json.Unmarshal([]byte(content), &configuration)
	if err != nil {
		fmt.Println("Failed to read json config.")
		panic(err)
	}
	cache = &configuration
	return configuration
}

func saveConfig(configuration *configuration) {
	cache = configuration
	content, err := json.Marshal(configuration)
	if err != nil {
		fmt.Println("Failed to save json config.")
		panic(err)
	}
	os.WriteFile(configFilePath, content, 0b110_110_100)
}

// 添加一个环境变量声明
func AddEnvDeclarition(declaritionName string) {
	config := readConfig()
	config.DeclariedLinkNames = append(config.DeclariedLinkNames, declaritionName)
	saveConfig(&config)
}

func (th *configuration) isDeclaritionExist(declarationName string) bool {
	for _, v := range th.DeclariedLinkNames {
		if v == declarationName {
			return true
		}
	}
	return false
}

// 添加一个环境变量的值
func AddEnvValue(env *Link) error {
	config := readConfig()
	if !config.isDeclaritionExist(env.Name) {
		return errors.New("对应的环境变量没有声明")
	}
	config.Envs = append(config.Envs, *env)
	saveConfig(&config)
	return nil
}

func ListEnv(name string) []Link {
	config := readConfig()
	result := make([]Link, 2)
	for _, v := range config.Envs {
		if v.Name == name {
			result = append(result, v)
		}
	}
	return result
}

func FindEnvByNameAndAlias(name, aliase string) *Link {
	envs := ListEnv(name)
	for _, v := range envs {
		if v.Alias == aliase {
			return &v
		}
	}
	return nil
}