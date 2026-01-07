package main

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		fileContent string
		createFile  bool
		wantErr     bool
	}{
		{
			name:    "missing arguments",
			args:    []string{"cmd"},
			wantErr: true,
		},
		{
			name:       "file not found",
			args:       []string{"cmd", "non-existent-file.txt"},
			createFile: false,
			wantErr:    true,
		},
		{
			name:        "valid commit message",
			args:        []string{"cmd", "DUMMY_PATH"},
			fileContent: "refactor(ui): improve code readability",
			createFile:  true,
			wantErr:     false,
		},
		{
			name:        "invalid commit message",
			args:        []string{"cmd", "DUMMY_PATH"},
			fileContent: "fixed bug",
			createFile:  true,
			wantErr:     true,
		},
		{
			name:        "missing scope",
			args:        []string{"cmd", "DUMMY_PATH"},
			fileContent: "perf: improve performance",
			createFile:  true,
			wantErr:     true,
		},
		{
			name:        "empty file",
			args:        []string{"cmd", "DUMMY_PATH"},
			fileContent: "",
			createFile:  true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := make([]string, len(tt.args))
			copy(args, tt.args)

			if tt.createFile {
				path := createTempFile(t, tt.fileContent)

				if len(args) > 1 {
					args[1] = path
				}
			}

			err := run(args)
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReadFirstLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "simple valid line",
			input:   "docs: update README",
			want:    "docs: update README",
			wantErr: nil,
		},
		{
			name:    "skip leading empty lines",
			input:   "\n\n\nfix: bug fix",
			want:    "fix: bug fix",
			wantErr: nil,
		},
		{
			name:    "skip comments (#)",
			input:   "# This is a comment\n# Another comment\nchore: clean up unused code",
			want:    "chore: clean up unused code",
			wantErr: nil,
		},
		{
			name:    "trim whitespace",
			input:   "   build: update dependencies   ",
			want:    "build: update dependencies",
			wantErr: nil,
		},
		{
			name:    "empty file",
			input:   "",
			want:    "",
			wantErr: errEmptyMessage,
		},
		{
			name:    "file with only comments",
			input:   "# comment 1\n# comment 2",
			want:    "",
			wantErr: errEmptyMessage,
		},
		{
			name:    "file with only whitespace",
			input:   "   \n  \t  ",
			want:    "",
			wantErr: errEmptyMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := readFirstLine(r)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("readFirstLine() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("readFirstLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		rawMsg  string
		want    commitMessage
		wantErr error
	}{
		{
			name:   "standard commit",
			rawMsg: "feat: add login",
			want: commitMessage{
				raw:        "feat: add login",
				typ:        feat,
				scope:      "",
				subject:    "add login",
				isBreaking: false,
			},
			wantErr: nil,
		},
		{
			name:   "scoped commit",
			rawMsg: "fix(ui): button color",
			want: commitMessage{
				raw:        "fix(ui): button color",
				typ:        fix,
				scope:      "ui",
				subject:    "button color",
				isBreaking: false,
			},
			wantErr: nil,
		},
		{
			name:   "breaking change without scope",
			rawMsg: "feat!: drop old api",
			want: commitMessage{
				raw:        "feat!: drop old api",
				typ:        feat,
				scope:      "",
				subject:    "drop old api",
				isBreaking: true,
			},
			wantErr: nil,
		},
		{
			name:   "breaking change with scope",
			rawMsg: "refactor(db)!: migrate schema",
			want: commitMessage{
				raw:        "refactor(db)!: migrate schema",
				typ:        refactor,
				scope:      "db",
				subject:    "migrate schema",
				isBreaking: true,
			},
			wantErr: nil,
		},
		{
			name:   "extra whitespace in subject",
			rawMsg: "feat:  add login",
			want: commitMessage{
				raw:        "feat:  add login",
				typ:        feat,
				scope:      "",
				subject:    " add login",
				isBreaking: false,
			},
			wantErr: nil,
		},
		{
			name:    "missing separator",
			rawMsg:  "feat add login",
			want:    commitMessage{},
			wantErr: errMissingSeparator,
		},
		{
			name:    "colon without space",
			rawMsg:  "feat:add login",
			want:    commitMessage{},
			wantErr: errMissingSeparator,
		},
		{
			name:    "error propagates from child parser",
			rawMsg:  "feat(ui: broken parens",
			want:    commitMessage{},
			wantErr: errMismatchedParens,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommitMessage(tt.rawMsg)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("parse error = %v, want %v", err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)

				return
			}

			if got != tt.want {
				t.Errorf("\ngot  %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

func TestParseTypeAndScope(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		wantType  commitType
		wantScope string
		wantErr   error
	}{
		{
			name:      "simple type",
			prefix:    "feat",
			wantType:  feat,
			wantScope: "",
			wantErr:   nil,
		},
		{
			name:      "type with scope",
			prefix:    "feat(auth)",
			wantType:  feat,
			wantScope: "auth",
			wantErr:   nil,
		},
		{
			name:      "scope contains confusing characters",
			prefix:    "fix(gui-v1.2)",
			wantType:  fix,
			wantScope: "gui-v1.2",
			wantErr:   nil,
		},
		{
			name:    "open paren without close",
			prefix:  "feat(auth",
			wantErr: errMismatchedParens,
		},
		{
			name:    "close paren without open",
			prefix:  "featauth)",
			wantErr: errMismatchedParens,
		},
		{
			name:    "empty scope",
			prefix:  "feat()",
			wantErr: errEmptyScopeParens,
		},
		{
			name:    "scope not at end",
			prefix:  "feat(auth)extra",
			wantErr: errInvalidScopeFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotScope, err := parseTypeAndScope(tt.prefix)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("parseTypeAndScope() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if err == nil {
				if gotType != tt.wantType {
					t.Errorf("gotType = %v, want %v", gotType, tt.wantType)
				}

				if gotScope != tt.wantScope {
					t.Errorf("gotScope = %v, want %v", gotScope, tt.wantScope)
				}
			}
		})
	}
}

func TestNewLinter(t *testing.T) {
	lint := newLinter()

	if lint == nil {
		t.Fatal("newLinter() returned nil")
	}

	if lint.scopeRegex == nil {
		t.Error("scopeRegex is nil")
	}

	if lint.allowedTypes == nil {
		t.Error("allowedTypes map is nil")
	}

	if len(lint.allowedTypes) == 0 {
		t.Error("allowedTypes map is empty")
	}

	if lint.allowedBreakingTypes == nil {
		t.Error("allowedBreakingTypes map is nil")
	}

	if len(lint.allowedBreakingTypes) == 0 {
		t.Error("allowedBreakingTypes map is empty")
	}

	if len(lint.rules) == 0 {
		t.Error("rules list is empty")
	}
}

func TestLoadConfig(t *testing.T) {
	const (
		testKeyMsgMaxLen     = "COMMITLINT_MSG_MAX_LEN"
		testKeyScopeMinLen   = "COMMITLINT_SCOPE_MIN_LEN"
		testKeyScopeMaxLen   = "COMMITLINT_SCOPE_MAX_LEN"
		testKeySubjectMinLen = "COMMITLINT_SUBJECT_MIN_LEN"
	)

	tests := []struct {
		name      string
		envValues map[string]string
		want      config
	}{
		{
			name:      "default values",
			envValues: map[string]string{},
			want: config{
				msgMaxLen:     defaultMsgMaxLen,
				scopeMinLen:   defaultScopeMinLen,
				scopeMaxLen:   defaultScopeMaxLen,
				subjectMinLen: defaultSubjectMinLen,
			},
		},
		{
			name: "environment variables override defaults",
			envValues: map[string]string{
				testKeyMsgMaxLen:     "120",
				testKeyScopeMinLen:   "3",
				testKeyScopeMaxLen:   "20",
				testKeySubjectMinLen: "5",
			},
			want: config{
				msgMaxLen:     120,
				scopeMinLen:   3,
				scopeMaxLen:   20,
				subjectMinLen: 5,
			},
		},
		{
			name: "invalid values",
			envValues: map[string]string{
				testKeyMsgMaxLen:     "not-a-number",
				testKeyScopeMinLen:   "0",
				testKeyScopeMaxLen:   "10",
				testKeySubjectMinLen: "0",
			},
			want: config{
				msgMaxLen:     defaultMsgMaxLen,
				scopeMinLen:   defaultScopeMinLen,
				scopeMaxLen:   10,
				subjectMinLen: defaultSubjectMinLen,
			},
		},
		{
			name: "negative values",
			envValues: map[string]string{
				testKeyMsgMaxLen:     "-10",
				testKeyScopeMinLen:   "10",
				testKeyScopeMaxLen:   "-20",
				testKeySubjectMinLen: "-100",
			},
			want: config{
				msgMaxLen:     defaultMsgMaxLen,
				scopeMinLen:   10,
				scopeMaxLen:   defaultScopeMaxLen,
				subjectMinLen: defaultSubjectMinLen,
			},
		},
		{
			name: "swapped scope min and max",
			envValues: map[string]string{
				testKeyScopeMinLen: "20",
				testKeyScopeMaxLen: "10",
			},
			want: config{
				scopeMinLen:   10,
				scopeMaxLen:   20,
				subjectMinLen: defaultSubjectMinLen,
				msgMaxLen:     defaultMsgMaxLen,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envValues {
				_ = os.Setenv(key, value)
			}

			defer func() {
				for key := range tt.envValues {
					_ = os.Unsetenv(key)
				}
			}()

			got := loadConfig()
			if got != tt.want {
				t.Errorf("loadConfig() = %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	const testKey = "TEST_ENV_INT_KEY"

	tests := []struct {
		name     string
		envValue string
		isSet    bool
		fallback int
		want     int
	}{
		{
			name:     "valid integer",
			envValue: "42",
			isSet:    true,
			fallback: 10,
			want:     42,
		},
		{
			name:     "variable not set",
			envValue: "",
			isSet:    false,
			fallback: 10,
			want:     10,
		},
		{
			name:     "variable set but empty",
			envValue: "",
			isSet:    true,
			fallback: 10,
			want:     10,
		},
		{
			name:     "invalid integer",
			envValue: "not-a-number",
			isSet:    true,
			fallback: 10,
			want:     10,
		},
		{
			name:     "float value",
			envValue: "3.14",
			isSet:    true,
			fallback: 10,
			want:     10,
		},
		{
			name:     "negative integer",
			envValue: "-42",
			isSet:    true,
			fallback: 10,
			want:     -42,
		},
		{
			name:     "zero value",
			envValue: "0",
			isSet:    true,
			fallback: 10,
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isSet {
				_ = os.Setenv(testKey, tt.envValue)
			}

			defer func() { _ = os.Unsetenv(testKey) }()

			got := getEnvAsInt(testKey, tt.fallback)
			if got != tt.want {
				t.Errorf("getEnvAsInt() = %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	errMock1 := errors.New("mock error 1")
	errMock2 := errors.New("mock error 2")

	rulePass := func(_ commitMessage, _ *linter) error { return nil }
	ruleFail1 := func(_ commitMessage, _ *linter) error { return errMock1 }
	ruleFail2 := func(_ commitMessage, _ *linter) error { return errMock2 }

	tests := []struct {
		name     string
		rules    []ruleFunc
		wantErr  bool
		wantErrs []error
	}{
		{
			name:     "no rules",
			rules:    []ruleFunc{},
			wantErr:  false,
			wantErrs: nil,
		},
		{
			name:     "one rule passes",
			rules:    []ruleFunc{rulePass},
			wantErr:  false,
			wantErrs: nil,
		},
		{
			name:     "one rule fails",
			rules:    []ruleFunc{ruleFail1},
			wantErr:  true,
			wantErrs: []error{errMock1},
		},
		{
			name:     "some rules pass, some fail",
			rules:    []ruleFunc{rulePass, ruleFail1, rulePass, ruleFail2},
			wantErr:  true,
			wantErrs: []error{errMock1, errMock2},
		},
		{
			name:     "all rules fail",
			rules:    []ruleFunc{ruleFail1, ruleFail2},
			wantErr:  true,
			wantErrs: []error{errMock1, errMock2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lint := &linter{rules: tt.rules}
			err := lint.validate(commitMessage{})

			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wanted %v", err, tt.wantErr)
			}

			if tt.wantErr {
				verifyLinterError(t, err, tt.wantErrs)
			}
		})
	}
}

func TestCheckLength(t *testing.T) {
	const msgMaxLen = 10

	sharedLinter := &linter{
		config: config{msgMaxLen: msgMaxLen},
	}

	tests := []struct {
		name string
		msg  commitMessage
		want error
	}{
		{
			name: "message below max length",
			msg:  commitMessage{raw: "a"},
			want: nil,
		},
		{
			name: "message at max length",
			msg:  commitMessage{raw: strings.Repeat("a", msgMaxLen)},
			want: nil,
		},
		{
			name: "message exceeds max length",
			msg:  commitMessage{raw: strings.Repeat("a", msgMaxLen+1)},
			want: errMsgTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkLength(tt.msg, sharedLinter)
			if !errors.Is(got, tt.want) {
				t.Errorf("checkLength() error = %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestCheckASCII(t *testing.T) {
	sharedLinter := &linter{}

	tests := []struct {
		name string
		msg  commitMessage
		want error
	}{
		{
			name: "valid ASCII with letters and numbers",
			msg:  commitMessage{raw: "add 123 login"},
			want: nil,
		},
		{
			name: "invalid ASCII with Turkish characters",
			msg:  commitMessage{raw: "add üğşçö login"},
			want: errNonASCII,
		},
		{
			name: "invalid ASCII with emojis",
			msg:  commitMessage{raw: "add 🚀 login"},
			want: errNonASCII,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkASCII(tt.msg, sharedLinter)
			if !errors.Is(got, tt.want) {
				t.Errorf("checkASCII() error = %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestCheckType(t *testing.T) {
	allowedTypes := map[commitType]struct{}{
		feat: {},
		fix:  {},
	}
	sharedLinter := &linter{
		allowedTypes: allowedTypes,
	}

	tests := []struct {
		name string
		msg  commitMessage
		want error
	}{
		{
			name: "valid type",
			msg:  commitMessage{typ: feat},
			want: nil,
		},
		{
			name: "invalid type",
			msg:  commitMessage{typ: commitType("something")},
			want: errTypeInvalid,
		},
		{
			name: "case sensitive type",
			msg:  commitMessage{typ: commitType("Feat")},
			want: errTypeInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkType(tt.msg, sharedLinter)
			if !errors.Is(got, tt.want) {
				t.Errorf("checkType() error = %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestCheckScope(t *testing.T) {
	const (
		scopeMinLen = 2
		scopeMaxLen = 15
	)

	scopeRegex := regexp.MustCompile(`^[a-z0-9_-]+$`)
	sharedLinter := &linter{
		scopeRegex: scopeRegex,
		config: config{
			scopeMinLen: scopeMinLen,
			scopeMaxLen: scopeMaxLen,
		},
	}

	tests := []struct {
		name string
		msg  commitMessage
		want error
	}{
		{
			name: "scope is required",
			msg:  commitMessage{scope: ""},
			want: errScopeRequired,
		},
		{
			name: "scope too short",
			msg:  commitMessage{scope: "a"},
			want: errScopeTooShort,
		},
		{
			name: "scope at min length",
			msg:  commitMessage{scope: "ab"},
			want: nil,
		},
		{
			name: "scope at max length",
			msg:  commitMessage{scope: strings.Repeat("a", scopeMaxLen)},
			want: nil,
		},
		{
			name: "scope exceeds max length",
			msg:  commitMessage{scope: strings.Repeat("a", scopeMaxLen+1)},
			want: errScopeTooLong,
		},
		{
			name: "scope with valid lowercase",
			msg:  commitMessage{scope: "auth"},
			want: nil,
		},
		{
			name: "scope with valid numbers",
			msg:  commitMessage{scope: "auth2"},
			want: nil,
		},
		{
			name: "scope with valid underscore",
			msg:  commitMessage{scope: "user_auth"},
			want: nil,
		},
		{
			name: "scope with valid hyphen",
			msg:  commitMessage{scope: "user-auth"},
			want: nil,
		},
		{
			name: "scope with uppercase",
			msg:  commitMessage{scope: "Auth"},
			want: errScopeInvalidChar,
		},
		{
			name: "scope with invalid character",
			msg:  commitMessage{scope: "auth@test"},
			want: errScopeInvalidChar,
		},
		{
			name: "scope with space",
			msg:  commitMessage{scope: "auth test"},
			want: errScopeInvalidChar,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkScope(tt.msg, sharedLinter)
			if !errors.Is(got, tt.want) {
				t.Errorf("checkScope() error = %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestCheckBreaking(t *testing.T) {
	allowedBreakingTypes := map[commitType]struct{}{
		feat:  {},
		build: {},
	}
	sharedLinter := &linter{
		allowedBreakingTypes: allowedBreakingTypes,
	}

	tests := []struct {
		name string
		msg  commitMessage
		want error
	}{
		{
			name: "valid breaking change",
			msg:  commitMessage{isBreaking: true, typ: feat},
			want: nil,
		},
		{
			name: "invalid breaking change",
			msg:  commitMessage{isBreaking: true, typ: fix},
			want: errInvalidBreakingType,
		},
		{
			name: "invalid breaking change type",
			msg:  commitMessage{isBreaking: true, typ: commitType("something")},
			want: errInvalidBreakingType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkBreaking(tt.msg, sharedLinter)
			if !errors.Is(got, tt.want) {
				t.Errorf("checkBreaking() error = %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestCheckSubject(t *testing.T) {
	const subjectMinLen = 3

	sharedLinter := &linter{
		config: config{
			subjectMinLen: subjectMinLen,
		},
	}

	tests := []struct {
		name string
		msg  commitMessage
		want error
	}{
		{
			name: "subject too short",
			msg:  commitMessage{subject: "ab"},
			want: errSubjectTooShort,
		},
		{
			name: "subject at min length",
			msg:  commitMessage{subject: "abc"},
			want: nil,
		},
		{
			name: "subject with extra leading whitespace",
			msg:  commitMessage{subject: " add login"},
			want: errSubjectLeadingSpace,
		},
		{
			name: "subject with double space",
			msg:  commitMessage{subject: "add  login"},
			want: errSubjectDoubleSpace,
		},
		{
			name: "subject starts with uppercase",
			msg:  commitMessage{subject: "Add login"},
			want: errSubjectNotLowercase,
		},
		{
			name: "subject ends with period",
			msg:  commitMessage{subject: "add login."},
			want: errSubjectEndsWithPeriod,
		},
		{
			name: "subject with tab character",
			msg:  commitMessage{subject: "add\tlogin"},
			want: nil,
		},
		{
			name: "subject starts with lowercase",
			msg:  commitMessage{subject: "add login"},
			want: nil,
		},
		{
			name: "subject starts with number",
			msg:  commitMessage{subject: "123 login"},
			want: errSubjectNotLowercase,
		},
		{
			name: "subject with special character",
			msg:  commitMessage{subject: "@login"},
			want: errSubjectNotLowercase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkSubject(tt.msg, sharedLinter)
			if !errors.Is(got, tt.want) {
				t.Errorf("checkSubject() error = %v, wanted %v", got, tt.want)
			}
		})
	}
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "commit-msg-*")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Remove(tmpFile.Name())
	})

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Could not write to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Could not close temp file: %v", err)
	}

	return tmpFile.Name()
}

func verifyLinterError(t *testing.T, err error, wantErrors []error) {
	t.Helper()

	var le linterError
	if !errors.As(err, &le) {
		t.Errorf("expected linterError type, got %T", err)

		return
	}

	if le.kind != errValidate {
		t.Errorf("expected kind %v, got %v", errValidate, le.kind)
	}

	if len(le.errors) != len(wantErrors) {
		t.Errorf("expected %d errors, got %d", len(wantErrors), len(le.errors))

		return
	}

	for i, want := range wantErrors {
		got := le.errors[i]
		if !errors.Is(got, want) {
			t.Errorf("error at index %d: got %v, want %v", i, got, want)
		}
	}
}
