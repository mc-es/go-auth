package main

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		fileContent string
		createFile  bool
		checkErr    assert.ErrorAssertionFunc
	}{
		{
			name:     "missing arguments",
			args:     []string{"cmd"},
			checkErr: assert.Error,
		},
		{
			name:       "file not found",
			args:       []string{"cmd", "non-existent-file.txt"},
			createFile: false,
			checkErr:   assert.Error,
		},
		{
			name:        "valid commit message",
			args:        []string{"cmd", "DUMMY_PATH"},
			fileContent: "refactor(ui): improve code readability",
			createFile:  true,
			checkErr:    assert.NoError,
		},
		{
			name:        "invalid commit message",
			args:        []string{"cmd", "DUMMY_PATH"},
			fileContent: "fixed bug",
			createFile:  true,
			checkErr:    assert.Error,
		},
		{
			name:        "missing scope",
			args:        []string{"cmd", "DUMMY_PATH"},
			fileContent: "perf: improve performance",
			createFile:  true,
			checkErr:    assert.Error,
		},
		{
			name:        "empty file",
			args:        []string{"cmd", "DUMMY_PATH"},
			fileContent: "",
			createFile:  true,
			checkErr:    assert.Error,
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

			tt.checkErr(t, run(args))
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
			name:  "simple valid line",
			input: "docs: update README",
			want:  "docs: update README",
		},
		{
			name:  "skip leading empty lines",
			input: "\n\n\nfix: bug fix",
			want:  "fix: bug fix",
		},
		{
			name:  "skip comments (#)",
			input: "# This is a comment\n# Another comment\nchore: clean up unused code",
			want:  "chore: clean up unused code",
		},
		{
			name:  "trim whitespace",
			input: "   build: update dependencies   ",
			want:  "build: update dependencies",
		},
		{
			name:    "empty file",
			input:   "",
			wantErr: errEmptyMessage,
		},
		{
			name:    "file with only comments",
			input:   "# comment 1\n# comment 2",
			wantErr: errEmptyMessage,
		},
		{
			name:    "file with only whitespace",
			input:   "   \n  \t  ",
			wantErr: errEmptyMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := readFirstLine(r)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
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
				raw:     "feat: add login",
				typ:     feat,
				subject: "add login",
			},
		},
		{
			name:   "scoped commit",
			rawMsg: "fix(ui): button color",
			want: commitMessage{
				raw:     "fix(ui): button color",
				typ:     fix,
				scope:   "ui",
				subject: "button color",
			},
		},
		{
			name:   "breaking change without scope",
			rawMsg: "feat!: drop old api",
			want: commitMessage{
				raw:        "feat!: drop old api",
				typ:        feat,
				subject:    "drop old api",
				isBreaking: true,
			},
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
		},
		{
			name:   "extra whitespace in subject",
			rawMsg: "feat:  add login",
			want: commitMessage{
				raw:     "feat:  add login",
				typ:     feat,
				subject: " add login",
			},
		},
		{
			name:    "missing separator",
			rawMsg:  "feat add login",
			wantErr: errMissingSeparator,
		},
		{
			name:    "colon without space",
			rawMsg:  "feat:add login",
			wantErr: errMissingSeparator,
		},
		{
			name:    "error propagates from child parser",
			rawMsg:  "feat(ui: broken parens",
			wantErr: errMismatchedParens,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommitMessage(tt.rawMsg)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
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
			name:     "simple type",
			prefix:   "feat",
			wantType: feat,
		},
		{
			name:      "type with scope",
			prefix:    "feat(auth)",
			wantType:  feat,
			wantScope: "auth",
		},
		{
			name:      "scope contains confusing characters",
			prefix:    "fix(gui-v1.2)",
			wantType:  fix,
			wantScope: "gui-v1.2",
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
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantType, gotType)
				assert.Equal(t, tt.wantScope, gotScope)
			}
		})
	}
}

func TestNewLinter(t *testing.T) {
	lint := newLinter()
	require.NotNil(t, lint, "newLinter() returned nil")
	assert.NotNil(t, lint.scopeRegex)
	assert.NotEmpty(t, lint.allowedTypes)
	assert.NotEmpty(t, lint.allowedBreakingTypes)
	assert.NotEmpty(t, lint.rules)
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		envValues map[string]string
		want      config
	}{
		{
			name: "default values",
			want: config{defaultMsgMaxLen, defaultScopeMinLen, defaultScopeMaxLen, defaultSubjectMinLen},
		},
		{
			name: "environment variables override defaults",
			envValues: map[string]string{
				"COMMITLINT_MSG_MAX_LEN":     "120",
				"COMMITLINT_SCOPE_MIN_LEN":   "3",
				"COMMITLINT_SCOPE_MAX_LEN":   "20",
				"COMMITLINT_SUBJECT_MIN_LEN": "5",
			},
			want: config{120, 3, 20, 5},
		},
		{
			name: "invalid values",
			envValues: map[string]string{
				"COMMITLINT_MSG_MAX_LEN":     "not-a-number",
				"COMMITLINT_SCOPE_MIN_LEN":   "0",
				"COMMITLINT_SCOPE_MAX_LEN":   "10",
				"COMMITLINT_SUBJECT_MIN_LEN": "0",
			},
			want: config{defaultMsgMaxLen, defaultScopeMinLen, 10, defaultSubjectMinLen},
		},
		{
			name: "negative values",
			envValues: map[string]string{
				"COMMITLINT_MSG_MAX_LEN":     "-10",
				"COMMITLINT_SCOPE_MIN_LEN":   "10",
				"COMMITLINT_SCOPE_MAX_LEN":   "-20",
				"COMMITLINT_SUBJECT_MIN_LEN": "-100",
			},
			want: config{defaultMsgMaxLen, 10, defaultScopeMaxLen, defaultSubjectMinLen},
		},
		{
			name: "swapped scope min and max",
			envValues: map[string]string{
				"COMMITLINT_SCOPE_MIN_LEN": "20",
				"COMMITLINT_SCOPE_MAX_LEN": "10",
			},
			want: config{defaultMsgMaxLen, 10, 20, defaultSubjectMinLen},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envValues {
				t.Setenv(key, value)
			}

			assert.Equal(t, tt.want, loadConfig())
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
			envValue: "NaN",
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
				t.Setenv(testKey, tt.envValue)
			}

			assert.Equal(t, tt.want, getEnvAsInt(testKey, tt.fallback))
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
		wantErrs []error
	}{
		{
			name:  "no rules",
			rules: []ruleFunc{},
		},
		{
			name:  "one rule passes",
			rules: []ruleFunc{rulePass},
		},
		{
			name:     "one rule fails",
			rules:    []ruleFunc{ruleFail1},
			wantErrs: []error{errMock1},
		},
		{
			name:     "mixed rules",
			rules:    []ruleFunc{rulePass, ruleFail1, rulePass, ruleFail2},
			wantErrs: []error{errMock1, errMock2},
		},
		{
			name:     "all rules fail",
			rules:    []ruleFunc{ruleFail1, ruleFail2},
			wantErrs: []error{errMock1, errMock2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lint := &linter{rules: tt.rules}
			err := lint.validate(commitMessage{})

			if len(tt.wantErrs) > 0 {
				require.Error(t, err)

				var le linterError
				require.ErrorAs(t, err, &le)
				assert.Equal(t, errValidate, le.kind)
				require.Len(t, le.errors, len(tt.wantErrs))

				for i, want := range tt.wantErrs {
					assert.ErrorIs(t, le.errors[i], want)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckRules(t *testing.T) {
	runCheck := func(t *testing.T, check func(commitMessage, *linter) error, tests []struct {
		name string
		msg  commitMessage
		want error
	},
	) {
		t.Helper()

		lint := &linter{
			config:               config{msgMaxLen: 10, scopeMinLen: 2, scopeMaxLen: 15, subjectMinLen: 3},
			scopeRegex:           regexp.MustCompile(`^[a-z0-9_-]+$`),
			allowedTypes:         map[commitType]struct{}{feat: {}, fix: {}},
			allowedBreakingTypes: map[commitType]struct{}{feat: {}, build: {}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := check(tt.msg, lint)
				if tt.want != nil {
					assert.ErrorIs(t, got, tt.want)
				} else {
					assert.NoError(t, got)
				}
			})
		}
	}

	t.Run("CheckLength", func(t *testing.T) {
		runCheck(t, checkLength, []struct {
			name string
			msg  commitMessage
			want error
		}{
			{
				name: "valid length",
				msg:  commitMessage{raw: "a"},
				want: nil,
			},
			{
				name: "max length",
				msg:  commitMessage{raw: strings.Repeat("a", 10)},
				want: nil,
			},
			{
				name: "exceeds length",
				msg:  commitMessage{raw: strings.Repeat("a", 11)},
				want: errMsgTooLong,
			},
		})
	})

	t.Run("CheckASCII", func(t *testing.T) {
		runCheck(t, checkASCII, []struct {
			name string
			msg  commitMessage
			want error
		}{
			{
				name: "valid ASCII",
				msg:  commitMessage{raw: "abc 123"},
				want: nil,
			},
			{
				name: "invalid ASCII",
				msg:  commitMessage{raw: "üçgen"},
				want: errNonASCII,
			},
			{
				name: "emoji",
				msg:  commitMessage{raw: "🚀"},
				want: errNonASCII,
			},
		})
	})

	t.Run("CheckType", func(t *testing.T) {
		runCheck(t, checkType, []struct {
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
				msg:  commitMessage{typ: "chore"},
				want: errTypeInvalid,
			},
			{
				name: "case sensitive",
				msg:  commitMessage{typ: "Feat"},
				want: errTypeInvalid,
			},
		})
	})

	t.Run("CheckScope", func(t *testing.T) {
		runCheck(t, checkScope, []struct {
			name string
			msg  commitMessage
			want error
		}{
			{
				name: "required",
				msg:  commitMessage{scope: ""},
				want: errScopeRequired,
			},
			{
				name: "too short",
				msg:  commitMessage{scope: "a"},
				want: errScopeTooShort,
			},
			{
				name: "valid min",
				msg:  commitMessage{scope: "ab"},
				want: nil,
			},
			{
				name: "valid max",
				msg:  commitMessage{scope: strings.Repeat("a", 15)},
				want: nil,
			},
			{
				name: "too long",
				msg:  commitMessage{scope: strings.Repeat("a", 16)},
				want: errScopeTooLong,
			},
			{
				name: "valid chars",
				msg:  commitMessage{scope: "auth-2_api"},
				want: nil,
			},
			{
				name: "invalid char",
				msg:  commitMessage{scope: "Auth"},
				want: errScopeInvalidChar,
			},
		})
	})

	t.Run("CheckBreaking", func(t *testing.T) {
		runCheck(t, checkBreaking, []struct {
			name string
			msg  commitMessage
			want error
		}{
			{
				name: "valid breaking",
				msg:  commitMessage{isBreaking: true, typ: feat},
				want: nil,
			},
			{
				name: "invalid type",
				msg:  commitMessage{isBreaking: true, typ: fix},
				want: errInvalidBreakingType,
			},
		})
	})

	t.Run("CheckSubject", func(t *testing.T) {
		runCheck(t, checkSubject, []struct {
			name string
			msg  commitMessage
			want error
		}{
			{
				name: "too short",
				msg:  commitMessage{subject: "ab"},
				want: errSubjectTooShort,
			},
			{
				name: "valid min",
				msg:  commitMessage{subject: "abc"},
				want: nil,
			},
			{
				name: "leading space",
				msg:  commitMessage{subject: " abc"},
				want: errSubjectLeadingSpace,
			},
			{
				name: "double space",
				msg:  commitMessage{subject: "a  bc"},
				want: errSubjectDoubleSpace,
			},
			{
				name: "uppercase start",
				msg:  commitMessage{subject: "Abc"},
				want: errSubjectNotLowercase,
			},
			{
				name: "ends with period",
				msg:  commitMessage{subject: "abc."},
				want: errSubjectEndsWithPeriod,
			},
		})
	})
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "commit-msg-*")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	_ = tmpFile.Close()

	return tmpFile.Name()
}
