package url

import (
	"fmt"
	"net/url"
)

func GetLocalFilePath(rawURL string) (localFilePath string, err error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if url.Scheme != "file" || url.Path == "" {
		err = fmt.Errorf("%s is not for local file", rawURL)
		return
	}

	localFilePath = url.Path
	return
}
