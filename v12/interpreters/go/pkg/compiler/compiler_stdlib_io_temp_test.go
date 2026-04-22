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

func TestCompilerStdlibFsReadLinesAfterWriteExecutes(t *testing.T) {
	t.Setenv("ABLE_STDLIB_ROOT", "/home/david/sync/projects/able-stdlib/src")
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-stdlib-fs-read-lines-", strings.Join([]string{
		"package demo",
		"",
		"import able.fs",
		"import able.io",
		"import able.io.path",
		"import able.io.temp",
		"",
		"fn main() -> void {",
		"  root := temp.dir(\"ablec-fs-read-lines-\").path",
		"  file := path.parse(root).join(\"data.txt\").to_string()",
		"  do {",
		"    fs.mkdir(root, true)",
		"    writer := fs.open(file, fs.write_only(true, true), nil)",
		"    io.write_all(writer, io.string_to_bytes(\"alpha\\r\\nbeta\\r\\n\"))",
		"    io.close(writer)",
		"    lines := fs.read_lines(file)",
		"    print(`${lines.len()} ${lines.get(0)!} ${lines.get(1)!}`)",
		"  } ensure {",
		"    fs.remove(root, true)",
		"  }",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "2 alpha beta" {
		t.Fatalf("expected stdlib fs read_lines output '2 alpha beta', got %q", stdout)
	}
}

func TestCompilerStdlibFsReadLinesAfterRewriteExecutes(t *testing.T) {
	t.Setenv("ABLE_STDLIB_ROOT", "/home/david/sync/projects/able-stdlib/src")
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-stdlib-fs-read-lines-rewrite-", strings.Join([]string{
		"package demo",
		"",
		"import able.fs",
		"import able.io",
		"import able.io.path",
		"import able.io.temp",
		"",
		"fn main() -> void {",
		"  root := temp.dir(\"ablec-fs-read-lines-rewrite-\").path",
		"  file := path.parse(root).join(\"data.txt\").to_string()",
		"  do {",
		"    fs.mkdir(root, true)",
		"    writer := fs.open(file, fs.write_only(true, true), nil)",
		"    io.write_all(writer, io.string_to_bytes(\"alpha\\r\\nbeta\\r\\n\"))",
		"    io.close(writer)",
		"    first := fs.read_lines(file)",
		"    writer = fs.open(file, fs.write_only(true, true), nil)",
		"    io.write_all(writer, io.string_to_bytes(\"gamma\\r\\ndelta\\r\\nepsilon\\r\\n\"))",
		"    io.close(writer)",
		"    second := fs.read_lines(file)",
		"    print(`${first.len()} ${first.get(0)!} ${second.len()} ${second.get(0)!} ${second.get(2)!}`)",
		"  } ensure {",
		"    fs.remove(root, true)",
		"  }",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "2 alpha 3 gamma epsilon" {
		t.Fatalf("expected stdlib fs read_lines rewrite output '2 alpha 3 gamma epsilon', got %q", stdout)
	}
}
