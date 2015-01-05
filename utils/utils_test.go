package utils

import (
	"fmt"
	"testing"
)

type TestData struct {
	Name  string
	Value string
}

func TestReadFile(t *testing.T) {
	content := ReadFile("./../test/samples/file1.txt")

	if content != "TEST\n" {
		t.Error("Error ocurred")
	}
}

func TestDecodeJson(t *testing.T) {
	var jsonData string = `{
    "name": "value",
    "name2": "value2"
}`
	mapDataInt, err := DecodeJson([]byte(jsonData))

	if err != nil {
		t.Error("Error when decoding ", jsonData)
	}

	mapData := mapDataInt.(map[string]interface{})

	if mapData["name"] != "value" {
		t.Error("mapData['name'] doesn't exists")
	}

	if mapData["name2"] != "value2" {
		t.Error("mapData['name2'] doesn't exists")
	}
}

func TestEncodeJson(t *testing.T) {
	tData := TestData{
		Name:  "TESTE",
		Value: "VALUE",
	}

	jsonDataBytes, err := EncodeJson(tData)

	if err != nil {
		t.Error("Failed to encode JSON")
	}

	jsonDataStr := string(jsonDataBytes)

	if jsonDataStr != `{"Name":"TESTE","Value":"VALUE"}` {
		t.Error("Encode error")
	}
}

func TestDecodeYaml(t *testing.T) {
	var yamlData string = `
name: value
name2: value2
`
	mapData, err := DecodeYaml([]byte(yamlData))

	if err != nil {
		t.Error("Failed to decode Yaml")
	}

	fmt.Println(mapData)
}

func TestEncodeYaml(t *testing.T) {
	var tData = TestData{
		Name:  "TESTE",
		Value: "VALUE",
	}

	yamlData, err := EncodeYaml(tData)

	if err != nil {
		t.Error("Failed to encode YAML")
	}

	yamlDataStr := string(yamlData)

	if yamlDataStr != `name: TESTE
value: VALUE
` {
		t.Error("Failed to encode YAML")
	}
}
