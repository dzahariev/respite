package repo

import (
	"fmt"
	"reflect"

	"github.com/dzahariev/respite/basemodel"
)

// Resource represent a resource entity in the system.
type Resource struct {
	Name     string
	IsGlobal bool
	Type     reflect.Type
}

// ResourceFactory is used to hold information about supported resources
type ResourceFactory struct {
	Resources map[string]Resource
}

// Register is used to register a resource type
func (resourceFactory *ResourceFactory) Register(object basemodel.Object) {
	name := object.ResourceName()
	isGlobal := object.IsGlobal()
	objectType := reflect.TypeOf(object).Elem()
	resourceFactory.Resources[name] = Resource{
		Name:     name,
		IsGlobal: isGlobal,
		Type:     objectType,
	}
}

// Names returns the names of all registered resources
func (resourceFactory *ResourceFactory) Names() []string {
	names := make([]string, 0, len(resourceFactory.Resources))
	for name := range resourceFactory.Resources {
		names = append(names, name)
	}
	return names
}

// New is used to create a new resource object
func (resourceFactory *ResourceFactory) New(name string) (basemodel.Object, error) {
	t, ok := resourceFactory.Resources[name]
	if !ok {
		return nil, fmt.Errorf("unrecognized resource name: %s", name)
	}

	obj, ok := reflect.New(t.Type).Interface().(basemodel.Object)
	if !ok {
		return nil, fmt.Errorf("type %s does not implement model.Object", t.Type)
	}
	return obj, nil
}
