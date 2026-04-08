package sharedrepo

func NormalizePageParams(page int, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}

	switch {
	case pageSize <= 0:
		pageSize = 20
	case pageSize > 100:
		pageSize = 100
	}

	return page, pageSize
}
