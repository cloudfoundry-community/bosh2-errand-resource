package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"io/ioutil"

	"github.com/cloudfoundry-community/bosh2-errand-resource/bosh"
	"github.com/cloudfoundry-community/bosh2-errand-resource/check"
	"github.com/cloudfoundry-community/bosh2-errand-resource/concourse"
	"github.com/cloudfoundry/socks5-proxy"
)

func main() {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read configuration: %s\n", err)
		os.Exit(1)
	}

	checkRequest, err := concourse.NewCheckRequest(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid parameters: %s\n", err)
		os.Exit(1)
	}

	var checkResponse []concourse.Version

	if checkRequest.Source.SkipCheck {
		checkResponse = []concourse.Version{}
	} else {
		hostKeyGetter := proxy.NewHostKey()
		socks5Proxy := proxy.NewSocks5Proxy(hostKeyGetter, log.New(ioutil.Discard, "", log.LstdFlags))
		cliCoordinator := bosh.NewCLICoordinator(checkRequest.Source, os.Stderr, socks5Proxy)
		commandRunner := bosh.NewCommandRunner(cliCoordinator)
		cliDirector, err := cliCoordinator.Director()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		director := bosh.NewBoshDirector(
			checkRequest.Source,
			commandRunner,
			cliDirector,
		)

		checkCommand := check.NewCheckCommand(director)
		checkResponse, err = checkCommand.Run(checkRequest)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	}

	concourseOutputFormatted, err := json.MarshalIndent(checkResponse, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not generate version: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "%s", concourseOutputFormatted)
}
