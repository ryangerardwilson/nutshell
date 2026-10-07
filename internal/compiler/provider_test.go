package compiler

import (
	"io"
	"reflect"
	"testing"
)

func TestCustomPromptTransports(t *testing.T) {
	for _, tc := range []struct{ mode, placeholder, want string }{
		{"file", "{prompt_file}", "/tmp/a path/prompt.txt"},
		{"argument", "{prompt}", "Say `hello` and $(do nothing).\nNext line."},
	} {
		a := Adapter{Command: []string{"tool", "generate", "{unsafe_args}", "--prompt", tc.placeholder}, UnsafeArgs: []string{"--unsafe", "--no-approval"}, Prompt: tc.mode}
		if err := a.validate(); err != nil {
			t.Fatal(err)
		}
		argv, stdin := a.invocation("Say `hello` and $(do nothing).\nNext line.", "/tmp/a path/prompt.txt")
		want := []string{"tool", "generate", "--unsafe", "--no-approval", "--prompt", tc.want}
		if !reflect.DeepEqual(argv, want) || stdin != nil {
			t.Fatalf("%q != %q", argv, want)
		}
	}
	a := Adapter{Command: []string{"tool"}, UnsafeArgs: []string{"--unsafe"}, Prompt: "stdin"}
	_, stdin := a.invocation("input", "unused")
	got, _ := io.ReadAll(stdin)
	if string(got) != "input" {
		t.Fatal(string(got))
	}
	for _, a := range []Adapter{{}, {Command: []string{"./tool"}, Prompt: "stdin"}, {Command: []string{"tool"}, UnsafeArgs: []string{"--unsafe"}, Prompt: "file"}} {
		if err := a.validate(); err == nil {
			t.Fatalf("accepted %+v", a)
		}
	}
}
