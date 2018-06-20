package main

import (
	"fmt"

	"github.com/samoslab/nebula/client/common"

	"github.com/spf13/pflag"
)

func main() {
	input := pflag.StringP("input", "i", "", "input file")
	output := pflag.StringP("output", "o", "", "output file")
	method := pflag.StringP("method", "m", "", "method")
	password := pflag.StringP("password", "p", "", "password")
	pflag.Parse()

	if *password == "" {
		fmt.Printf("need password -p\n")
		pflag.PrintDefaults()
		return
	}
	if len(*password) != 16 {
		fmt.Printf("password length need 16\n")
		pflag.PrintDefaults()
		return
	}
	if *input == "" {
		fmt.Printf("need input -i\n")
		pflag.PrintDefaults()
		return
	}
	if *output == "" {
		fmt.Printf("need output -o\n")
		pflag.PrintDefaults()
		return
	}
	if *method == "" {
		fmt.Printf("need method -m\n")
		pflag.PrintDefaults()
		return
	}

	var err error

	switch *method {
	case "enc":
		err = common.EncryptFile(*input, []byte(*password), *output)
	case "dec":
		err = common.DecryptFile(*input, []byte(*password), *output)
	default:
		err = fmt.Errorf("only support enc|dec\n")
	}

	if err != nil {
		fmt.Printf("error %v\n", err)
		return
	}

	fmt.Printf("success\n")

}
