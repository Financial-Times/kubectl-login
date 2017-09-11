package main

import (
	"golang.org/x/oauth2"
	"github.com/coreos/go-oidc"
	"context"
	"fmt"
	"log"
	"os/exec"
	"os"
	"syscall"
	"strings"
	"golang.org/x/crypto/ssh/terminal"
	"flag"

	. "github.com/logrusorgru/aurora"
	"encoding/json"
	"io/ioutil"
)

var logger = log.New(os.Stdout, "", log.LUTC)

type Configuration struct {
	Issuer      string `json:"issuer"`
	RedirectURL string `json:"redirectUrl"`
	LoginSecret string `json:"loginSecret"`
	Cluster     string `json:"cluster"`
}

func getConfigFile(clusterName string) Configuration {
	file, err := os.Open(os.Getenv("HOME") + "/.kubectl-login.json")
	if err != nil {
		logger.Fatal("error:", err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		logger.Fatal("error:", err)
	}
	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(data, &objmap)
	if err != nil {
		logger.Fatal("error:", err)
	}
	var configuration Configuration

	if val, ok := objmap[clusterName]; ok {
		err = json.Unmarshal(*val, &configuration)
		return configuration
	} else {
		logger.Fatal(fmt.Sprintf("Cluster %s not found in ~/.kubectl-login.json file", Bold(Cyan(clusterName))))
		return Configuration{}
	}
}

func main() {

	clusterPtr := flag.String("cluster", "", "The Cluster to login to")
	flag.Parse()

	var cluster string

	if *clusterPtr != "" {
		cluster = *clusterPtr
	} else {
		logger.Fatal(fmt.Sprintf("Cluster flag is mandatory. try '%s' to get this value.", Bold(Cyan("kubectl config get-contexts"))))
	}

	config := getConfigFile(cluster)

	var kl string

	if os.Getenv("KUBELOGIN") != "" {
		kl = os.Getenv("KUBELOGIN")
	} else if config.LoginSecret != "" {
		kl = config.LoginSecret
	} else {
		logger.Fatal("KUBELOGIN is not set. You Can also set this in your ~/.kubectl-login.json file.")
	}

	ctx := context.Background()
	// Initialize a provider by specifying dex's issuer URL.
	provider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		logger.Fatal(err)
	}

	// Configure the OAuth2 config with the client values.
	oauth2Config := oauth2.Config{
		// client_id and client_secret of the client.
		ClientID:     "kubectl-login",
		ClientSecret: kl,

		// The redirectURL.
		RedirectURL: config.RedirectURL,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		//
		// Other scopes, such as "groups" can be requested.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	}

	// Create an ID token parser.

	idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: "kubectl-login"})

	acu := oauth2Config.AuthCodeURL("some state")

	cmd := exec.Command("open", acu)
	err = cmd.Start()
	if err != nil {
		logger.Fatal(err)
	}

	rawToken := getToken()

	_, err = idTokenVerifier.Verify(ctx, rawToken)

	if err != nil {
		logger.Printf("token is invalid, error: %s\n", err.Error())
		return
	}

	setCreds(rawToken)
	switchContext(cluster)
	notifyAndPrompt()
}

func notifyAndPrompt() {
	fmt.Printf("\nLogged in. Now try `%s` to get your context or '%s' to get started.\n", Cyan("kubectl config get-contexts"), Cyan("kubectl get pods"))
}

func getToken() string {
	fmt.Print(Cyan("Enter token: "))
	byteToken, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		logger.Fatal(err)
	}
	token := string(byteToken)

	return strings.TrimSpace(token)
}

func setCreds(token string) {
	tstr := fmt.Sprintf("--token=%s", token)
	cmd := exec.Command("kubectl", "config", "set-credentials", "kubectl-login", tstr)
	err := cmd.Run()
	if err != nil {
		logger.Fatal(err)
	}
}

func switchContext(cluster string) {

	clusterArg := fmt.Sprintf("--cluster=%s", cluster)
	cmd := exec.Command("kubectl", "config", "set-context", "kubectl-login-context", "--user=kubectl-login", clusterArg, "--namespace=default")
	err := cmd.Run()
	if err != nil {
		logger.Fatal(err)
	}

	cmd = exec.Command("kubectl", "config", "use-context", "kubectl-login-context")
	err = cmd.Run()
	if err != nil {
		logger.Fatal(err)
	}
}
