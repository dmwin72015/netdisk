package db

import (
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name         string
		input        ListFilesParams
		wantPage     int
		wantPageSize int
		wantSortBy   string
		wantSortDir  string
		wantCategory *string
	}{
		{
			name:         "defaults for zero values",
			input:        ListFilesParams{},
			wantPage:     1,
			wantPageSize: 50,
			wantSortBy:   "created_at",
			wantSortDir:  "DESC",
			wantCategory: nil,
		},
		{
			name:         "valid values preserved",
			input:        ListFilesParams{Page: 3, PageSize: 25, SortBy: "file_name", SortDir: "ASC"},
			wantPage:     3,
			wantPageSize: 25,
			wantSortBy:   "file_name",
			wantSortDir:  "ASC",
		},
		{
			name:         "page clamped to 1",
			input:        ListFilesParams{Page: -5, PageSize: 10, SortBy: "file_size", SortDir: "ASC"},
			wantPage:     1,
			wantPageSize: 10,
			wantSortBy:   "file_size",
			wantSortDir:  "ASC",
		},
		{
			name:         "page size too large resets to 50",
			input:        ListFilesParams{Page: 1, PageSize: 500},
			wantPage:     1,
			wantPageSize: 50,
			wantSortBy:   "created_at",
			wantSortDir:  "DESC",
		},
		{
			name:         "page size zero resets to 50",
			input:        ListFilesParams{Page: 1, PageSize: 0},
			wantPage:     1,
			wantPageSize: 50,
			wantSortBy:   "created_at",
			wantSortDir:  "DESC",
		},
		{
			name:         "invalid sort column resets to created_at",
			input:        ListFilesParams{SortBy: "hack_col", SortDir: "ASC"},
			wantPage:     1,
			wantPageSize: 50,
			wantSortBy:   "created_at",
			wantSortDir:  "ASC",
		},
		{
			name:         "invalid sort dir resets to DESC",
			input:        ListFilesParams{SortBy: "file_name", SortDir: "RANDOM"},
			wantPage:     1,
			wantPageSize: 50,
			wantSortBy:   "file_name",
			wantSortDir:  "DESC",
		},
		{
			name:         "empty sort dir resets to DESC",
			input:        ListFilesParams{SortBy: "file_name", SortDir: ""},
			wantPage:     1,
			wantPageSize: 50,
			wantSortBy:   "file_name",
			wantSortDir:  "DESC",
		},
		{
			name:         "valid category preserved",
			input:        ListFilesParams{Category: strPtr("video")},
			wantPage:     1,
			wantPageSize: 50,
			wantSortBy:   "created_at",
			wantSortDir:  "DESC",
			wantCategory: strPtr("video"),
		},
		{
			name:         "invalid category set to nil",
			input:        ListFilesParams{Category: strPtr("malware")},
			wantPage:     1,
			wantPageSize: 50,
			wantSortBy:   "created_at",
			wantSortDir:  "DESC",
			wantCategory: nil,
		},
		{
			name:         "page size exactly 100 is valid",
			input:        ListFilesParams{PageSize: 100},
			wantPage:     1,
			wantPageSize: 100,
			wantSortBy:   "created_at",
			wantSortDir:  "DESC",
		},
		{
			name:         "all valid sort columns",
			input:        ListFilesParams{SortBy: "trashed_at", SortDir: "ASC"},
			wantPage:     1,
			wantPageSize: 50,
			wantSortBy:   "trashed_at",
			wantSortDir:  "ASC",
		},
		{
			name:         "all valid categories",
			input:        ListFilesParams{Category: strPtr("archive")},
			wantPage:     1,
			wantPageSize: 50,
			wantSortBy:   "created_at",
			wantSortDir:  "DESC",
			wantCategory: strPtr("archive"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.input
			p.normalize()
			if p.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
			}
			if p.PageSize != tt.wantPageSize {
				t.Errorf("PageSize = %d, want %d", p.PageSize, tt.wantPageSize)
			}
			if p.SortBy != tt.wantSortBy {
				t.Errorf("SortBy = %q, want %q", p.SortBy, tt.wantSortBy)
			}
			if p.SortDir != tt.wantSortDir {
				t.Errorf("SortDir = %q, want %q", p.SortDir, tt.wantSortDir)
			}
			if tt.wantCategory == nil {
				if p.Category != nil {
					t.Errorf("Category = %v, want nil", *p.Category)
				}
			} else {
				if p.Category == nil || *p.Category != *tt.wantCategory {
					t.Errorf("Category = %v, want %v", p.Category, *tt.wantCategory)
				}
			}
		})
	}
}

func TestBuildListFilesQuery_BasicOutput(t *testing.T) {
	params := ListFilesParams{
		UserID:    42,
		ParentID:  int64Ptr(10),
		IsTrashed: false,
		Page:      1,
		PageSize:  20,
	}

	sql, args, countSql, countArgs, err := BuildListFilesQuery(params)
	if err != nil {
		t.Fatalf("BuildListFilesQuery() error: %v", err)
	}

	// SQL should reference the correct table and user_id
	if !strings.Contains(sql, "user_files") {
		t.Errorf("list SQL missing table reference: %s", sql)
	}
	if !strings.Contains(countSql, "COUNT(*)") {
		t.Errorf("count SQL missing COUNT(*): %s", countSql)
	}

	// Args should contain user_id 42 and parent_id 10
	found42 := false
	found10 := false
	for _, a := range args {
		if a == int64(42) {
			found42 = true
		}
		if a == int64(10) {
			found10 = true
		}
	}
	if !found42 {
		t.Errorf("args missing user_id 42: %v", args)
	}
	if !found10 {
		t.Errorf("args missing parent_id 10: %v", args)
	}

	// Count args should also have user_id and parent_id
	found42 = false
	found10 = false
	for _, a := range countArgs {
		if a == int64(42) {
			found42 = true
		}
		if a == int64(10) {
			found10 = true
		}
	}
	if !found42 {
		t.Errorf("countArgs missing user_id 42: %v", countArgs)
	}
	if !found10 {
		t.Errorf("countArgs missing parent_id 10: %v", countArgs)
	}
}

func TestBuildListFilesQuery_SortOrder(t *testing.T) {
	// With parent_id set and not trashed: should have "is_dir DESC" prefix
	params := ListFilesParams{
		UserID:    1,
		ParentID:  int64Ptr(5),
		IsTrashed: false,
		SortBy:    "file_name",
		SortDir:   "ASC",
		Page:      1,
		PageSize:  10,
	}

	sql, _, _, _, err := BuildListFilesQuery(params)
	if err != nil {
		t.Fatalf("BuildListFilesQuery() error: %v", err)
	}

	if !strings.Contains(sql, "is_dir DESC") {
		t.Errorf("expected 'is_dir DESC' in ORDER BY for directory browsing, got: %s", sql)
	}
	if !strings.Contains(sql, "file_name ASC") {
		t.Errorf("expected 'file_name ASC' in ORDER BY, got: %s", sql)
	}

	// Without parent_id: no is_dir prefix
	params.ParentID = nil
	params.IgnoreParentID = true
	sql, _, _, _, err = BuildListFilesQuery(params)
	if err != nil {
		t.Fatalf("BuildListFilesQuery() error: %v", err)
	}

	if strings.Contains(sql, "is_dir DESC") {
		t.Errorf("unexpected 'is_dir DESC' when no parent_id, got: %s", sql)
	}
}

func TestBuildListFilesQuery_MimePrefixEscaping(t *testing.T) {
	// LIKE wildcards in mime prefix should be escaped
	mimePrefix := "image/%test_value"
	params := ListFilesParams{
		UserID:         1,
		MimePrefix:     &mimePrefix,
		IsTrashed:      false,
		IgnoreParentID: true,
		Page:           1,
		PageSize:       10,
	}

	sql, args, _, _, err := BuildListFilesQuery(params)
	if err != nil {
		t.Fatalf("BuildListFilesQuery() error: %v", err)
	}

	// The escaped LIKE pattern should appear in the args
	found := false
	for _, a := range args {
		if s, ok := a.(string); ok && strings.Contains(s, `\%`) && strings.Contains(s, `\_`) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected escaped LIKE wildcards in args, SQL: %s, args: %v", sql, args)
	}
}

func TestBuildListFilesQuery_TrashedFilter(t *testing.T) {
	params := ListFilesParams{
		UserID:         1,
		IsTrashed:      true,
		IgnoreParentID: true,
		Page:           1,
		PageSize:       10,
	}

	_, args, _, _, err := BuildListFilesQuery(params)
	if err != nil {
		t.Fatalf("BuildListFilesQuery() error: %v", err)
	}

	// is_trashed should be true
	found := false
	for _, a := range args {
		if b, ok := a.(bool); ok && b {
			found = true
		}
	}
	if !found {
		t.Errorf("expected is_trashed=true in args: %v", args)
	}
}

func TestBuildListFilesQuery_OnlyDirs(t *testing.T) {
	params := ListFilesParams{
		UserID:      1,
		OnlyDirs:    true,
		IncludeDirs: true,
		Page:        1,
		PageSize:    10,
	}

	sql, args, _, _, err := BuildListFilesQuery(params)
	if err != nil {
		t.Fatalf("BuildListFilesQuery() error: %v", err)
	}

	if !strings.Contains(sql, "is_dir") {
		t.Errorf("expected is_dir filter in SQL: %s", sql)
	}

	found := false
	for _, a := range args {
		if b, ok := a.(bool); ok && b {
			found = true
		}
	}
	if !found {
		t.Errorf("expected is_dir=true in args: %v", args)
	}
}

func strPtr(s string) *string {
	return &s
}

func int64Ptr(n int64) *int64 {
	return &n
}
