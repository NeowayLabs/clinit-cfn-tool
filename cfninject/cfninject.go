package cfninject

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/jteeuwen/go-pkg-optarg"
	"github.com/neowaylabs/clinit-cfn-tool/utils"
)

type CloudInit struct {
	CurrentDate         string
	Hostname            string
	Description         string
	MasterCloudInitData string
	CloudInitData       string
}

func getAwsVar(variable string) (string, error) {
	varParts := strings.Split(variable, ".")

	if len(varParts) >= 3 {
		return `, {"Fn::GetAtt": ["` + strings.Join(varParts[2:], `","`) + `"]}, "`, nil
	} else {
		return "", errors.New("Invalid variable '" + variable + "'")
	}
}

func ValidateAwsTemplate(content string) bool {
	return strings.LastIndexAny(content, " .CloudInitData ") >= 0
}

func applyJoinTemplate(cloudData []string) string {
	var buffer bytes.Buffer

	fns := template.FuncMap{
		"last": func(x int, a interface{}) bool {
			return x == reflect.ValueOf(a).Len()-1
		},
		"quote": func(str string) string {
			return strconv.Quote(str)
		},
		"concat": func(str1 string, str2 string) string {
			return str1 + str2
		},
		"quote_concat": func(str1 string, str2 string) string {
			return strconv.Quote(str1 + str2)
		},
		"quote_replace_vars": func(str string) string {
			// in:          --etcd_servers=http://{{ .GetAtt.KubernetesMasterInstance.PrivateIp }}:4001\
			// out: "        --etcd_servers=http://", {"Fn::GetAtt" :["KubernetesMasterInstance" , "PrivateIp"]}, ":4001\\\n",

			str = str + "\n"
			retStr := str
			idx1 := strings.Index(str, "{{")

			if idx1 != -1 {
				idx2 := strings.Index(str[idx1:], "}}")

				if idx2 != -1 {
					variable := strings.Trim(str[idx1+2:(idx1+idx2-1)], " ")
					variableSubst, err := getAwsVar(variable)

					if err != nil {
						retStr = strconv.Quote(str)
					} else {
						retStr = strconv.Quote(str[0:idx1]) + variableSubst + strconv.Quote(str[idx1+idx2+2:])
					}
				}
			} else {
				retStr = strconv.Quote(retStr)
			}

			return retStr
		},
	}

	joinTemplate := `{ "Fn::Base64": {"Fn::Join" : ["", [
	    {{ range $index, $element := . }}{{ quote_replace_vars $element }}{{if last $index $ | not}},
	    {{end}}{{ end }}
          ]]}
        }`

	tmpl, err := template.New("aws-template-join").Funcs(fns).Parse(joinTemplate)

	utils.Check(err)

	err = tmpl.Execute(&buffer, cloudData)
	utils.Check(err)

	return buffer.String()
}

func ApplyTemplate(cloudInit CloudInit, masterCloudInitData []string, cloudInitData []string, awsContentStr string) bool {
	if len(masterCloudInitData) > 0 && masterCloudInitData != nil {
		cloudInit.MasterCloudInitData = applyJoinTemplate(masterCloudInitData)
	}

	cloudInit.CloudInitData = applyJoinTemplate(cloudInitData)

	tmpl, err := template.New("aws-cloud-formation").Parse(awsContentStr)

	utils.Check(err)

	err = tmpl.Execute(os.Stdout, cloudInit)

	utils.Check(err)

	return true
}

func Conversion(masterClinitPath string, clinitPath string, awsFormationPath string) bool {
	var masterClinitContentStr, clinitContentStr, awsFormationContentStr string
	var masterClinitParts, clinitParts []string

	if clinitPath == "" || awsFormationPath == "" {
		return false
	}

	if masterClinitPath != "" {
		masterClinitContentStr = utils.ReadFile(masterClinitPath)
	}

	clinitContentStr = utils.ReadFile(clinitPath)
	awsFormationContentStr = utils.ReadFile(awsFormationPath)

	clinitParts = strings.Split(clinitContentStr, "\n")

	if len(clinitParts) <= 0 {
		return false
	}

	if masterClinitContentStr != "" {
		masterClinitParts = strings.Split(masterClinitContentStr, "\n")
	} else {
		masterClinitParts = nil
	}

	if !ValidateAwsTemplate(awsFormationContentStr) {
		return false
	}

	cloudInit := CloudInit{
		Hostname:    "core-master",
		CurrentDate: "2010-01-01",
		Description: "Kubernetes AWS Formation",
	}

	return ApplyTemplate(cloudInit, masterClinitParts, clinitParts, awsFormationContentStr)
}

func Inject() {
	var cloudinitOpt, awsFormationTpl, masterCloudinitOpt string
	var helpOpt, missingOpts bool

	optarg.Add("h", "help", "Displays this help", false)
	optarg.Add("c", "cloudinit", "Cloudinit input file", "")
	optarg.Add("m", "master-cloudinit", "Cloudinit of Master Instance", "")
	optarg.Add("f", "aws-cloud-formation-tpl", "AWS CloudFormation template", "")

	for opt := range optarg.Parse() {
		switch opt.ShortName {
		case "c":
			cloudinitOpt = opt.String()
		case "h":
			helpOpt = opt.Bool()
		case "f":
			awsFormationTpl = opt.String()
		case "m":
			masterCloudinitOpt = opt.String()
		}
	}

	if helpOpt {
		optarg.Usage()
		os.Exit(0)
	}

	if cloudinitOpt == "" {
		fmt.Println("-c is required...")
		missingOpts = true
	}

	if awsFormationTpl == "" {
		fmt.Println("-f is required...")
		missingOpts = true
	}

	if missingOpts {
		optarg.Usage()
		os.Exit(1)
	}

	if os.Getenv("DEBUG_OPTS") != "" {
		fmt.Println(cloudinitOpt)
		fmt.Println(awsFormationTpl)
	}

	Conversion(masterCloudinitOpt, cloudinitOpt, awsFormationTpl)
}
