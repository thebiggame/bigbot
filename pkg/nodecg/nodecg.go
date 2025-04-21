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
	"net/http"
	"reflect"
	"time"
)

const (
	nodecgMessagePrefix   = "/message"
	nodecgReplicantPrefix = "/replicant"
	nodecgRestPrefix      = "/rest"

	StatusSuccess = "OK"
	StatusError   = "ERROR"
)

var (
	ErrNodeCGInternalError = errors.New("internal NodeCG error")
	ErrNodeCGUnknownError  = errors.New("unknown NodeCG error")
	ErrNodeCGGeneralError  = errors.New("NodeCG error")

	ErrNotBool = errors.New("non-boolean returned")
)

type NodeCGServer struct {
	Hostname string
	Key      string
}

func New(host string) *NodeCGServer {
	return &NodeCGServer{
		Hostname: host,
	}
}

func (s *NodeCGServer) WithKey(key string) *NodeCGServer {
	s.Key = key
	return s
}

type ReplicantResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`

	Name   string          `json:"name"`
	Bundle string          `json:"bundle"`
	Value  json.RawMessage `json:"value"`
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
func (s *NodeCGServer) ReplicantGetBool(ctx context.Context, bundle, replicant string) (result bool, err error) {
	rep, err := s.ReplicantGet(ctx, bundle, replicant)
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
func (s *NodeCGServer) ReplicantGet(ctx context.Context, bundle, replicant string) (result interface{}, err error) {
	var resp interface{}
	err = s.ReplicantGetDecode(ctx, bundle, replicant, &resp)
	return resp, err
}

// ReplicantGetDecode fetches the current state of a given Replicant, then decodes it to your provided pointer.
func (s *NodeCGServer) ReplicantGetDecode(ctx context.Context, bundle, replicant string, target any) (err error) {
	tv := reflect.ValueOf(target)
	if tv.Kind() != reflect.Pointer || tv.IsNil() {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(target)}
	}
	// Build URL.
	url := s.Hostname + nodecgRestPrefix + nodecgReplicantPrefix + "/" + bundle + "/" + replicant

	body := &requestReplicant{
		requestAuth: requestAuth{
			Key: s.Key,
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	tCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(tCtx, "POST", url, bytes.NewBuffer(data))
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

	var resp ReplicantResponse
	err = json.NewDecoder(respData.Body).Decode(&resp)
	if err != nil {
		return err
	}
	if resp.Status != StatusSuccess {
		return fmt.Errorf("%w: %s", ErrNodeCGGeneralError, resp.Message)
	}

	if resp.Value != nil {
		err = json.Unmarshal(resp.Value, target)
	}

	return err
}

// ReplicantSet sets the current state of a remote Replicant.
// value MUST be serialisable as JSON in some fashion.
func (s *NodeCGServer) ReplicantSet(ctx context.Context, bundle string, replicant string, value interface{}) (err error) {
	// Build URL.
	url := s.Hostname + nodecgRestPrefix + nodecgReplicantPrefix + "/" + bundle + "/" + replicant

	body := &requestReplicant{
		requestAuth: requestAuth{
			Key: s.Key,
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
	var resp ReplicantResponse
	err = json.NewDecoder(respData.Body).Decode(&resp)
	if err != nil {
		return err
	}
	if resp.Status != StatusSuccess {
		return fmt.Errorf("%w: %s", ErrNodeCGGeneralError, resp.Message)
	}
	return nil
}

// MessageSend sends a NodeCG message.
// value is optional, but MUST be serialisable as JSON in some fashion.
func (s *NodeCGServer) MessageSend(ctx context.Context, bundle, messageChannel string, value interface{}) (err error) {
	// Build URL.
	url := s.Hostname + nodecgRestPrefix + nodecgMessagePrefix + "/" + bundle + "/" + messageChannel

	body := &requestReplicant{
		requestAuth: requestAuth{
			Key: s.Key,
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
	var resp ReplicantResponse
	err = json.NewDecoder(respData.Body).Decode(&resp)
	if err != nil {
		return err
	}
	if resp.Status != StatusSuccess {
		return fmt.Errorf("%w: %s", ErrNodeCGGeneralError, resp.Message)
	}
	return nil
}
