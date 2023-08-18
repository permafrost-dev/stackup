package scripting

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ScriptFs struct {
}

func CreateScriptFsObject(e *JavaScriptEngine) {
	obj := &ScriptFs{}
	e.Vm.Set("fs", obj)
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
	if err != nil {
		return err
	}

	return nil
}

func (fs *ScriptFs) ReadJSON(filename string) (interface{}, error) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, err
	}

	var data interface{}

	// Read the contents of the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(content, &data)
	if err != nil {
		return data, err
	}

	return data, nil
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
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

func (fs *ScriptFs) IsDirectory(filename string) bool {
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return fileInfo.IsDir()
}

func (fs *ScriptFs) IsFile(filename string) bool {
	return !fs.IsDirectory(filename)
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
