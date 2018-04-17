package main

import (
	"testing"

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
func TestGetAlias(t *testing.T) {

}
func TestGetConfigByAlias(t *testing.T) {

}

func TestGetKubeLogin(t *testing.T) {

}
func TestContainsAlias(t *testing.T) {

}
