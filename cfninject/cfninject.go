package cfninject

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/jteeuwen/go-pkg-optarg"
	"github.com/tiago4orion/cloudinit-convert/utils"
)

type CloudInit struct {
	CurrentDate         string
	Hostname            string
	Description         string
	MasterCloudInitData string
	CloudInitData       string
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
	}

	joinTemplate := `{ "Fn::Base64": {"Fn::Join" : ["", [
	    {{ range $index, $element := . }}"{{ $element }}\n"{{if last $index $ | not}},
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
	cloudInit.CloudInitData = applyJoinTemplate(cloudInitData)

	if len(masterCloudInitData) > 0 && masterCloudInitData != nil {
		cloudInit.MasterCloudInitData = applyJoinTemplate(masterCloudInitData)
	}

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
