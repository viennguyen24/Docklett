/*
Maps each Docker instruction keyword to its corresponding LLB state mutation.
Switches on DockerStatement.Keyword.Lexeme and applies the operation to the
translator's current LLB state.

All LLB state operations follow the immutability rule:

	llb.State is a value type — methods return NEW states, never mutate in place.

This mirrors the dispatch pattern from BuildKit's dockerfile2llb/convert.go.
*/
package translator

import (
	"docklett/compiler/ast"
	"fmt"
	"strings"
)

// translateDocker routes a DockerStatement to its keyword-specific LLB handler.
// Variable interpolation is applied to args before dispatch.
func (t *Translator) translateDocker(stmt *ast.DockerStatement) error {
	keyword := strings.ToUpper(stmt.Keyword.Lexeme)
	args := t.interpolateVariables(stmt.Args)

	switch keyword {
	case "FROM":
		return t.translateFrom(args)
	case "RUN":
		return t.translateRun(args)
	case "WORKDIR":
		return t.translateWorkdir(args)
	case "ENV":
		return t.translateEnv(args)
	case "COPY":
		return t.translateCopy(args)
	case "ADD":
		return t.translateAdd(args)

	// image config metadata — stored for image manifest, no LLB state mutation
	case "EXPOSE", "CMD", "ENTRYPOINT", "LABEL", "USER",
		"VOLUME", "SHELL", "STOPSIGNAL", "ARG",
		"HEALTHCHECK", "MAINTAINER", "ONBUILD":
		return nil

	default:
		return fmt.Errorf("[line %d] unknown Docker instruction: %s", stmt.Keyword.Line, keyword)
	}
}

// interpolateVariables replaces ${name} references in args with compile-time variable values.
// Unresolved variables are left as-is for runtime resolution by the container engine.
func (t *Translator) interpolateVariables(args string) string {
	result := args
	for name, val := range t.env.Bindings {
		placeholder := "${" + name + "}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", val))
	}
	return result
}

// translateFrom sets the base image. "scratch" produces an empty state.
func (t *Translator) translateFrom(args string) error {
	// placeholder — LLB: llb.Image(args) or llb.Scratch()
	_ = args
	return nil
}

// translateRun appends a shell command execution to the current state.
func (t *Translator) translateRun(args string) error {
	// placeholder — LLB: state.Run(llb.Shlex(args)).Root()
	_ = args
	return nil
}

// translateWorkdir sets the working directory for subsequent operations.
func (t *Translator) translateWorkdir(args string) error {
	// placeholder — LLB: state.Dir(args)
	_ = args
	return nil
}

// translateEnv parses "KEY=VALUE" or "KEY VALUE" and adds an environment variable.
func (t *Translator) translateEnv(args string) error {
	// placeholder — LLB: state.AddEnv(key, value)
	_ = args
	return nil
}

// translateCopy copies files from the build context into the image.
func (t *Translator) translateCopy(args string) error {
	// placeholder — LLB: state.File(llb.Copy(buildContext, src, dst))
	_ = args
	return nil
}

// translateAdd copies files with optional URL/tarball extraction support.
func (t *Translator) translateAdd(args string) error {
	// placeholder — LLB: state.File(llb.Copy(buildContext, src, dst))
	_ = args
	return nil
}
