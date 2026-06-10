package interpreter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStdlibResultAcceptsImportedErrorImplementation(t *testing.T) {
	root := repositoryRoot()
	stdlibRoot := filepath.Join(filepath.Dir(root), "able-stdlib", "src")
	if _, err := os.Stat(filepath.Join(stdlibRoot, "package.yml")); err != nil {
		t.Fatalf("canonical stdlib source: %v", err)
	}
	t.Setenv("ABLE_STDLIB_ROOT", stdlibRoot)

	const source = `
package demo

import able.core.errors.{IndexError}

fn failure() -> !nil {
  IndexError { index: -1_i64, length: 0_i64 }
}

fn propagate() -> !i32 {
  failure()!
  1
}

fn main() -> void {
  propagate() or { err =>
    err match {
      case _: IndexError => {},
      case _ => { raise "expected IndexError" }
    }
  }
}
`

	for _, mode := range []testExecMode{testExecTreewalker, testExecBytecode} {
		t.Run(string(mode), func(t *testing.T) {
			stdout, err := runStdlibExecSource(t, source, mode)
			if err != nil {
				t.Fatalf("execute Result propagation: %v", err)
			}
			if len(stdout) != 0 {
				t.Fatalf("unexpected stdout: %v", stdout)
			}
		})
	}
}

func TestStdlibRaiseUsesImportedErrorMessage(t *testing.T) {
	root := repositoryRoot()
	stdlibRoot := filepath.Join(filepath.Dir(root), "able-stdlib", "src")
	if _, err := os.Stat(filepath.Join(stdlibRoot, "package.yml")); err != nil {
		t.Fatalf("canonical stdlib source: %v", err)
	}
	t.Setenv("ABLE_STDLIB_ROOT", stdlibRoot)

	const source = `
package demo

import able.core.interfaces.{Error}

struct SampleError { label: String }

impl Error for SampleError {
  fn message(self: Self) -> String { self.label }
  fn cause(self: Self) -> ?Error { nil }
}

fn fail() -> void {
  raise SampleError { label: "boom" }
}

fn error_message(err: Error) -> String {
  err.message()
}

fn main() -> void {
  if error_message(SampleError { label: "direct" }) != "direct" {
    raise "wrong direct Error coercion"
  }
  casted := SampleError { label: "cast" } as Error
  if casted.message() != "cast" { raise "wrong explicit Error cast" }
  do {
    fail()
  } rescue {
    case err: Error => {
      if err.message() != "boom" { raise "wrong error message" }
    }
  }
}
`

	for _, mode := range []testExecMode{testExecTreewalker, testExecBytecode} {
		t.Run(string(mode), func(t *testing.T) {
			stdout, err := runStdlibExecSource(t, source, mode)
			if err != nil {
				t.Fatalf("execute imported Error raise: %v", err)
			}
			if len(stdout) != 0 {
				t.Fatalf("unexpected stdout: %v", stdout)
			}
		})
	}
}
