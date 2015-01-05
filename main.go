package main

import (
	"fmt"
	"os"

	"github.com/tiago4orion/cloudinit-convert/cfnextract"
	"github.com/tiago4orion/cloudinit-convert/cfninject"
)

var BANNER string = `
         ,_---~~~~~----._         
  _,,_,*^____      _____` + "``" + `*g*\"*, 
 / __/ /'     ^.  /      \ ^@q   f 
[  @f | @))    |  | @))   l  0 _/  
 \` + "`" + `/   \~____ / __ \_____/    \   
  |           _l__l_           I   
  }          [______]           I  
  ]            | | |            |  
  ]             ~ ~             |  
  |                             | cloudinit-aws-tools v0.1  
  |                             | Author: Tiago Natel de Moura
---------------------------------------------------------------
`
var USAGE string = `Usage: %s <command> arg1, argN
For each command you can use the --help flag to obtain details about
the usage.
    %s <command> --help

Available commands:
  inject    - Inject a cloudinit file into a AWS CloudFormation at UserData
              section
  extract   - Extract the UserData section of the CloudFormation into a output file
  help      - Displays this help information.

`

func main() {
	var command string

	if len(os.Args) <= 1 {
		fmt.Println(BANNER)
		fmt.Printf(USAGE, os.Args[0], os.Args[0])
		os.Exit(1)
	}

	command = os.Args[1]

	switch command {
	case "inject":
		cfninject.Inject()
	case "extract":
		cfnextract.Extract()
	default:
		fmt.Println("Unknown command: ", command)
	}
}
