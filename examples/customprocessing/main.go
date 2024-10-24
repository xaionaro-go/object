package main

import (
	"fmt"
	"reflect"

	"github.com/xaionaro-go/object"
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

	censoredValue := object.DeepCopy(value, object.OptionWithVisitorFunc(func(_ *object.ProcContext, v reflect.Value, sf *reflect.StructField) (reflect.Value, bool, error) {
		if sf == nil {
			return v, true, nil
		}
		switch sf.Name {
		case "PublicData":
			return reflect.ValueOf("true == false"), true, nil
		case "SecretData":
			return reflect.ValueOf("this is the nuance, sometimes"), true, nil
		}
		return v, true, nil
	}))
	fmt.Println(censoredValue)
}
