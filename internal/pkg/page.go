package pkg

func TotalPages(total, perPage int) int {
	return (total + perPage - 1) / perPage
}
