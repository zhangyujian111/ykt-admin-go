package repo

import (
	"context"
	"database/sql"
	"errors"

	"ykt.dev/admin/internal/auth"
)

type User struct {
	UserID   int64
	Username string
	Nickname sql.NullString
	Password string
	Status   string
	DeptID   sql.NullInt64
}

type UserRepo struct{ DB *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{DB: db} }

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*User, error) {
	u := &User{}
	row := r.DB.QueryRowContext(ctx, `
		SELECT userId, username, nickname, password, status, deptId
		FROM ykt_sys_user
		WHERE username = ?
		LIMIT 1`, username)
	if err := row.Scan(&u.UserID, &u.Username, &u.Nickname, &u.Password, &u.Status, &u.DeptID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) VerifyPassword(u *User, plain string) bool {
	return u != nil && auth.Verify(u.Password, plain)
}
