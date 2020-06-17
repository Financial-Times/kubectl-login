# kubectl-login

[![Circle CI](https://circleci.com/gh/Financial-Times/kubectl-login/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/kubectl-login/tree/master)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/kubectl-login)](https://goreportcard.com/report/github.com/Financial-Times/kubectl-login) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/kubectl-login/badge.svg)](https://coveralls.io/github/Financial-Times/kubectl-login)

## Config file

`$HOME/.kubectl-login.json`

```
{
  "cluster-1": {
    "issuer": "https://dex-for-cluster-1.example.com",
    "redirectUrl": "https://dex-redirect-for-cluster-1.example.com/callback",
    "loginSecret": "some shared secret",
    "aliases": ["test"]
  },
  "cluster-2": {
    "issuer": "https://dex-for-cluster-2.example.com",
    "redirectUrl": "https://dex-redirect-for-cluster-2.example.com/callback",
    "loginSecret": "some shared secret",
    "aliases": ["prod"]
  }
}
```

# Releases

## Install dep

### Mac
`brew install dep` / `brew upgrade dep`

### Other Platforms
`curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh`

## Run dep ensure
`dep ensure`

## Release
### Build binaries
* `GOOS=linux GOARCH=amd64 go build -o kubectl-login-linux .`
* `GOOS=darwin GOARCH=amd64 go build -o kubectl-login-darwin .`
* `GOOS=windows GOARCH=amd64 go build -o kubectl-login-windows.exe .`
### Create Github Release
* Upload the binaries and the cluster-login.sh on the release

## How to use locally
* rename binary to kubectl-login and put in on your PATH
* run `source ./cluster-login.sh  cluster-x` or `. ./cluster-login.sh  cluster-x`

### How to [Fish](https://fishshell.com/) locally
* put the following lines in `~/.config/fish/config.fish`:
    * `set -x KUBECONFIG <PATH_TO membership-developer-setup>/kubeconfig`
    * `alias k8s-login="source <PATH_TO kubectl-login>/cluster-login.fish $argv"`
* Use with `k8s-login` (or whatever name you alias for the command)

### How to use with ZSH
* put `export KUBECONFIG=[path-to-repo]/content-k8s-auth-setup/kubeconfig`  in `~/.zshrc`
* execute `source cluster-login.zsh udde`

### Only if you are connecting to EKS

When provisioning a new EKS cluster the provisioning script will upload the corresponding kubeconfig
to S3 bucket named "upp-kubeconfig-ACCOUNT-NUMBER". In order to use it you need to get this kubeconifg and merge it
with the kubeconfig files of the rest of EKS clusters. This way you will have a single united
kubeconfig for all the EKS clusters. Inside `update-eks-kubeconfig` directory there is a script
named `update-eks-kubeconfig.sh`. The script will do this merge for you and will store the merged
eks-kubecofig under $HOME/.kube
Connect to the Restricted VPN and execute:

```shell
cd update-eks-kubeconfig/
bash update-eks-kubeconfig.sh
```

If you want create an alias to easily change the KUBECONFIG:

```shell
alias export-eks-kubeconfig="export KUBECONFIG=/$HOME/.kube/eks-kubeconfig"
```

The script `ops-eks-kubeconfig.sh` inside `update-eks-kubeconfig` directory is intended to be used
from OPS on the jumpbox (Upp Jumpbox p). The script sits in `/usr/local/bin` and is invoked by
`/etc/skel/.bashrc` everytime a user is connected. The purpose of the script is similar - it will download and
merge the kubeconfigs for OPS and will get `kubectx` tool on the jumpbox.

#### Step-by-Step guide how to connect to EKS clusters

1. Connect to Restricted VPN
2. Checkout the kubectl-login repo
```shell
git clone git@github.com:Financial-Times/kubectl-login.git
```
3. Get in `update-eks-kubeconfig/` folder
```shell
cd update-eks-kubeconfig/
```
4. Edit `update-eks-kubeconfig.sh` script and fill in the EKS cluster names in PROD and TEST accounts
```shell
PROD_ACCOUNT_CLUSTERS=(
  eks-pac-staging-eu
  eks-pac-staging-us

)
TEST_ACCOUNT_CLUSTERS=(
  eks-publish-staging-eu
  eks-delivery-staging-eu
)
```
5. Execute `update-eks-kubeconfig.sh`
```shell
bash update-eks-kubeconfig.sh
```
Restricted VPN is now no longer needed.

6. Export KUBECONFIG
```shell
export KUBECONFIG=$HOME/.kube/eks-kubeconfig
```
7. Install [kubectx](https://github.com/ahmetb/kubectx)
```shell
brew install kubectx
```
8. Run kubectx
```shell
laptop$ kubectx
eks-delivery-test-eu
eks-pac-test-eu
laptop$
```
9. Connect to EKS cluster
```shell
laptop$ kubectx eks-delivery-test-eu
```
10. Profit
```shell
kubectl get pods
```
