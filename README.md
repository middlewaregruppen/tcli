# tcli
[![Release](https://github.com/middlewaregruppen/tcli/actions/workflows/release.yaml/badge.svg)](https://github.com/middlewaregruppen/tcli/actions/workflows/release.yaml)

`tcli` is a command line utility for managing VMWare Tanzu namespaces and clusters

**Work in progress** *tcli is still under active development and most features are still in an idéa phase. Please check in from time to time to follow the progress* 🧡


## Login to Tanzu Supervisor Cluster
The `login` command will authenticate the user using username/password. This will store session credentials in a temporary kubeconfig file. It's temporary because we will use the kubeconfig file with other commands without having to provide credentials each time.

```bash
tcli login -s https://supervisor.local -u bobby -p 'MyP5ssW0rD'
```

## Flag values in env vars
You can use environment variables prefixed with `TCLI_` so that you don't have to provide flags for each command. For example:
```bash
export TCLI_SERVER="https://supervisor.local"
export TCLI_USERNAME="bobby"
export TCLI_PASSWORD="MyP5ssW0rD"
$ tcli login
$ tcli clusters
```

## List clusters within a namespace
After you have logged in, the cli will print a list of namespaces your account has access to. You may list clusters within each namespace with following. If no namespace is defined, then "default" will be used.
```bash
tcli clusters -n dev
```

## Login to a cluster
The architecture of Tanzu does not allow you to use the same credentials for the supervisor cluster and guest clusters. So we have to log in to each cluster separately. You can do this easily with following

```bash
tcli login -n mynamespace -c mycluster
```

The login command will update your `kubeconfig` by adding data so you can continue interacting with the clusters using `kubectl`

## CLI Usage
```
Usage:
  tcli [command]

Available Commands:
  clusters    List clusters within a Tanzu namespace
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  login       Authenticate user with Tanzu namespaces and clusters
  logout      Logout user and remove credentials

Flags:
  -h, --help                help for tcli
  -i, --insecure            Skip certificate verification (this is insecure). (default true)
      --kubeconfig string   Path to kubeconfig file. (default "/Users/amir/.kube/tcli")
  -p, --password string     Password to use for authentication.
  -s, --server string       Address of the server to authenticate against.
  -u, --username string     Username to authenticate.
  -v, --verbosity string    number for the log level verbosity (debug, info, warn, error, fatal, panic) (default "info")
```

## Getting started

Download pre-built binaries from [Releases](https://github.com/middlewaregruppen/tcli/releases)
