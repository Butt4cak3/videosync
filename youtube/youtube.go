package youtube

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
	"videosync/media"

	"google.golang.org/api/googleapi"
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

func FetchVideoInfo(videoId string) (media.Video, error) {
	client, err := GetClient()
	if err != nil {
		return media.Video{}, err
	}
	fields := []googleapi.Field{
		"items/id",
		"items/snippet(publishedAt,title,channelTitle,thumbnails(medium(url)))",
		"items/contentDetails(duration)",
		"items/statistics(viewCount)",
	}
	res, err := client.Videos.List([]string{"snippet", "contentDetails", "statistics"}).Fields(fields...).Id(videoId).Do()
	if err != nil {
		return media.Video{}, err
	}
	if len(res.Items) == 0 {
		return media.Video{}, fmt.Errorf("video id \"%s\" not found", videoId)
	}
	item := res.Items[0]
	duration, err := parseDuration(item.ContentDetails.Duration)
	if err != nil {
		return media.Video{}, fmt.Errorf("could not parse video duration \"%s\": %v", item.ContentDetails.Duration, err)
	}
	video := media.Video{
		Id:          item.Id,
		Title:       item.Snippet.Title,
		Duration:    duration,
		Thumbnail:   item.Snippet.Thumbnails.Medium.Url,
		Channel:     item.Snippet.ChannelTitle,
		PublishedAt: item.Snippet.PublishedAt,
		Views:       item.Statistics.ViewCount,
	}
	return video, nil
}

func ParseUrl(urlString string) (string, bool) {
	u, err := url.Parse(urlString)
	if err != nil {
		return "", false
	}
	q := u.Query()
	switch u.Host {
	case "youtube.com", "www.youtube.com", "m.youtube.com":
		if u.Path == "/watch" && q.Has("v") {
			return q.Get("v"), true
		} else if strings.HasPrefix(u.Path, "/watch/") {
			return u.Path[7:], true
		} else if strings.HasPrefix(u.Path, "/v/") {
			return u.Path[3:], true
		} else if strings.HasPrefix(u.Path, "/shorts/") {
			return u.Path[8:], true
		}
	case "youtu.be":
		return u.Path[1:], true
	}
	return "", false
}

func parseDuration(ytDuration string) (time.Duration, error) {
	return time.ParseDuration(strings.ToLower(ytDuration[2:]))
}
