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
	"unicode"
	"unicode/utf8"
)

// commitType enumerates the allowed commit categories.
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

const (
	defaultMaxHeaderLen  = 100
	defaultScopeMinLen   = 2
	defaultScopeMaxLen   = 15
	defaultSubjectMinLen = 3
)

// commitMessage holds the parsed commit header information.
type commitMessage struct {
	raw        string
	typ        commitType
	scope      string
	subject    string
	isBreaking bool
}

// ruleFunc defines the function signature for validation rules.
type ruleFunc func(msg commitMessage, lint *linter) string

// linter defines the validation rules for commit messages.
type linter struct {
	allowedTypes  map[commitType]struct{}
	scopeRegex    *regexp.Regexp
	rules         []ruleFunc
	maxHeaderLen  int
	scopeMinLen   int
	scopeMaxLen   int
	subjectMinLen int
}

// main is the entry point for the commitlint tool.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: commitlint <path-to-msg-file>")
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error opening file: %v\n", err)
		os.Exit(1)
	}

	header, err := readFirstLine(file)
	if closeErr := file.Close(); closeErr != nil {
		fmt.Fprintf(os.Stderr, "❌ Error closing file: %v\n", closeErr)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error reading file: %v\n", err)
		os.Exit(1)
	}

	msg, err := parseHeader(header)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Format error: %v\n", err)
		fmt.Fprintln(os.Stderr, "👉 Usage: type(scope): subject")
		fmt.Fprintln(os.Stderr, "👉 Example: fix(auth): handle jwt expiration")
		os.Exit(1)
	}

	lint := defaultLinter()
	if err := lint.validate(msg); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
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

	return "", errors.New("commit message is empty")
}

// parseHeader parses the header of the commit message.
func parseHeader(header string) (commitMessage, error) {
	prefix, subject, found := strings.Cut(header, ": ")
	if !found {
		return commitMessage{}, errors.New("missing colon separator and space")
	}

	isBreaking := false
	if strings.HasSuffix(prefix, "!") {
		isBreaking = true
		prefix = strings.TrimSuffix(prefix, "!")
	}

	typ, scope, err := parseTypeAndScope(prefix)
	if err != nil {
		return commitMessage{}, err
	}

	return commitMessage{
		typ:        typ,
		scope:      scope,
		subject:    subject,
		raw:        header,
		isBreaking: isBreaking,
	}, nil
}

// parseTypeAndScope parses the type and scope information of the commit message.
func parseTypeAndScope(prefix string) (commitType, string, error) {
	hasOpen := strings.Contains(prefix, "(")
	hasClose := strings.Contains(prefix, ")")

	if hasOpen != hasClose {
		return "", "", errors.New("mismatched parentheses in scope definition")
	}

	if hasOpen {
		if !strings.HasSuffix(prefix, ")") {
			return "", "", errors.New("invalid scope format")
		}

		prefixWithoutClosing := strings.TrimSuffix(prefix, ")")
		parts := strings.SplitN(prefixWithoutClosing, "(", 2)

		typ := commitType(parts[0])
		scope := parts[1]

		if scope == "" {
			return "", "", errors.New("scope cannot be empty when parentheses are used")
		}

		return typ, scope, nil
	}

	return commitType(prefix), "", nil
}

// getEnvInt retrieves an integer value from the environment variable with a default fallback.
func getEnvInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}

	return parsed
}

// defaultLinter initializes a default linter with predefined validation rules.
func defaultLinter() *linter {
	lint := &linter{
		maxHeaderLen:  getEnvInt("COMMITLINT_MAX_HEADER_LEN", defaultMaxHeaderLen),
		scopeMinLen:   getEnvInt("COMMITLINT_SCOPE_MIN_LEN", defaultScopeMinLen),
		scopeMaxLen:   getEnvInt("COMMITLINT_SCOPE_MAX_LEN", defaultScopeMaxLen),
		subjectMinLen: getEnvInt("COMMITLINT_SUBJECT_MIN_LEN", defaultSubjectMinLen),
		scopeRegex:    regexp.MustCompile(`^[a-z0-9_\-]+$`),
	}

	types := []commitType{
		feat, fix, chore, docs, style,
		refactor, perf, test, build,
	}

	lint.allowedTypes = make(map[commitType]struct{}, len(types))
	for _, t := range types {
		lint.allowedTypes[t] = struct{}{}
	}

	lint.rules = []ruleFunc{
		checkLength,
		checkType,
		checkScope,
		checkSubject,
	}

	return lint
}

// validate runs all validation rules against the commit message.
func (lint *linter) validate(msg commitMessage) error {
	var errs []string

	for _, rule := range lint.rules {
		if errMsg := rule(msg, lint); errMsg != "" {
			errs = append(errs, errMsg)
		}
	}

	if len(errs) > 0 {
		return errors.New("❌ Validation errors:\n  - " + strings.Join(errs, "\n  - "))
	}

	return nil
}

// checkLength checks if the commit message header exceeds the maximum allowed length.
func checkLength(msg commitMessage, lint *linter) string {
	if count := utf8.RuneCountInString(msg.raw); count > lint.maxHeaderLen {
		return fmt.Sprintf("header is too long (max %d chars)", lint.maxHeaderLen)
	}

	return ""
}

// checkType ensures the commit message type is one of the allowed types.
func checkType(msg commitMessage, lint *linter) string {
	if _, ok := lint.allowedTypes[msg.typ]; !ok {
		allowed := make([]string, 0, len(lint.allowedTypes))
		for t := range lint.allowedTypes {
			allowed = append(allowed, string(t))
		}

		sort.Strings(allowed)

		return fmt.Sprintf("type is invalid (allowed: %s)", strings.Join(allowed, ", "))
	}

	return ""
}

// checkScope checks if the scope is within the allowed length and format.
func checkScope(msg commitMessage, lint *linter) string {
	if msg.scope == "" {
		return "scope is required"
	}

	count := utf8.RuneCountInString(msg.scope)
	if count < lint.scopeMinLen {
		return fmt.Sprintf("scope is too short (min %d chars)", lint.scopeMinLen)
	}

	if count > lint.scopeMaxLen {
		return fmt.Sprintf("scope is too long (max %d chars)", lint.scopeMaxLen)
	}

	if !lint.scopeRegex.MatchString(msg.scope) {
		return "scope contains invalid characters (lowercase, numbers, -, _)"
	}

	return ""
}

// checkSubject ensures the subject meets all formatting and style requirements.
func checkSubject(msg commitMessage, lint *linter) string {
	if strings.HasPrefix(msg.subject, " ") {
		return "subject contains extra leading whitespace"
	}

	if strings.Contains(msg.subject, "  ") {
		return "subject contains consecutive whitespaces"
	}

	if utf8.RuneCountInString(msg.subject) < lint.subjectMinLen {
		return fmt.Sprintf("subject is too short (min %d chars)", lint.subjectMinLen)
	}

	if strings.HasSuffix(msg.subject, ".") {
		return "subject cannot end with a period"
	}

	if first, _ := utf8.DecodeRuneInString(msg.subject); !unicode.IsLower(first) {
		return fmt.Sprintf("subject must start with a lowercase letter (got '%c')", first)
	}

	return ""
}
