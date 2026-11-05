//go:build java || all
// +build java all

package examples

import "testing"

// The Java program resolves the pulumiservice SDK from the local Maven
// repository, where `make install_java_sdk` publishes it.
func TestJavaSmokeExample(t *testing.T) {
	runSmokeTest(t, "java")
}
