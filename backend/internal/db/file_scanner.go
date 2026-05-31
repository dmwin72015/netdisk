package db

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// FileRow holds the columns needed for file listing.
type FileRow struct {
	ID           int64              `json:"id"`
	Slug         string             `json:"slug"`
	FileName     string             `json:"fileName"`
	IsDir        bool               `json:"isDir"`
	FileSize     int64              `json:"fileSize"`
	MimeType     pgtype.Text        `json:"mimeType"`
	FileCategory string             `json:"fileCategory"`
	IsStarred    bool               `json:"isStarred"`
	IsSystem     bool               `json:"isSystem"`
	SystemKind   pgtype.Text        `json:"systemKind"`
	CreatedAt    pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt    pgtype.Timestamptz `json:"updatedAt"`
	ParentSlug   pgtype.Text        `json:"parentSlug"`
	ParentName   pgtype.Text        `json:"parentName"`
}

// ScanFileRows collects all rows from pgx.Rows into []FileRow.
func ScanFileRows(rows pgx.Rows) ([]FileRow, error) {
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FileRow, error) {
		var f FileRow
		err := row.Scan(
			&f.ID,
			&f.Slug,
			&f.FileName,
			&f.IsDir,
			&f.FileSize,
			&f.MimeType,
			&f.FileCategory,
			&f.IsStarred,
			&f.IsSystem,
			&f.SystemKind,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.ParentSlug,
			&f.ParentName,
		)
		return f, err
	})
}
