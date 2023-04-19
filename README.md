# tcli
[![Release](https://github.com/middlewaregruppen/tcli/actions/workflows/release.yaml/badge.svg)](https://github.com/middlewaregruppen/tcli/actions/workflows/release.yaml)

`tcli` is a command line utility for managing VMWare Tanzu namespaces and clusters

**Work in progress** *tcli is still under active development and most features are still in an idéa phase. Please check in from time to time for follow the progress* 🧡


## Login to Tanzu Supervisor Cluster
The `login` command will authenticate the user using username/password. This will store session credentials in a temporary kubeconfig file. It's temporary because we will use the kubeconfig file with other commands without having to provide credentials each time.

```bash
tcli login -s https://supervisor.local -u bobby -p 'MyP5ssW0rD'
```

## List clusters within a namespace
After you have logged in, the cli will print a list of namespaces your account has access to. You may list clusters within each namespace with following
```bash
tcli clusters -n dev
```

## Login to a cluster
The architecture of Tanzu does not allow you to use the same credentials for the supervisor cluster and guest clusters. So we have to log in to each cluster separately. You can do this easily with following

```bash
tcli login -s https://supervisor.local -u bobby -p 'MyP5ssW0rD' -n mynamespace -c mycluster
```

The login command will update your `kubeconfig` by adding data so you can continue interacting with the clusters using `kubectl`

## Getting startet

Download tcli from [Releases](https://github.com/middlewaregruppen/tcli/releases)