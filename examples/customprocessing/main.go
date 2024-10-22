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
