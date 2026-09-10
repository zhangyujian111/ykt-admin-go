package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// QueryRows 通用查询助手：把 sql.Rows 折叠成 []map[string]any。
func QueryRows(ctx context.Context, q Queryer, query string, args ...any) ([]map[string]any, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var out []map[string]any
	for rows.Next() {
		holders := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range holders {
			ptrs[i] = &holders[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, c := range cols {
			v := holders[i]
			if b, ok := v.([]byte); ok {
				row[c] = string(b)
			} else {
				row[c] = v
			}
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// TimeRange 通用时间范围参数。
type TimeRange struct {
	From time.Time
	To   time.Time
}

func (r TimeRange) Where(col string) (string, []any) {
	return fmt.Sprintf(" AND %s BETWEEN ? AND ? ", col), []any{r.From, r.To}
}
