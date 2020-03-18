// Copyright 2019 Cloudbase Solutions SRL
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package auth

import (
	"context"
	"coriolis-logger/config"
	"fmt"
	"net/http"

	"github.com/databus23/keystone"
	"github.com/pkg/errors"
)

type keystoneAuth struct {
	auth *keystone.Auth
	cfg  *config.KeystoneAuth
}

func (k keystoneAuth) rolesAsMap() map[string]bool {
	ret := map[string]bool{}
	if k.cfg == nil {
		return ret
	}

	for _, val := range k.cfg.AdminRoles {
		ret[val] = true
	}
	return ret
}

func (k keystoneAuth) Authenticate(req *http.Request) (context.Context, error) {
	authToken := req.Header.Get("X-Auth-Token")
	if authToken == "" {
		authType := req.URL.Query().Get("auth_type")
		if authType == config.AuthenticationKeystone {
			authToken = req.URL.Query().Get("auth_token")
		}
		if authToken == "" {
			return nil, fmt.Errorf("missing token in headers")
		}
	}

	keystoneContext, err := k.auth.Validate(authToken)
	if err != nil {
		return nil, errors.Wrap(err, "authenticating token")
	}

	roles := k.rolesAsMap()
	var isAdmin bool
	for _, val := range keystoneContext.Roles {
		if _, ok := roles[val.Name]; ok {
			isAdmin = true
			break
		}
	}
	authDetails := AuthDetails{
		UserID:    keystoneContext.User.ID,
		IsAdmin:   isAdmin,
		ExpiresAt: keystoneContext.ExpiresAt,
	}

	ctx := req.Context()

	return context.WithValue(ctx, AuthDetailsKey, authDetails), nil
}
