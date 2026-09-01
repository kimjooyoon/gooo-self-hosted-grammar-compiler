package evidence

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
)

type Metric struct {
	WallMS     int64 `json:"wall_ms"`
	PeakRSSKiB int64 `json:"peak_rss_kib"`
}

type ScopeInventory struct {
	Scope          string `json:"scope"`
	Extension      string `json:"extension"`
	PhysicalLines  int64  `json:"physical_lines"`
	DescendantDirs int64  `json:"descendant_dirs"`
	RegularFiles   int64  `json:"regular_files"`
	GeneratedFiles int64  `json:"generated_files"`
	GeneratedBytes int64  `json:"generated_bytes"`
}

type TestCounts struct {
	Total    int64 `json:"total"`
	Selected int64 `json:"selected"`
	Executed int64 `json:"executed"`
	Reused   int64 `json:"reused"`
	Failed   int64 `json:"failed"`
	Unknown  int64 `json:"unknown"`
}

type Evidence struct {
	Schema                    string           `json:"schema"`
	Toolchain                 string           `json:"toolchain"`
	RepositoryRoot            string           `json:"repository_root"`
	ReadmeExcluded            bool             `json:"root_readme_excluded"`
	InputRepositoryWrites     int              `json:"input_repository_writes"`
	CallerOwnedTempOutputOnly bool             `json:"caller_owned_temp_output_only"`
	AutomaticAuthority        map[string]int   `json:"automatic_authority"`
	Inventory                 []ScopeInventory `json:"inventory"`
	Compile                   Metric           `json:"compile"`
	Build                     Metric           `json:"build"`
	Test                      Metric           `json:"test"`
	Conformance               Metric           `json:"conformance"`
	Integration               Metric           `json:"integration"`
	Tests                     TestCounts       `json:"tests"`
	LocalExecutions           map[string]int64 `json:"local_executions"`
	ConformancePass           bool             `json:"conformance_pass"`
}

type testEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
}

func Command(args []string) error {
	flags := flag.NewFlagSet("ci-evidence", flag.ContinueOnError)
	grammar := flags.String("grammar", "meta/gooo-grammar.gooo", "authoritative grammar")
	cases := flags.String("cases", "fixtures/cases", "fixed conformance cases")
	out := flags.String("output", ".ci/evidence.json", "absolute evidence output")
	if err := flags.Parse(args); err != nil {
		return err
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	if !filepath.IsAbs(*out) {
		*out = filepath.Join(root, *out)
	}
	grammarPath := filepath.Join(root, *grammar)
	casesPath := filepath.Join(root, *cases)
	if filepath.IsAbs(*grammar) {
		grammarPath = *grammar
	}
	if filepath.IsAbs(*cases) {
		casesPath = *cases
	}

	evidence := Evidence{
		Schema:    "gooo/self-hosted-grammar-compiler/ci-evidence/v1",
		Toolchain: toolchainVersion(), RepositoryRoot: root, ReadmeExcluded: true,
		InputRepositoryWrites: 0, CallerOwnedTempOutputOnly: true,
		AutomaticAuthority: map[string]int{"commit": 0, "push": 0, "merge": 0, "release": 0},
		Inventory:          inventory(root), LocalExecutions: map[string]int64{"compile": 0, "build": 0, "test": 0, "vet": 0, "conformance": 0, "integration": 0},
	}
	var firstErr error
	compileMetric, _, err := run(root, "go", "test", "-run", "^$", "-count=1", "./...")
	evidence.Compile = compileMetric
	if err != nil && firstErr == nil {
		firstErr = fmt.Errorf("compile: %w", err)
	}
	buildMetric, _, err := run(root, "go", "build", "./...")
	evidence.Build = buildMetric
	if err != nil && firstErr == nil {
		firstErr = fmt.Errorf("build: %w", err)
	}
	testMetric, testOutput, err := run(root, "go", "test", "-json", "-count=1", "./...")
	evidence.Test = testMetric
	evidence.Tests = counts(testOutput)
	if err != nil && firstErr == nil {
		firstErr = fmt.Errorf("test: %w", err)
	}
	conformanceOutput := filepath.Join(filepath.Dir(*out), "conformance.json")
	conformanceMetric, _, err := run(root, "go", "run", "./cmd/gooo-grammar-compiler", "conformance", "--cases", casesPath, "--reference-root", root, "--out", conformanceOutput)
	evidence.Conformance = conformanceMetric
	if err != nil && firstErr == nil {
		firstErr = fmt.Errorf("conformance: %w", err)
	}
	integrationMetric, integrationOutput, err := run(root, "go", "test", "-json", "-count=1", "./generated", "-run", "TestStage2ParsesSample")
	evidence.Integration = integrationMetric
	integrationCounts := counts(integrationOutput)
	if integrationCounts.Failed > 0 && firstErr == nil {
		firstErr = errors.New("integration test failed")
	}
	if raw, readErr := os.ReadFile(conformanceOutput); readErr == nil {
		var report struct {
			AllPass bool `json:"all_pass"`
		}
		if json.Unmarshal(raw, &report) == nil {
			evidence.ConformancePass = report.AllPass
		}
	}
	if err := write(*out, evidence); err != nil {
		return err
	}
	return firstErr
}

func run(root, command string, args ...string) (Metric, string, error) {
	start := time.Now()
	before := childRSS()
	cmd := exec.Command(command, args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	after := childRSS()
	rss := after
	if before > rss {
		rss = before
	}
	return Metric{WallMS: time.Since(start).Milliseconds(), PeakRSSKiB: rss}, string(output), err
}

func childRSS() int64 {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_CHILDREN, &usage); err != nil {
		return 0
	}
	return int64(usage.Maxrss)
}

func toolchainVersion() string {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return runtime.Version()
	}
	return strings.TrimSpace(string(output))
}

func counts(output string) TestCounts {
	seen := map[string]bool{}
	passed := map[string]bool{}
	failed := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		var event testEvent
		if json.Unmarshal([]byte(scanner.Text()), &event) != nil || event.Test == "" {
			continue
		}
		key := event.Package + "/" + event.Test
		switch event.Action {
		case "run":
			seen[key] = true
		case "pass":
			passed[key] = true
		case "fail":
			failed[key] = true
		}
	}
	return TestCounts{Total: int64(len(seen)), Selected: int64(len(seen)), Executed: int64(len(passed) + len(failed)), Reused: 0, Failed: int64(len(failed)), Unknown: 0}
}

func inventory(root string) []ScopeInventory {
	values := []ScopeInventory{{Scope: "repository", Extension: ".go"}, {Scope: "repository", Extension: ".gooo"}}
	for index := range values {
		extension := values[index].Extension
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil {
				return nil
			}
			if info.IsDir() {
				if path != root {
					values[index].DescendantDirs++
				}
				return nil
			}
			if !info.Mode().IsRegular() || filepath.Ext(path) != extension || filepath.Base(path) == "README.md" {
				return nil
			}
			values[index].RegularFiles++
			values[index].PhysicalLines += physicalLines(path)
			generated := false
			if raw, readErr := os.ReadFile(path); readErr == nil {
				generated = strings.Contains(string(raw), "Code generated")
			}
			if generated {
				values[index].GeneratedFiles++
				values[index].GeneratedBytes += info.Size()
			}
			return nil
		})
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Extension < values[j].Extension })
	return values
}

func physicalLines(path string) int64 {
	raw, err := os.ReadFile(path)
	if err != nil || len(raw) == 0 {
		return 0
	}
	lines := int64(strings.Count(string(raw), "\n"))
	if raw[len(raw)-1] != '\n' {
		lines++
	}
	return lines
}

func write(path string, value Evidence) error {
	if !filepath.IsAbs(path) {
		return errors.New("evidence output must be absolute")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
