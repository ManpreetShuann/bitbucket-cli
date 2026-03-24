package cmd

import (
	"fmt"

	"github.com/ManpreetShuann/bitbucket-cli/internal/client"
	"github.com/ManpreetShuann/bitbucket-cli/internal/config"
	"github.com/ManpreetShuann/bitbucket-cli/internal/output"
	"github.com/spf13/cobra"
)

type GlobalFlags struct {
	Profile string
	JSON    bool
	Format  string
	NoColor bool
	Debug   bool
}

func NewRootCmd(version string) *cobra.Command {
	flags := &GlobalFlags{}

	rootCmd := &cobra.Command{
		Use:     "bb",
		Short:   "Bitbucket Server CLI",
		Long:    "A command-line interface for Bitbucket Server",
		Version: version,
	}

	rootCmd.PersistentFlags().StringVar(&flags.Profile, "profile", "default", "config profile to use")
	rootCmd.PersistentFlags().BoolVar(&flags.JSON, "json", false, "output as JSON")
	rootCmd.PersistentFlags().StringVar(&flags.Format, "format", "", "Go template for custom output")
	rootCmd.PersistentFlags().BoolVar(&flags.NoColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "print HTTP request/response details to stderr")

	rootCmd.AddCommand(NewAuthCmd(flags))
	rootCmd.AddCommand(NewProjectCmd(flags))
	rootCmd.AddCommand(NewRepoCmd(flags))
	rootCmd.AddCommand(NewBranchCmd(flags))
	rootCmd.AddCommand(NewTagCmd(flags))
	rootCmd.AddCommand(NewPRCmd(flags))
	rootCmd.AddCommand(NewCommitCmd(flags))
	rootCmd.AddCommand(NewFileCmd(flags))
	rootCmd.AddCommand(NewSearchCmd(flags))
	rootCmd.AddCommand(NewUserCmd(flags))
	rootCmd.AddCommand(NewDashboardCmd(flags))
	rootCmd.AddCommand(NewAttachmentCmd(flags))

	return rootCmd
}

func newClient(flags *GlobalFlags) (*client.Client, *config.Config, error) {
	cfg, err := config.Load(flags.Profile, "")
	if err != nil {
		return nil, nil, err
	}
	if cfg.URL == "" {
		return nil, nil, fmt.Errorf("no Bitbucket URL configured. Run 'bb auth login' or set BITBUCKET_URL")
	}
	if cfg.Token == "" {
		return nil, nil, fmt.Errorf("no token configured. Run 'bb auth login' or set BITBUCKET_TOKEN")
	}
	c := client.New(cfg.URL, cfg.Token)
	c.SetDebug(flags.Debug)
	return c, cfg, nil
}

func resolveProjectRepo(cfg *config.Config, args []string, minArgs int) (string, string, []string, error) {
	project := cfg.DefaultProject
	repo := cfg.DefaultRepo
	remaining := args

	if len(args) >= 2 {
		project = args[0]
		repo = args[1]
		remaining = args[2:]
	} else if len(args) == 1 && project != "" {
		repo = args[0]
		remaining = args[1:]
	}

	if project == "" {
		return "", "", nil, fmt.Errorf("project is required (provide as argument or set with 'bb repo use')")
	}
	if minArgs >= 2 && repo == "" {
		return "", "", nil, fmt.Errorf("repository is required (provide as argument or set with 'bb repo use')")
	}

	return project, repo, remaining, nil
}

func printFormatted(flags *GlobalFlags, data any) error {
	f := output.NewFormatter(flags.JSON, flags.Format, flags.NoColor)
	if f != nil {
		out, err := f.Format(data)
		if err != nil {
			return err
		}
		fmt.Println(out)
	}
	return nil
}
