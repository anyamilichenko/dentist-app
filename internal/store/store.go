package store

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrInUse     = errors.New("referenced by other rows")
	ErrDuplicate = errors.New("duplicate")
	ErrOverlap   = errors.New("time slot overlaps another appointment")
	ErrBadRef    = errors.New("referenced row does not exist")
)

// Коды ошибок PostgreSQL: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	codeForeignKeyViolation = "23503"
	codeUniqueViolation     = "23505"
	codeExclusionViolation  = "23P01"
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// containsPattern превращает строку поиска в шаблон LIKE "%...%".
func containsPattern(s string) string {
	return "%" + likeEscaper.Replace(s) + "%"
}
