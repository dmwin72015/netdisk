package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/storage"
)

type ShareService struct {
	pg    *pgxpool.Pool
	store *storage.Local
}

func NewShareService(pg *pgxpool.Pool, store *storage.Local) *ShareService {
	return &ShareService{pg: pg, store: store}
}

type CreateShareInput struct {
	FileSlugs    []string   `json:"fileSlugs"`
	PasswordCode *string    `json:"passwordCode"`
	ExpiresAt    *time.Time `json:"expiresAt"`
}

type UpdateShareInput struct {
	PasswordCode    *string    `json:"passwordCode"`
	PasswordCodeSet bool       `json:"-"`
	ExpiresAt       *time.Time `json:"expiresAt"`
	ExpiresAtSet    bool       `json:"-"`
}

type ShareFileItem struct {
	Slug         string  `json:"slug"`
	FileSlug     string  `json:"fileSlug"`
	FileName     string  `json:"fileName"`
	FileSize     int64   `json:"fileSize"`
	MimeType     *string `json:"mimeType"`
	FileCategory string  `json:"fileCategory"`
	ParentSlug   string  `json:"parentSlug"`
}

type ShareItem struct {
	Slug         string          `json:"slug"`
	Files        []ShareFileItem `json:"files"`
	HasPassword  bool            `json:"hasPassword"`
	PasswordCode *string         `json:"passwordCode,omitempty"`
	ExpiresAt    *string         `json:"expiresAt"`
	DisabledAt   *string         `json:"disabledAt"`
	CreatedAt    string          `json:"createdAt"`
	UpdatedAt    string          `json:"updatedAt"`
	IsExpired    bool            `json:"isExpired"`
}

type PublicShareInfo struct {
	Slug        string          `json:"slug"`
	Files       []ShareFileItem `json:"files"`
	HasPassword bool            `json:"hasPassword"`
	ExpiresAt   *string         `json:"expiresAt"`
	CreatedAt   string          `json:"createdAt"`
}

type sharedFileInfo struct {
	FileSlug     string `json:"fileSlug"`
	FileName     string `json:"fileName"`
	FileHash     string `json:"fileHash"`
	FileSize     int64  `json:"fileSize"`
	MimeType     string `json:"mimeType"`
	FileCategory string `json:"fileCategory"`
	ParentSlug   string `json:"parentSlug"`
}

type shareRow struct {
	Slug         string
	PasswordCode pgtype.Text
	ExpiresAt    pgtype.Timestamptz
	DisabledAt   pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	IsExpired    bool
	FilesJSON    []byte
}

const shareQuerySelect = `
	s.slug,
	s.password_code,
	s.expires_at,
	s.disabled_at,
	s.created_at,
	s.updated_at,
	(s.expires_at IS NOT NULL AND s.expires_at <= NOW()) AS is_expired,
	COALESCE(
		json_agg(
			json_build_object(
				'slug', f.slug,
				'fileSlug', f.slug,
				'fileName', f.file_name,
				'fileSize', f.file_size,
				'mimeType', f.mime_type,
				'fileCategory', f.file_category,
				'parentSlug', COALESCE(f.parent_slug, '')
			) ORDER BY f.id
		) FILTER (WHERE f.id IS NOT NULL),
		'[]'::json
	) AS files`

const shareQueryFrom = `
	FROM file_shares s
	LEFT JOIN file_share_items si ON si.share_id = s.id
	LEFT JOIN user_files f ON f.id = si.file_id`

func (s *ShareService) CreateShare(ctx context.Context, userID int64, input CreateShareInput) (ShareItem, error) {
	if len(input.FileSlugs) == 0 {
		return ShareItem{}, model.ErrInvalidInput
	}

	passwordCode, err := normalizeSharePassword(input.PasswordCode)
	if err != nil {
		return ShareItem{}, err
	}

	var fileIDs []int64
	for _, fileSlug := range input.FileSlugs {
		var fileID int64
		var isDir, isTrashed bool
		err := s.pg.QueryRow(ctx, `
			SELECT id, is_dir, is_trashed
			FROM user_files
			WHERE slug = $1 AND user_id = $2
		`, strings.TrimSpace(fileSlug), userID).Scan(&fileID, &isDir, &isTrashed)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ShareItem{}, model.ErrNotFound
			}
			return ShareItem{}, fmt.Errorf("get share file: %w", err)
		}
		if isDir || isTrashed {
			return ShareItem{}, model.ErrNotFound
		}
		fileIDs = append(fileIDs, fileID)
	}

	slug, err := gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", 21)
	if err != nil {
		return ShareItem{}, fmt.Errorf("generate share slug: %w", err)
	}

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return ShareItem{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var shareID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO file_shares (slug, user_id, password_code, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, slug, userID, textOrNil(passwordCode), timeOrNil(input.ExpiresAt)).Scan(&shareID)
	if err != nil {
		return ShareItem{}, fmt.Errorf("insert share: %w", err)
	}

	for _, fid := range fileIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO file_share_items (share_id, file_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, shareID, fid)
		if err != nil {
			return ShareItem{}, fmt.Errorf("insert share item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return ShareItem{}, fmt.Errorf("commit: %w", err)
	}

	return s.querySingleShare(ctx, slug, userID)
}

func (s *ShareService) ListShares(ctx context.Context, userID int64, limit int, offset int) ([]ShareItem, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`SELECT %s %s WHERE s.user_id = $1 AND s.deleted_at IS NULL GROUP BY s.id ORDER BY s.created_at DESC LIMIT $2 OFFSET $3`, shareQuerySelect, shareQueryFrom)
	rows, err := s.pg.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list shares: %w", err)
	}
	defer rows.Close()

	shares := make([]ShareItem, 0, limit)
	for rows.Next() {
		share, scanErr := scanShareRow(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		shares = append(shares, share)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate shares: %w", err)
	}

	var total int64
	if err := s.pg.QueryRow(ctx, `SELECT COUNT(*) FROM file_shares WHERE user_id = $1 AND deleted_at IS NULL`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count shares: %w", err)
	}

	return shares, total, nil
}

func (s *ShareService) querySingleShare(ctx context.Context, slug string, userID int64) (ShareItem, error) {
	query := fmt.Sprintf(`SELECT %s %s WHERE s.slug = $1 AND s.user_id = $2 AND s.deleted_at IS NULL GROUP BY s.id`, shareQuerySelect, shareQueryFrom)
	row := s.pg.QueryRow(ctx, query, slug, userID)
	return scanShareRow(row)
}

func (s *ShareService) UpdateShare(ctx context.Context, userID int64, slug string, input UpdateShareInput) (ShareItem, error) {
	shareSlug := strings.TrimSpace(slug)
	if shareSlug == "" || (!input.PasswordCodeSet && !input.ExpiresAtSet) {
		return ShareItem{}, model.ErrInvalidInput
	}

	var passwordCode *string
	var err error
	if input.PasswordCodeSet {
		passwordCode, err = normalizeSharePassword(input.PasswordCode)
		if err != nil {
			return ShareItem{}, err
		}
	}

	switch {
	case input.PasswordCodeSet && input.ExpiresAtSet:
		_, err = s.pg.Exec(ctx, `
			UPDATE file_shares
			SET password_code = $1, expires_at = $2, updated_at = NOW()
			WHERE slug = $3 AND user_id = $4
		`, textOrNil(passwordCode), timeOrNil(input.ExpiresAt), shareSlug, userID)
	case input.PasswordCodeSet:
		_, err = s.pg.Exec(ctx, `
			UPDATE file_shares
			SET password_code = $1, updated_at = NOW()
			WHERE slug = $2 AND user_id = $3
		`, textOrNil(passwordCode), shareSlug, userID)
	default:
		_, err = s.pg.Exec(ctx, `
			UPDATE file_shares
			SET expires_at = $1, updated_at = NOW()
			WHERE slug = $2 AND user_id = $3
		`, timeOrNil(input.ExpiresAt), shareSlug, userID)
	}
	if err != nil {
		return ShareItem{}, fmt.Errorf("update share: %w", err)
	}

	share, err := s.querySingleShare(ctx, shareSlug, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ShareItem{}, model.ErrNotFound
		}
		return ShareItem{}, fmt.Errorf("scan updated share: %w", err)
	}
	return share, nil
}

func (s *ShareService) CancelShare(ctx context.Context, userID int64, slug string) error {
	shareSlug := strings.TrimSpace(slug)
	if shareSlug == "" {
		return model.ErrInvalidInput
	}
	commandTag, err := s.pg.Exec(ctx, `
		UPDATE file_shares
		SET disabled_at = COALESCE(disabled_at, NOW()), updated_at = NOW()
		WHERE slug = $1 AND user_id = $2 AND deleted_at IS NULL
	`, shareSlug, userID)
	if err != nil {
		return fmt.Errorf("cancel share: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (s *ShareService) DeleteShare(ctx context.Context, userID int64, slug string) error {
	shareSlug := strings.TrimSpace(slug)
	if shareSlug == "" {
		return model.ErrInvalidInput
	}
	commandTag, err := s.pg.Exec(ctx, `
		UPDATE file_shares
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE slug = $1 AND user_id = $2
	`, shareSlug, userID)
	if err != nil {
		return fmt.Errorf("delete share: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (s *ShareService) GetPublicInfo(ctx context.Context, slug string) (PublicShareInfo, error) {
	_, info, err := s.getPublicShareData(ctx, slug, nil, false)
	return info, err
}

func (s *ShareService) VerifyPublicPassword(ctx context.Context, slug string, passwordCode string) (PublicShareInfo, error) {
	passwordCode = strings.TrimSpace(passwordCode)
	record, info, err := s.getPublicShareData(ctx, slug, nil, false)
	if err != nil {
		return PublicShareInfo{}, err
	}
	if !matchesSharePassword(record.PasswordCode, passwordCode) {
		return PublicShareInfo{}, model.ErrForbidden
	}
	return info, nil
}

type SharedFileResult struct {
	File     *os.File
	Name     string
	MimeType string
	FileHash string
}

func (s *ShareService) OpenSharedFile(ctx context.Context, slug string, passwordCode string, fileSlug string) (*SharedFileResult, error) {
	passwordCode = strings.TrimSpace(passwordCode)

	record, _, err := s.getPublicShareData(ctx, slug, nil, false)
	if err != nil {
		return nil, err
	}
	if !matchesSharePassword(record.PasswordCode, passwordCode) {
		return nil, model.ErrForbidden
	}

	targetFileSlug := strings.TrimSpace(fileSlug)
	targetFile := record.Files[0]
	if targetFileSlug != "" {
		for _, f := range record.Files {
			if f.FileSlug == targetFileSlug {
				targetFile = f
				break
			}
		}
		if targetFile.FileSlug == "" && targetFileSlug != record.Files[0].FileSlug {
			return nil, model.ErrNotFound
		}
	}

	file, err := s.store.Open(targetFile.FileHash)
	if err != nil {
		return nil, fmt.Errorf("open shared file: %w", err)
	}

	mimeType := targetFile.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return &SharedFileResult{
		File:     file,
		Name:     safeFilename(targetFile.FileName),
		MimeType: mimeType,
		FileHash: targetFile.FileHash,
	}, nil
}

type publicShareRecord struct {
	Slug         string
	PasswordCode pgtype.Text
	ExpiresAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	Files        []sharedFileInfo
}

func (s *ShareService) getPublicShareData(ctx context.Context, slug string, passwordCode *string, verifyOnly bool) (publicShareRecord, PublicShareInfo, error) {
	shareSlug := strings.TrimSpace(slug)
	if shareSlug == "" {
		return publicShareRecord{}, PublicShareInfo{}, model.ErrInvalidInput
	}

	query := `
		SELECT
			s.slug,
			s.password_code,
			s.expires_at,
			s.created_at,
			COALESCE(
		json_agg(
			json_build_object(
				'fileSlug', f.slug,
				'fileName', f.file_name,
				'fileSize', f.file_size,
				'mimeType', COALESCE(f.mime_type, ''),
				'fileCategory', f.file_category,
				'parentSlug', COALESCE(f.parent_slug, ''),
				'fileHash', pf.file_hash
					) ORDER BY f.id
				) FILTER (WHERE f.id IS NOT NULL),
				'[]'::json
			) AS files
		FROM file_shares s
		JOIN file_share_items si ON si.share_id = s.id
		JOIN user_files f ON f.id = si.file_id
		JOIN physical_files pf ON pf.id = f.physical_file_id
		WHERE s.slug = $1
		  AND s.disabled_at IS NULL
		  AND s.deleted_at IS NULL
		  AND f.is_trashed = FALSE
		  AND f.is_dir = FALSE
		  AND (s.expires_at IS NULL OR s.expires_at > NOW())
		GROUP BY s.id
	`

	var record publicShareRecord
	var filesJSON []byte
	err := s.pg.QueryRow(ctx, query, shareSlug).Scan(
		&record.Slug,
		&record.PasswordCode,
		&record.ExpiresAt,
		&record.CreatedAt,
		&filesJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return publicShareRecord{}, PublicShareInfo{}, model.ErrNotFound
		}
		return publicShareRecord{}, PublicShareInfo{}, fmt.Errorf("get public share: %w", err)
	}

	if err := json.Unmarshal(filesJSON, &record.Files); err != nil {
		return publicShareRecord{}, PublicShareInfo{}, fmt.Errorf("parse share files: %w", err)
	}

	if len(record.Files) == 0 {
		return publicShareRecord{}, PublicShareInfo{}, model.ErrNotFound
	}

	info := PublicShareInfo{
		Slug:        record.Slug,
		HasPassword: record.PasswordCode.Valid && record.PasswordCode.String != "",
		ExpiresAt:   formatOptionalTime(record.ExpiresAt),
		CreatedAt:   formatRequiredTime(record.CreatedAt),
	}
	for _, f := range record.Files {
		mimeType := f.MimeType
		mimePtr := &mimeType
		if mimeType == "" {
			mimePtr = nil
		}
		info.Files = append(info.Files, ShareFileItem{
			Slug:         f.FileSlug,
			FileSlug:     f.FileSlug,
			FileName:     f.FileName,
			FileSize:     f.FileSize,
			MimeType:     mimePtr,
			FileCategory: f.FileCategory,
			ParentSlug:   f.ParentSlug,
		})
	}

	return record, info, nil
}

func scanShareRow(scanner interface{ Scan(dest ...any) error }) (ShareItem, error) {
	var row shareRow
	err := scanner.Scan(
		&row.Slug,
		&row.PasswordCode,
		&row.ExpiresAt,
		&row.DisabledAt,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.IsExpired,
		&row.FilesJSON,
	)
	if err != nil {
		return ShareItem{}, fmt.Errorf("scan share: %w", err)
	}

	var files []ShareFileItem
	if err := json.Unmarshal(row.FilesJSON, &files); err != nil {
		return ShareItem{}, fmt.Errorf("parse share files: %w", err)
	}

	share := ShareItem{
		Slug:       row.Slug,
		Files:      files,
		IsExpired:  row.IsExpired,
		ExpiresAt:  formatOptionalTime(row.ExpiresAt),
		DisabledAt: formatOptionalTime(row.DisabledAt),
		CreatedAt:  formatRequiredTime(row.CreatedAt),
		UpdatedAt:  formatRequiredTime(row.UpdatedAt),
	}
	if row.PasswordCode.Valid && row.PasswordCode.String != "" {
		share.HasPassword = true
		share.PasswordCode = &row.PasswordCode.String
	}

	return share, nil
}

func normalizeSharePassword(passwordCode *string) (*string, error) {
	if passwordCode == nil {
		return nil, nil
	}
	value := strings.TrimSpace(*passwordCode)
	if value == "" {
		return nil, nil
	}
	if len(value) > 16 {
		return nil, model.ErrInvalidInput
	}
	return &value, nil
}

func matchesSharePassword(expected pgtype.Text, actual string) bool {
	if !expected.Valid || expected.String == "" {
		return true
	}
	return expected.String == strings.TrimSpace(actual)
}

func textOrNil(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func timeOrNil(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}

func formatOptionalTime(value pgtype.Timestamptz) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.UTC().Format(time.RFC3339)
	return &formatted
}

func formatRequiredTime(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}
