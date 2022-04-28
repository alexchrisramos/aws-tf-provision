# bunk

## How-To

1) Install pre-requistes and ensure they are in your `$PATH`:
* Docker: https://docs.docker.com/engine/install/
* k3d 1.7.0 (`bunk` does not currently support k3d 3.x): https://github.com/rancher/k3d/releases/tag/v1.7.0
2) Download the latest `bunk` [release](https://github.com/some-things/bunk/releases) and add it to your `$PATH`.
3) Extract Konvoy diagnostic bundle: `bunk extract <bundle-file>`
4) `cd` to the extracted bundle directory.
5) Create k3d cluster and inject bundle resources: `bunk up`
6) Analyze bundle resources with kubectl: `export KUBECONFIG="$(k3d get-kubeconfig --name='k3s-default')" && kubectl get po -A`
7) Once finished, tear down the cluster and its resources: `bunk down`
