package client

// Project represents a Bitbucket Server project.
type Project struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
	Type        string `json:"type"`
	Links       Links  `json:"links"`
}

// Repository represents a Bitbucket Server repository.
type Repository struct {
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	State       string  `json:"state"`
	ScmID       string  `json:"scmId"`
	Project     Project `json:"project"`
	Forkable    bool    `json:"forkable"`
	Links       Links   `json:"links"`
}

// PullRequest represents a Bitbucket Server pull request.
type PullRequest struct {
	ID          int           `json:"id"`
	Version     int           `json:"version"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	State       string        `json:"state"`
	Draft       bool          `json:"draft"`
	Author      Participant   `json:"author"`
	Reviewers   []Participant `json:"reviewers"`
	FromRef     Ref           `json:"fromRef"`
	ToRef       Ref           `json:"toRef"`
	CreatedDate int64         `json:"createdDate"`
	UpdatedDate int64         `json:"updatedDate"`
	Links       Links         `json:"links"`
}

// Participant represents a PR participant with role and approval status.
type Participant struct {
	User     User   `json:"user"`
	Role     string `json:"role"`
	Approved bool   `json:"approved"`
	Status   string `json:"status"`
}

// User represents a Bitbucket Server user.
type User struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Slug         string `json:"slug"`
	Active       bool   `json:"active"`
	Links        Links  `json:"links"`
}

// Ref represents a git reference (branch/tag).
type Ref struct {
	ID           string     `json:"id"`
	DisplayID    string     `json:"displayId"`
	LatestCommit string     `json:"latestCommit"`
	Repository   Repository `json:"repository"`
}

// Branch represents a branch in a repository.
type Branch struct {
	ID              string `json:"id"`
	DisplayID       string `json:"displayId"`
	LatestCommit    string `json:"latestCommit"`
	IsDefault       bool   `json:"isDefault"`
}

// Tag represents a git tag.
type Tag struct {
	ID           string `json:"id"`
	DisplayID    string `json:"displayId"`
	LatestCommit string `json:"latestCommit"`
	Hash         string `json:"hash"`
}

// Commit represents a git commit.
type Commit struct {
	ID              string   `json:"id"`
	DisplayID       string   `json:"displayId"`
	Message         string   `json:"message"`
	Author          Person   `json:"author"`
	AuthorTimestamp int64    `json:"authorTimestamp"`
	Committer       Person   `json:"committer"`
	Parents         []Commit `json:"parents"`
}

// Person represents a git author/committer.
type Person struct {
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
}

// Comment represents a PR comment.
type Comment struct {
	ID          int       `json:"id"`
	Version     int       `json:"version"`
	Text        string    `json:"text"`
	Author      User      `json:"author"`
	Severity    string    `json:"severity"`
	State       string    `json:"state"`
	Anchor      *Anchor   `json:"anchor,omitempty"`
	Comments    []Comment `json:"comments"`
	CreatedDate int64     `json:"createdDate"`
	UpdatedDate int64     `json:"updatedDate"`
}

// Anchor represents the file/line location of an inline comment.
type Anchor struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	LineType string `json:"lineType"`
	FileType string `json:"fileType"`
}

// Task represents a PR task.
type Task struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	State string `json:"state"`
}

// Activity represents a PR activity feed entry.
type Activity struct {
	ID          int      `json:"id"`
	Action      string   `json:"action"`
	Comment     *Comment `json:"comment,omitempty"`
	CreatedDate int64    `json:"createdDate"`
	User        User     `json:"user"`
}

// Links holds HATEOAS links from the API.
type Links struct {
	Clone []Link `json:"clone"`
	Self  []Link `json:"self"`
}

// Link is a single HATEOAS link.
type Link struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

// MergeStatus represents the merge readiness check result.
type MergeStatus struct {
	CanMerge   bool   `json:"canMerge"`
	Conflicted bool   `json:"conflicted"`
	Vetoes     []Veto `json:"vetoes"`
}

// Veto represents a merge veto reason.
type Veto struct {
	SummaryMessage  string `json:"summaryMessage"`
	DetailedMessage string `json:"detailedMessage"`
}

// FileEntry represents a file or directory in a browse response.
type FileEntry struct {
	Path      FileEntryPath `json:"path,omitempty"`
	Type      string        `json:"type"`
	Size      int64         `json:"size"`
	ContentID string        `json:"contentId"`
}

// FileEntryPath contains the path components of a file entry.
type FileEntryPath struct {
	Components []string `json:"components"`
	Name       string   `json:"name"`
	ToString   string   `json:"toString"`
}

// Change represents a file changed in a commit.
type Change struct {
	ContentID  string         `json:"contentId"`
	FromHash   string         `json:"fromHash"`
	ToHash     string         `json:"toHash"`
	Path       FileEntryPath  `json:"path"`
	SrcPath    *FileEntryPath `json:"srcPath,omitempty"`
	Type       string         `json:"type"`
	NodeType   string         `json:"nodeType"`
}

// SearchResult represents a code search result.
type SearchResult struct {
	File        SearchFile    `json:"file,omitempty"`
	HitCount    int           `json:"hitCount"`
	PathMatches []SearchMatch `json:"pathMatches,omitempty"`
	HitContexts []HitContext  `json:"hitContexts,omitempty"`
}

// SearchFile represents the file info in a search result.
type SearchFile struct {
	Path       string     `json:"path"`
	Repository Repository `json:"repository"`
}

// SearchMatch represents a matched path segment.
type SearchMatch struct {
	Text  string `json:"text"`
	Match bool   `json:"match"`
}

// HitContext represents surrounding lines of a code search hit.
type HitContext struct {
	Lines []HitLine `json:"lines"`
}

// HitLine represents a single line in a code search hit context.
type HitLine struct {
	Text string `json:"text"`
	Line int    `json:"line"`
}

// Diff represents a diff response.
type Diff struct {
	Diffs []DiffEntry `json:"diffs"`
}

// DiffEntry represents a single file diff.
type DiffEntry struct {
	Source      *DiffPath  `json:"source,omitempty"`
	Destination *DiffPath `json:"destination,omitempty"`
	Hunks       []Hunk    `json:"hunks"`
	Truncated   bool      `json:"truncated"`
}

// DiffPath represents a path in a diff.
type DiffPath struct {
	ToString string `json:"toString"`
}

// Hunk represents a diff hunk.
type Hunk struct {
	SourceLine      int       `json:"sourceLine"`
	SourceSpan      int       `json:"sourceSpan"`
	DestinationLine int       `json:"destinationLine"`
	DestinationSpan int       `json:"destinationSpan"`
	Segments        []Segment `json:"segments"`
}

// Segment represents a diff segment (context, added, removed).
type Segment struct {
	Type  string     `json:"type"`
	Lines []DiffLine `json:"lines"`
}

// DiffLine represents a single line in a diff.
type DiffLine struct {
	Source      int    `json:"source"`
	Destination int   `json:"destination"`
	Line        string `json:"line"`
}

// DiffStatEntry represents a single file in diffstat response.
type DiffStatEntry struct {
	Path     FileEntryPath  `json:"path"`
	SrcPath  *FileEntryPath `json:"srcPath,omitempty"`
	Type     string         `json:"type"`
	NodeType string         `json:"nodeType"`
}
