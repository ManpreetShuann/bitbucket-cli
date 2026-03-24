package validation

import (
	"strings"
	"testing"
)

func TestValidateProjectKey(t *testing.T) {
	valid := []string{"PROJ", "my_proj", "~jsmith", "A", strings.Repeat("a", 128)}
	for _, v := range valid {
		if err := ValidateProjectKey(v); err != nil {
			t.Errorf("ValidateProjectKey(%q) = %v, want nil", v, err)
		}
	}

	invalid := []string{"", "has spaces", "special@char", strings.Repeat("a", 129), "proj/key"}
	for _, v := range invalid {
		if err := ValidateProjectKey(v); err == nil {
			t.Errorf("ValidateProjectKey(%q) = nil, want error", v)
		}
	}
}

func TestValidateRepoSlug(t *testing.T) {
	valid := []string{"my-repo", "repo123", "my.repo", "A"}
	for _, v := range valid {
		if err := ValidateRepoSlug(v); err != nil {
			t.Errorf("ValidateRepoSlug(%q) = %v, want nil", v, err)
		}
	}

	invalid := []string{"", "-starts-with-dash", ".starts-with-dot", "has spaces"}
	for _, v := range invalid {
		if err := ValidateRepoSlug(v); err == nil {
			t.Errorf("ValidateRepoSlug(%q) = nil, want error", v)
		}
	}
}

func TestValidateCommitID(t *testing.T) {
	valid := []string{"abcd", "abc123def456", "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"}
	for _, v := range valid {
		if err := ValidateCommitID(v); err != nil {
			t.Errorf("ValidateCommitID(%q) = %v, want nil", v, err)
		}
	}

	invalid := []string{"", "abc", "xyz123", "not-hex-chars!"}
	for _, v := range invalid {
		if err := ValidateCommitID(v); err == nil {
			t.Errorf("ValidateCommitID(%q) = nil, want error", v)
		}
	}
}

func TestValidateBranchName(t *testing.T) {
	valid := []string{"main", "feature/my-branch", "release/v1.0", "a"}
	for _, v := range valid {
		if err := ValidateBranchName(v); err != nil {
			t.Errorf("ValidateBranchName(%q) = %v, want nil", v, err)
		}
	}

	invalid := []string{"", "-bad", "has//double", "trailing/", "../traversal", strings.Repeat("a", 257)}
	for _, v := range invalid {
		if err := ValidateBranchName(v); err == nil {
			t.Errorf("ValidateBranchName(%q) = nil, want error", v)
		}
	}
}

func TestValidatePath(t *testing.T) {
	valid := []string{"", "src/main.go", "path/to/file"}
	for _, v := range valid {
		if err := ValidatePath(v); err != nil {
			t.Errorf("ValidatePath(%q) = %v, want nil", v, err)
		}
	}

	invalid := []string{"/absolute", "path/../traversal", "has\x00null"}
	for _, v := range invalid {
		if err := ValidatePath(v); err == nil {
			t.Errorf("ValidatePath(%q) = nil, want error", v)
		}
	}
}

func TestValidateEnum(t *testing.T) {
	tests := []struct {
		fn    func(string) error
		valid []string
		bad   string
	}{
		{ValidatePRState, []string{"OPEN", "DECLINED", "MERGED", "ALL", "open"}, "INVALID"},
		{ValidatePRDirection, []string{"INCOMING", "OUTGOING", "incoming"}, "INVALID"},
		{ValidatePRRole, []string{"AUTHOR", "REVIEWER", "PARTICIPANT"}, "INVALID"},
		{ValidatePROrder, []string{"OLDEST", "NEWEST"}, "INVALID"},
		{ValidateTaskState, []string{"OPEN", "RESOLVED"}, "INVALID"},
	}
	for _, tt := range tests {
		for _, v := range tt.valid {
			if err := tt.fn(v); err != nil {
				t.Errorf("validate(%q) = %v, want nil", v, err)
			}
		}
		if err := tt.fn(tt.bad); err == nil {
			t.Errorf("validate(%q) = nil, want error", tt.bad)
		}
	}
}

func TestClampLimit(t *testing.T) {
	tests := []struct{ in, want int }{
		{0, 1}, {1, 1}, {25, 25}, {1000, 1000}, {1001, 1000}, {-5, 1},
	}
	for _, tt := range tests {
		if got := ClampLimit(tt.in); got != tt.want {
			t.Errorf("ClampLimit(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestClampContextLines(t *testing.T) {
	tests := []struct{ in, want int }{
		{-1, 0}, {0, 0}, {10, 10}, {100, 100}, {101, 100},
	}
	for _, tt := range tests {
		if got := ClampContextLines(tt.in); got != tt.want {
			t.Errorf("ClampContextLines(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
