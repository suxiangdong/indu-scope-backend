package pagination

const (
	DefaultPage     uint = 1
	DefaultPageSize uint = 20
	MaxPageSize     uint = 50
)

func Normalize(page, pageSize uint) (uint, uint) {
	if page == 0 {
		page = DefaultPage
	}
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}
