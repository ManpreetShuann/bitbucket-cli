package validation

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	projectKeyRE = regexp.MustCompile(`^~?[A-Za-z0-9_]{1,128}$`)
	repoSlugRE   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	commitIDRE   = regexp.MustCompile(`^[0-9a-fA-F]{4,40}$`)

	validPRStates      = map[string]bool{"OPEN": true, "DECLINED": true, "MERGED": true, "ALL": true}
	validPRDirections  = map[string]bool{"INCOMING": true, "OUTGOING": true}
	validPRRoles       = map[string]bool{"AUTHOR": true, "REVIEWER": true, "PARTICIPANT": true}
	validPROrders      = map[string]bool{"OLDEST": true, "NEWEST": true}
	validTaskStates    = map[string]bool{"OPEN": true, "RESOLVED": true}
)

const (
	MaxLimit        = 1000
	MaxContextLines = 100
	MaxBranchLen    = 256
)

func ValidateProjectKey(key string) error {
	if !projectKeyRE.MatchString(key) {
		return fmt.Errorf("invalid project key: %q (must be alphanumeric/underscores, 1-128 chars, optional ~ prefix)", key)
	}
	return nil
}

func ValidateRepoSlug(slug string) error {
	if !repoSlugRE.MatchString(slug) {
		return fmt.Errorf("invalid repo slug: %q (must start with alphanumeric, contain only [A-Za-z0-9._-])", slug)
	}
	return nil
}

func ValidateCommitID(id string) error {
	if !commitIDRE.MatchString(id) {
		return fmt.Errorf("invalid commit ID: %q (must be a hex SHA, 4-40 chars)", id)
	}
	return nil
}

func ValidateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name must not be empty")
	}
	if len(name) > MaxBranchLen {
		return fmt.Errorf("branch name too long: %d chars (max %d)", len(name), MaxBranchLen)
	}
	if name[0] < '0' || (name[0] > '9' && name[0] < 'A') || (name[0] > 'Z' && name[0] < 'a') || name[0] > 'z' {
		return fmt.Errorf("branch name must start with alphanumeric character: %q", name)
	}
	if strings.Contains(name, "//") {
		return fmt.Errorf("branch name must not contain '//': %q", name)
	}
	if strings.HasSuffix(name, "/") {
		return fmt.Errorf("branch name must not end with '/': %q", name)
	}
	for _, seg := range strings.Split(name, "/") {
		if seg == ".." {
			return fmt.Errorf("branch name must not contain path traversal: %q", name)
		}
	}
	return nil
}

func ValidateTagName(name string) error {
	return ValidateBranchName(name)
}

func ValidatePath(path string) error {
	if path == "" {
		return nil
	}
	if strings.ContainsRune(path, 0) {
		return fmt.Errorf("path must not contain null bytes")
	}
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must not start with '/'")
	}
	for _, seg := range strings.Split(path, "/") {
		if seg == ".." {
			return fmt.Errorf("path traversal ('..') is not permitted")
		}
	}
	return nil
}

func ValidatePositiveInt(value int, name string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be a positive integer, got %d", name, value)
	}
	return nil
}

func validateEnum(value string, valid map[string]bool, name string) error {
	upper := strings.ToUpper(value)
	if !valid[upper] {
		keys := make([]string, 0, len(valid))
		for k := range valid {
			keys = append(keys, k)
		}
		return fmt.Errorf("invalid %s: %q (must be one of %v)", name, value, keys)
	}
	return nil
}

func ValidatePRState(state string) error     { return validateEnum(state, validPRStates, "PR state") }
func ValidatePRDirection(dir string) error    { return validateEnum(dir, validPRDirections, "PR direction") }
func ValidatePRRole(role string) error        { return validateEnum(role, validPRRoles, "PR role") }
func ValidatePROrder(order string) error      { return validateEnum(order, validPROrders, "PR order") }
func ValidateTaskState(state string) error    { return validateEnum(state, validTaskStates, "task state") }

func ClampLimit(limit int) int {
	if limit < 1 {
		return 1
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

func ClampStart(start int) int {
	if start < 0 {
		return 0
	}
	return start
}

func ClampContextLines(lines int) int {
	if lines < 0 {
		return 0
	}
	if lines > MaxContextLines {
		return MaxContextLines
	}
	return lines
}
