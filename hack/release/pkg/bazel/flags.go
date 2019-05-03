/*
Copyright 2019 The Jetstack cert-manager contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bazel

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/go-logr/logr"
	flag "github.com/spf13/pflag"

	"github.com/jetstack/cert-manager/hack/release/pkg/flags"
	logf "github.com/jetstack/cert-manager/hack/release/pkg/log"
	"github.com/jetstack/cert-manager/hack/release/pkg/util"
)

var (
	Default = &Bazel{}
	log     = logf.Log.WithName("bazel")
)

type Bazel struct {
	// Path to the bazel binary
	bazel string
}

func (g *Bazel) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&g.bazel, "bazel.path", "bazel", "path to the bazel command")
}

func (g *Bazel) Validate() []error {
	var errs []error

	if g.bazel == "" {
		errs = append(errs, fmt.Errorf("--bazel-path must be specified"))
	}

	return errs
}

func (g *Bazel) Complete() error {
	return nil
}

func (g *Bazel) Cmd(ctx context.Context, args ...string) *exec.Cmd {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, g.bazel, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DOCKER_REPO=%s", flags.Default.DockerRepo),
		fmt.Sprintf("APP_VERSION=%s", flags.Default.AppVersion),
		fmt.Sprintf("APP_GIT_COMMIT=%s", flags.Default.GitCommitRef),
	)
	log.Info("set command environment variables", "env", cmd.Env)
	cmd.Dir = flags.Default.RepoRoot
	return cmd
}

// Build will construct a new exec.Cmd that will build the given targets.
// It will produce artifacts suitable for running on the current host OS.
// The --platforms flag *will not* be set.
func (g *Bazel) Build(ctx context.Context, targets ...string) *exec.Cmd {
	return g.Cmd(ctx, append([]string{"build"}, targets...)...)
}

// BuildPlatform will construct a new exec.Cmd that will build the given
// targets for the provided os and architecture.
// This is useful when producing cross-builds.
// Bazel's --platforms variable will be set automatically.
func (g *Bazel) BuildPlatform(ctx context.Context, os, arch string, targets ...string) *exec.Cmd {
	platform := fmt.Sprintf("--platforms=@io_bazel_rules_go//go/toolchain:%s_%s", os, arch)
	return g.Cmd(ctx, append([]string{"build", platform}, targets...)...)
}

// BuildE will build the given targets for the current host OS.
// The --platforms flag *will not* be set.
func (g *Bazel) BuildE(ctx context.Context, log logr.Logger, targets ...string) error {
	return util.RunE(log, g.Build(ctx, targets...))
}

// BuildPlatformE will build the given targets for the provided os and
// architecture.
// This is useful when producing cross-builds.
// Bazel's --platforms variable will be set automatically.
func (g *Bazel) BuildPlatformE(ctx context.Context, log logr.Logger, os, arch string, targets ...string) error {
	return util.RunE(log, g.BuildPlatform(ctx, os, arch, targets...))
}

// Run will construct a new exec.Cmd that will run the given targets
// It will produce artifacts suitable for running on the current host OS.
// The --platforms flag *will not* be set.
func (g *Bazel) Run(ctx context.Context, target string) *exec.Cmd {
	return g.Cmd(ctx, []string{"run", target}...)
}

// RunPlatform will construct a new exec.Cmd that will build the given
// target for the provided os and architecture.
// This is useful when producing cross-builds.
// Bazel's --platforms variable will be set automatically.
func (g *Bazel) RunPlatform(ctx context.Context, os, arch string, targets ...string) *exec.Cmd {
	platform := fmt.Sprintf("--platforms=@io_bazel_rules_go//go/toolchain:%s_%s", os, arch)
	return g.Cmd(ctx, append([]string{"run", platform}, targets...)...)
}

// RunE will run the given targets on the current host OS.
// The --platforms flag *will not* be set.
func (g *Bazel) RunE(ctx context.Context, log logr.Logger, target string) error {
	return util.RunE(log, g.Run(ctx, target))
}

// RunPlatformE will run the given target, built for the provided os and
// architecture.
// This is useful when producing cross-builds.
// Bazel's --platforms variable will be set automatically.
func (g *Bazel) RunPlatformE(ctx context.Context, log logr.Logger, os, arch string, target string) error {
	return util.RunE(log, g.RunPlatform(ctx, os, arch, target))
}
