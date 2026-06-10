package db

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var allowedSortColumns = map[string]bool{
	"file_name":  true,
	"created_at": true,
	"updated_at": true,
	"trashed_at": true,
	"file_size":  true,
}

var allowedCategories = map[string]bool{
	"folder":   true,
	"video":    true,
	"audio":    true,
	"image":    true,
	"document": true,
	"archive":  true,
	"other":    true,
}

// ListFilesParams holds all optional filters for listing user files.
type ListFilesParams struct {
	UserID         int64
	ParentID       *int64
	MimePrefix     *string
	Category       *string
	SearchQuery    *string
	IsStarred      *bool
	IsTrashed      bool
	IncludeDirs    bool
	OnlyDirs       bool
	ExcludeSystem  bool
	IgnoreParentID bool   // true = skip parent_id filter (for recent files, search, etc.)
	SortBy         string // "file_name" | "created_at" | "updated_at" | "trashed_at"
	SortDir        string // "ASC" | "DESC"
	Page           int
	PageSize       int
}

func (p *ListFilesParams) normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 50
	}
	if !allowedSortColumns[p.SortBy] {
		p.SortBy = "created_at"
	}
	if p.SortDir != "ASC" && p.SortDir != "DESC" {
		p.SortDir = "DESC"
	}
	if p.Category != nil && !allowedCategories[*p.Category] {
		p.Category = nil
	}
	if p.SearchQuery != nil && *p.SearchQuery != "" {
		p.IgnoreParentID = true
	}
}

func buildWhere(p ListFilesParams) sq.And {
	w := sq.And{
		sq.Eq{"f.user_id": p.UserID},
		sq.Eq{"f.is_trashed": p.IsTrashed},
	}

	if !p.IgnoreParentID {
		if p.ParentID != nil {
			w = append(w, sq.Eq{"f.parent_id": *p.ParentID})
		} else if !p.IsTrashed {
			// browsing mode: root-level (parent_id IS NULL)
			w = append(w, sq.Eq{"f.parent_id": nil})
		}
	}

	if p.MimePrefix != nil {
		// Escape LIKE wildcards to prevent unintended pattern matching
		escaped := strings.ReplaceAll(*p.MimePrefix, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, "%", `\%`)
		escaped = strings.ReplaceAll(escaped, "_", `\_`)
		w = append(w, sq.Like{"f.mime_type": escaped + "%"})
	}
	if p.Category != nil {
		w = append(w, sq.Eq{"f.file_category": *p.Category})
	}
	if p.IsStarred != nil {
		w = append(w, sq.Eq{"f.is_starred": *p.IsStarred})
	}
	if p.OnlyDirs {
		w = append(w, sq.Eq{"f.is_dir": true})
	} else if !p.IncludeDirs {
		w = append(w, sq.Eq{"f.is_dir": false})
	}
	if p.ExcludeSystem {
		w = append(w, sq.Eq{"f.is_system": false})
	}

	if p.SearchQuery != nil && *p.SearchQuery != "" {
		escaped := strings.ReplaceAll(*p.SearchQuery, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, "%", `\%`)
		escaped = strings.ReplaceAll(escaped, "_", `\_`)
		w = append(w, sq.ILike{"f.file_name": "%" + escaped + "%"})
	}

	return w
}

// BuildListFilesQuery returns the SELECT and COUNT SQL strings with args.
func BuildListFilesQuery(p ListFilesParams) (sql string, args []any, countSql string, countArgs []any, err error) {
	p.normalize()

	where := buildWhere(p)

	orderBy := fmt.Sprintf("f.%s %s", p.SortBy, p.SortDir)
	// System directories always appear first
	if !p.ExcludeSystem {
		orderBy = fmt.Sprintf("f.is_system DESC, %s", orderBy)
	}
	// Directory browsing: always show folders first
	if !p.IsTrashed && !p.IgnoreParentID && p.ParentID != nil {
		orderBy = fmt.Sprintf("is_dir DESC, %s", orderBy)
	}

	offset := uint64((p.Page - 1) * p.PageSize)

	listQuery := psql.Select(
		"f.id", "f.slug", "f.file_name", "f.is_dir", "f.file_size",
		"f.mime_type", "f.file_category", "f.is_starred",
		"f.is_system", "f.system_kind",
		"f.created_at", "f.updated_at", "f.parent_slug",
		"p.file_name AS parent_name",
	).
		From("user_files f").
		LeftJoin("user_files p ON f.parent_id = p.id").
		Where(where).
		OrderBy(orderBy).
		Limit(uint64(p.PageSize)).
		Offset(offset)

	sql, args, err = listQuery.ToSql()
	if err != nil {
		return "", nil, "", nil, fmt.Errorf("build list query: %w", err)
	}

	countQuery := psql.Select("COUNT(*)").
		From("user_files f").
		Where(where)

	countSql, countArgs, err = countQuery.ToSql()
	if err != nil {
		return "", nil, "", nil, fmt.Errorf("build count query: %w", err)
	}

	return sql, args, countSql, countArgs, nil
}
