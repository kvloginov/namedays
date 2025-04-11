package fetch

import "github.com/kvloginov/namedays/internal/domain"

type Fetcher interface {
	FetchAllNamedays() (domain.NamedaysDataList, error)
}
