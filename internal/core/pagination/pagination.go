package pagination

const defaultLimit = 20

// Page holds the resolved pagination values ready to pass to a repository.
type Page struct {
	Limit   int
	Offset  int
	Current int
}

// Parse resolves page/limit to safe defaults and computes the offset.
func Parse(page, limit int) Page {
	if limit <= 0 {
		limit = defaultLimit
	}
	if page <= 0 {
		page = 1
	}
	return Page{
		Limit:   limit,
		Offset:  (page - 1) * limit,
		Current: page,
	}
}
