// A Dagger module for the Elasticsearch Go client CI.
//
// It mirrors the opensearch client CI: build, vet, lint, format, speccheck,
// unit tests, acceptance tests against a containerized Elasticsearch 9 with
// security enabled, CodeCov upload, and git push-back of coverage.out.

package main

import (
	"context"
	"fmt"
	"time"

	"dagger/elasticsearch/internal/dagger"

	"emperror.dev/errors"
	"github.com/disaster37/dagger-library-go/lib/helper"
)

const (
	// ElasticsearchVersion is the pinned 9.x patch used for acceptance tests.
	ElasticsearchVersion string = "9.0.0"
	username             string = "elastic"
	password             string = "vLPeJYa8.3RqtZCcAK6jNz"
	gitUsername          string = "ci"
	gitEmail             string = "ci@localhost"
	defaultGitBranch     string = "release-branch.v9"
)

// Elasticsearch is the Dagger module entrypoint.
type Elasticsearch struct {
	// Src is a directory that contains the project source code.
	// +private
	Src *dagger.Directory

	// GolangModule is the reusable Go module helper.
	// +private
	GolangModule *dagger.Golang
}

// New constructs the Elasticsearch module.
func New(
	ctx context.Context,
	// a path to a directory containing the source code
	// +required
	src *dagger.Directory,
) (*Elasticsearch, error) {
	return &Elasticsearch{
		Src:          src,
		GolangModule: dag.Golang(src.WithoutDirectory("ci")),
	}, nil
}

// Ci runs the full CI pipeline: build, lint, format, speccheck, unit tests,
// acceptance tests, and (when ci=true) CodeCov upload + git push-back.
func (h *Elasticsearch) Ci(
	ctx context.Context,

	// Set true when running on CI (enables CodeCov + git push-back)
	// +default=false
	ci bool,

	// The CodeCov token
	// +optional
	codeCoveToken *dagger.Secret,

	// The git branch to push coverage.out to
	// +optional
	gitBranch string,

	// Set true if the current build is a tag
	// +optional
	isTag bool,

	// The git token
	// +optional
	gitToken *dagger.Secret,
) (dir *dagger.Directory, err error) {
	var stdout string

	h.GolangModule = dag.Golang(h.Src.WithoutDirectory("ci"), dagger.GolangOpts{Base: h.GolangModule.Container().WithExec([]string{"go", "mod", "tidy"})})

	// Build
	if _, err = h.Build(ctx).Sync(ctx); err != nil {
		return nil, errors.Wrap(err, "Error when build project")
	}

	// Vet
	if stdout, err = h.Vet(ctx); err != nil {
		return nil, errors.Wrapf(err, "Error when vet project: %s", stdout)
	}

	// Lint
	if stdout, err = h.Lint(ctx); err != nil {
		return nil, errors.Wrapf(err, "Error when lint project: %s", stdout)
	}

	// Format
	dir = h.Format(ctx)

	// SpecCheck
	if stdout, err = h.SpecCheck(ctx); err != nil {
		return nil, errors.Wrapf(err, "Error when speccheck: %s", stdout)
	}

	// Unit tests
	reportFile := h.Test(ctx, "")
	dir = dir.WithFile("coverage.out", reportFile)

	// Acceptance tests (merged coverage)
	acceptanceFile := h.TestAcceptance(ctx, "")
	dir = dir.WithFile("coverage_acceptance.out", acceptanceFile)

	if ci {
		if codeCoveToken == nil {
			return nil, errors.New("You need to provide CodeCov token")
		}
		if stdout, err = h.CodeCov(ctx, dir, codeCoveToken); err != nil {
			return nil, errors.Wrapf(err, "Error when upload report on CodeCov: %s", stdout)
		}

		git := dag.GitModule(dir.WithDirectory("ci", h.Src.Directory("ci")), dagger.GitModuleOpts{Ci: "github"}).
			SetConfig(dagger.GitModuleSetConfigOpts{
				Username: gitUsername,
				Email:    gitEmail,
			})

		if isTag {
			gitBranch = defaultGitBranch
		}

		if _, err = git.CommitAndPush(
			ctx,
			gitToken,
			dagger.GitModuleCommitAndPushOpts{
				BranchName: gitBranch,
				GitRepoURL: "https://github.com/disaster37/elasticsearch.git",
				Message:    "Commit from CI",
			},
		); err != nil {
			return nil, errors.Wrap(err, "Error when commit and push files change")
		}
	}

	return dir, nil
}

// Lint lints the Go code.
func (h *Elasticsearch) Lint(ctx context.Context) (string, error) {
	return h.GolangModule.Lint(ctx)
}

// Format formats the Go code.
func (h *Elasticsearch) Format(ctx context.Context) *dagger.Directory {
	return h.GolangModule.Format()
}

// Vet runs go vet.
func (h *Elasticsearch) Vet(ctx context.Context) (string, error) {
	return h.GolangModule.Container().WithExec([]string{"go", "vet", "./..."}).Stdout(ctx)
}

// SpecCheck clones the ES spec and runs the speccheck audit.
func (h *Elasticsearch) SpecCheck(ctx context.Context) (string, error) {
	return h.GolangModule.Container().
		WithExec([]string{"sh", "-c", "git clone --depth 1 --branch v" + ElasticsearchVersion + " https://github.com/elastic/elasticsearch /tmp/es-spec"}).
		WithExec([]string{"go", "run", "./internal/speccheck", "-spec", "/tmp/es-spec/rest-api-spec/src/main/resources/rest-api-spec/api"}).
		Stdout(ctx)
}

// Test runs the unit tests (no integration tag) and returns coverage.out.
func (h *Elasticsearch) Test(
	ctx context.Context,
	// run select tests only, defined using a regex
	// +optional
	run string,
) *dagger.File {
	expectedRunTest := ""
	if run != "" {
		expectedRunTest = fmt.Sprintf("-run %s", run)
	}
	return h.GolangModule.Container().
		WithExec(helper.ForgeScript(`
go install gotest.tools/gotestsum@latest
gotestsum --format testname -- -covermode=atomic -coverprofile coverage.out ./... %s
		`, expectedRunTest)).
		File("coverage.out")
}

// TestAcceptance runs the integration tests against a containerized ES 9 and
// returns coverage.out.
func (h *Elasticsearch) TestAcceptance(
	ctx context.Context,
	// run select tests only, defined using a regex
	// +optional
	run string,
) *dagger.File {
	expectedRunTest := ""
	if run != "" {
		expectedRunTest = fmt.Sprintf("-run %s", run)
	}
	esService := h.Elasticsearch(ctx)
	return h.GolangModule.Container().
		WithServiceBinding("elasticsearch.svc", esService).
		WithEnvVariable("ELASTICSEARCH_URL", "https://elasticsearch.svc:9200").
		WithEnvVariable("ELASTICSEARCH_USERNAME", username).
		WithSecretVariable("ELASTICSEARCH_PASSWORD", dag.SetSecret("es-password", password)).
		WithExec(helper.ForgeScript(`
curl --fail -k -u elastic:`+password+` "https://elasticsearch.svc:9200/_cluster/health?wait_for_status=yellow&timeout=500s"
curl --fail -XPUT -k -u elastic:`+password+` -H 'Content-Type: application/json' "https://elasticsearch.svc:9200/_index_template/ci-base" -d '{"index_patterns":["test-*"],"priority":500,"template":{"settings":{"number_of_shards":1,"number_of_replicas":0}}}'
go install gotest.tools/gotestsum@latest
gotestsum --format testname -- -tags integration -covermode=atomic -coverprofile coverage.out ./... %s
		`, expectedRunTest)).
		File("coverage.out")
}

// DebugTest runs the integration tests under delve for interactive debugging.
func (h *Elasticsearch) DebugTest(
	ctx context.Context,
	// run select tests only, defined using a regex
	// +optional
	run string,
) *dagger.Service {
	expectedRunTest := ""
	if run != "" {
		expectedRunTest = fmt.Sprintf("-- -test.run %s", run)
	}
	esService := h.Elasticsearch(ctx)
	defer esService.Stop(ctx)
	return h.GolangModule.Container().
		WithServiceBinding("elasticsearch.svc", esService).
		WithExposedPort(4000).
		WithExec(helper.ForgeScript(`
curl --fail -k -u elastic:`+password+` "https://elasticsearch.svc:9200/_cluster/health?wait_for_status=yellow&timeout=500s"
go install github.com/go-delve/delve/cmd/dlv@latest
dlv test --listen=:4000 --log=true --headless=true --accept-multiclient --api-version=2 %s
		`, expectedRunTest)).
		AsService()
}

// Build builds the project.
func (h *Elasticsearch) Build(ctx context.Context) *dagger.Directory {
	return h.GolangModule.Build()
}

// Elasticsearch returns the containerized Elasticsearch 9 service with security
// enabled (HTTPS, fixed ELASTIC_PASSWORD, trial license, path.repo for
// snapshot tests).
func (h *Elasticsearch) Elasticsearch(ctx context.Context) *dagger.Service {
	svc := dag.Container().
		From(fmt.Sprintf("docker.elastic.co/elasticsearch/elasticsearch:%s", ElasticsearchVersion)).
		WithUser("root").
		WithExec([]string{"mkdir", "-p", "/usr/share/elasticsearch/backup"}).
		WithExec([]string{"chown", "-R", "elasticsearch:root", "/usr/share/elasticsearch/backup"}).
		WithUser("elasticsearch").
		WithEnvVariable("cluster.name", "elasticsearch-test").
		WithEnvVariable("node.name", "es-node1").
		WithEnvVariable("discovery.type", "single-node").
		WithEnvVariable("network.publish_host", "0.0.0.0").
		WithEnvVariable("bootstrap.memory_lock", "false").
		WithEnvVariable("xpack.security.enabled", "true").
		WithEnvVariable("ELASTIC_PASSWORD", password).
		WithEnvVariable("xpack.license.self_generated.type", "trial").
		WithEnvVariable("path.repo", "/usr/share/elasticsearch/backup").
		WithEnvVariable("ES_JAVA_OPTS", "-Xms1g -Xmx1g").
		WithEnvVariable("CLEAN_CACHE", time.Now().String()).
		WithExposedPort(9200).
		AsService()

	// Health wait via a curl sidecar.
	if _, err := dag.Container().
		From("alpine/curl").
		WithEntrypoint([]string{"sh", "-c"}).
		WithServiceBinding("elasticsearch.svc", svc).
		WithEnvVariable("CLEAN_CACHE", time.Now().String()).
		WithExec(helper.ForgeScript(`
sleep 10
curl --fail -k -u elastic:` + password + ` "https://elasticsearch.svc:9200/_cluster/health?wait_for_status=yellow&timeout=500s"
		`)).
		WithExec([]string{"sh", "-c", "echo done"}).
		Stdout(ctx); err != nil {
		panic(err)
	}

	return svc
}

// CodeCov uploads coverage.out to CodeCov.
func (h *Elasticsearch) CodeCov(
	ctx context.Context,
	// Optional directory
	// +optional
	src *dagger.Directory,
	// The Codecov token
	// +required
	token *dagger.Secret,
) (stdout string, err error) {
	if src == nil {
		src = h.Src
	}
	return dag.Codecov().Upload(
		ctx,
		src,
		token,
		dagger.CodecovUploadOpts{
			Files: []string{"coverage.out"},
		},
	)
}
