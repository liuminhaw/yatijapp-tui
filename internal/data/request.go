package data

import "net/url"

type ListRequestInfo struct {
	ServerURL    string
	QueryStrings map[string]string
	SrcUUID      string
}

func (i ListRequestInfo) requestUrl(path string) (string, error) {
	u, err := url.Parse(i.ServerURL)
	if err != nil {
		return "", err
	}
	u.Path = path
	q := u.Query()
	for key, value := range i.QueryStrings {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}
