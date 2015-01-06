package cfninject

import (
	"testing"

	"github.com/neowaylabs/clinit-cfn-tool/utils"
)

func TestValidateAwsTemplate(t *testing.T) {
	content := utils.ReadFile("./../test/samples/file2.txt")

	if !ValidateAwsTemplate(content) {
		t.Error("File not validate as AWS CloudFormation with CloudInitData variable")
	}
}

func TestGetAwsVars(t *testing.T) {
	var1 := ".GetAtt.MasterInstance.PrivateIp"
	awsVar, err := getAwsVar(var1)

	utils.Check(err)

	if awsVar != `, {"Fn::GetAtt": ["MasterInstance","PrivateIp"]}, "` {
		t.Error("Error when converting variable '", var1, "'")
	}

	var2 := ".GetAtt.NodeInstance.Name"
	awsVar, err = getAwsVar(var2)

	utils.Check(err)

	if awsVar != `, {"Fn::GetAtt": ["NodeInstance","Name"]}, "` {
		t.Error("Error when converting variable '", var2, "'")
	}

	var3 := ".GetAtt"
	awsVar, err = getAwsVar(var3)

	if err == nil {
		t.Error("Test should fail....., variable '", var3, "'")
	}
}
