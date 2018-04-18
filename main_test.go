package main

import (
	"os/exec"
	"testing"

	"os"

	"github.com/stretchr/testify/assert"
)

func TestIsMasterConfig(t *testing.T) {
	var testCases = []struct {
		config         string
		expectedResult bool
	}{
		{
			config:         "kubeconfig",
			expectedResult: true,
		},
		{
			config:         "kubeconfig_k8s-dev-delivery",
			expectedResult: false,
		},
		{
			config:         "kubeconfig_k8s_dev_delivery",
			expectedResult: false,
		},
		{
			config:         "",
			expectedResult: false,
		},
	}
	for _, tc := range testCases {
		actualResult := isMasterConfig(tc.config)
		assert.Equal(t, tc.expectedResult, actualResult)
	}
}

func TestSwitchConfig(t *testing.T) {

}
func TestCopyConfig(t *testing.T) {

}
func TestGetRawConfig(t *testing.T) {

}
func TestGetAliasSuccessfully(t *testing.T) {
	var testCases = []struct {
		args          []string
		expectedAlias string
	}{
		{
			args:          []string{"alias"},
			expectedAlias: "alias",
		},
		{
			args:          []string{"arg1", "arg2"},
			expectedAlias: "arg1",
		},
	}
	for _, tc := range testCases {
		actualAlias := getAlias(tc.args)
		assert.Equal(t, tc.expectedAlias, actualAlias)
	}
}

func TestGetAliasFailure(t *testing.T) {
	if os.Getenv("CRASH") == "true" {
		getAlias([]string{})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestGetAliasFailure")
	cmd.Env = append(os.Environ(), "CRASH=true")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("getAlias should exit on empty args list")
}

func TestGetConfigByAliasEmptyConfigMap(t *testing.T) {
	if os.Getenv("CRASH") == "true" {
		getConfigByAlias("alias", map[string]*configuration{})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestGetConfigByAliasEmptyConfigMap")
	cmd.Env = append(os.Environ(), "CRASH=true")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("getConfigByAlias should exit on empty configmap")
}

func TestGetConfigByAliasNotFound(t *testing.T) {
	if os.Getenv("CRASH") == "true" {
		getConfigByAlias("alias1", map[string]*configuration{"config1": {Aliases: []string{"alias2", "alias3"}}})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestGetConfigByAliasNotFound")
	cmd.Env = append(os.Environ(), "CRASH=true")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("getConfigByAlias should exit on alias not found in configmap")
}

func TestGetConfigByAliasSuccessCases(t *testing.T) {
	var testCases = []struct {
		configs            map[string]*configuration
		expectedConfig     *configuration
		expectedConfigName string
		alias              string
	}{
		{
			configs: map[string]*configuration{
				"config1": {Aliases: []string{"alias1", "alias2"}}},
			expectedConfig:     &configuration{Aliases: []string{"alias1", "alias2"}},
			expectedConfigName: "config1",
			alias:              "alias1",
		},
		{
			configs: map[string]*configuration{
				"config1": {Aliases: []string{"alias1", "alias2"}},
				"config2": {Aliases: []string{"alias3", "alias4"}}},
			expectedConfig:     &configuration{Aliases: []string{"alias1", "alias2"}},
			expectedConfigName: "config1",
			alias:              "alias1",
		},
		{
			configs: map[string]*configuration{
				"config1": {Aliases: []string{"alias1", "alias2"}},
				"config2": {Aliases: []string{"alias3", "alias4"}}},
			expectedConfig:     &configuration{Aliases: []string{"alias3", "alias4"}},
			expectedConfigName: "config2",
			alias:              "alias3",
		},
	}
	for _, tc := range testCases {
		actualConfig, actualConfigName := getConfigByAlias(tc.alias, tc.configs)
		assert.Equal(t, tc.expectedConfig, actualConfig)
		assert.Equal(t, tc.expectedConfigName, actualConfigName)
	}
}

func TestGetKubeLogin(t *testing.T) {

}

func TestContainsAlias(t *testing.T) {
	var testCases = []struct {
		config         *configuration
		alias          string
		expectedResult bool
	}{
		{
			config:         &configuration{},
			alias:          "cluster-1",
			expectedResult: false,
		},
		{
			config:         &configuration{Aliases: []string{"cluster-2"}},
			alias:          "cluster-1",
			expectedResult: false,
		},
		{
			config:         &configuration{Aliases: []string{"cluster-2", "cluster-3"}},
			alias:          "cluster-1",
			expectedResult: false,
		},
		{
			config:         &configuration{Aliases: []string{"cluster-1"}},
			alias:          "cluster-1",
			expectedResult: true,
		},
		{
			config:         &configuration{Aliases: []string{"cluster-2", "cluster-1"}},
			alias:          "cluster-1",
			expectedResult: true,
		},
	}
	for _, tc := range testCases {
		actualResult := containsAlias(tc.config, tc.alias)
		assert.Equal(t, tc.expectedResult, actualResult)
	}
}
