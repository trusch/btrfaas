Getting started with nodejs
===========================

## 1. Install `btrfaasctl`
```bash
curl -sL https://github.com/trusch/btrfaas/releases/download/v0.1.0/btrfaasctl > /tmp/btrfaasctl
chmod +x /tmp/btrfaasctl
sudo mv /tmp/btrfaasctl /usr/bin/
```

## 2. Init your deployment
```bash
btrfaasctl init
```

## 3. Create and build your function
```bash
btrfaasctl function init my-function --template python
# edit my-function/Server.py to fit your needs
btrfaasctl function build my-function
```

## 4. Deploy and test your function
```bash
btrfaasctl function deploy my-function/function.yaml
echo "Hello World" | btrfaasctl function invoke my-function
```
