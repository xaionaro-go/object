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
