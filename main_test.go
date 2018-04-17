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

func TestGetConfigByAlias(t *testing.T) {

}

func TestGetKubeLogin(t *testing.T) {

}
func TestContainsAlias(t *testing.T) {

}
