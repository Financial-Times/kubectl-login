package main

import (
	"io/ioutil"
	"os/exec"
	"testing"

	"os"

	"encoding/json"

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
	src, _ := ioutil.TempFile(os.TempDir(), "test")
	defer os.Remove(src.Name())
	expectedData, _ := json.Marshal(validConfig)
	src.Write(expectedData)

	masterConfig := src.Name()
	clusterConfig := switchConfig(masterConfig, "cluster-test")

	assert.True(t, clusterConfig == masterConfig+"_cluster-test")

	actualData, err := ioutil.ReadFile(clusterConfig)
	if err != nil {
		t.Fatalf("cannot read cluster config")
	}
	assert.Equal(t, expectedData, actualData)
	os.Remove(clusterConfig)
}

func TestCopyConfigSuccess(t *testing.T) {
	src, _ := ioutil.TempFile(os.TempDir(), "test")
	defer os.Remove(src.Name())
	expectedData, _ := json.Marshal(validConfig)
	src.Write(expectedData)

	dst := os.TempDir() + string(os.PathSeparator) + "test_dst"
	defer os.Remove(dst)

	copyConfig(src.Name(), dst)

	actualData, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatalf("cannot read destination file")
	}

	assert.Equal(t, expectedData, actualData)
}

func TestCopyConfigWrongSrcPath(t *testing.T) {
	if os.Getenv("CRASH") == "true" {
		copyConfig("", "")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCopyConfigWrongSrcPath")
	cmd.Env = append(os.Environ(), "CRASH=true")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("copyConfig should exit if cannot open source file")
}

func TestCopyConfigWrongDstPath(t *testing.T) {
	if os.Getenv("CRASH") == "true" {
		file, _ := ioutil.TempFile(os.TempDir(), "prefix")
		defer os.Remove(file.Name())
		copyConfig(file.Name(), "")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCopyConfigWrongDstPath")
	cmd.Env = append(os.Environ(), "CRASH=true")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("copyConfig should exit if cannot create destination file")
}

func TestGetRawConfigFileNotFound(t *testing.T) {
	if os.Getenv("CRASH") == "true" {
		getRawConfig()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestGetRawConfigFileNotFound")
	cmd.Env = append(os.Environ(), "CRASH=true")
	cmd.Env = append(cmd.Env, "HOME=.")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("getRawConfig should exit on config file not found")
}

func TestGetRawConfigInvalidContents(t *testing.T) {
	if os.Getenv("CRASH") == "true" {
		file, _ := ioutil.TempFile(os.TempDir(), "prefix")
		defer os.Remove(file.Name())
		file.Write([]byte("this is not a {valid} json content"))
		file.Sync()
		getRawConfig()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestGetRawConfigFileNotFound")
	cmd.Env = append(os.Environ(), "CRASH=true", "HOME="+os.TempDir())
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("getRawConfig should exit on invalid config file contents")
}

func TestGetRawConfigValidConfig(t *testing.T) {
	testConfigFile := os.TempDir() + string(os.PathSeparator) + configFile
	marshaledConfig, _ := json.Marshal(validConfig)
	ioutil.WriteFile(testConfigFile, marshaledConfig, 0644)
	defer os.Remove(testConfigFile)

	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", os.TempDir())

	actualConfig := getRawConfig()
	assert.Equal(t, validConfig, actualConfig)
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

func TestGetKubeLoginNotSet(t *testing.T) {
	if os.Getenv("CRASH") == "true" {
		getKubeLogin(&configuration{})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestGetKubeLoginNotSet")
	cmd.Env = append(os.Environ(), "CRASH=true")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("getKubeLogin should exit on kubelogin not founds")
}

func TestGetKubeLoginSuccessCases(t *testing.T) {
	var testCases = []struct {
		config            *configuration
		envVar            string
		expectedKubeLogin string
	}{
		{
			config:            &configuration{LoginSecret: "secret1"},
			envVar:            "",
			expectedKubeLogin: "secret1",
		},
		{
			config:            &configuration{LoginSecret: ""},
			envVar:            "secret1",
			expectedKubeLogin: "secret1",
		},
	}
	for _, tc := range testCases {
		if len(tc.envVar) > 0 {
			os.Setenv("KUBELOGIN", tc.expectedKubeLogin)
		}
		actualKubeLogin := getKubeLogin(tc.config)
		assert.Equal(t, tc.expectedKubeLogin, actualKubeLogin)
		os.Unsetenv("KUBELOGIN")
	}
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

var validConfig = map[string]*configuration{
	"config1": {
		Issuer:      "https://upp-k8s-cluster.ft.com",
		RedirectURL: "https://cluster-redirect.ft.com/callback",
		LoginSecret: "terces",
		Aliases:     []string{"alias1", "alias2"},
	},
	"config2": {
		Issuer:      "https://upp-k8s-cluster2.ft.com",
		RedirectURL: "https://cluster-redirect2.ft.com/callback",
		LoginSecret: "2terces",
		Aliases:     []string{"alias3", "alias4"},
	}}
