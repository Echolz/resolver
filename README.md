# Resolver
[![Build Status](https://travis-ci.org/Echolz/resolver.svg?branch=master)](https://travis-ci.org/Echolz/resolver)
[![Go Report Card](https://goreportcard.com/badge/github.com/Echolz/resolver)](https://goreportcard.com/report/github.com/Echolz/resolver)

## Overview

Resolver - Go package for extracting values of a `maps/structs/arrays` by a given expression.

## Pattern

### Usage1

Use `NewResolver` to create a new resolver with some values in the map, that you want to resolve

    res := resolver.NewResolver(args map[string]interface)

Use `Resolve` to access the value you're interested in. You can use dot and array
notation too:
    
    v, err := res.Resolve("something.Field1.Field2[2].Field3")
    
The resolver has a state, that you can change with the `AddValue`

    if err != nil { res.AddValue("resolvedValue", v) }

Now you can resolve the value that you already resolved.

    v, err := res.Resolve("resolvedValue.field1.field2[2].field3")

### Usage2

Use `resolver.DirectResolve(v interface{}, s string) (interface{}, error)` to access the value you're interested, without creating or using a Resolver struct.
    
        type Person struct {
        //it does not matter if the given element is a pointer, resolver always uses the value that the pointer points to
                Name    *string 
        }
        
        name := "somename"
        p := Person{
            Name: &name,
        }
                        
        v, err := resolver.DirectResolve(p, "person.Name")
        fmt.Println(v) // Output: somename
        vStr := v.(string)
        fmt.Println(len(vStr)) // Output: 8
    
Here we have a exaple with resolving an array:

		arr := []int{1, 2, 3}
		v, err := resolver.DirectResolve(arr, "arr[1]")
		fmt.Println(v, err) // Output: 2 <nil>

It works with slices:
        
		arr := make([]int, 0, 2)
		arr = append(arr, 1, 2)
		v, err := DirectResolve(arr, "arr[1]")
		fmt.Println(v, err) // Output: 2 <nil>
		
It works with maps:

		m := make(map[string]interface{})
		m["mapvalue"] = map[string]interface{}{"somevalue": "newvalue"}
		v, err := DirectResolve(m, "mapvalue.somevalue")
		fmt.Println(v, err) // Output: newvalue <nil>

## Installation
To install Resolver, use go get:

    go get github.com/Echolz/resolver

### Staying up to date
To update Resolver to the latest version, run:

    go get -u github.com/Echolz/resolver

## Contributing
Please feel free to submit issues, fork the repository and send pull requests!