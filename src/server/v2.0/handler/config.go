//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package handler

import (
	"context"
	"encoding/json"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/config"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/configure"
)

type configAPI struct {
	BaseAPI
	controller config.Controller
}

func newConfigAPI() *configAPI {
	return &configAPI{controller: config.Ctl}
}

func (c *configAPI) GetConfigurations(ctx context.Context, params configure.GetConfigurationsParams) middleware.Responder {
	if sec, exist := security.FromContext(ctx); exist {
		if sec.IsSolutionUser() {
			cfg, err := c.controller.AllConfigs(ctx)
			if err != nil {
				return c.SendError(ctx, err)
			}
			var result []byte
			result, err = json.Marshal(cfg)
			if err != nil {
				return c.SendError(ctx, err)
			}
			var cfgResp models.ConfigurationsResponse
			err = json.Unmarshal(result, &cfgResp)
			if err != nil {
				return c.SendError(ctx, err)
			}
			return configure.NewGetConfigurationsOK().WithPayload(&cfgResp)
		}
	}
	if err := c.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceConfiguration); err != nil {
		return c.SendError(ctx, err)
	}
	cfg, err := c.controller.UserConfigs(ctx)
	if err != nil {
		return c.SendError(ctx, err)
	}
	payload, err := toResponseModel(cfg)
	if err != nil {
		c.SendError(ctx, err)
	}
	return configure.NewGetConfigurationsOK().WithPayload(payload)
}

func (c *configAPI) PutConfigurations(ctx context.Context, params configure.PutConfigurationsParams) middleware.Responder {
	if err := c.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceConfiguration); err != nil {
		return c.SendError(ctx, err)
	}
	if params.Configurations == nil {
		return c.SendError(ctx, errors.BadRequestError(nil).WithMessage("Missing configure item"))
	}
	conf := params.Configurations
	err := c.controller.UpdateUserConfigs(ctx, conf)
	if err != nil {
		return c.SendError(ctx, err)
	}
	return configure.NewPutConfigurationsOK()
}

func (c *configAPI) GetInternalconfig(ctx context.Context, params configure.GetInternalconfigParams) middleware.Responder {
	if err := c.RequireSolutionUserAccess(ctx); err != nil {
		return c.SendError(ctx, err)
	}
	cfg, err := c.controller.AllConfigs(ctx)
	resultCfg, err := config.ConvertForGet(ctx, cfg, true)
	if err != nil {
		return c.SendError(ctx, err)
	}
	return configure.NewGetInternalconfigOK().WithPayload(resultCfg)
}

func toResponseModel(cfg map[string]*cfgModels.Value) (*models.ConfigurationsResponse, error) {
	var result []byte
	result, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	var cfgResp models.ConfigurationsResponse
	err = json.Unmarshal(result, &cfgResp)
	if err != nil {
		return nil, err
	}
	return &cfgResp, nil
}
