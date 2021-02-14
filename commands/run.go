/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/justaugustus/ggreconcile/auth"
	"github.com/justaugustus/ggreconcile/groups"
	"github.com/justaugustus/ggreconcile/util"
)

// TODO: Move this into cobra command
func Usage() {
	fmt.Fprintf(os.Stderr, `
Usage: %s [-config <config-yaml-file>] [--confirm]
Command line flags override config values.
`, os.Args[0])
	flag.PrintDefaults()
}

var config util.Config
var groupsConfig groups.Config

func run(opts *Options) error {
	configPath := opts.config
	printConfig := opts.print

	err := readConfig(opts)
	if err != nil {
		log.Fatal(err)
	}

	// rootDir contains groups.yaml files
	rootDir := filepath.Dir(configPath)
	if config.GroupsPath != nil {
		if !filepath.IsAbs(*config.GroupsPath) {
			log.Fatalf("groups-path \"%s\" must be an absolute path", *config.GroupsPath)
		}
		rootDir = *config.GroupsPath
	}

	err = groups.ReadConfig(rootDir, &groupsConfig)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Set this in a client instead
	authOptions := &auth.Options{
		Ctx:           context.Background(),
		BotID:         config.BotID,
		SecretVersion: config.SecretVersion,
		Scopes:        auth.DefaultScopes,
	}

	authClient, err := auth.NewClient(authOptions)
	if err != nil {
		return errors.Wrap(err, "creating an auth client")
	}

	// TODO: Set this in a client instead
	gc := &groups.Client{
		Config:     &groupsConfig,
		UtilConfig: &config,
		AuthClient: authClient,
	}

	if printConfig {
		err = gc.PrintMembersAndSettings()
		if err != nil {
			return errors.Wrap(err, "printing group members and settings")
		}

		return nil
	}

	log.Println(" ======================= Updates =========================")
	for _, g := range groupsConfig.Groups {
		if g.EmailId == "" {
			log.Fatal(fmt.Sprintf("Group has no email-id: %#v", g))
		}

		err = gc.CreateOrUpdateIfNecessary(g.EmailId, g.Name, g.Description)
		if err != nil {
			log.Fatal(err)
		}

		err = gc.UpdateSettings(g.EmailId, g.Settings)
		if err != nil {
			log.Fatal(err)
		}

		err = gc.AddOrUpdateMember(g.EmailId, g.Owners, "OWNER")
		if err != nil {
			log.Fatal(err)
		}

		err = gc.AddOrUpdateMember(g.EmailId, g.Managers, "MANAGER")
		if err != nil {
			log.Fatal(err)
		}

		err = gc.AddOrUpdateMember(g.EmailId, g.Members, "MEMBER")
		if err != nil {
			log.Println(err)
		}

		if g.Settings["ReconcileMembers"] == "true" {
			members := append(g.Owners, g.Managers...)
			members = append(members, g.Members...)

			err = gc.RemoveMembers(g.EmailId, members)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			members := append(g.Owners, g.Managers...)

			err = gc.RemoveOwnerOrManagers(g.EmailId, members)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	err = gc.DeleteIfNecessary()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func readConfig(opts *Options) error {
	configPath := opts.config
	confirmChanges := opts.confirm

	logrus.Debugf("reading config file %s", configPath)

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file %s: %v", configPath, err)
	}

	if err = yaml.Unmarshal(content, &config); err != nil {
		return fmt.Errorf("error parsing config file %s: %v", configPath, err)
	}

	config.ConfirmChanges = confirmChanges
	return err
}
