package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/coreos/go-oidc"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"

	"encoding/json"
	"io/ioutil"

	"runtime"

	. "github.com/logrusorgru/aurora"
)

var logger = log.New(os.Stdout, "", log.LUTC)

const (
	clientID   = "kubectl-login"
	configFile = ".kubectl-login.json"
	state      = "csrf-protection-state"
)

type configuration struct {
	Issuer      string   `json:"issuer"`
	RedirectURL string   `json:"redirectUrl"`
	LoginSecret string   `json:"loginSecret"`
	Aliases     []string `json:"aliases"`
}

func main() {
	rawConfig := getRawConfig()
	alias := getAlias(os.Args[1:])
	config, cluster := getConfigByAlias(alias, rawConfig)

	currentKubeconfig := os.Getenv("KUBECONFIG")
	masterKubeconfig := currentKubeconfig
	if !isMasterConfig(currentKubeconfig) {
		if isCurrentContext(cluster) && isLoggedIn(currentKubeconfig) {
			logger.Printf("Already logged in to cluster %s", cluster)
			//exit with error code so wrapper script will output this message
			os.Exit(1)
		} else {
			masterKubeconfig = strings.Split(currentKubeconfig, "_")[0]
		}
	}

	newKubeconfig := switchConfig(masterKubeconfig, cluster)
	kubeLogin := getKubeLogin(config)
	ctx := context.Background()

	// Initialize a provider by specifying dex's issuer URL.
	provider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		logger.Fatalf("error: cannot initialize OIDC provider for issuer %s:%v", config.Issuer, err)
	}

	// Configure the OAuth2 config with the client values.
	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: kubeLogin,
		RedirectURL:  config.RedirectURL,
		Endpoint:     provider.Endpoint(),                                      // Discovery returns the OAuth2 endpoints.
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"}, // "openid" is a required scope for OpenID Connect flows.
	}

	if err = openBrowser(oauth2Config.AuthCodeURL(state)); err != nil {
		logger.Fatalf("error: cannot open browser: %v", err)
	}

	idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: clientID})
	rawToken := getToken()
	if _, err = idTokenVerifier.Verify(ctx, rawToken); err != nil {
		logger.Fatalf("error: token is invalid: %v", err)
	}

	setCreds(rawToken, newKubeconfig)
	switchContext(cluster, newKubeconfig)
	if !isLoggedIn(newKubeconfig) {
		logger.Fatal("error: kubectl command didn't work, even after login!")
	}
	//output the new kubeconfig path, used in the wrapper to set the env variable
	logger.Printf(newKubeconfig)
}

func isMasterConfig(kubeconfigPath string) bool {
	return len(kubeconfigPath) > 0 && !strings.Contains(kubeconfigPath, "_")
}

func switchConfig(masterConfig, cluster string) string {
	clusterKubeconfig := masterConfig + "_" + cluster
	copyConfig(masterConfig, clusterKubeconfig)
	return clusterKubeconfig
}

func copyConfig(srcPath string, dstPath string) {
	src, err := os.Open(srcPath)
	if err != nil {
		logger.Fatalf("error: could not open kubeconfig %s: %v", srcPath, err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		logger.Fatalf("error: could not create kubeconfig %s: %v", dstPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		logger.Fatalf("error: could not copy kubeconfig %s to %s: %v", srcPath, dstPath, err)
	}
}

func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform: %v", runtime.GOOS)
	}
	return err
}

func getRawConfig() map[string]*configuration {
	configPath := os.Getenv("HOME") + string(os.PathSeparator) + configFile

	file, err := os.Open(configPath)
	if err != nil {
		logger.Fatalf("error: cannot open config file at %s: %v", configPath, err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		closeFile(file)
		logger.Fatalf("error: cannot read config file at %s: %v", configPath, err)
	}

	var cfg map[string]*configuration
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		closeFile(file)
		logger.Fatalf("error: cannot unmarshal contents of config file at %s: %v", configPath, err)
	}

	closeFile(file)
	return cfg
}

func getAlias(args []string) string {
	if len(args) == 0 {
		logger.Fatalf("Alias is mandatory i.e %s. try '%s' to get this value.",
			Bold(Cyan("kubectl-login <ALIAS>")), Bold(Cyan("cat $HOME/"+configFile)))
	}
	return args[0]
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		logger.Printf("warning: couldn't close config file: %v", err)
	}
}

func getConfigByAlias(alias string, rawConfig map[string]*configuration) (*configuration, string) {
	for k, v := range rawConfig {
		if containsAlias(v, alias) {
			return v, k
		}
	}
	logger.Fatalf("Alias \"%s\" not found. Try '%s' to get this value.",
		Bold(Cyan(alias)), Bold(Cyan("cat $HOME/"+configFile)))
	return nil, ""
}

func getKubeLogin(config *configuration) string {
	if os.Getenv("KUBELOGIN") != "" {
		return os.Getenv("KUBELOGIN")
	} else if config.LoginSecret != "" {
		return config.LoginSecret
	} else {
		logger.Fatal("KUBELOGIN is not set. You Can also set this in your ~/" + configFile + " file.")
		return ""
	}
}

func containsAlias(c *configuration, s string) bool {
	for _, val := range c.Aliases {
		if val == s {
			return true
		}
	}
	return false
}

func getToken() string {
	switch runtime.GOOS {
	case "windows":
		return getTokenClearText()
	default:
		return getTokenHidden()
	}
}

func getTokenHidden() string {
	// handle restoring terminal
	stdinFd := int(os.Stdin.Fd())
	state, err := terminal.GetState(stdinFd)
	defer terminal.Restore(stdinFd, state)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	go func() {
		for range sigch {
			terminal.Restore(stdinFd, state)
			os.Exit(1)
		}
	}()
	byteToken, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		logger.Fatalf("error: cannot read token from terminal: %v", err)
	}
	token := string(byteToken)

	return strings.TrimSpace(token)
}

func getTokenClearText() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func setCreds(token, config string) {
	tstr := fmt.Sprintf("--token=%s", token)
	cfg := fmt.Sprintf("--kubeconfig=%s", config)
	cmd := exec.Command("kubectl", "config", "set-credentials", clientID, tstr, cfg)
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("error: cannot set kubectl credentials: %v", err)
	}
}

func switchContext(cluster, config string) {
	clusterArg := fmt.Sprintf("--cluster=%s", cluster)
	user := fmt.Sprintf("--user=%s", clientID)
	cfg := fmt.Sprintf("--kubeconfig=%s", config)
	cmd := exec.Command("kubectl", "config", "set-context", "kubectl-login-context", user, clusterArg, "--namespace=default", cfg)
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("error: cannot set kubectl login context: %v", err)
	}

	cmd = exec.Command("kubectl", "config", "use-context", "kubectl-login-context", cfg)
	err = cmd.Run()
	if err != nil {
		logger.Fatalf("error: cannot switch to kubectl login context: %v", err)
	}
}

func isCurrentContext(cluster string) bool {
	output, err := exec.Command("kubectl", "config", "view",
		`--output=jsonpath='{.contexts[?(@.name == "kubectl-login-context")].context.cluster}'`).CombinedOutput()
	if err != nil {
		fmt.Printf("error: cannot check current context: %v\n", err)
	}
	currentContext := strings.Trim(string(output), "'")
	return currentContext == cluster
}

func isLoggedIn(config string) bool {
	cfg := fmt.Sprintf("--kubeconfig=%s", config)
	err := exec.Command("kubectl", "get", "configmap", cfg).Run()
	if err != nil {
		logger.Fatal(err)
	}
	return err == nil
}
