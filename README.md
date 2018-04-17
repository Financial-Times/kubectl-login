# kubectl-login


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
