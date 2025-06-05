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

// Resources is used to hold information about supported resources
type Resources struct {
	Resources map[string]Resource
}

// Register is used to register a resource type
func (resources *Resources) Register(object basemodel.Object) {
	name := object.ResourceName()
	isGlobal := object.IsGlobal()
	objectType := reflect.TypeOf(object).Elem()
	resources.Resources[name] = Resource{
		Name:     name,
		IsGlobal: isGlobal,
		Type:     objectType,
	}
}

// Names returns the names of all registered resources
func (resources *Resources) Names() []string {
	names := make([]string, 0, len(resources.Resources))
	for name := range resources.Resources {
		names = append(names, name)
	}
	return names
}

// New is used to create a new resource object
func (resources *Resources) New(name string) (basemodel.Object, error) {
	t, ok := resources.Resources[name]
	if !ok {
		return nil, fmt.Errorf("unrecognized resource name: %s", name)
	}

	obj, ok := reflect.New(t.Type).Interface().(basemodel.Object)
	if !ok {
		return nil, fmt.Errorf("type %s does not implement model.Object", t.Type)
	}
	return obj, nil
}

// IsGlobal is used to check if a resource is global
func (resources *Resources) IsGlobal(name string) bool {
	resource, ok := resources.Resources[name]
	if !ok {
		return false
	}
	return resource.IsGlobal
}
