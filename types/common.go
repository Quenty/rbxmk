package types

import (
	"fmt"
	"github.com/robloxapi/rbxapi"
	"github.com/robloxapi/rbxfile"
)

func isAlnum(b byte) bool {
	return ('0' <= b && b <= '9') ||
		('A' <= b && b <= 'Z') ||
		('a' <= b && b <= 'z') ||
		(b == '_')
}

func isDigit(b byte) bool {
	return ('0' <= b && b <= '9')
}

type ParseError struct {
	Index int
	Err   error
}

func (err ParseError) Error() string {
	return fmt.Sprintf("@%d: %s", err.Index, err.Err)
}

func typeOfProperty(api *rbxapi.API, className, propName string) rbxfile.Type {
	if api == nil {
		return rbxfile.TypeInvalid
	}
	class := api.Classes[className]
	if class == nil {
		return rbxfile.TypeInvalid
	}
	prop, ok := class.Members[propName].(*rbxapi.Property)
	if !ok {
		return rbxfile.TypeInvalid
	}
	return rbxfile.TypeFromAPIString(api, prop.ValueType)
}

func propertyIsOfType(api *rbxapi.API, inst *rbxfile.Instance, propName string, typ rbxfile.Type) bool {
	if api == nil {
		v, ok := inst.Properties[propName]
		if !ok {
			// Type cannot be determined, assume given type is correct.
			return true
		}
		return v.Type() == typ
	}
	class := api.Classes[inst.ClassName]
	if class == nil {
		// Unknown class, assume given type is correct.
		return true
	}
	member := class.Members[propName]
	prop, ok := member.(*rbxapi.Property)
	if !ok {
		if member != nil {
			// Incorrect member type.
			return false
		}
		// Unknown property, assume given type is correct.
		return true
	}
	return rbxfile.TypeFromAPIString(api, prop.ValueType) == typ
}

type RegionError string

func (err RegionError) Error() string {
	return fmt.Sprintf("failed to find region \"%s\"", string(err))
}
