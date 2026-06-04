package youtube

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
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

func ParseUrl(urlString string) (videoId string, timestamp string, ok bool) {
	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return
	}

	path := parsedUrl.Path
	query := parsedUrl.Query()
	timestamp = query.Get("t")

	switch parsedUrl.Host {
	case "youtube.com", "www.youtube.com", "m.youtube.com":
		if path == "/watch" && query.Has("v") {
			videoId = query.Get("v")
		} else if strings.HasPrefix(path, "/watch/") {
			videoId = path[7:]
		} else if strings.HasPrefix(path, "/v/") {
			videoId = path[3:]
		} else if strings.HasPrefix(path, "/shorts/") {
			videoId = path[8:]
		}
	case "youtu.be":
		videoId = path[1:]
	}

	return videoId, timestamp, videoId != ""
}

func ParseTimestamp(timestamp string) float32 {
	if seconds, err := strconv.ParseFloat(timestamp, 32); err == nil {
		return float32(seconds)
	}
	if duration, err := time.ParseDuration(timestamp); err == nil {
		return float32(duration.Seconds())
	}
	return 0.0
}

func parseDuration(ytDuration string) (time.Duration, error) {
	return time.ParseDuration(strings.ToLower(ytDuration[2:]))
}
