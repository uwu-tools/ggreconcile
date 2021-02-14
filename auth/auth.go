/*
Copyright 2021 The Kubernetes Authors.

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

package auth

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/groupssettings/v1"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type Client struct {
	AdminSvc          *admin.Service
	GroupsSettingsSvc *groupssettings.Service
}

type Options struct {
	Ctx           context.Context
	BotID         string
	SecretVersion string
	Scopes        []string
}

var DefaultScopes = []string{
	admin.AdminDirectoryUserReadonlyScope,
	admin.AdminDirectoryGroupScope,
	admin.AdminDirectoryGroupMemberScope,
	groupssettings.AppsGroupsSettingsScope,
}

func NewClient(opts *Options) (client *Client, err error) {
	svcAccountKey, err := getSvcAccountKey(opts.SecretVersion)
	if err != nil {
		return client, errors.Wrap(
			err,
			"getting service account key",
		)
	}

	if len(opts.Scopes) == 0 {
		opts.Scopes = DefaultScopes
	}

	credential, err := google.JWTConfigFromJSON(
		svcAccountKey,
		opts.Scopes...,
	)
	if err != nil {
		return client, errors.Wrap(
			err,
			fmt.Sprintf(
				"authenticating using key in secret-version %s",
				opts.SecretVersion,
			),
		)
	}

	credential.Subject = opts.BotID

	ctx := opts.Ctx
	credClient := credential.Client(ctx)
	clientOption := option.WithHTTPClient(credClient)

	adminSvc, err := admin.NewService(ctx, clientOption)
	if err != nil {
		return client, errors.Wrap(
			err,
			"retrieving directory service",
		)
	}

	groupsSettingsSvc, err := groupssettings.NewService(ctx, clientOption)
	if err != nil {
		return client, errors.Wrap(
			err,
			"retrieving groupssettings service",
		)
	}

	client.AdminSvc = adminSvc
	client.GroupsSettingsSvc = groupsSettingsSvc

	return client, nil
}

// getSvcAccountKey accesses the payload for the given secret version if one exists
// secretVersion is of the form projects/{project}/secrets/{secret}/versions/{version}
// Usage: svcAccountKey, err := util.getSvcAccountKey(config.SecretVersion)
func getSvcAccountKey(secretVersion string) ([]byte, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "creating secretmanager client")
	}

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretVersion,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "accessing secret version")
	}

	return result.Payload.Data, nil
}
