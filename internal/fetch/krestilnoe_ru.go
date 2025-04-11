package fetch

import (
	"net/http"
	"time"

	"github.com/kvloginov/namedays/internal/domain"
)

type KrestilnoeFetcher struct {
	baseURL string
	client  *http.Client
}

func NewKrestilnoeFetcher() *KrestilnoeFetcher {
	return &KrestilnoeFetcher{
		baseURL: "https://krestilnoe.ru/svyattsy-kalendar-god/",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f *KrestilnoeFetcher) FetchAllNamedays() (domain.NamedaysDataList, error) {
	// not implemented
	return nil, nil
}
