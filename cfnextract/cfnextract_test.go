package cfnextract

import "testing"
import "github.com/neowaylabs/clinit-cfn-tool/utils"

func TestGetAwsUserData(t *testing.T) {
	cfnData := utils.ReadFile("../test/samples/cfn1.json")
	cfnMap, err := utils.DecodeJson([]byte(cfnData))

	utils.Check(err)

	userDataArr := getAwsUserData(cfnMap.(map[string]interface{}))

	if len(userDataArr) != 2 {
		t.Error("Error ocurred getting UserData from cfn1.json")
	}

	cfnData = utils.ReadFile("../test/samples/cfn2.json")
	cfnMap, err = utils.DecodeJson([]byte(cfnData))

	utils.Check(err)

	userDataArr = getAwsUserData(cfnMap.(map[string]interface{}))

	if len(userDataArr) != 1 {
		t.Error("Error ocurred getting UserData from cfn2.json")
	}
}
