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
	// elasticsearchURL is the HTTPS endpoint exposed by the Elasticsearch
	// service. Elasticsearch generates its own self-signed certificates on
	// first start, so consumers skip certificate verification.
	elasticsearchURL string = "https://elasticsearch.svc:9200"
	username         string = "elastic"
	password         string = "vLPeJYa8.3RqtZCcAK6jNz"
	gitUsername      string = "ci"
	gitEmail         string = "ci@localhost"
	defaultGitBranch string = "release-branch.v9"
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
	acceptanceFile, err := h.TestAcceptance(ctx, "")
	if err != nil {
		return nil, err
	}
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
// returns coverage.out. It waits for authenticated HTTPS health, applies the
// shared index template and then runs the integration-tagged tests. Any
// readiness, template, installation or test failure is surfaced here instead of
// returning a partial coverage artifact.
func (h *Elasticsearch) TestAcceptance(
	ctx context.Context,
	// run select tests only, defined using a regex
	// +optional
	run string,
) (*dagger.File, error) {
	expectedRunTest := ""
	if run != "" {
		expectedRunTest = fmt.Sprintf("-run %s", run)
	}
	esService := h.Elasticsearch(ctx)
	container := h.GolangModule.Container().
		WithServiceBinding("elasticsearch.svc", esService).
		WithEnvVariable("ELASTICSEARCH_URL", elasticsearchURL).
		WithEnvVariable("ELASTICSEARCH_USERNAME", username).
		WithSecretVariable("ELASTICSEARCH_PASSWORD", dag.SetSecret("es-password", password)).
		WithExec(helper.ForgeScript("%s", readinessScript()+acceptanceScript(expectedRunTest)))

	synced, err := container.Sync(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "acceptance tests failed against %s", elasticsearchURL)
	}
	return synced.File("coverage.out"), nil
}

// acceptanceScript validates the shared index template and runs the
// integration-tagged tests. It is appended to the shared readiness script so
// the service is healthy before the template is applied.
func acceptanceScript(testArgs string) string {
	return `
template_body=$(mktemp)
template_err=$(mktemp)
# Replaces the readiness stage's EXIT trap; include its temp files so the
# combined readiness+acceptance script still cleans them up on exit.
trap 'rm -f "$readiness_body" "$readiness_err" "$template_body" "$template_err"' EXIT
template_rc=0
template_status=$(curl --silent --show-error --insecure --connect-timeout 2 \
	--max-time 30 \
	--user "$ELASTICSEARCH_USERNAME:$ELASTICSEARCH_PASSWORD" \
	-X PUT -H 'Content-Type: application/json' \
	--output "$template_body" --write-out '%{http_code}' \
	"$ELASTICSEARCH_URL/_index_template/ci-base" \
	-d '{"index_patterns":["test-*"],"priority":500,"template":{"settings":{"number_of_shards":1,"number_of_replicas":0}}}') || template_rc=$?
if [ "$template_rc" -ne 0 ] || [ "$template_status" != "200" ] || \
	! grep -Eq '"acknowledged"[[:space:]]*:[[:space:]]*true' "$template_body"; then
	printf 'Elasticsearch index template PUT failed (curl_rc=%s http=%s)\n' "$template_rc" "$template_status" >&2
	bounded_dump "curl error" "$template_err"
	bounded_dump "response body" "$template_body"
	exit 1
fi
printf 'Elasticsearch index template ci-base acknowledged\n'

go install gotest.tools/gotestsum@latest
gotestsum --format testname -- -tags integration -timeout=10m -covermode=atomic -coverprofile coverage.out ./... ` + testArgs + `
`
}

// readinessScript returns a POSIX shell fragment that waits for Elasticsearch
// to answer authenticated HTTPS cluster-health requests. It is reused by both
// the acceptance and debug consumers so they share the same bounded wait.
func readinessScript() string {
	return `set -eu

bounded_dump() {
	if [ -s "$2" ]; then
		printf '  %s: ' "$1" >&2
		head -c 8192 "$2" >&2
		printf '\n' >&2
	fi
}

readiness_fail() {
	printf 'Elasticsearch HTTPS readiness failed: %s\n' "$1" >&2
	printf '  expected URL: %s\n' "$ELASTICSEARCH_URL" >&2
	bounded_dump "curl error" "$readiness_err"
	bounded_dump "response body" "$readiness_body"
	exit 1
}

wait_for_elasticsearch() {
	readiness_deadline=$(( $(date +%s) + 180 ))
	readiness_body=$(mktemp)
	readiness_err=$(mktemp)
	trap 'rm -f "$readiness_body" "$readiness_err"' EXIT
	readiness_attempt=0
	readiness_rc=0
	readiness_status=000
	readiness_hinted=0
	while :; do
		readiness_now=$(date +%s)
		readiness_remaining=$(( readiness_deadline - readiness_now ))
		if [ "$readiness_remaining" -le 0 ]; then
			break
		fi
		if [ "$readiness_remaining" -lt 5 ]; then
			readiness_timeout=$readiness_remaining
		else
			readiness_timeout=5
		fi
		readiness_attempt=$(( readiness_attempt + 1 ))
		: > "$readiness_body"
		: > "$readiness_err"
		readiness_rc=0
		readiness_status=$(curl --silent --show-error --insecure --connect-timeout 2 \
			--max-time "$readiness_timeout" \
			--user "$ELASTICSEARCH_USERNAME:$ELASTICSEARCH_PASSWORD" \
			--output "$readiness_body" --write-out '%{http_code}' \
			"$ELASTICSEARCH_URL/_cluster/health?wait_for_status=yellow&timeout=2s" 2>"$readiness_err") || readiness_rc=$?
		readiness_elapsed=$(( $(date +%s) - (readiness_deadline - 180) ))
		printf 'elasticsearch readiness: attempt=%s elapsed=%ss curl_rc=%s http=%s\n' "$readiness_attempt" "$readiness_elapsed" "$readiness_rc" "$readiness_status"

		if [ "$readiness_rc" -eq 0 ]; then
			case "$readiness_status" in
			200)
				if grep -Eq '"status"[[:space:]]*:[[:space:]]*"(yellow|green)"' "$readiness_body" && \
					grep -Eq '"timed_out"[[:space:]]*:[[:space:]]*false' "$readiness_body"; then
					printf 'elasticsearch readiness: ready (HTTP 200, status yellow/green, timed_out false)\n'
					return 0
				fi
				if ! grep -Eq '"status"[[:space:]]*:[[:space:]]*"' "$readiness_body" || \
					! grep -Eq '"timed_out"[[:space:]]*:' "$readiness_body"; then
					readiness_fail "malformed health response (missing status/timed_out)"
				fi
				;;
			401|403)
				readiness_fail "authentication rejected (HTTP $readiness_status); check ELASTICSEARCH_USERNAME and ELASTICSEARCH_PASSWORD"
				;;
			408|429|5*)
				;;
			3*|4*)
				readiness_fail "unexpected HTTP $readiness_status"
				;;
			esac
		else
			case "$readiness_rc" in
			6|7|16|18|28|35|52|55|56)
				if [ "$readiness_rc" -eq 35 ] || [ "$readiness_rc" -eq 52 ]; then
					if [ "$readiness_hinted" -eq 0 ]; then
						printf 'elasticsearch readiness: hint: HTTPS request failed (curl rc=%s); the listener may be serving plain HTTP because automatic security/TLS configuration was skipped. Verify ELASTICSEARCH_URL and Elasticsearch auto-configuration. Refusing to fall back to HTTP.\n' "$readiness_rc" >&2
						readiness_hinted=1
					fi
				fi
				;;
			*)
				readiness_fail "non-transient curl error (rc=$readiness_rc)"
				;;
			esac
		fi

		readiness_now=$(date +%s)
		readiness_remaining=$(( readiness_deadline - readiness_now ))
		if [ "$readiness_remaining" -le 0 ]; then
			break
		fi
		if [ "$readiness_remaining" -lt 2 ]; then
			sleep "$readiness_remaining"
		else
			sleep 2
		fi
	done

	printf 'Elasticsearch HTTPS readiness failed after 180s\n' >&2
	printf '  expected URL: %s\n' "$ELASTICSEARCH_URL" >&2
	printf '  last curl rc=%s http=%s\n' "$readiness_rc" "$readiness_status" >&2
	bounded_dump "last curl error" "$readiness_err"
	bounded_dump "last response body" "$readiness_body"
	exit 1
}

wait_for_elasticsearch
`
}

// DebugTest runs the integration tests under delve for interactive debugging.
// The integration build tag is forwarded through --build-flags because delve
// has no direct -tags flag; --build-flags="-tags=integration" makes dlv pass
// -tags=integration to the underlying go build/test.
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
	return h.GolangModule.Container().
		WithServiceBinding("elasticsearch.svc", esService).
		WithEnvVariable("ELASTICSEARCH_URL", elasticsearchURL).
		WithEnvVariable("ELASTICSEARCH_USERNAME", username).
		WithSecretVariable("ELASTICSEARCH_PASSWORD", dag.SetSecret("es-password", password)).
		WithExposedPort(4000).
		WithExec(helper.ForgeScript("%s", readinessScript()+`
go install github.com/go-delve/delve/cmd/dlv@latest
dlv test --listen=:4000 --log=true --headless=true --accept-multiclient --api-version=2 --build-flags="-tags=integration" `+expectedRunTest)).
		AsService()
}

// Build builds the project.
func (h *Elasticsearch) Build(ctx context.Context) *dagger.Directory {
	return h.GolangModule.Build()
}

// Elasticsearch returns the containerized Elasticsearch 9 service. HTTPS,
// basic authentication and self-signed certificates are configured
// automatically by Elasticsearch on first start; explicitly overriding
// security or SSL settings can disable that automatic configuration.
// Certificates, configuration and data are ephemeral per service container, so
// do not persist them or rely on certificates from a previous run.
func (h *Elasticsearch) Elasticsearch(ctx context.Context) *dagger.Service {
	return dag.Container().
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
		WithEnvVariable("ELASTIC_PASSWORD", password).
		WithEnvVariable("xpack.license.self_generated.type", "trial").
		WithEnvVariable("path.repo", "/usr/share/elasticsearch/backup").
		WithEnvVariable("ES_JAVA_OPTS", "-Xms1g -Xmx1g").
		WithEnvVariable("CLEAN_CACHE", time.Now().String()).
		WithExposedPort(9200).
		AsService()
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
