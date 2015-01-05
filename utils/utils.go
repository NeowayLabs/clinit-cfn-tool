// Utils functions
package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func ReadFile(path string) string {
	contentBytes, err := ioutil.ReadFile(path)

	Check(err)
	return string(contentBytes)
}

func SaveOutput(path string, content string) error {
	fileOut, err := os.Create(path)

	if err != nil {
		return err
	}

	// close fo on exit and check for its returned error
	defer func() {
		if err := fileOut.Close(); err != nil {
			panic(err)
		}
	}()

	// write a chunk
	if _, err := fileOut.Write([]byte(content)); err != nil {
		panic(err)
	}

	return nil
}

func DecodeJson(data []byte) (interface{}, error) {
	var mapDataVoid interface{}
	var mapData map[string]interface{}

	err := json.Unmarshal(data, &mapDataVoid)

	if err == nil {
		mapData = mapDataVoid.(map[string]interface{})
	}

	return mapData, err
}

func EncodeJson(mapData interface{}) ([]byte, error) {
	data, err := json.Marshal(mapData)

	return data, err
}

func DecodeYaml(data []byte) (map[string]interface{}, error) {
	var mapData map[string]interface{}

	err := yaml.Unmarshal(data, &mapData)

	return mapData, err
}

func EncodeYaml(mapData interface{}) ([]byte, error) {
	data, err := yaml.Marshal(mapData)

	return data, err
}
