package update

// Tests for extractBinary — the pure archive-extraction half of the update
// pipeline. It was previously the only function in update.go with no direct
// coverage. All fixtures are built in-memory (no network, no temp files).

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
)

// zipFixture builds an in-memory zip archive containing name -> content.
func zipFixture(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// targzFixture builds an in-memory tar.gz archive containing name -> content.
func targzFixture(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var gzBuf bytes.Buffer
	gw := gzip.NewWriter(&gzBuf)
	tw := tar.NewWriter(gw)
	for name, content := range entries {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o755,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return gzBuf.Bytes()
}

func TestExtractBinary_Zip(t *testing.T) {
	archive := zipFixture(t, map[string]string{"app.exe": "MZ-BINARY"})
	got, err := extractBinary(archive, "zip", "app")
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if string(got) != "MZ-BINARY" {
		t.Errorf("got %q, want the exe content", got)
	}
}

func TestExtractBinary_ZipNestedPath(t *testing.T) {
	// Windows archives often nest the binary in a versioned dir; the lookup
	// must match on the basename.
	archive := zipFixture(t, map[string]string{"dist/v1.2/app.exe": "NESTED"})
	got, err := extractBinary(archive, "zip", "app")
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if string(got) != "NESTED" {
		t.Errorf("got %q, want the nested exe content", got)
	}
}

func TestExtractBinary_ZipMissing(t *testing.T) {
	archive := zipFixture(t, map[string]string{"other.txt": "x"})
	_, err := extractBinary(archive, "zip", "app")
	if err == nil || !strings.Contains(err.Error(), "not found in update archive") {
		t.Fatalf("err = %v, want not-found", err)
	}
}

func TestExtractBinary_ZipInvalid(t *testing.T) {
	_, err := extractBinary([]byte("not-a-zip"), "zip", "app")
	if err == nil || !strings.Contains(err.Error(), "reading update archive") {
		t.Fatalf("err = %v, want reading-update-archive", err)
	}
}

func TestExtractBinary_TarGz(t *testing.T) {
	archive := targzFixture(t, map[string]string{"bluefin-cli": "ELF-BINARY"})
	got, err := extractBinary(archive, "tar.gz", "bluefin-cli")
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if string(got) != "ELF-BINARY" {
		t.Errorf("got %q, want the binary content", got)
	}
}

func TestExtractBinary_TarGzNestedPath(t *testing.T) {
	archive := targzFixture(t, map[string]string{"bluefin-cli/bin/bluefin-cli": "NESTED-ELF"})
	got, err := extractBinary(archive, "tar.gz", "bluefin-cli")
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if string(got) != "NESTED-ELF" {
		t.Errorf("got %q, want the nested binary content", got)
	}
}

func TestExtractBinary_TarGzMissing(t *testing.T) {
	archive := targzFixture(t, map[string]string{"README.md": "hi"})
	_, err := extractBinary(archive, "tar.gz", "bluefin-cli")
	if err == nil || !strings.Contains(err.Error(), "not found in update archive") {
		t.Fatalf("err = %v, want not-found", err)
	}
}

func TestExtractBinary_TarGzInvalid(t *testing.T) {
	_, err := extractBinary([]byte("definitely-not-gzip"), "tar.gz", "bluefin-cli")
	if err == nil || !strings.Contains(err.Error(), "reading update archive") {
		t.Fatalf("err = %v, want reading-update-archive", err)
	}
}

func TestExtractBinary_TarSkipsDirectoryEntries(t *testing.T) {
	// A directory entry whose basename collides with the target must not be
	// treated as the binary (TypeReg filter): only the real file wins.
	var gzBuf bytes.Buffer
	gw := gzip.NewWriter(&gzBuf)
	tw := tar.NewWriter(gw)
	dir := &tar.Header{Name: "bluefin-cli", Typeflag: tar.TypeDir, Mode: 0o755}
	if err := tw.WriteHeader(dir); err != nil {
		t.Fatal(err)
	}
	bin := &tar.Header{Name: "bin/bluefin-cli", Mode: 0o755, Size: 2}
	if err := tw.WriteHeader(bin); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("OK")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := extractBinary(gzBuf.Bytes(), "tar.gz", "bluefin-cli")
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if string(got) != "OK" {
		t.Errorf("got %q, want the real file (directory entry skipped)", got)
	}
}
