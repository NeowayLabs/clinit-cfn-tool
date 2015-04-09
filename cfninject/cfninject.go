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

	"github.com/NeowayLabs/clinit-cfn-tool/utils"
	"github.com/jteeuwen/go-pkg-optarg"
)

type CloudInitInfo struct {
	Variable string
	Path     string
}

func getAwsVar(variable string) (string, error) {
	varParts := strings.Split(variable, ".")

	if len(varParts) <= 2 {
		return "", errors.New("Invalid variable '" + variable + "'")
	}

	if varParts[1] == "Ref" {
		return `, {"Ref": "` + varParts[2] + `"}, `, nil
	}

	if varParts[1] == "GetAtt" && len(varParts) >= 3 {
		return `, {"Fn::GetAtt": ["` + strings.Join(varParts[2:], `","`) + `"]}, `, nil
	}

	return "", errors.New("Invalid variable '" + variable + "'")
}

func ValidateAwsTemplate(content, variable string) bool {
	return strings.LastIndexAny(content, " ."+variable+" ") >= 0
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

func ApplyTemplate(cloudInit map[string]string, awsContentStr string) (bool, string) {
	var buf bytes.Buffer
	tmpl, err := template.New("aws-cloud-formation").Parse(awsContentStr)

	utils.Check(err)

	err = tmpl.Execute(&buf, cloudInit)
	utils.Check(err)

	return true, string(buf.Bytes())
}

func Conversion(clinitConfig []CloudInitInfo, awsFormationPath string) bool {
	var clinitContentStr, awsFormationContentStr string
	var clinitParts []string
	var ok bool

	if awsFormationPath == "" {
		return false
	}

	awsFormationContentStr = utils.ReadFile(awsFormationPath)
	cloudInit := make(map[string]string)

	for _, cl := range clinitConfig {
		clinitPath := cl.Path
		clInitVar := cl.Variable

		clinitContentStr = utils.ReadFile(clinitPath)
		clinitParts = strings.Split(clinitContentStr, "\n")

		if len(clinitParts) <= 0 {
			return false
		}

		if !ValidateAwsTemplate(awsFormationContentStr, clInitVar) {
			return false
		}

		cloudInit[clInitVar] = applyJoinTemplate(clinitParts)
	}

	ok, awsFormationContentStr = ApplyTemplate(cloudInit, awsFormationContentStr)

	if ok {
		fmt.Println(awsFormationContentStr)
		return true
	}

	return false
}

func configureCloudInit(cloudinitPairs string) ([]CloudInitInfo, error) {
	clParts := strings.Split(cloudinitPairs, ",")
	cloudinitInfo := make([]CloudInitInfo, len(clParts))

	for i, cl := range clParts {
		varPath := strings.Split(cl, ":")

		if len(varPath) != 2 {
			return cloudinitInfo, errors.New("Invalid -c option")
		}

		cloudinitInfo[i].Variable = varPath[0]
		cloudinitInfo[i].Path = varPath[1]
	}

	return cloudinitInfo, nil
}

func Inject() {
	var cloudinitOpt, awsFormationTpl string
	var helpOpt, missingOpts bool

	optarg.Add("h", "help", "Displays this help", false)
	optarg.Add("c", "cloudinit", "Cloudinit pairs: <VARIABLE_NAME>:<file-path>[,<VARIABLE_NAME>,<file-path>]", "")
	optarg.Add("f", "aws-cloud-formation-tpl", "AWS CloudFormation template", "")

	for opt := range optarg.Parse() {
		switch opt.ShortName {
		case "c":
			cloudinitOpt = opt.String()
		case "h":
			helpOpt = opt.Bool()
		case "f":
			awsFormationTpl = opt.String()
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

	cloudInitPairs, err := configureCloudInit(cloudinitOpt)

	if err != nil {
		fmt.Println(err.Error())
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

	Conversion(cloudInitPairs, awsFormationTpl)
}
