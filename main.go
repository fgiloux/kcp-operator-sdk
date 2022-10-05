// https://github.com/operator-framework/operator-sdk/blob/master/cmd/operator-sdk/main.go

package main

import (
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that `exec-entrypoint` and `run` can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	log "github.com/sirupsen/logrus"

	"github.com/fgiloux/kcp-operator-sdk/internal/cmd/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		log.Fatal(err)
	}
}
