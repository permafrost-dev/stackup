package commands

import (
	"fmt"
	"os"

	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/utils"
)

func CreateNewConfigFile() {
	filename := "stackup.yaml"

	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("%s already exists.\n", filename)
		return
	}

	var dependencyBin string = "php"

	if utils.IsFile("composer.json") {
		dependencyBin = "php"
	} else if utils.IsFile("package.json") {
		dependencyBin = "node"
	} else if utils.IsFile("requirements.txt") {
		dependencyBin = "python"
	}

	contents := fmt.Sprintf(consts.INIT_CONFIG_FILE_CONTENTS, dependencyBin)
	os.WriteFile(filename, []byte(contents), 0644)
}
