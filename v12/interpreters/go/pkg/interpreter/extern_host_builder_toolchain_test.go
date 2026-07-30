//go:build !(js && wasm)

package interpreter

import (
	"reflect"
	"testing"
)

func TestExternPluginBuildEnvironmentUsesRunningToolchain(t *testing.T) {
	environ := []string{
		"PATH=/usr/bin",
		"GOTOOLCHAIN=local",
		"HOME=/example",
		"GOTOOLCHAIN=go1.25.0",
	}
	got := externPluginBuildEnvironment(environ, "go1.26.5")
	want := []string{
		"PATH=/usr/bin",
		"HOME=/example",
		"GOTOOLCHAIN=go1.26.5",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extern plugin build environment = %#v, want %#v", got, want)
	}
}

func TestExternPluginBuildEnvironmentPreservesDevelopmentToolchain(t *testing.T) {
	environ := []string{"PATH=/usr/bin", "GOTOOLCHAIN=local"}
	got := externPluginBuildEnvironment(environ, "devel go1.27-deadbeef")
	if !reflect.DeepEqual(got, environ) {
		t.Fatalf("development build environment = %#v, want %#v", got, environ)
	}
}
