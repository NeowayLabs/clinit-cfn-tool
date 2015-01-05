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
