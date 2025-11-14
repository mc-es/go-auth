// Package commitcheck implements the commit message checker executed by the commit-msg hook.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mc-es/go-auth/pkg/logger"
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
	defaultMaxCommitLen = 100
	defaultScopeMaxLen  = 10
	defaultScopeMinLen  = 2

	defaultSubjectMinLen = 3
)

// commitMessage holds the parsed commit header information.
type commitMessage struct {
	typ     commitType
	scope   string
	subject string
}

// scopeRule collects scope-related validation requirements.
type scopeRule struct {
	maxLen  int
	minLen  int
	pattern *regexp.Regexp
}

// subjectRule collects subject-related validation requirements.
type subjectRule struct {
	minLen  int
	pattern *regexp.Regexp
}

// commitRule bundles every validation knob applied to a commit message.
type commitRule struct {
	allowedTypes    []commitType
	allowedTypesSet map[commitType]struct{}
	maxLen          int
	scope           scopeRule
	subject         subjectRule
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

// run orchestrates logger initialization, message parsing, and validation.
func run() error {
	if err := logger.Init(
		logger.WithDevelopmentMode(),
		logger.WithLevel("info"),
		logger.WithoutStacktrace(),
	); err != nil {
		panic(err)
	}

	defer func() {
		_ = logger.Sync()
	}()

	log := logger.S()

	if len(os.Args) < 2 {
		return fmt.Errorf("missing commit message file path")
	}

	message, err := readMessage(os.Args[1])
	if err != nil {
		log.Errorw("commit message read failed", "error", err)

		return err
	}

	parsed, err := parseMessage(message)
	if err != nil {
		log.Errorw("commit message parse failed", "error", err)

		return err
	}

	if err := validateMessage(parsed, defaultConfig(), message); err != nil {
		log.Errorw("commit message validation failed", "error", err)

		return err
	}

	return nil
}

// defaultConfig returns the strict default rule set enforced by the hook.
func defaultConfig() commitRule {
	rules := commitRule{
		allowedTypes: []commitType{
			feat, fix, chore,
			docs, style, refactor,
			perf, test, build,
		},
		maxLen: defaultMaxCommitLen,
		scope: scopeRule{
			maxLen:  defaultScopeMaxLen,
			minLen:  defaultScopeMinLen,
			pattern: regexp.MustCompile(`^[a-z0-9_.-]+$`),
		},
		subject: subjectRule{
			minLen:  defaultSubjectMinLen,
			pattern: regexp.MustCompile(`^[a-zA-Z0-9 :,_\-(){}\[\].'"\/]+$`),
		},
	}

	m := make(map[commitType]struct{}, len(rules.allowedTypes))
	for _, t := range rules.allowedTypes {
		m[t] = struct{}{}
	}

	rules.allowedTypesSet = m

	return rules
}

// readMessage reads and cleans the first non-comment line from the commit file.
func readMessage(path string) (msg string, err error) {
	cleanPath := filepath.Clean(path)

	file, err := os.Open(cleanPath)
	if err != nil {
		return "", fmt.Errorf("closing %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		return line, nil
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return "", fmt.Errorf("scanning %w", scanErr)
	}

	return "", fmt.Errorf("commit message is empty")
}

// parseMessage splits the raw message into structured components.
func parseMessage(message string) (commitMessage, error) {
	var msg commitMessage

	header, subject, err := splitCommitMessage(message)
	if err != nil {
		return msg, err
	}

	typ, scope, err := parseHeader(header)
	if err != nil {
		return msg, err
	}

	msg.typ = typ
	msg.scope = scope
	msg.subject = subject

	return msg, nil
}

// splitCommitMessage validates and separates the header from the subject.
func splitCommitMessage(message string) (string, string, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return "", "", fmt.Errorf("empty commit message")
	}

	colonIdx := strings.IndexByte(message, ':')
	if colonIdx == -1 {
		return "", "", fmt.Errorf(
			"missing ':' separator (expected type(scope): subject)",
		)
	}

	if colonIdx+1 >= len(message) || message[colonIdx+1] != ' ' {
		return "", "", fmt.Errorf(
			"expected a space after ':' (use `type(scope): subject`)",
		)
	}

	header := strings.TrimSpace(message[:colonIdx])
	subject := strings.TrimSpace(message[colonIdx+2:])

	if header == "" {
		return "", "", fmt.Errorf("missing commit type and scope")
	}

	if subject == "" {
		return "", "", fmt.Errorf("subject is empty")
	}

	return header, subject, nil
}

// parseHeader extracts the commit type and scope segments from the header.
func parseHeader(header string) (commitType, string, error) {
	openIdx := strings.IndexByte(header, '(')
	closeIdx := strings.LastIndexByte(header, ')')

	if openIdx == -1 || closeIdx == -1 {
		return "", "", fmt.Errorf(
			"scope is required (expected type(scope): subject)",
		)
	}

	if closeIdx <= openIdx+1 || closeIdx != len(header)-1 {
		return "", "", fmt.Errorf(
			"invalid scope format (expected type(scope))",
		)
	}

	typePart := strings.TrimSpace(header[:openIdx])
	if typePart == "" {
		return "", "", fmt.Errorf("missing commit type")
	}

	scopePart := strings.TrimSpace(header[openIdx+1 : closeIdx])
	if scopePart == "" {
		return "", "", fmt.Errorf("scope is empty")
	}

	return commitType(typePart), scopePart, nil
}

// validateMessage enforces every configured rule on the parsed message.
func validateMessage(msg commitMessage, rules commitRule, raw string) error {
	l := utf8.RuneCountInString(raw)
	if l > rules.maxLen {
		return fmt.Errorf(
			"commit message length %d exceeds max %d",
			l,
			rules.maxLen,
		)
	}

	if err := validateType(msg, rules); err != nil {
		return err
	}

	if err := validateScope(msg.scope, rules); err != nil {
		return err
	}

	if err := validateSubject(msg.subject, rules); err != nil {
		return err
	}

	return nil
}

// validateType ensures that the commit type belongs to the allowed list.
func validateType(msg commitMessage, rules commitRule) error {
	if _, ok := rules.allowedTypesSet[msg.typ]; !ok {
		return fmt.Errorf("commit type %q is not allowed", msg.typ)
	}

	return nil
}

// validateScope ensures the scope length and characters stay within limits.
func validateScope(scope string, rules commitRule) error {
	scopeLen := utf8.RuneCountInString(scope)
	if scopeLen < rules.scope.minLen {
		return fmt.Errorf(
			"scope %q shorter than minimum %d characters",
			scope,
			rules.scope.minLen,
		)
	}

	if scopeLen > rules.scope.maxLen {
		return fmt.Errorf(
			"scope %q longer than maximum %d characters",
			scope,
			rules.scope.maxLen,
		)
	}

	if !rules.scope.pattern.MatchString(scope) {
		return fmt.Errorf("scope %q contains invalid characters", scope)
	}

	return nil
}

// validateSubject enforces subject formatting, character set, and casing.
func validateSubject(subject string, rules commitRule) error {
	subjectLen := utf8.RuneCountInString(subject)
	if subjectLen < rules.subject.minLen {
		return fmt.Errorf(
			"subject %q shorter than minimum %d characters",
			subject,
			rules.subject.minLen,
		)
	}

	if !rules.subject.pattern.MatchString(subject) {
		return fmt.Errorf("subject %q contains invalid characters", subject)
	}

	if strings.HasSuffix(subject, ".") {
		return fmt.Errorf("subject must not end with a full stop")
	}

	first, _ := utf8.DecodeRuneInString(subject)
	if unicode.IsUpper(first) {
		return fmt.Errorf("subject must start with a lowercase letter")
	}

	return nil
}
