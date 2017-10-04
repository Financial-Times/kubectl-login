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