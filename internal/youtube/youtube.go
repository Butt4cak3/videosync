package youtube

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var client *youtube.Service

func GetClient() (*youtube.Service, error) {
	var err error
	if client == nil {
		ctx := context.Background()
		client, err = youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func GetVideoDuration(videoId string) (time.Duration, error) {
	client, err := GetClient()
	if err != nil {
		return 0, err
	}
	call := client.Videos.List([]string{"contentDetails"}).Id(videoId)
	res, err := call.Do()
	if err != nil {
		return 0, err
	}
	if len(res.Items) == 0 {
		return 0, fmt.Errorf("video id \"%s\" not found", videoId)
	}
	return parseDuration(res.Items[0].ContentDetails.Duration)
}

func ParseUrl(url string) (string, bool) {
	re := regexp.MustCompile("v=([^&]+)")
	match := re.FindStringSubmatch(url)
	if match == nil {
		return "", false
	}

	return match[1], true
}

func parseDuration(ytDuration string) (time.Duration, error) {
	return time.ParseDuration(strings.ToLower(ytDuration[2:]))
}
