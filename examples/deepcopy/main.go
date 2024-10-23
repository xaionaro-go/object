package main

import (
	"fmt"

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

	censoredValue := object.DeepCopy(value)
	fmt.Println(censoredValue)
}
