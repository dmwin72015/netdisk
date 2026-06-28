package service

import "github.com/jackc/pgx/v5/pgtype"

func textStr(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func strText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}
