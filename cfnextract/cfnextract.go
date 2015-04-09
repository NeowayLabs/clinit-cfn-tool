package cfnextract

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/NeowayLabs/clinit-cfn-tool/utils"
	"github.com/jteeuwen/go-pkg-optarg"
)

func getAwsResources(awsMap map[string]interface{}) (map[string]interface{}, error) {
	for k, v := range awsMap {
		if k == "Resources" {
			return v.(map[string]interface{}), nil
		}
	}

	return awsMap, errors.New("Resources not found...")
}

func getAwsUserData(awsMap map[string]interface{}) []map[string]interface{} {
	var tmp map[string]interface{}
	userDataArr := make([]map[string]interface{}, 0, 0)
	userDataArr2 := make([]map[string]interface{}, 0, 0)

	resources := awsMap["Resources"].(map[string]interface{})

	if resources == nil {
		fmt.Println("AWS CloudFormation Resources not found...")
		return userDataArr
	}

	for kk, vv := range resources {
		tmp = vv.(map[string]interface{})
		if tmp["Properties"] != nil {
			tmp := tmp["Properties"].(map[string]interface{})
			if tmp == nil {
				fmt.Printf("Resource '%s' doesn't have UserData\n", kk)
				continue
			}

			if tmp["UserData"] != nil {
				tmp = tmp["UserData"].(map[string]interface{})
				userDataArr2 = make([]map[string]interface{}, len(userDataArr)+1, len(userDataArr)+1)
				for i := range userDataArr {
					userDataArr2[i] = userDataArr[i]
				}

				userDataArr2[len(userDataArr)] = tmp
				userDataArr = userDataArr2
			}
		} else {
			fmt.Printf("Resource '%s' doesn't have Properties\n", kk)
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
			specialAwsFunc, ok := elem.(map[string]interface{})

			if !ok {
				return "", fmt.Errorf("Unsupported value: ", specialAwsFunc)
			}

			if len(specialAwsFunc) != 1 {
				return "", fmt.Errorf("Unsupported special variable: %s", specialAwsFunc)
			}

			for f, awsValue := range specialAwsFunc {
				v := reflect.TypeOf(awsValue)

				switch v.Kind() {
				case reflect.Slice:
					funcArr := awsValue.([]interface{})
					funcArrStr := make([]string, len(funcArr))
					for ii := range funcArr {
						funcArrStr[ii] = funcArr[ii].(string)
					}

					cloudInitData[i] = "{{ ." + f + "." + strings.Join(funcArrStr, ".") + " }}"
				case reflect.String:
					cloudInitData[i] = "{{ ." + f + "." + awsValue.(string) + " }}"
				default:
					return "", fmt.Errorf("Unsupported special variable of type '%s': %s: %s", v, f, awsValue)
				}
			}
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
	awsFormationContentStr := utils.ReadFile(awsFormationPath)
	awsMapInt, err := utils.DecodeJson([]byte(awsFormationContentStr))

	if err != nil {
		fmt.Println("Failed to decode JSON: %s", err.Error())
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
