package resolver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	firstPersonName  = "first person name"
	secondPersonName = "second person name"
)

//TODO add tests with slices
func TestResolver_Resolve_SimpleValues(t *testing.T) {
	tests := []struct {
		name         string
		expression   string
		resolverArgs map[string]interface{}
		wantedValue  interface{}
		wantedError  error
	}{
		{
			name:       "Test resolving that accesses the first element in the map",
			expression: "user",
			resolverArgs: map[string]interface{}{
				"user": "value",
			},
			wantedValue: "value",
			wantedError: nil,
		},
		{
			name:       "Test resolving that accesses the first element in the map that is array",
			expression: "user[0]",
			resolverArgs: map[string]interface{}{
				"user": []int{1, 2},
			},
			wantedValue: 1,
			wantedError: nil,
		},
		{
			name:       "Test resolving that accesses the value of a field that is nested one level in a struct",
			expression: "user.Name",
			resolverArgs: map[string]interface{}{
				"user": struct {
					Name string
				}{
					Name: "userName",
				},
			},
			wantedValue: "userName",
			wantedError: nil,
		},
		{
			name:       "Test resolving that accesses the value of a field that is nested two levels in a struct",
			expression: "user.Info.Age",
			resolverArgs: map[string]interface{}{
				"user": struct {
					Info struct {
						Age int
					}
				}{
					Info: struct {
						Age int
					}{
						Age: 12,
					},
				},
			},
			wantedValue: 12,
			wantedError: nil,
		},
		{
			name:       "Test resolving that accesses the value of a field that is an array element nested two levels",
			expression: "user.Info.Age[2]",
			resolverArgs: map[string]interface{}{
				"user": struct {
					Info struct {
						Age []int
					}
				}{
					Info: struct {
						Age []int
					}{
						Age: []int{1, 2, 3},
					},
				},
			},
			wantedValue: 3,
			wantedError: nil,
		},
		{
			name:       "Test resolving that accesses the value of a field that belongs to an array element nested three levels",
			expression: "user.Info.Parents[0].Name",
			resolverArgs: map[string]interface{}{
				"user": struct {
					Info struct {
						Parents []struct {
							Name string
						}
					}
				}{
					Info: struct {
						Parents []struct {
							Name string
						}
					}{
						Parents: []struct{ Name string }{
							{
								Name: "someone1",
							},
							{
								Name: "someone2",
							},
						},
					},
				},
			},
			wantedValue: "someone1",
			wantedError: nil,
		},
		{
			name:       "Test resolving that accesses the value of a key that belongs to a map nested one level",
			expression: "user.name",
			resolverArgs: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "somename",
				},
			},
			wantedValue: "somename",
			wantedError: nil,
		},
		{
			name:       "Test resolving that accesses the value of a key that belongs to a map nested two levels",
			expression: "user.name.value",
			resolverArgs: map[string]interface{}{
				"user": map[string]interface{}{
					"name": map[string]interface{}{
						"value": "somename",
					},
				},
			},
			wantedValue: "somename",
			wantedError: nil,
		},
		{
			name:       "Test resolving that accesses the value of an array element that belongs to a map nested two levels",
			expression: "user.name.value[0]",
			resolverArgs: map[string]interface{}{
				"user": map[string]interface{}{
					"name": map[string]interface{}{
						"value": []int{1, 2, 3},
					},
				},
			},
			wantedValue: 1,
			wantedError: nil,
		},
		{
			name:       "Test resolving that accesses the value of an array element that belongs to a map nested two levels",
			expression: "user.name.value[0].Name",
			resolverArgs: map[string]interface{}{
				"user": map[string]interface{}{
					"name": map[string]interface{}{
						"value": []struct {
							Name string
						}{
							{Name: "string"},
						},
					},
				},
			},
			wantedValue: "string",
			wantedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			r := NewResolver(tc.resolverArgs)
			v, err := r.Resolve(tc.expression)
			assert.Equal(t, tc.wantedValue, v)
			assert.Equal(t, tc.wantedError, err)

			v, err = DirectResolve(tc.resolverArgs, tc.expression)
			assert.Equal(t, tc.wantedValue, v)
			assert.Equal(t, tc.wantedError, err)
		})
	}
}

func TestResolver_Resolve_Errors(t *testing.T) {
	tests := []struct {
		name         string
		expression   string
		resolverArgs map[string]interface{}
		wantedValue  interface{}
		wantedError  error
	}{
		{
			name:         "Test resolving that accesses the first element in the map not found",
			expression:   "user",
			resolverArgs: map[string]interface{}{},
			wantedValue:  nil,
			wantedError:  errors.New("could not resolve user: user was not found"),
		},
		{
			name:         "Test resolving with empty expression",
			expression:   "",
			resolverArgs: map[string]interface{}{},
			wantedValue:  nil,
			wantedError:  errors.New("could not resolve: expression is empty"),
		},
		{
			name:       "Test resolving with invalid array index",
			expression: "user[asd]",
			resolverArgs: map[string]interface{}{
				"user": []int{1, 2, 3},
			},
			wantedValue: nil,
			wantedError: errors.New("could not parse index of array"),
		},
		{
			name:       "Test resolving with array index out of range",
			expression: "user[100]",
			resolverArgs: map[string]interface{}{
				"user": []int{1, 2, 3},
			},
			wantedValue: nil,
			wantedError: errors.New("index is out of range"),
		},
		{
			name:       "Test resolving that accesses non existing element in a map",
			expression: "user.nonexisting.value",
			resolverArgs: map[string]interface{}{
				"user": map[string]interface{}{
					"name": map[string]interface{}{
						"value": "somename",
					},
				},
			},
			wantedValue: nil,
			wantedError: errors.New("map[string]interface {} does not have field nonexisting: value left to resolve"),
		},
		{
			name:       "Test resolving that accesses a field of a primitive type",
			expression: "user.username",
			resolverArgs: map[string]interface{}{
				"user": "somevalue",
			},
			wantedValue: nil,
			wantedError: errors.New("string does not have field username: nothing left to resolve"),
		},
		{
			name:       "Test resolving that accesses a field of a primitive type that is nested in map",
			expression: "user.username.somevalue2",
			resolverArgs: map[string]interface{}{
				"user": "somevalue",
			},
			wantedValue: nil,
			wantedError: errors.New("string does not have field username: somevalue2 left to resolve"),
		},
		{
			name:       "Test resolving that accesses a field of a primitive type that is nested in a struct",
			expression: "user.Name.somevalue2",
			resolverArgs: map[string]interface{}{
				"user": struct {
					Name string
				}{
					Name: "somename",
				},
			},
			wantedValue: nil,
			wantedError: errors.New("string does not have field somevalue2: nothing left to resolve"),
		},
		{
			name:       "Test resolving that accesses a field of a struct that is not defined",
			expression: "user.firstname",
			resolverArgs: map[string]interface{}{
				"user": struct {
					Name string
				}{
					Name: "somename",
				},
			},
			wantedValue: nil,
			//TODO find if this way of outputting the error is good, because a struct can have many fields
			wantedError: errors.New("struct { Name string } does not have field firstname: nothing left to resolve"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewResolver(tc.resolverArgs)
			v, err := r.Resolve(tc.expression)

			assert.Equal(t, tc.wantedValue, v)
			assert.Equal(t, tc.wantedError, err)
		})
	}
}

func TestResolver_Resolve_JSON(t *testing.T) {
	tests := []struct {
		name             string
		expression       string
		expressionPrefix string
		jsonAsStr        string
		wantedValue      interface{}
		wantedError      error
	}{
		{
			name:             "Test resolving that accesses two array elements",
			expression:       "resources[0].document_paths[1].type",
			expressionPrefix: "invoice",
			jsonAsStr:        "{\"meta\":{\"has_more\":false,\"count\":1,\"limit\":50,\"page\":1,\"pages\":1},\"resources\":[{\"status\":\"completed\",\"document_path\":\"\\/invoice_doc_url\\/fee18c3f-eecc-42f7-a1d3-a42fbb34462f\",\"service\":\"transportation\",\"invoice_date\":\"Thu, 02 Mar 2017 15:51:37 UTC\",\"tax_point_date\":\"Tue, 25 Oct 2016 18:13:08 UTC\",\"invoice_uuid\":\"fee18c3f-eecc-42f7-a1d3-a42fbb34462f\",\"invoice_number\":\"4AE7723E-41DF-4CCF-86F5-250471BD1950\",\"trip_uid\":\"95f6b0e4-8f0c-4432-8021-a5acf082a7a9\",\"document_paths\":[{\"path\":\"\\/invoice_doc_url\\/fee18c3f-eecc-42f7-a1d3-a42fbb34462f\",\"type\":\"PDF\"},{\"path\":\"\\/invoice_doc_url\\/fee18c3f-eecc-42f7-a1d3-a42fbb34462f?document_type=XML\",\"type\":\"XML\"}]}]}",
			wantedValue:      "XML",
			wantedError:      nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := ioutil.ReadAll(bytes.NewBufferString(tc.jsonAsStr))
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("%s: Expected nil error but got: %s", tc.name, err.Error()))
			}

			var jsonMap map[string]interface{}
			if err := json.Unmarshal(data, &jsonMap); err != nil {
				assert.FailNow(t, fmt.Sprintf("%s: Expected nil error but got: %s", tc.name, err.Error()))
			}

			r := NewResolver(make(map[string]interface{}))
			r.AddValue(tc.expressionPrefix, jsonMap)

			v, err := r.Resolve(fmt.Sprintf("%s.%s", tc.expressionPrefix, tc.expression))
			assert.Equal(t, tc.wantedValue, v)
			assert.Equal(t, tc.wantedError, err)

			v, err = DirectResolve(jsonMap, tc.expression)
			assert.Equal(t, tc.wantedValue, v)
			assert.Equal(t, tc.wantedError, err)
		})
	}
}

func TestResolver_Resolve_ValuesWithPointers(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		wantedValue interface{}
		wantedError error
	}{
		{
			name:        "Test resolving that accesses a field that is a pointer",
			expression:  "user.Name",
			wantedValue: firstPersonName,
			wantedError: nil,
		},
		{
			name:        "Test resolving that accesses two level nested fields both pointers",
			expression:  "user.Parent.Name",
			wantedValue: secondPersonName,
			wantedError: nil,
		},
		{
			name:        "Test resolving that accesses a field of array of pointers",
			expression:  "user.Parents[0].Name",
			wantedValue: secondPersonName,
			wantedError: nil,
		},
		{
			name:        "Test resolving that accesses a field that is a pointer to struct of array of pointers",
			expression:  "user.Parents[0].Parent.Name",
			wantedValue: firstPersonName,
			wantedError: nil,
		},
		{
			name:        "Test resolving that accesses a field that is a pointer to struct of array of pointers",
			expression:  "user.Parents[0].Parents[0].Parent.Name",
			wantedValue: firstPersonName,
			wantedError: nil,
		},
		{
			name:        "Test resolving that accesses a field that is a pointer to struct of array of pointers",
			expression:  "user.Parents[0].Parents[0].Name",
			wantedValue: secondPersonName,
			wantedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fPersonName := firstPersonName
			sPersonName := secondPersonName

			user := Person{
				Name: &fPersonName,
				Parent: &Person{
					Name: &sPersonName,
				},
				Parents: []*Person{
					{
						Name: &sPersonName,
						Parent: &Person{
							Name: &fPersonName,
						},
						Parents: []*Person{
							{
								Name: &sPersonName,
								Parent: &Person{
									Name: &fPersonName,
								},
							},
						},
					},
				},
			}

			resolverArgs := map[string]interface{}{
				"user": user,
			}

			r := NewResolver(resolverArgs)
			v, err := r.Resolve(tc.expression)
			strVal := v.(*string)
			assert.Equal(t, tc.wantedValue, *strVal)
			assert.Equal(t, tc.wantedError, err)

			v, err = DirectResolve(user, tc.expression)
			strVal = v.(*string)
			assert.Equal(t, tc.wantedValue, *strVal)
			assert.Equal(t, tc.wantedError, err)
		})
	}
}

type Person struct {
	Name    *string
	Parent  *Person
	Parents []*Person
}
