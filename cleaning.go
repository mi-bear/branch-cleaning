package main

import (
	"context"
	"os"

	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/joho/godotenv"
	"github.com/mi-bear/go-github/github" // fork dip/githubenterprise
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
)

type (
	// GitHubEnvironment is GitHubEnterprise Environments.
	gitHubEnvironment struct {
		Owner string
		Token string
	}
)

var (
	// GitHubEnvironments is GitHubEnterprise Environments.
	gitHubEnvironments = func() *gitHubEnvironment {
		return &gitHubEnvironment{
			Owner: os.Getenv("GHE_OWNER"),
			Token: os.Getenv("GHE_TOKEN"),
		}
	}

	protectBranches = []string{"master", "staging", "develop"}
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// env load for development
	_ = godotenv.Load()
}

func main() {
	app := cli.NewApp()
	app.Name = "branch-cleaning"
	app.Usage = "Clean the merged branch!"
	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "repository",
			Value: "repository-name",
			Usage: "Repository name you want to clean.",
		},
	}

	app.Action = func(c *cli.Context) error {

		owner := gitHubEnvironments().Owner
		repo := c.String("repository")
		ctx := context.Background()

		service, err := gitHubInit(ctx, owner, repo)
		if err != nil {
			logrus.Fatal(err)
		}

		list, err := branchList(ctx, owner, repo, service)
		if err != nil {
			logrus.Fatal(err)
		}

		for _, row := range list {
			slice := strings.SplitN(*row.Ref, "/", 3)

			in := func(str string, list []string) bool {
				for _, v := range list {
					if v == str {
						return true
					}
				}
				return false
			}

			if in(slice[2], protectBranches) {
				continue
			}

			if err := deleteBranch(ctx, owner, repo, slice[2], service); err != nil {
				logrus.Debug(err) // debug only
			}
		}

		return nil
	}

	app.Run(os.Args)
}

func gitHubInit(ctx context.Context, owner, repo string) (service *github.GitService, err error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: gitHubEnvironments().Token,
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	service = client.Git

	// base branch
	base := "heads/master"
	_, _, err = service.GetRef(ctx, owner, repo, base)
	if err != nil {
		return nil, errors.Wrap(err, "GitHubEnterprise init")
	}
	return service, nil
}

func branchList(ctx context.Context, owner, repo string, service *github.GitService) ([]*github.Reference, error) {
	opt := &github.ReferenceListOptions{
		Type: "heads",
	}
	refs, _, err := service.ListRefs(ctx, owner, repo, opt)
	if err != nil {
		return nil, errors.Wrap(err, "GitHubEnterprise BranchList[ListRefs]")
	}
	return refs, nil
}

func deleteBranch(ctx context.Context, owner, repo, branch string, service *github.GitService) error {
	ref := "refs/heads/" + branch
	_, err := service.DeleteRef(ctx, owner, repo, ref)
	if err != nil {
		return errors.Wrap(err, "GitHubEnterprise DeleteBranch[CreateRef]")
	}
	return nil
}
