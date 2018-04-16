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

## Compile binaries
* Cross compile binary for linux with with `GOOS=linux GOARCH=amd64 go build -o kubectl-login-linux .`
* Cross compile binary for windows with `GOOS=windows GOARCH=amd64 go build -o kubectl-login-windows.exe .`
* Upload the binaries to github and create a release

## How to login
* rename binary to kubectl-login
* run `source ./kubectl-login.sh  cluster-x` or `. ./kubectl-login.sh  cluster-x`