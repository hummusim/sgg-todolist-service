package repository

import sq "github.com/Masterminds/squirrel"

var (
	limitPage uint64 = 20
	limitOne  uint64 = 1
)

var (
	// Squirrel for Postgres
	psql sq.StatementBuilderType = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)

// GetOffset get a offset from page and pagesize
func GetOffset(page int32, pageSize uint64) uint64 {
	var value int32 = page

	if value < 1 {
		value = 1
	}
	return uint64((value - 1)) * pageSize
}
