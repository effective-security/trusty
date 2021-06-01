package appcontainer

import (
	"fmt"
	"reflect"

	"github.com/juju/errors"
)

type serviceInfo struct {
	ServerName string
	Service    interface{}
	Type       reflect.Type
}

// Discovery provides service discovery interface
type Discovery interface {
	Register(server string, service interface{}) error
	Find(v interface{}) error
	ForEach(v interface{}, f func(typ string) error) error
}

type disco struct {
	reg map[string]serviceInfo
}

// NewDiscovery return new Discovery
func NewDiscovery() Discovery {
	return &disco{
		reg: make(map[string]serviceInfo),
	}
}

// Register interface
func (d *disco) Register(server string, service interface{}) error {
	typ := reflect.TypeOf(service)

	logger.Infof("src=Register, server=%s, type=%v", server, typ)
	key := fmt.Sprintf("%s/%s", server, typ.String())

	if _, ok := d.reg[key]; ok {
		return errors.Errorf("already registered: %s", key)
	}

	d.reg[key] = serviceInfo{
		ServerName: server,
		Service:    service,
		Type:       typ,
	}

	return nil
}

// Find interface
func (d *disco) Find(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.Errorf("a pointer to interface is required, invalid type: %v", rv)
	}

	logger.Debugf("src=Find, type=%v", rv.String())

	rv = rv.Elem()
	if !rv.IsValid() || rv.Kind() != reflect.Interface {
		return errors.Errorf("non interface type: %s", reflect.TypeOf(v))
	}

	for _, reg := range d.reg {
		if reg.Type.Implements(rv.Type()) {
			rv.Set(reflect.ValueOf(reg.Service))
			return nil
		}
	}

	return errors.Errorf("not implemented: " + rv.String())
}

// ForEach interface
func (d *disco) ForEach(v interface{}, f func(typ string) error) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.Errorf("a pointer to interface is required, invalid type: %v", rv)
	}

	rv = rv.Elem()
	if !rv.IsValid() || rv.Kind() != reflect.Interface {
		return errors.Errorf("non interface type: %s", reflect.TypeOf(v))
	}

	for key, reg := range d.reg {
		if reg.Type.Implements(rv.Type()) {
			rv.Set(reflect.ValueOf(reg.Service))
			err := f(key)
			if err != nil {
				return errors.Annotatef(err, "failed to execute callback for %s", reg.Type.String())
			}
		}
	}
	return nil
}
