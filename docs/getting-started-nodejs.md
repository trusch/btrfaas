Getting started with nodejs
===========================

## 1. Install `btrfaasctl`
```bash
curl -sL https://raw.githubusercontent.com/trusch/btrfaas/master/install.sh | sh
```

## 2. Init your deployment
```bash
btrfaasctl init
```

## 3. Create and build your function
```bash
btrfaasctl function init my-function --template nodejs
# edit my-function/server.js to fit your needs
btrfaasctl function build my-function
```

## 4. Deploy and test your function
```bash
btrfaasctl function deploy my-function/function.yaml
echo "Hello World" | btrfaasctl function invoke my-function
```
