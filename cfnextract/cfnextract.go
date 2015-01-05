package cfnextract

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jteeuwen/go-pkg-optarg"
	"github.com/tiago4orion/cloudinit-convert/utils"
)

func getAwsUserData(awsMap map[string]interface{}) []map[string]interface{} {
	var tmp, tmp3, tmp4 map[string]interface{}
	userDataArr := make([]map[string]interface{}, 0, 0)
	userDataArr2 := make([]map[string]interface{}, 0, 0)

	for k, v := range awsMap {
		if k == "Resources" {
			tmp = v.(map[string]interface{})
			for kk, vv := range tmp {
				var _ string = kk // unused variable
				tmp = vv.(map[string]interface{})
				if tmp["Properties"] != nil {
					tmp2 := tmp["Properties"].(map[string]interface{})
					tmp3 = tmp2
					if tmp3["UserData"] != nil {
						tmp4 = tmp3["UserData"].(map[string]interface{})
						userDataArr2 = make([]map[string]interface{}, len(userDataArr)+1, len(userDataArr)+1)
						for i := range userDataArr {
							userDataArr2[i] = userDataArr[i]
						}

						userDataArr2[len(userDataArr)] = tmp4
						userDataArr = userDataArr2
					}
				}
			}
		}
	}

	return userDataArr
}

func JoinCfnUserData(userData map[string]interface{}) (string, error) {
	var tmp map[string]interface{}

	if userData["Fn::Base64"] == nil {
		return "", errors.New("UserData doesn't have Fn::Base64 field")
	}

	tmp = userData["Fn::Base64"].(map[string]interface{})

	if tmp["Fn::Join"] == nil {
		return "", errors.New("UserData doesn-t have Fn::Join field")
	}

	tmpArr := tmp["Fn::Join"].([]interface{})

	// TODO: Review this assertion
	if len(tmpArr) <= 0 {
		return "", errors.New("Empty UserData string")
	}

	tmpArr2 := tmpArr[1].([]interface{})

	cloudInitData := make([]string, len(tmpArr2))

	for i, elem := range tmpArr2 {
		if elemStr, ok := elem.(string); ok {
			cloudInitData[i] = elemStr
		} else {
			specialAwsFunc := elem.(map[string]interface{})
			funcArr := specialAwsFunc["Fn::GetAtt"].([]interface{})

			funcArrStr := make([]string, len(funcArr))

			for ii := range funcArr {
				funcArrStr[ii] = funcArr[ii].(string)
			}

			cloudInitData[i] = "$" + strings.Join(funcArrStr, "_")
		}
	}

	return strings.Join(cloudInitData, ""), nil
}

func handleUserData(userData map[string]interface{}) string {
	userDataStr, err := JoinCfnUserData(userData)

	if err != nil {
		utils.Check(err)
	}

	return userDataStr
}

func ExtractCloudinit(baseCloudinitPath string, awsFormationPath string) bool {
	fmt.Println(baseCloudinitPath, awsFormationPath)

	awsFormationContentStr := utils.ReadFile(awsFormationPath)
	awsMapInt, err := utils.DecodeJson([]byte(awsFormationContentStr))

	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	awsMap := awsMapInt.(map[string]interface{})

	if err != nil {
		fmt.Errorf("Failed to decode Json")
		return false
	}

	userData := getAwsUserData(awsMap)

	cloudInitDataArr := make([]string, len(userData), len(userData))

	for i := range userData {
		cloudInitDataArr[i] = handleUserData(userData[i])
	}

	for i := range cloudInitDataArr {
		outPath := baseCloudinitPath + strconv.Itoa(i+1) + ".yaml"
		fmt.Printf("Generating file '%s'\n", outPath)
		err := utils.SaveOutput(outPath, cloudInitDataArr[i])

		utils.Check(err)
	}

	return true
}

func Extract() {
	var baseCloudinitPath, awsFormationPath string
	var helpOpt, missingOpts bool

	optarg.Add("h", "help", "Displays this help", false)
	optarg.Add("o", "output-base-path", "Output base path name.", "")
	optarg.Add("i", "cloud-formation", "CloudFormation input file", "")

	for opt := range optarg.Parse() {
		switch opt.ShortName {
		case "o":
			baseCloudinitPath = opt.String()
		case "i":
			awsFormationPath = opt.String()
		case "h":
			helpOpt = opt.Bool()

		default:
			fmt.Println("Invalid flag: ", opt)
			optarg.Usage()
			os.Exit(1)
		}
	}

	if helpOpt {
		optarg.Usage()
		os.Exit(0)
	}

	if baseCloudinitPath == "" {
		fmt.Println("-o is required...")
		missingOpts = true
	}

	if awsFormationPath == "" {
		fmt.Println("-i is required...")
		missingOpts = true
	}

	if missingOpts {
		optarg.Usage()
		os.Exit(1)
	}

	if os.Getenv("DEBUG_OPTS") != "" {
		fmt.Println(baseCloudinitPath)
		fmt.Println(awsFormationPath)
	}

	ExtractCloudinit(baseCloudinitPath, awsFormationPath)
}