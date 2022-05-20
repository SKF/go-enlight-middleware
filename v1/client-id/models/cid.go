package models

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/SKF/go-utility/v2/uuid"
	"gopkg.in/yaml.v3"
)

type ClientIDs map[uuid.UUID]ClientID

type ClientID struct {
	Identifier   uuid.UUID              `yaml:"-,omitempty"`
	Name         string                 `yaml:""`
	Description  string                 `yaml:",omitempty"`
	Owner        string                 `yaml:""`
	Environments Environments           `yaml:",flow,omitempty"`
	NotBefore    time.Time              `yaml:"notBefore,omitempty"`
	Expires      time.Time              `yaml:",omitempty"`
	Properties   map[string]interface{} `yaml:",omitempty"`
}

func (cid *ClientID) ExtractProperty(key string, value interface{}) error {
	to := reflect.ValueOf(value)
	if to.Kind() == reflect.Ptr && !to.IsNil() {
		to = to.Elem()
	}

	if !to.CanSet() {
		return fmt.Errorf("value must be an addressable value, is it a pointer?")
	}

	property := cid.Properties[key]

	r, w := io.Pipe()
	defer r.Close()

	go func() {
		if err := yaml.NewEncoder(w).Encode(property); err != nil {
			panic(err)
		}

		w.Close()
	}()

	if err := yaml.NewDecoder(r).Decode(value); err != nil {
		return err
	}

	return nil
}

func (cid *ClientID) IsEmpty() bool {
	return cid == nil || cid.Identifier == ""
}

type contextKey struct{}

func (cid *ClientID) EmbedIntoContext(parent context.Context) context.Context {
	return context.WithValue(parent, contextKey{}, cid)
}

func FromContext(ctx context.Context) (*ClientID, bool) {
	cid, ok := ctx.Value(contextKey{}).(*ClientID)
	return cid, ok
}
