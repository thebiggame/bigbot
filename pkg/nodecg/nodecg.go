// Package nodecg defines an API for communicating with a NodeCG graphics server.
// Right now it uses nodecg-rest (with authentication) - if you can refactor this to use
// WebSockets or anything less hacky, please do so.
package nodecg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thebiggame/bigbot/internal/config"
	"net/http"
	"reflect"
	"time"
)

const (
	nodecgMessagePrefix   = "/message"
	nodecgReplicantPrefix = "/replicant"
	nodecgRestPrefix      = "/rest"

	nodecgStatusSuccess = "OK"
	nodecgStatusError   = "ERROR"
)

var (
	ErrNodeCGInternalError = errors.New("internal NodeCG error")
	ErrNodeCGUnknownError  = errors.New("unknown NodeCG error")
	ErrNodeCGGeneralError  = errors.New("NodeCG error")

	ErrNotBool = errors.New("non-boolean returned")
)

type replicantResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`

	Name   string      `json:"name"`
	Bundle string      `json:"bundle"`
	Value  interface{} `json:"value"`
}

type requestAuth struct {
	Key string `json:"key,omitempty"`
}

type requestReplicant struct {
	requestAuth
	Data interface{} `json:"data,omitempty"`
}

// ReplicantGetBool is a shortcut to ReplicantGet for retrieving the current state of a Replicant,
// where the content is a boolean value.
func ReplicantGetBool(ctx context.Context, replicant string) (result bool, err error) {
	rep, err := ReplicantGet(ctx, replicant)
	if err != nil {
		return false, err
	}
	// test before returning (otherwise we panic)
	if reflect.TypeOf(rep).Kind() != reflect.Bool {
		return false, ErrNotBool
	} else {
		return reflect.ValueOf(rep).Bool(), nil
	}
}

// ReplicantGet fetches the current state of a given Replicant.
func ReplicantGet(ctx context.Context, replicant string) (result interface{}, err error) {
	// Build URL.
	url := config.RuntimeConfig.AV.NodeCG.Hostname + nodecgRestPrefix + nodecgReplicantPrefix + "/" + config.RuntimeConfig.AV.NodeCG.BundleName + "/" + replicant

	body := &requestReplicant{
		requestAuth: requestAuth{
			Key: config.RuntimeConfig.AV.NodeCG.AuthenticationKey,
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	tCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(tCtx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	respData, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer respData.Body.Close()

	if respData.StatusCode != http.StatusOK {
		if respData.StatusCode == http.StatusInternalServerError {
			return nil, ErrNodeCGInternalError
		} else {
			return nil, ErrNodeCGUnknownError
		}
	}

	var resp replicantResponse
	err = json.NewDecoder(respData.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	if resp.Status != nodecgStatusSuccess {
		return nil, fmt.Errorf("%w: %s", ErrNodeCGGeneralError, resp.Message)
	}

	return resp.Value, nil
}

// ReplicantSet sets the current state of a remote Replicant.
// value MUST be serialisable as JSON in some fashion.
func ReplicantSet(ctx context.Context, replicant string, value interface{}) (err error) {
	// Build URL.
	url := config.RuntimeConfig.AV.NodeCG.Hostname + nodecgRestPrefix + nodecgReplicantPrefix + "/" + config.RuntimeConfig.AV.NodeCG.BundleName + "/" + replicant

	body := &requestReplicant{
		requestAuth: requestAuth{
			Key: config.RuntimeConfig.AV.NodeCG.AuthenticationKey,
		},
		Data: value,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	tCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(tCtx, http.MethodPut, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	respData, err := HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer respData.Body.Close()

	if respData.StatusCode != http.StatusOK {
		if respData.StatusCode == http.StatusInternalServerError {
			return ErrNodeCGInternalError
		} else {
			return ErrNodeCGUnknownError
		}
	}
	var resp replicantResponse
	err = json.NewDecoder(respData.Body).Decode(&resp)
	if err != nil {
		return err
	}
	if resp.Status != nodecgStatusSuccess {
		return fmt.Errorf("%w: %s", ErrNodeCGGeneralError, resp.Message)
	}
	return nil
}

// MessageSend sends a NodeCG message.
// value is optional, but MUST be serialisable as JSON in some fashion.
func MessageSend(ctx context.Context, messageChannel string, value interface{}) (err error) {
	// Build URL.
	url := config.RuntimeConfig.AV.NodeCG.Hostname + nodecgRestPrefix + nodecgMessagePrefix + "/" + config.RuntimeConfig.AV.NodeCG.BundleName + "/" + messageChannel

	body := &requestReplicant{
		requestAuth: requestAuth{
			Key: config.RuntimeConfig.AV.NodeCG.AuthenticationKey,
		},
		Data: value,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	tCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(tCtx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	respData, err := HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer respData.Body.Close()

	if respData.StatusCode != http.StatusOK {
		if respData.StatusCode == http.StatusInternalServerError {
			return ErrNodeCGInternalError
		} else {
			return ErrNodeCGUnknownError
		}
	}
	var resp replicantResponse
	err = json.NewDecoder(respData.Body).Decode(&resp)
	if err != nil {
		return err
	}
	if resp.Status != nodecgStatusSuccess {
		return fmt.Errorf("%w: %s", ErrNodeCGGeneralError, resp.Message)
	}
	return nil
}
