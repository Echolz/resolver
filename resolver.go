package resolver

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

//Resolver is the interface, that some resolver with a state needs to satisfy
type Resolver interface {
	Resolve(string) (interface{}, error)
	AddValue(string, interface{})
}

type concreteResolver struct {
	m    sync.Mutex
	args map[string]interface{}
}

//NewResolver returns a concreteResolver, in the form of a Resolver interface
func NewResolver(args map[string]interface{}) Resolver {
	return &concreteResolver{args: args}
}

//Resolve is a method of the concreteResolver that returns a value by a given string
func (r *concreteResolver) Resolve(s string) (interface{}, error) {
	r.m.Lock()
	defer r.m.Unlock()

	s = formatExpression(s)
	expr := strings.SplitN(s, ".", 2)

	if len(expr) == 0 || expr[0] == "" {
		return nil, errors.New("could not resolve: expression is empty")
	}

	currentField := expr[0]

	v, found := r.args[currentField]

	if !found {
		return nil, fmt.Errorf("could not resolve %s: %s was not found", s, currentField)
	}

	if len(expr) == 1 {
		return v, nil
	}

	notResolvedFields := expr[1]
	reflV := reflect.ValueOf(v)
	return resolve(reflV, notResolvedFields)
}

//DirectResolve is a function that resolves directly a value by a given string and an interface, without the need of a Resolver
func DirectResolve(v interface{}, s string) (interface{}, error) {
	s = formatExpression(s)
	expr := strings.SplitN(s, ".", 2)

	if len(expr) == 0 || expr[0] == "" {
		return nil, errors.New("could not resolve: expression is empty")
	}

	value := reflect.ValueOf(v)

	if value.Kind() == reflect.Map {
		return resolve(value, s)
	}

	if len(expr) == 1 {
		return v, nil
	}

	notResolvedFields := expr[1]
	return resolve(value, notResolvedFields)
}

func resolve(value reflect.Value, s string) (interface{}, error) {
	if value.Kind() == reflect.Interface {
		value = reflect.ValueOf(value.Interface())
	}

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	expr := strings.SplitN(s, ".", 2)

	currentField := expr[0]

	var newValue reflect.Value

	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		index, err := strconv.Atoi(currentField)

		if err != nil {
			return nil, errors.New("could not parse index of array")
		}

		if index >= value.Len() || index < 0 {
			return nil, errors.New("index is out of range")
		}

		newValue = value.Index(index)
	}

	if value.Kind() == reflect.Struct {
		newValue = value.FieldByName(currentField)
	}

	if value.Kind() == reflect.Map {
		newValue = value.MapIndex(reflect.ValueOf(currentField))
	}

	if !newValue.IsValid() {
		if len(expr) == 1 {
			return nil, fmt.Errorf("%s does not have field %s: nothing left to resolve", value.Type(), currentField)
		}

		notResolvedFields := expr[1]

		return nil, fmt.Errorf("%s does not have field %s: %s left to resolve", value.Type(), currentField, notResolvedFields)
	}

	//this is the last string in the whole expression
	if len(expr) == 1 {
		return newValue.Interface(), nil
	}

	notResolvedFields := expr[1]

	return resolve(newValue, notResolvedFields)
}

func (r *concreteResolver) AddValue(s string, i interface{}) {
	r.m.Lock()
	defer r.m.Unlock()

	r.args[s] = i
}

func formatExpression(s string) string {
	s = strings.Replace(s, "[", ".", -1)
	return strings.Replace(s, "]", "", -1)
}
