package youtube

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestUrlParsing(t *testing.T) {
	// Examples taken from https://gist.github.com/rodrigoborgesdeoliveira/987683cfbfcc8d800192da1e73adc486
	file, err := os.Open("example_urls.txt")
	if err != nil {
		t.Errorf("error reading example file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			_, ok := ParseUrl(line)
			if !ok {
				t.Errorf("%s could not be parsed as a valid YouTube URL", line)
			}
		}
	}
}
