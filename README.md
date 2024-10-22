# About
[![Go Reference](https://godoc.org/github.com/xaionaro-go/deepcopy?status.svg)](https://godoc.org/github.com/xaionaro-go/deepcopy)
[![Go Report Card](https://goreportcard.com/badge/github.com/xaionaro-go/deepcopy?branch=main)](https://goreportcard.com/report/github.com/xaionaro-go/deepcopy)

This package provides functions for deep copying an arbitrary object in Go. The difference of this deepcopier from others is that this allows to use a custom function that modifies the copied data. Personally I use it to erase all the secrets from my data while doing a copy (that in turn is used for logging).

# How to use

### JUST DEEP COPY
```go
package main

import (
	"fmt"

	"github.com/xaionaro-go/deepcopy"
)

type myStruct struct {
	PublicData string
	SecretData string `secret:""`
}

func main() {
	value := myStruct{
		PublicData: "true == true",
		SecretData: "but there is a nuance",
	}

	censoredValue := deepcopy.DeepCopy(value)
	fmt.Println(censoredValue)
}

```
```sh
$ go run ./examples/deepcopy/
{true == true but there is a nuance}
```


### REMOVE MY SECRETS
```go
package main

import (
	"fmt"

	"github.com/xaionaro-go/deepcopy"
)

type myStruct struct {
	PublicData string
	SecretData string `secret:""`
}

func main() {
	value := myStruct{
		PublicData: "true == true",
		SecretData: "but there is a nuance",
	}

	censoredValue := deepcopy.DeepCopyWithoutSecrets(value)
	fmt.Println(censoredValue)
}
```
```sh
$ go run ./examples/censoredvalue/
{true == true }
```

### CUSTOM PROCESSING
```go
package main

import (
	"fmt"
	"reflect"

	"github.com/xaionaro-go/deepcopy"
)

type myStruct struct {
	PublicData string
	SecretData string `secret:""`
}

func main() {
	value := myStruct{
		PublicData: "true == true",
		SecretData: "but there is a nuance",
	}

	censoredValue := deepcopy.DeepCopyWithProcessing(value, func(v reflect.Value, sf *reflect.StructField) reflect.Value {
		if sf == nil {
			return v
		}
		switch sf.Name {
		case "PublicData":
			return reflect.ValueOf("true == false")
		case "SecretData":
			return reflect.ValueOf("this is the nuance, sometimes")
		}
		return v
	})
	fmt.Println(censoredValue)
}
```
```sh
$ go run ./examples/customprocessing/
{true == false this is the nuance, sometimes}
```