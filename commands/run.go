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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/groupssettings/v1"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"

	"github.com/justaugustus/ggreconcile/groups"
	"github.com/justaugustus/ggreconcile/util"
)

func Usage() {
	fmt.Fprintf(os.Stderr, `
Usage: %s [-config <config-yaml-file>] [--confirm]
Command line flags override config values.
`, os.Args[0])
	flag.PrintDefaults()
}

var config util.Config
var groupsConfig groups.Config

var verbose = flag.Bool("v", false, "log extra information")

func Run() {
	configFilePath := flag.String("config", "config.yaml", "the config file in yaml format")
	confirmChanges := flag.Bool("confirm", false, "false by default means that we do not push anything to google groups")
	printConfig := flag.Bool("print", false, "print the existing group information")

	flag.Usage = Usage
	flag.Parse()

	err := readConfig(*configFilePath, *confirmChanges)
	if err != nil {
		log.Fatal(err)
	}

	// rootDir contains groups.yaml files
	rootDir := filepath.Dir(*configFilePath)
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

	serviceAccountKey, err := util.AccessSecretVersion(config.SecretVersion)

	credential, err := google.JWTConfigFromJSON(serviceAccountKey, admin.AdminDirectoryUserReadonlyScope,
		admin.AdminDirectoryGroupScope,
		admin.AdminDirectoryGroupMemberScope,
		groupssettings.AppsGroupsSettingsScope)
	if err != nil {
		log.Fatalf("Unable to authenticate using key in secret-version %s, %v", config.SecretVersion, err)
	}
	credential.Subject = config.BotID

	ctx := context.Background()
	client := credential.Client(ctx)
	clientOption := option.WithHTTPClient(client)

	srv, err := admin.NewService(ctx, clientOption)
	if err != nil {
		log.Fatalf("Unable to retrieve directory Client %v", err)
	}

	srv2, err := groupssettings.NewService(ctx, clientOption)
	if err != nil {
		log.Fatalf("Unable to retrieve groupssettings Service %v", err)
	}

	if *printConfig {
		err = groups.PrintMembersAndSettings(srv, srv2)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Println(" ======================= Updates =========================")
	for _, g := range groupsConfig.Groups {
		if g.EmailId == "" {
			log.Fatal(fmt.Sprintf("Group has no email-id: %#v", g))
		}

		err = groups.CreateOrUpdateIfNecessary(config, srv, g.EmailId, g.Name, g.Description)
		if err != nil {
			log.Fatal(err)
		}
		err = groups.UpdateSettings(config, srv2, g.EmailId, g.Settings)
		if err != nil {
			log.Fatal(err)
		}
		err = groups.AddOrUpdateMember(config, srv, g.EmailId, g.Owners, "OWNER")
		if err != nil {
			log.Fatal(err)
		}
		err = groups.AddOrUpdateMember(config, srv, g.EmailId, g.Managers, "MANAGER")
		if err != nil {
			log.Fatal(err)
		}
		err = groups.AddOrUpdateMember(config, srv, g.EmailId, g.Members, "MEMBER")
		if err != nil {
			log.Println(err)
		}
		if g.Settings["ReconcileMembers"] == "true" {
			members := append(g.Owners, g.Managers...)
			members = append(members, g.Members...)
			err = groups.RemoveMembers(config, srv, g.EmailId, members)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			members := append(g.Owners, g.Managers...)
			err = groups.RemoveOwnerOrManagers(config, srv, g.EmailId, members)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	err = groups.DeleteIfNecessary(config, groupsConfig, srv)
	if err != nil {
		log.Fatal(err)
	}
}

func readConfig(configFilePath string, confirmChanges bool) error {
	if *verbose {
		log.Printf("reading config file %s", configFilePath)
	}
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("error reading config file %s: %v", configFilePath, err)
	}
	if err = yaml.Unmarshal(content, &config); err != nil {
		return fmt.Errorf("error parsing config file %s: %v", configFilePath, err)
	}
	config.ConfirmChanges = confirmChanges
	return err
}
