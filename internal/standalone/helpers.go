package standalone

import (
	"database/sql"
	"time"
)

func nowStr() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z", "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

func parseTimePtr(ns sql.NullString) *time.Time {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	t := parseTime(ns.String)
	return &t
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func isFKError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, sub := range []string{"FOREIGN KEY", "foreign key"} {
		for i := 0; i+len(sub) <= len(msg); i++ {
			if msg[i:i+len(sub)] == sub {
				return true
			}
		}
	}
	return false
}
