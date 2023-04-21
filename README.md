# tcli
[![Go](https://github.com/middlewaregruppen/tcli/actions/workflows/go.yaml/badge.svg)](https://github.com/middlewaregruppen/tcli/actions/workflows/go.yaml)

`tcli` is a command line utility for managing VMWare Tanzu namespaces and clusters

**Work in progress** *tcli is still under active development and most features are still in an idéa phase. Please check in from time to time to follow the progress* 🧡

## Installing
Download pre-built binaries from [Releases](https://github.com/middlewaregruppen/tcli/releases)

## Using
The first thing you'll want to do is probably logging in. The `login` command will authenticate the user using their SSO credentials. 
```bash
tcli login -s https://supervisor.local -u beyonce -p 'MyP5ssW0rD'
```

Too many flags? You can use environment variables prefixed with `TCLI_` so you don't have to provide them each time. For example
```bash
export TCLI_SERVER=https://supervisor.local
export TCLI_USERNAME=beyonce
export TCLI_PASSWORD="MyP5ssW0rD"
```

Other useful things you can do
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

# Logging in to a cluster will add a new context to your kubectl config file (kubeconfig)
$ tcli login beyonce-prod -n beyonces-ns
$ kubectl get pods -A
```

*The architecture of Tanzu does not allow you to use the same credentials for the supervisor cluster and guest clusters. So we have to log in to each cluster separately*

## Contributing
We love feedback! Please let us know if we've made an oopsie somewhere. The easiest and best way to provide feedback, report bugs or discussing features is to open an [Issue](https://github.com/middlewaregruppen/tcli/issues). Also, you are more than welcome to open a PR to submit contributions.