package model

import (
	"fmt"

	"emotibot.com/emotigo/pkg/misc/mathutil"
)

// Pagination is the abstract layer for sql offset part.
//	Limit and Page are corresponding to limit and offset.
type Pagination struct {
	Limit int
	Page  int
}

// offsetSQL generate the raw limit sql.
func (p *Pagination) offsetSQL() string {
	limit := mathutil.MaxInt(p.Limit, 0)
	page := mathutil.MaxInt(p.Page-1, 0)
	offset := limit * page
	return fmt.Sprintf(" LIMIT %d, %d", offset, limit)
}
