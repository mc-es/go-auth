// Package main validates commit messages.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type commitType string

const (
	feat     commitType = "feat"
	fix      commitType = "fix"
	chore    commitType = "chore"
	docs     commitType = "docs"
	style    commitType = "style"
	refactor commitType = "refactor"
	perf     commitType = "perf"
	test     commitType = "test"
	build    commitType = "build"
)

type commitMessage struct {
	raw        string
	typ        commitType
	scope      string
	subject    string
	isBreaking bool
}

type config struct {
	msgMaxLen     int
	scopeMinLen   int
	scopeMaxLen   int
	subjectMinLen int
}

type ruleFunc func(msg commitMessage, lint *linter) error

type linter struct {
	allowedTypes         map[commitType]struct{}
	allowedBreakingTypes map[commitType]struct{}
	scopeRegex           *regexp.Regexp
	rules                []ruleFunc
	config               config
}

type errorKind int

const (
	errParse errorKind = iota
	errValidate
)

type linterError struct {
	kind   errorKind
	errors []error
}

func (e linterError) Error() string {
	parts := make([]string, 0, len(e.errors))
	for _, err := range e.errors {
		parts = append(parts, err.Error())
	}

	return strings.Join(parts, "; ")
}

var (
	// Parse errors.
	errEmptyMessage       = errors.New("commit message is empty")
	errMissingSeparator   = errors.New("missing colon separator and space")
	errMismatchedParens   = errors.New("mismatched parentheses in scope definition")
	errInvalidScopeFormat = errors.New("invalid scope format")
	errEmptyScopeParens   = errors.New("scope cannot be empty when parentheses are used")

	// Validation errors.
	errMsgTooLong            = errors.New("commit message is too long")
	errNonASCII              = errors.New("commit message contains non-ASCII characters")
	errTypeInvalid           = errors.New("type is invalid")
	errScopeRequired         = errors.New("scope is required")
	errScopeTooShort         = errors.New("scope is too short")
	errScopeTooLong          = errors.New("scope is too long")
	errScopeInvalidChar      = errors.New("scope contains invalid characters (lowercase, numbers, -, _)")
	errInvalidBreakingType   = errors.New("breaking changes (!) are not allowed for this commit type")
	errSubjectTooShort       = errors.New("subject is too short")
	errSubjectLeadingSpace   = errors.New("subject contains leading whitespace")
	errSubjectDoubleSpace    = errors.New("subject contains double spaces")
	errSubjectNotLowercase   = errors.New("subject must start with a lowercase letter")
	errSubjectEndsWithPeriod = errors.New("subject cannot end with a period")
)

const (
	defaultMsgMaxLen     = 100
	defaultScopeMinLen   = 2
	defaultScopeMaxLen   = 15
	defaultSubjectMinLen = 3
)

// main is the entry point.
func main() {
	if err := run(os.Args); err != nil {
		if le, ok := err.(linterError); ok {
			switch le.kind {
			case errParse:
				fmt.Fprintln(os.Stderr, "❌ Format error:")
			case errValidate:
				fmt.Fprintln(os.Stderr, "❌ Validation errors:")
			}

			for _, e := range le.errors {
				fmt.Fprintf(os.Stderr, "  - %s\n", e)
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
		}

		os.Exit(1)
	}
}

// run runs the commitlint command.
func run(args []string) error {
	if len(args) < 2 {
		return errors.New("usage: commitlint <path-to-msg-file>")
	}

	file, err := os.Open(args[1])
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	rawCommitMsg, err := readFirstLine(file)
	if err != nil {
		return err
	}

	parsedMsg, err := parseCommitMessage(rawCommitMsg)
	if err != nil {
		return linterError{
			kind:   errParse,
			errors: []error{err},
		}
	}

	return newLinter().validate(parsedMsg)
}

// readFirstLine reads the first line of the commit message.
func readFirstLine(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		return line, nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", errEmptyMessage
}

// parseCommitMessage parses the commit message.
func parseCommitMessage(rawCommitMsg string) (commitMessage, error) {
	prefix, subject, found := strings.Cut(rawCommitMsg, ": ")
	if !found {
		return commitMessage{}, errMissingSeparator
	}

	isBreaking := strings.HasSuffix(prefix, "!")
	prefix = strings.TrimSuffix(prefix, "!")

	typ, scope, err := parseTypeAndScope(prefix)
	if err != nil {
		return commitMessage{}, err
	}

	return commitMessage{
		raw:        rawCommitMsg,
		typ:        typ,
		scope:      scope,
		subject:    subject,
		isBreaking: isBreaking,
	}, nil
}

// parseTypeAndScope parses the type and scope from the commit message.
func parseTypeAndScope(prefix string) (commitType, string, error) {
	hasOpen := strings.Contains(prefix, "(")
	hasClose := strings.Contains(prefix, ")")

	if hasOpen != hasClose {
		return "", "", errMismatchedParens
	}

	if hasOpen {
		if !strings.HasSuffix(prefix, ")") {
			return "", "", errInvalidScopeFormat
		}

		p := strings.TrimSuffix(prefix, ")")
		parts := strings.SplitN(p, "(", 2)

		if parts[1] == "" {
			return "", "", errEmptyScopeParens
		}

		return commitType(parts[0]), parts[1], nil
	}

	return commitType(prefix), "", nil
}

// newLinter creates a new linter.
func newLinter() *linter {
	cfg := loadConfig()

	lint := &linter{
		config:     cfg,
		scopeRegex: regexp.MustCompile(`^[a-z0-9_-]+$`),
	}

	allTypes := []commitType{
		feat, fix, chore, docs, style,
		refactor, perf, test, build,
	}

	breakingTypes := []commitType{
		feat, build,
	}

	lint.allowedTypes = make(map[commitType]struct{})
	for _, t := range allTypes {
		lint.allowedTypes[t] = struct{}{}
	}

	lint.allowedBreakingTypes = make(map[commitType]struct{})
	for _, t := range breakingTypes {
		lint.allowedBreakingTypes[t] = struct{}{}
	}

	lint.rules = []ruleFunc{
		checkLength,
		checkASCII,
		checkType,
		checkScope,
		checkBreaking,
		checkSubject,
	}

	return lint
}

// loadConfig loads the configuration.
func loadConfig() config {
	cfg := config{
		msgMaxLen:     getEnvAsInt("COMMITLINT_MSG_MAX_LEN", defaultMsgMaxLen),
		scopeMinLen:   getEnvAsInt("COMMITLINT_SCOPE_MIN_LEN", defaultScopeMinLen),
		scopeMaxLen:   getEnvAsInt("COMMITLINT_SCOPE_MAX_LEN", defaultScopeMaxLen),
		subjectMinLen: getEnvAsInt("COMMITLINT_SUBJECT_MIN_LEN", defaultSubjectMinLen),
	}

	if cfg.msgMaxLen <= 0 {
		cfg.msgMaxLen = defaultMsgMaxLen
	}

	if cfg.scopeMinLen <= 0 {
		cfg.scopeMinLen = defaultScopeMinLen
	}

	if cfg.scopeMaxLen <= 0 {
		cfg.scopeMaxLen = defaultScopeMaxLen
	}

	if cfg.subjectMinLen <= 0 {
		cfg.subjectMinLen = defaultSubjectMinLen
	}

	if cfg.scopeMinLen > cfg.scopeMaxLen {
		cfg.scopeMinLen, cfg.scopeMaxLen = cfg.scopeMaxLen, cfg.scopeMinLen
	}

	return cfg
}

// getEnvAsInt gets an environment variable as an integer.
func getEnvAsInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return n
}

// validate runs all the rules on the commit message.
func (lint *linter) validate(msg commitMessage) error {
	var errs []error

	for _, rule := range lint.rules {
		if err := rule(msg, lint); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return linterError{
			kind:   errValidate,
			errors: errs,
		}
	}

	return nil
}

// checkLength checks if the commit message is too long.
func checkLength(msg commitMessage, lint *linter) error {
	if len(msg.raw) > lint.config.msgMaxLen {
		return fmt.Errorf("%w (max %d chars)", errMsgTooLong, lint.config.msgMaxLen)
	}

	return nil
}

// checkASCII checks if the commit message contains non-ASCII characters.
func checkASCII(msg commitMessage, _ *linter) error {
	for i := 0; i < len(msg.raw); i++ {
		if msg.raw[i] > 127 {
			return fmt.Errorf("%w (got '%c')", errNonASCII, msg.raw[i])
		}
	}

	return nil
}

// checkType checks if the commit type is allowed.
func checkType(msg commitMessage, lint *linter) error {
	if _, ok := lint.allowedTypes[msg.typ]; !ok {
		var allowed []string
		for t := range lint.allowedTypes {
			allowed = append(allowed, string(t))
		}

		sort.Strings(allowed)

		return fmt.Errorf("%w (allowed: %s)", errTypeInvalid, strings.Join(allowed, ", "))
	}

	return nil
}

// checkScope checks if the scope is allowed.
func checkScope(msg commitMessage, lint *linter) error {
	if msg.scope == "" {
		return errScopeRequired
	}

	n := len(msg.scope)
	if n < lint.config.scopeMinLen {
		return fmt.Errorf("%w (min %d chars)", errScopeTooShort, lint.config.scopeMinLen)
	}

	if n > lint.config.scopeMaxLen {
		return fmt.Errorf("%w (max %d chars)", errScopeTooLong, lint.config.scopeMaxLen)
	}

	if !lint.scopeRegex.MatchString(msg.scope) {
		return errScopeInvalidChar
	}

	return nil
}

// checkBreaking checks if the commit is a breaking change.
func checkBreaking(msg commitMessage, lint *linter) error {
	if msg.isBreaking {
		if _, ok := lint.allowedBreakingTypes[msg.typ]; !ok {
			return fmt.Errorf("%w (got '%s')", errInvalidBreakingType, msg.typ)
		}
	}

	return nil
}

// checkSubject checks if the subject is allowed.
func checkSubject(msg commitMessage, lint *linter) error {
	if len(msg.subject) < lint.config.subjectMinLen {
		return fmt.Errorf("%w (min %d chars)", errSubjectTooShort, lint.config.subjectMinLen)
	}

	if strings.HasPrefix(msg.subject, " ") {
		return errSubjectLeadingSpace
	}

	if strings.Contains(msg.subject, "  ") {
		return errSubjectDoubleSpace
	}

	first := msg.subject[0]
	if first < 'a' || first > 'z' {
		return fmt.Errorf("%w (got '%c')", errSubjectNotLowercase, first)
	}

	if strings.HasSuffix(msg.subject, ".") {
		return errSubjectEndsWithPeriod
	}

	return nil
}
