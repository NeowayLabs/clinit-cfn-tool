package cfninject

import (
	"testing"

	"github.com/tiago4orion/cloudinit-convert/utils"
)

func TestValidateAwsTemplate(t *testing.T) {
	content := utils.ReadFile("./../test/samples/file2.txt")

	if !ValidateAwsTemplate(content) {
		t.Error("File not validate as AWS CloudFormation with CloudInitData variable")
	}
}
