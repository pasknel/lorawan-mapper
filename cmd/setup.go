package cmd

import (
	"fmt"

	"github.com/pasknel/lorawan-mapper/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup neo4j database",
	Long:  `Setup neo4j database`,

	Run: func(cmd *cobra.Command, args []string) {
		util.PrintBanner()

		if err := util.SetupNeo4j(); err != nil {
			log.Fatal(err)
		}

		fmt.Println("[*] Neo4j setup finished!")
	},
}

func init() {
}
