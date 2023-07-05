package cmd

import (
	"github.com/permafrost-dev/stack-supervisor/lib"
	"github.com/permafrost-dev/stack-supervisor/server"
	"github.com/permafrost-dev/stack-supervisor/utils"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start the development stack",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: startCmdRun,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	lib.InitGlobals()

	// 	config, err := configuration.LoadStackConfig(utils.WorkingDir("/stack-supervisor.config.yaml"))

	// 	if err != nil {
	// 		log.Fatalln(err)
	// 		os.Exit(0)
	// 	}

	// 	app := app.Application{Config: &config}
	// 	app.Run(&config, nil)
	// },
}

func startCmdRun(cmd *cobra.Command, args []string) {
	lib.InitGlobals()

	srv := server.WebServer{}
	srv.Start()

	app := lib.Application{}
	app.LoadStackConfig(utils.WorkingDir("/stack-supervisor.config.yaml"))
	app.Run(cmd)
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define yourS flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	startCmd.Flags().BoolP("seed", "s", false, "Rebuild and seed the database")
}
