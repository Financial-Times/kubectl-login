package main

import (
	"context"
	"fmt"
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
	clientId   = "kubectl-login"
	configFile = ".kubectl-login.json"
)

type Configuration struct {
	Issuer      string   `json:"issuer"`
	RedirectURL string   `json:"redirectUrl"`
	LoginSecret string   `json:"loginSecret"`
	Aliases     []string `json:"aliases"`
}

func main() {
	rawConfig := getRawConfig()
	alias := getAlias()
	config, cluster := getConfigByAlias(alias, rawConfig)
	kubeLogin := getKubeLogin(config)
	ctx := context.Background()

	// Initialize a provider by specifying dex's issuer URL.
	provider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		logger.Fatalf("error: cannot initialize OIDC provider for issuer %s:%v", config.Issuer, err)
	}

	// Configure the OAuth2 config with the client values.
	oauth2Config := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: kubeLogin,
		RedirectURL:  config.RedirectURL,
		Endpoint:     provider.Endpoint(),                                      // Discovery returns the OAuth2 endpoints.
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"}, // "openid" is a required scope for OpenID Connect flows.
	}

	// Create an ID token parser.
	idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: clientId})

	cmd := exec.Command(getOpenCmd(), oauth2Config.AuthCodeURL("some state"))
	err = cmd.Start()
	if err != nil {
		logger.Fatalf("error: cannnot open browser using command %s: %v", getOpenCmd(), err)
	}

	rawToken := getToken()
	_, err = idTokenVerifier.Verify(ctx, rawToken)
	if err != nil {
		logger.Fatalf("error: token is invalid: %s\n", err.Error())
	}

	setCreds(rawToken)
	switchContext(cluster)
	notifyAndPrompt()
}

func getOpenCmd() string {
	if runtime.GOOS == "darwin" {
		return "open"
	} else {
		return "sensible-browser"
	}
}
func getRawConfig() map[string]*Configuration {
	configPath := os.Getenv("HOME") + "/" + configFile

	file, err := os.Open(configPath)
	if err != nil {
		logger.Fatalf("error: cannot open config file at %s: %v", configPath, err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		closeFile(file)
		logger.Fatalf("error: cannot read config file at %s: %v", configPath, err)
	}

	var cfg map[string]*Configuration
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		closeFile(file)
		logger.Fatalf("error: cannot unmarshal contents of config file at %s: %v", configPath, err)
	}

	closeFile(file)
	return cfg
}

func getAlias() string {
	args := os.Args[1:]
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

func getConfigByAlias(alias string, rawConfig map[string]*Configuration) (*Configuration, string) {
	for k, v := range rawConfig {
		if containsAlias(v, alias) {
			return v, k
		}
	}
	logger.Fatalf("Alias \"%s\" not found. Try '%s' to get this value.",
		Bold(Cyan(alias)), Bold(Cyan("cat $HOME/"+configFile)))
	return nil, ""
}

func getKubeLogin(config *Configuration) string {
	if os.Getenv("KUBELOGIN") != "" {
		return os.Getenv("KUBELOGIN")
	} else if config.LoginSecret != "" {
		return config.LoginSecret
	} else {
		logger.Fatal("KUBELOGIN is not set. You Can also set this in your ~/" + configFile + " file.")
		return ""
	}
}

func containsAlias(c *Configuration, s string) bool {
	for _, val := range c.Aliases {
		if val == s {
			return true
		}
	}
	return false
}

func notifyAndPrompt() {
	fmt.Printf("\nLogged in. Now try `%s` to get your context or '%s' to get started.\n",
		Cyan("kubectl config get-contexts"), Cyan("kubectl get pods"))
}

func getToken() string {
	fmt.Print(Cyan("Enter token: "))

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

func setCreds(token string) {
	tstr := fmt.Sprintf("--token=%s", token)
	cmd := exec.Command("kubectl", "config", "set-credentials", clientId, tstr)
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("error: cannot set kubectl credentials: %v", err)
	}
}

func switchContext(cluster string) {
	clusterArg := fmt.Sprintf("--cluster=%s", cluster)
	user := fmt.Sprintf("--user=%s", clientId)
	cmd := exec.Command("kubectl", "config", "set-context", "kubectl-login-context", user, clusterArg, "--namespace=default")
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("error: cannot set kubectl login context: %v", err)
	}

	cmd = exec.Command("kubectl", "config", "use-context", "kubectl-login-context")
	err = cmd.Run()
	if err != nil {
		logger.Fatalf("error: cannot switch to kubectl login context: %v", err)
	}
}
