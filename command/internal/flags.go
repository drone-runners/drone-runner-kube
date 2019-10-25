// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"time"

	"github.com/drone/drone-go/drone"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags maps
type Flags struct {
	Build  *drone.Build
	Netrc  *drone.Netrc
	Repo   *drone.Repo
	Stage  *drone.Stage
	System *drone.System
}

// ParseFlags parses the flags from the command args.
func ParseFlags(cmd *kingpin.CmdClause) *Flags {
	f := &Flags{
		Build:  &drone.Build{},
		Netrc:  &drone.Netrc{},
		Repo:   &drone.Repo{},
		Stage:  &drone.Stage{},
		System: &drone.System{},
	}

	now := fmt.Sprint(
		time.Now().Unix(),
	)

	cmd.Flag("repo-id", "repo id").Default("1").Int64Var(&f.Repo.ID)
	cmd.Flag("repo-namespace", "repo namespace").Default("").StringVar(&f.Repo.Namespace)
	cmd.Flag("repo-name", "repo name").Default("").StringVar(&f.Repo.Name)
	cmd.Flag("repo-slug", "repo slug").Default("").StringVar(&f.Repo.Slug)
	cmd.Flag("repo-http", "repo http clone url").Default("").StringVar(&f.Repo.HTTPURL)
	cmd.Flag("repo-ssh", "repo ssh clone url").Default("").StringVar(&f.Repo.SSHURL)
	cmd.Flag("repo-link", "repo link").Default("").StringVar(&f.Repo.Link)
	cmd.Flag("repo-branch", "repo branch").Default("").StringVar(&f.Repo.Branch)
	cmd.Flag("repo-private", "repo private").Default("false").BoolVar(&f.Repo.Private)
	cmd.Flag("repo-visibility", "repo visibility").Default("").StringVar(&f.Repo.Visibility)
	cmd.Flag("repo-trusted", "repo trusted").Default("false").BoolVar(&f.Repo.Trusted)
	cmd.Flag("repo-protected", "repo protected").Default("false").BoolVar(&f.Repo.Protected)
	cmd.Flag("repo-timeout", "repo timeout in minutes").Default("60").Int64Var(&f.Repo.Timeout)
	cmd.Flag("repo-created", "repo created").Default(now).Int64Var(&f.Repo.Created)
	cmd.Flag("repo-updated", "repo updated").Default(now).Int64Var(&f.Repo.Updated)

	cmd.Flag("build-id", "build id").Default("1").Int64Var(&f.Build.ID)
	cmd.Flag("build-number", "build number").Default("1").Int64Var(&f.Build.Number)
	cmd.Flag("build-parent", "build parent").Default("0").Int64Var(&f.Build.Parent)
	cmd.Flag("build-event", "build event").Default("push").StringVar(&f.Build.Event)
	cmd.Flag("build-action", "build action").Default("").StringVar(&f.Build.Action)
	cmd.Flag("build-cron", "build cron trigger").Default("").StringVar(&f.Build.Cron)
	cmd.Flag("build-target", "build deploy target").Default("").StringVar(&f.Build.Deploy)
	cmd.Flag("build-created", "build created").Default(now).Int64Var(&f.Build.Created)
	cmd.Flag("build-updated", "build updated").Default(now).Int64Var(&f.Build.Updated)

	cmd.Flag("commit-sender", "commit sender").Default("").StringVar(&f.Build.Sender)
	cmd.Flag("commit-link", "commit link").Default("").StringVar(&f.Build.Link)
	cmd.Flag("commit-title", "commit title").Default("").StringVar(&f.Build.Title)
	cmd.Flag("commit-message", "commit message").Default("").StringVar(&f.Build.Message)
	cmd.Flag("commit-before", "commit before").Default("").StringVar(&f.Build.Before)
	cmd.Flag("commit-after", "commit after").Default("").StringVar(&f.Build.After)
	cmd.Flag("commit-ref", "commit ref").Default("").StringVar(&f.Build.Ref)
	cmd.Flag("commit-fork", "commit fork").Default("").StringVar(&f.Build.Fork)
	cmd.Flag("commit-source", "commit source branch").Default("").StringVar(&f.Build.Source)
	cmd.Flag("commit-target", "commit target branch").Default("").StringVar(&f.Build.Target)

	cmd.Flag("author-login", "commit author login").Default("").StringVar(&f.Build.Author)
	cmd.Flag("author-name", "commit author name").Default("").StringVar(&f.Build.AuthorName)
	cmd.Flag("author-email", "commit author email").Default("").StringVar(&f.Build.AuthorEmail)
	cmd.Flag("author-avatar", "commit author avatar").Default("").StringVar(&f.Build.AuthorAvatar)

	cmd.Flag("stage-id", "stage id").Default("1").Int64Var(&f.Stage.ID)
	cmd.Flag("stage-number", "stage number").Default("1").IntVar(&f.Stage.Number)
	cmd.Flag("stage-kind", "stage kind").Default("").StringVar(&f.Stage.Kind)
	cmd.Flag("stage-type", "stage type").Default("").StringVar(&f.Stage.Type)
	cmd.Flag("stage-name", "stage name").Default("default").StringVar(&f.Stage.Name)
	cmd.Flag("stage-os", "stage os").Default("").StringVar(&f.Stage.OS)
	cmd.Flag("stage-arch", "stage arch").Default("").StringVar(&f.Stage.Arch)
	cmd.Flag("stage-variant", "stage variant").Default("").StringVar(&f.Stage.Variant)
	cmd.Flag("stage-kernel", "stage kernel").Default("").StringVar(&f.Stage.Kernel)
	cmd.Flag("stage-created", "stage created").Default(now).Int64Var(&f.Stage.Created)
	cmd.Flag("stage-updated", "stage updated").Default(now).Int64Var(&f.Stage.Updated)

	cmd.Flag("netrc-username", "netrc username").Default("").StringVar(&f.Netrc.Login)
	cmd.Flag("netrc-password", "netrc password").Default("").StringVar(&f.Netrc.Password)
	cmd.Flag("netrc-machine", "netrc machine").Default("").StringVar(&f.Netrc.Machine)

	cmd.Flag("system-host", "server host").Default("").StringVar(&f.System.Host)
	cmd.Flag("system-proto", "server proto").Default("").StringVar(&f.System.Proto)
	cmd.Flag("system-link", "server link").Default("").StringVar(&f.System.Link)
	cmd.Flag("system-version", "server version").Default("").StringVar(&f.System.Version)

	return f
}
