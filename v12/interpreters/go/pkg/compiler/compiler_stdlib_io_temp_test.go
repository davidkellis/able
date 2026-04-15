package compiler

import (
	"strings"
	"testing"
)

func stdlibFsReadTextSource() string {
	return strings.Join([]string{
		"package demo",
		"",
		"import able.fs",
		"import able.io",
		"import able.io.path",
		"import able.io.temp",
		"",
		"fn main() -> void {",
		"  root := temp.dir(\"ablec-fs-read-text-\").path",
		"  file := path.parse(root).join(\"data.txt\").to_string()",
		"  do {",
		"    fs.mkdir(root, true)",
		"    writer := fs.open(file, fs.write_only(true, true), nil)",
		"    io.write_all(writer, io.string_to_bytes(\"alpha\"))",
		"    io.close(writer)",
		"    print(fs.read_text(file))",
		"  } ensure {",
		"    fs.remove(root, true)",
		"  }",
		"}",
		"",
	}, "\n")
}

func TestCompilerStdlibIoWriteAllTempExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-stdlib-io-write-all-", strings.Join([]string{
		"package demo",
		"",
		"import able.io",
		"import able.io.temp",
		"",
		"fn main() -> void {",
		"  tmp := temp.file(\"ablec-io-write-all-\")",
		"  do {",
		"    io.write_all(tmp.handle, io.string_to_bytes(\"alpha\"))",
		"    io.close(tmp.handle)",
		"    print(\"ok\")",
		"  } ensure {",
		"    temp.cleanup(tmp)",
		"  }",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "ok" {
		t.Fatalf("expected stdlib io write_all temp output ok, got %q", stdout)
	}
}

func TestCompilerStdlibFsReadTextAfterWriteExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-stdlib-fs-read-text-", stdlibFsReadTextSource(), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "alpha" {
		t.Fatalf("expected stdlib fs read_text output alpha, got %q", stdout)
	}
}
