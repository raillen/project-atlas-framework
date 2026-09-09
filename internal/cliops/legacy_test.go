package cliops

import "testing"

func TestParseLegacyYAML(t *testing.T) {
	value, err := parseLegacyYAML("version: 1\nproject:\n  name: Legacy\n  type: [cli]\nai:\n  orchestrator: native\n  preferred_models:\n    - id: test/model\n      provider: test\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	project, _ := value["project"].(map[string]any)
	if project["name"] != "Legacy" {
		t.Fatalf("project name: %v", value)
	}
}

func TestMigrateLegacyProject(t *testing.T) {
	root := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		if err := writeText(root+"/"+name, content); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	write(".atlas/project-profile.yaml", "version: 1\nproject:\n  name: Legacy\n  type: [cli]\nai:\n  orchestrator: native\n  autonomy: agentic\n  preferred_models:\n    - id: test/model\n      provider: test\n")
	write("PROJECT_MANIFEST.yaml", "framework: {name: project-atlas-framework, version: 0.1.0}\n")
	write(".ai/goals/P00/P00-G01.goal.yaml", "id: P00-G01\ntitle: Foundation\nphase: P00\nstate: DRAFT\nobjective: Foundation\nacceptance: [Works]\ngates: {tests: required}\ndependencies: []\nevidence: []\n")
	svc := New(repoRoot(t))
	report, err := svc.Migrate(root, false)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if report["from_version"] != 1 || report["to_version"] != 3 {
		t.Fatalf("report: %v", report)
	}
}
