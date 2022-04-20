// Copyright 2016-2022, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package pulumiapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type Webhook struct {
	Active      bool
	DisplayName string
	PayloadUrl  string
	Secret      string
	Name        string
}

func (c *Client) CreateWebhook(orgName, displayName, payLoadUrl, secret string, active bool) (*Webhook, error) {

	if len(orgName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}

	if len(displayName) == 0 {
		return nil, errors.New("displayname must not be empty")
	}

	if len(payLoadUrl) == 0 {
		return nil, errors.New("payloadurl must not be empty")
	}

	path := fmt.Sprintf("orgs/%s/hooks", orgName)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	values := map[string]interface{}{"organizationName": orgName, "displayName": displayName, "payloadUrl": payLoadUrl, "secret": secret, "active": active}
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpt.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200, 201:
		var webhook Webhook
		err := json.NewDecoder(res.Body).Decode(&webhook)
		if err != nil {
			return nil, err
		}

		return &webhook, nil
	case 400, 401, 403, 404, 500:
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			panic(err)
		}

		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return nil, &errRes
	default:
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

}

func (c *Client) ListWebhooks(orgName string) (*[]Webhook, error) {
	if len(orgName) == 0 {
		return nil, errors.New("orgName must not be empty")
	}

	path := fmt.Sprintf("orgs/%s/hooks", orgName)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	req, err := http.NewRequest("GET", endpt.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		var webhooks []Webhook
		err = json.NewDecoder(res.Body).Decode(&webhooks)
		if err != nil {
			return nil, err
		}

		return &webhooks, nil

	case 400, 401, 403, 404, 500:
		var errRes ErrorResponse
		err := json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			panic(err)
		}

		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return nil, &errRes
	default:
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
}

func (c *Client) GetWebhook(orgName, webhookName string) (*Webhook, error) {
	if len(orgName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}

	if len(webhookName) == 0 {
		return nil, errors.New("webhookname must not be empty")
	}

	path := fmt.Sprintf("orgs/%s/hooks/%s", orgName, webhookName)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	req, err := http.NewRequest("GET", endpt.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		var webhook Webhook
		err := json.NewDecoder(res.Body).Decode(&webhook)
		if err != nil {
			return nil, err
		}
		return &webhook, nil
	case 400, 401, 403, 404, 500:
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			panic(err)
		}

		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return nil, &errRes

	default:
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
}

func (c *Client) UpdateWebhook(name, orgName, displayName, payloadUrl, secret string, active bool) (*Webhook, error) {
	if len(name) == 0 {
		return nil, errors.New("name must not be empty")
	}
	if len(orgName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}
	if len(displayName) == 0 {
		return nil, errors.New("displayname must not be empty")
	}
	if len(payloadUrl) == 0 {
		return nil, errors.New("payloadurl must not be empty")
	}

	path := fmt.Sprintf("orgs/%s/hooks/%s", orgName, name)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	values := map[string]interface{}{
		"organizationName": orgName,
		"name":             name,
		"displayName":      displayName,
		"payloadUrl":       payloadUrl,
		"secret":           secret,
		"active":           active,
	}
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", endpt.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)

	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 204:
		var webhook Webhook
		webhook.DisplayName = displayName
		webhook.Secret = secret
		webhook.Active = active
		webhook.Name = name
		webhook.PayloadUrl = payloadUrl
		return &webhook, nil
	case 400, 401, 403, 404, 405, 500:
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			panic(err)
		}

		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return nil, &errRes
	default:
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

}

func (c *Client) DeleteWebhook(orgName, name string) error {
	if len(name) == 0 {
		return errors.New("name must not be empty")
	}
	if len(orgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	path := fmt.Sprintf("orgs/%s/hooks/%s", orgName, name)
	endpt := c.baseurl.ResolveReference(&url.URL{Path: path})

	req, err := http.NewRequest("DELETE", endpt.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)

	res, err := c.c.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 204:
		return nil
	case 400, 401, 403, 404, 405, 500:
		var errRes ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			panic(err)
		}

		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return &errRes
	default:
		return fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
}
