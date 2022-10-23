# kcp-operator-sdk

This project contains code generators to help with building cluster / [kcp](https://github.com/kcp-dev/kcp) aware operators.
## Description

After running a few commands the sdk generates all the plumbing so that operator authors can focus on designing their API and implementing the business logic.

kcp-operator-sdk will generate for you:

- Code doing the json conversion for the API (custom resource) you have defined
- Configuration files for
  - creating the CRD or the APIResourceSchema for your API
  - creating a kcp APIExport resource for the service
  - deploying the new controller on kcp or plain Kubernetes
  - setting up RBAC
- Code for the controller and its manager
- A Dockerfile for creating a container image with the controller
- A Makefile
- A skeleton for end to end tests

The project is based on [Kubebuilder](https://book.kubebuilder.io/) CLI extensions (plugins) and most of the features describe in the Kubebuilder documentation for go based operators are also available with the kcp-operator-sdk.

Here is a  ~8 minutes demo  of the operator.
[![asciicast](https://asciinema.org/a/531109.svg)](https://asciinema.org/a/531109)

## Test it out!

Generate the plumbing for your operator.

~~~
$ kcp-operator-sdk init --component-config --domain tutorial.kubebuilder.io --repo github.com/yourrepo/kb-kcp-tutorial
$ kcp-operator-sdk create api --version v1alpha1 --kind Widget
$ make manifests apiresourceschemas
~~~

After you have done modifications to the API, rerun the manifest generation.

~~~
$ make manifests apiresourceschemas
~~~

Build some reconciliation logic and test it with the end-to-end tests

~~~
$ make test-e2e
~~~

**NOTE:** Run `make --help` for more information on all potential `make` targets

## License

Operator SDK is under Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.

