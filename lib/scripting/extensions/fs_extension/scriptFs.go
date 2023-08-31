package fsextension

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
)

type ScriptFs struct {
	types.ScriptExtensionContract
}

func Create() *ScriptFs {
	return &ScriptFs{}
}

func (fs *ScriptFs) GetName() string {
	return "fs"
}

func (ex *ScriptFs) OnInstall(engine types.JavaScriptEngineContract) {
	engine.GetVm().Set(ex.GetName(), ex)
}

func (fs *ScriptFs) ReadFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (fs *ScriptFs) WriteFile(filename string, content string) error {
	err := os.WriteFile(filename, []byte(content), 0644)

	return err
}

func (fs *ScriptFs) ReadJSON(filename string) (interface{}, error) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, err
	}

	var data interface{}

	content, err := os.ReadFile(filename)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(content, &data)

	return data, err
}

func (fs *ScriptFs) WriteJSON(filename string, data interface{}) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return err
	}

	content, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, content, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (fs *ScriptFs) Exists(filename string) bool {
	return utils.FileExists(filename)
}

func (fs *ScriptFs) IsDirectory(filename string) bool {
	return utils.IsDir(filename)
}

func (fs *ScriptFs) IsFile(filename string) bool {
	return utils.IsFile(filename)
}

func (fs *ScriptFs) GetFiles(directory string) ([]string, error) {
	var files []string

	fileInfos, err := os.ReadDir(directory)
	if err != nil {
		return files, err
	}

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			files = append(files, filepath.Join(directory, fileInfo.Name()))
		}
	}

	return files, nil
}
