# kubectl-login

[![Circle CI](https://circleci.com/gh/Financial-Times/kubectl-login/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/kubectl-login/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/kubectl-login)](https://goreportcard.com/report/github.com/Financial-Times/kubectl-login)
[![Coverage Status](https://coveralls.io/repos/github/Financial-Times/kubectl-login/badge.svg)](https://coveralls.io/github/Financial-Times/kubectl-login)

## Config file

`$HOME/.kubectl-login.json`

```json
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

## Releases

### Install dep

#### Mac

`brew install dep` / `brew upgrade dep`

#### Other Platforms

`curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh`

### Run dep ensure

`dep ensure`

### Release

#### Build binaries

- `GOOS=linux GOARCH=amd64 go build -o kubectl-login-linux .`
- `GOOS=darwin GOARCH=amd64 go build -o kubectl-login-darwin .`
- `GOOS=windows GOARCH=amd64 go build -o kubectl-login-windows.exe .`

#### Create Github Release

- Upload the binaries and the cluster-login.sh on the release

### How to use locally

- rename binary to kubectl-login and put in on your PATH
- run `source ./cluster-login.sh  cluster-x` or `. ./cluster-login.sh  cluster-x`

#### How to [Fish](https://fishshell.com/) locally

- put the following lines in `~/.config/fish/config.fish`:
  - `set -x KUBECONFIG <PATH_TO membership-developer-setup>/kubeconfig`
  - `alias k8s-login="source <PATH_TO kubectl-login>/cluster-login.fish $argv"`
- Use with `k8s-login` (or whatever name you alias for the command)

#### How to use with ZSH

- put `export KUBECONFIG=[path-to-repo]/content-k8s-auth-setup/kubeconfig`  in `~/.zshrc`
- execute `source cluster-login.zsh udde`

## EKS

### Info

When provisioning a new EKS cluster the provisioning script will upload the
corresponding kubeconfig to S3 bucket named `upp-kubeconfig-ACCOUNT-NUMBER`.
In order to use it you need to get this kubeconfig and merge it with the
kubeconfig files of the rest of EKS clusters. This way you will have a single
united kubeconfig for all the EKS clusters.

If you want to do that automatically, there is this script
`./update-eks-kubeconfig/update-eks-kubeconfig.sh`.
The script will do the merge for you. It will store the merged
`eks-kubecofig` under `$HOME/.kube`.
More details on that in the [#step-by-step guide](#connect-to-eks-cluster-step-by-step-guide) below.

#### Ops specifics

The script `ops-eks-kubeconfig.sh` inside `update-eks-kubeconfig` directory
is intended to be used from OPS on the jumpbox (Upp Jumpbox p).
The script sits in `/usr/local/bin` and is invoked by `/etc/skel/.bashrc`
everytime a user is connected. The purpose of the script is similar - it will
download and merge the kubeconfigs for OPS and will get `kubectx` tool
on the jumpbox.

### Connect to EKS cluster (Step-by-step guide)

IMPORTANT: This guide has been moved. To read the most resent version go to [upp-docs](https://github.com/Financial-Times/upp-docs/tree/master/guides/howto/setup-eks-kubeconfig-login). 

1. Checkout the kubectl-login repo and get into the proper folder:

    ```shell
    git clone git@github.com:Financial-Times/kubectl-login.git
    cd update-eks-kubeconfig/
    ```

1. Check the script `update-eks-kubeconfig.sh` and update the EKS cluster names
  for PROD and TEST accounts, if needed:

    ```shell
    PROD_ACCOUNT_CLUSTERS=(
      ...
    )
    TEST_ACCOUNT_CLUSTERS=(
      ...
    )
    ```

    How to check what are the EKS clusters' names?

    - Log into the [AWS console](https://awslogin.in.ft.com/adfs/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices).
    - Check accordingly test and prod accounts.
    - Look for `EKS` and check the cluster names there.
      Or if you are already looged in, you can directly check the
      [EKS Cluster names](https://eu-west-1.console.aws.amazon.com/eks/home?region=eu-west-1#/clusters).

1. Once happy with the cluster names setup:

   - Make sure that you are connected to the Restricted VPN.
   - Then execute `update-eks-kubeconfig.sh`
     (you should be in its folder):

    ```shell
    bash update-eks-kubeconfig.sh
    ```

    - Restricted VPN is now no longer needed after this step.

1. Set KUBECONFIG to its proper EKS value

    ```shell
    export KUBECONFIG=$HOME/.kube/eks-kubeconfig
    ```

1. Install [kubectx](https://github.com/ahmetb/kubectx) and run it:

    ```shell
    brew install kubectx
    ```

    ```shell
    kubectx
    ```

1. Connect to an EKS cluster by specifying its name:

    ```shell
    kubectx eks-publish-dev-eu
    ```

1. Check that you are able to execute commands:

    ```shell
    kubectl get pods
    ```

### Navigation between K8S and EKS

#### Aliases

The main difference is the `KUBECONFIG` value.
So for easier access to both K8S and EKS of clusters,
you can set up some aliases to quickly change that:

```shell
alias export-k8s-kubeconfig='export KUBECONFIG=$HOME/content-k8s-auth-setup/kubeconfig'
alias export-eks-kubeconfig='export KUBECONFIG=$HOME/.kube/eks-kubeconfig'
```

Alter the path to your `content-k8s-auth-setup` repo accordingly.

Name your aliases according to your personal preference, above are just example names.
