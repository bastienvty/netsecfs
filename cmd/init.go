package cmd

import (
	"github.com/bastienvty/netsecfs/internal/db"
	"github.com/bastienvty/netsecfs/utils"
	"github.com/spf13/cobra"
)

var logger = utils.GetLogger("juicefs")

// initCmd represents the client command
var initCmd = &cobra.Command{
	Use:   "init [flags] NAME",
	Short: "Initialize the filesystem.",
	Long: `Initialize the filesystem by creating all necessary 
databases.`,
	Args:    cobra.ExactArgs(1),
	Example: "netsecfs init --storage /path/to/storage --meta /path/to/meta.db NAME",
	Run:     initialize,
}

func initialize(cmd *cobra.Command, args []string) {
	storage, _ := cmd.Flags().GetString("storage")
	addr, _ := cmd.Flags().GetString("meta")

	m := db.RegisterMeta(addr)
	blob, err := db.CreateStorage(storage)
	if err != nil {
		panic(err)
	}
	logger.Infof("Data use %s", blob)

	if err := m.Init(); err != nil {
		panic(err)
	}
}

func init() {
	initCmd.Flags().StringP("storage", "s", "", "Path to the storage database.")
	initCmd.Flags().StringP("meta", "m", "", "Path to the meta database.")
	initCmd.MarkFlagRequired("storage")
	initCmd.MarkFlagRequired("meta")
}
