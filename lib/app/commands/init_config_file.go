package commands

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
)

func CreateNewConfigFile(gw types.GatewayContract) {
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

	templateText, err := utils.GetUrlContents(consts.APP_NEW_CONFIG_TEMPLATE_URL, &gw)
	if err != nil {
		fmt.Printf("error retrieving configuration file template: %v\n", err)
		return
	}

	var writer strings.Builder

	tmpl, _ := template.New(filename).Parse(templateText)
	tmpl.Execute(&writer, map[string]string{"ProjectType": dependencyBin})

	os.WriteFile(filename, []byte(writer.String()), 0644)
}
