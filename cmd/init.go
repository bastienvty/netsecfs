package cmd

import (
	"path/filepath"
	"regexp"

	"github.com/bastienvty/netsecfs/internal/db/meta"
	"github.com/bastienvty/netsecfs/internal/db/object"
	"github.com/bastienvty/netsecfs/utils"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const (
	BlockSize = 4096
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
	name := args[0]
	validName := regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{1,61}[a-z0-9]$`)
	if !validName.MatchString(name) {
		logger.Fatalf("invalid name: %s, only alphabet, number and - are allowed, and the length should be 3 to 63 characters.", name)
	}

	m := meta.RegisterMeta(addr)

	format := &meta.Format{
		Name:    name,
		UUID:    uuid.New().String(),
		Storage: storage,
		// Capacity:  utils.ParseBytes(c, "capacity", 'G'),
		BlockSize: BlockSize,
	}
	p, err := filepath.Abs(format.Storage)
	if err != nil {
		logger.Fatalf("Failed to get absolute path of %s: %s", format.Storage, err)
	}
	format.Storage = p

	blob, err := object.CreateStorage(storage)
	logger.Infof("Data use %s", blob)
	if err != nil {
		panic(err)
	}

	if err := m.Init(format); err != nil {
		panic(err)
	}
	logger.Infof("Volume is formatted as %s", format)
}

func init() {
	initCmd.Flags().StringP("storage", "s", "", "Path to the storage database.")
	initCmd.Flags().StringP("meta", "m", "", "Path to the meta database.")
	initCmd.MarkFlagRequired("storage")
	initCmd.MarkFlagRequired("meta")
}
