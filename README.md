# tcli
[![Release](https://github.com/middlewaregruppen/tcli/actions/workflows/release.yaml/badge.svg)](https://github.com/middlewaregruppen/tcli/actions/workflows/release.yaml)

`tcli` is a command line utility for managing VMWare Tanzu namespaces and clusters

**Work in progress** *tcli is still under active development and most features are still in an idéa phase. Please check in from time to time to follow the progress* 🧡


## Login to Tanzu Supervisor Cluster
The `login` command will authenticate the user using their SSO credentials. 
```bash
tcli login -s https://supervisor.local -u beyonce -p 'MyP5ssW0rD'
```

**Pro Tip!** You can use environment variables prefixed with `TCLI_` so that you don't have to provide them for each command. For example:

```bash
# Listing namespaces 
$ tcli list namespaces
beyonces-ns
cardis-ns

# Listing clusters
$ tcli list clusters -n beyonces-ns
NAME          CONTROL PLANE   WORKER   TKR NAME                           AGE     READY   TKR COMPATIBLE   UPDATES AVAILABLE
beyonce-test   1               2        v1.22.9---vmware.1-tkg.1.cc71bc8   21d     True    True             [1.23.8+vmware.3-tkg.1]
beyonce-prod   1               2        v1.21.6---vmware.1-tkg.1.b3d708a   15d     True    True             [1.22.9+vmware.1-tkg.1.cc71bc8]

# Login to a cluster 
$ tcli login beyonce-prod -n beyonces-ns
```

*The architecture of Tanzu does not allow you to use the same credentials for the supervisor cluster and guest clusters. So we have to log in to each cluster separately*

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
