package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/meinto/git-semver/file"
	"github.com/meinto/git-semver/git"

	"github.com/gobuffalo/packr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmdFlags struct {
	shellPath         string
	verbose         bool
	push            bool
	createTag       bool
	versionFile     string
	versionFileType string
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rootCmdFlags.shellPath, "shellPath", "/bin/bash", "path to shell executor")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdFlags.verbose, "verbose", "v", false, "more logs")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdFlags.push, "push", "P", false, "push git tags")
	rootCmd.PersistentFlags().BoolVarP(&rootCmdFlags.createTag, "tag", "T", false, "create a git tag")
	rootCmd.PersistentFlags().StringVarP(&rootCmdFlags.versionFile, "versionFile", "f", "VERSION", "name of version file")
	rootCmd.PersistentFlags().StringVarP(&rootCmdFlags.versionFileType, "versionFileType", "t", "raw", "type of version file (json, raw)")

	viper.BindPFlag("shellPath", rootCmd.PersistentFlags().Lookup("shellPath"))
	viper.BindPFlag("pushChanges", rootCmd.PersistentFlags().Lookup("push"))
	viper.BindPFlag("tagVersions", rootCmd.PersistentFlags().Lookup("tag"))
	viper.BindPFlag("versionFile", rootCmd.PersistentFlags().Lookup("versionFile"))
	viper.BindPFlag("versionFileType", rootCmd.PersistentFlags().Lookup("versionFileType"))
}

var rootCmd = &cobra.Command{
	Use:   "semver",
	Short: "standalone tool to version your gitlab repo with semver",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		g := git.NewGitService(viper.GetString("shellPath"))
		repoPath, _ := g.GitRepoPath()

		viper.SetConfigName("semver.config")
		viper.SetConfigType("json")
		viper.AddConfigPath(repoPath)
		err := viper.ReadInConfig()
		if err != nil {
			log.Println("there is no semver.config file: ", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		var g git.Service
		var fs file.VersionFileService
		var repoPath string
		if rootCmdFlags.push || rootCmdFlags.createTag {
			g = git.NewGitService(viper.GetString("shellPath"))
			rp, err := g.GitRepoPath()
			if err != nil {
				log.Fatal(err)
			}
			repoPath = rp

			versionFilepath := repoPath + "/" + viper.GetString("versionFile")
			fs = file.NewVersionFileService(versionFilepath)
		}

		if rootCmdFlags.push {
			g.AddVersionChanges(viper.GetString("versionFile"))
			currentVersion, err := fs.ReadVersionFromFile(viper.GetString("versionFileType"))
			if err != nil {
				log.Fatal(err)
			}
			g.CommitVersionChanges(currentVersion)
		}

		if rootCmdFlags.createTag {
			fs := file.NewVersionFileService(repoPath + "/" + viper.GetString("versionFile"))
			currentVersion, err := fs.ReadVersionFromFile(viper.GetString("versionFileType"))
			if err != nil {
				log.Fatal(err)
			}
			if err = g.CreateTag(currentVersion); err != nil {
				log.Fatal(err)
			}
		}

		if rootCmdFlags.push {
			g.Push()
		}

		if !rootCmdFlags.createTag && !rootCmdFlags.push {
			box := packr.NewBox("../../../buildAssets")
			version, err := box.FindString("VERSION")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Version of git-semver: %s\n", version)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
