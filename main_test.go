package main

import (
	"io/ioutil"
	"os/exec"
	"testing"

	"os"

	"encoding/json"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
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

func TestExtractTokens(t *testing.T) {
	idToken := "Hjjhhasdft.ADDGfaerrgg.asdf"
	refreshToken := "HLKKDFfdgggAAA"
	var testCases = []struct {
		description          string
		input                string
		expectedIdToken      string
		expectedRefreshToken string
	}{
		{
			description:          "idtoken + refresh token happy case",
			input:                idToken + ";" + refreshToken,
			expectedIdToken:      idToken,
			expectedRefreshToken: refreshToken,
		},
		{
			description:          "idtoken only",
			input:                idToken,
			expectedIdToken:      idToken,
			expectedRefreshToken: "",
		},
		{
			description:          "empty input",
			input:                "",
			expectedIdToken:      "",
			expectedRefreshToken: "",
		},
		{
			description:          "more than 3 tokens included. one is unknown, but we should ignore it",
			input:                idToken + ";" + refreshToken + ";" + "some other nonsense",
			expectedIdToken:      idToken,
			expectedRefreshToken: refreshToken,
		},
	}
	for _, tc := range testCases {
		actualIdToken, actualRefreshToken := extractTokens(tc.input)
		assert.Equal(t, tc.expectedIdToken, actualIdToken, "Scenario: "+tc.description)
		assert.Equal(t, tc.expectedRefreshToken, actualRefreshToken, "Scenario: "+tc.description)
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

func TestSetIdTokenCreds(t *testing.T) {
	if os.Getenv("KUBECTL_AVAILABLE") == "FALSE" {
		t.Skip("skipping test: kubectl is not available")
	}

	kubeConfig, _ := ioutil.TempFile(os.TempDir(), "test")
	defer os.Remove(kubeConfig.Name())
	kubeConfig.Write([]byte(testKubeconfig))
	kubeConfig.Sync()

	expectedToken := "WQ1NDZiOGZkMTA4NWFkMzExZ"
	setIdTokenCreds(expectedToken, kubeConfig.Name())

	newKubeconfigRaw, _ := ioutil.ReadFile(kubeConfig.Name())
	newKubeconfig := parseIdTokenConfig(newKubeconfigRaw, t)
	assert.True(t, len(newKubeconfig.Users) > 0)
	assert.Equal(t, expectedToken, newKubeconfig.Users[0].UserData.Token)
}

func TestSetOIDCCreds(t *testing.T) {
	if os.Getenv("KUBECTL_AVAILABLE") == "FALSE" {
		t.Skip("skipping test: kubectl is not available")
	}

	kubeConfig, _ := ioutil.TempFile(os.TempDir(), "test")
	defer os.Remove(kubeConfig.Name())
	kubeConfig.Write([]byte(testKubeconfig))
	kubeConfig.Sync()

	expToken := "WQ1NDZiOGZkMTA4NWFkMzExZ"
	expClientSecret := "llldGgadfgkKjadfllj"
	expRefreshToken := "GHHHDLKJLKJDOIIKL"
	expIdpIssuerUrl := "https://upp-k8s-dev-delivery-eu-dex.ft.com"
	setOIDCAuth(expClientSecret, expToken, expRefreshToken, expIdpIssuerUrl, kubeConfig.Name())

	newKubeconfigRaw, _ := ioutil.ReadFile(kubeConfig.Name())
	newKubeconfig := parseOIDCAuthConfig(newKubeconfigRaw, t)
	assert.True(t, len(newKubeconfig.Users) > 0)
	assert.Equal(t, clientID, newKubeconfig.Users[0].Name)

	oauthConfig := newKubeconfig.Users[0].OIDCUserData.OIDCAuthProvider.OIDCAuthProviderConfig
	assert.Equal(t, expToken, oauthConfig.IDToken)
	assert.Equal(t, expClientSecret, oauthConfig.ClientSecret)
	assert.Equal(t, clientID, oauthConfig.ClientID)
	assert.Equal(t, expRefreshToken, oauthConfig.RefreshToken)
	assert.Equal(t, expIdpIssuerUrl, oauthConfig.IDPIssuerURL)
}

func TestSetSwitchContext(t *testing.T) {
	if os.Getenv("KUBECTL_AVAILABLE") == "FALSE" {
		t.Skip("skipping test: kubectl is not available")
	}

	kubeConfig, _ := ioutil.TempFile(os.TempDir(), "test")
	defer os.Remove(kubeConfig.Name())
	kubeConfig.Write([]byte(testKubeconfig))
	kubeConfig.Sync()

	expectedCluster := Context{
		Name: "k8s-test-publishing-cluster",
		ContextData: ContextData{
			ClusterName: "k8s-test-publishing-cluster",
			Namespace: "default",
			User: "kubectl-login",
		},
	}
	switchContext(expectedCluster.Name, kubeConfig.Name())

	newKubeconfigRaw, _ := ioutil.ReadFile(kubeConfig.Name())
	newKubeconfig := parseIdTokenConfig(newKubeconfigRaw, t)

	assert.NotEmpty(t, newKubeconfig.Contexts)
	assert.Equal(t, expectedCluster.Name, newKubeconfig.CurrentContext)
	assert.Contains(t, newKubeconfig.Contexts, expectedCluster )
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

const testKubeconfig = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority: ca.pem
    server: https://test-delivery.ft.com
  name: k8s-test-delivery-cluster
- cluster:
    certificate-authority: ca.pem
    server: https://test-publishing.ft.com
  name: k8s-test-publishing-cluster
contexts:
- context:
    cluster: k8s-test-delivery-cluster
    namespace: default
    user: ""
  name: k8s-test-delivery-context
- context:
    cluster: k8s-test-publishing-cluster
    namespace: default
    user: ""
  name: k8s-test-publishing-context
- context:
    cluster: random
    namespace: default
    user: kubectl-login
  name: kubectl-login-context
current-context: kubectl-login-context
kind: Config
preferences: {}
users:
- name: kubectl-login
  user:
    token: foobar
`

type OIDCKubeConfig struct {
	ApiVersion     string     `yaml:"apiVersion"`
	Clusters       []Cluster  `yaml:"clusters"`
	Contexts       []Context  `yaml:"contexts"`
	CurrentContext string     `yaml:"current-context"`
	Kind           string     `yaml:"kind"`
	Users          []OIDCUser `yaml:"users"`
}

type OIDCUser struct {
	Name         string       `yaml:"name"`
	OIDCUserData OIDCUserData `yaml:"user"`
}

type OIDCUserData struct {
	OIDCAuthProvider OIDCAuthProvider `yaml:"auth-provider"`
}

type OIDCAuthProvider struct {
	Name                   string                 `yaml:"name"`
	OIDCAuthProviderConfig OIDCAuthProviderConfig `yaml:"config"`
}

type OIDCAuthProviderConfig struct {
	ClientID     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
	IDToken      string `yaml:"id-token"`
	IDPIssuerURL string `yaml:"idp-issuer-url"`
	RefreshToken string `yaml:"refresh-token"`
}

type IdTokenKubeConfig struct {
	ApiVersion     string        `yaml:"apiVersion"`
	Clusters       []Cluster     `yaml:"clusters"`
	Contexts       []Context     `yaml:"contexts"`
	CurrentContext string        `yaml:"current-context"`
	Kind           string        `yaml:"kind"`
	Users          []IdTokenUser `yaml:"users"`
}

type IdTokenUser struct {
	Name     string   `yaml:"name"`
	UserData UserData `yaml:"user"`
}

type UserData struct {
	Token string `yaml:"token"`
}
type Cluster struct {
	Name        string      `yaml:"name"`
	ClusterData ClusterData `yaml:"cluster"`
}

type ClusterData struct {
	CertificateAuthority string `yaml:"certificate-authority"`
	Server               string `yaml:"server"`
}

type Context struct {
	Name        string      `yaml:"name"`
	ContextData ContextData `yaml:"context"`
}

type ContextData struct {
	ClusterName string `yaml:"cluster"`
	Namespace   string `yaml:"namespace"`
	User        string `yaml:"user"`
}

func parseIdTokenConfig(rawConfig []byte, t *testing.T) IdTokenKubeConfig {
	var config IdTokenKubeConfig
	if err := yaml.Unmarshal(rawConfig, &config); err != nil {
		t.Fatal(err)
	}
	return config
}

func parseOIDCAuthConfig(rawConfig []byte, t *testing.T) OIDCKubeConfig {
	var config OIDCKubeConfig
	if err := yaml.Unmarshal(rawConfig, &config); err != nil {
		t.Fatal(err)
	}
	return config
}
