package install

import "testing"

func TestParseBrewSearch(t *testing.T) {
	raw := "==> Formulae\nripgrep\nripgrep-all\n\n"
	pkgs := parseBrewSearch(raw, "brew")
	if len(pkgs) != 2 || pkgs[0].ID != "ripgrep" || pkgs[0].Kind != "brew" {
		t.Errorf("unexpected: %+v", pkgs)
	}
}

func TestParseWingetSearch(t *testing.T) {
	raw := "Name                     Id                          Version   Source\n" +
		"-----------------------------------------------------------------\n" +
		"Visual Studio Code       Microsoft.VisualStudioCode  1.90.0    winget\n" +
		"VS Code Insiders         Microsoft.VSCode.Insiders   1.91.0    winget\n"
	pkgs := parseWingetSearch(raw)
	if len(pkgs) != 2 {
		t.Fatalf("got %d results: %+v", len(pkgs), pkgs)
	}
	if pkgs[0].ID != "Microsoft.VisualStudioCode" || pkgs[0].Name != "Visual Studio Code" || pkgs[0].Kind != "winget" {
		t.Errorf("unexpected first result: %+v", pkgs[0])
	}
}

func TestParseScoopSearch(t *testing.T) {
	raw := "Results from local buckets...\n\nName     Version  Source Binaries\n----     -------  ------ --------\nripgrep  14.1.0   main\n"
	pkgs := parseScoopSearch(raw)
	if len(pkgs) != 1 || pkgs[0].ID != "ripgrep" || pkgs[0].Kind != "scoop" {
		t.Errorf("unexpected: %+v", pkgs)
	}
}

func TestParseChocoSearch(t *testing.T) {
	raw := "ripgrep|14.1.0\r\nrg-fzf|1.0.0\r\n"
	pkgs := parseChocoSearch(raw)
	if len(pkgs) != 2 || pkgs[0].ID != "ripgrep" || pkgs[0].Kind != "choco" {
		t.Errorf("unexpected: %+v", pkgs)
	}
}
