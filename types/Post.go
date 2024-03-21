package types

import (
	"fmt"
	"time"
)

type Post struct {
	Title        string
	Date         time.Time
	Content      string
	Slug         string
	Type         string
	Tags         []string
	Status       string
	Keywords     string
	Description  string
	Summary      string
	SummaryClean string
	Reminder     string
	Author       string
	SourceUrl    string
	Cover        string
	Image        string
}

type MarkdownPost struct {
	Title       string    `yaml:"title"`
	Date        time.Time `yaml:"date"`
	Tags        []string  `yaml:"tags"`
	Draft       bool      `yaml:"draft"`
	Description string    `yaml:"description"`
	Author      string    `yaml:"author"`
	SourceUrl   string    `yaml:"source_url"`
	Cover       string    `yaml:"cover"`
	Image       string    `yaml:"image"`
}

type Posts []Post

type PostsByDate []Post

func (posts Posts) FindByTag(tag string) Posts {

	var foundPosts Posts

	for _, post := range posts {
		for _, t := range post.Tags {
			if t == tag {
				foundPosts = append(foundPosts, post)
			}
		}
	}

	return foundPosts
}

func (post *Post) Permarlink() string {
	return fmt.Sprintf("/%s/%s.html", post.Type, post.Slug)
}

func (d PostsByDate) Len() int {
	return len(d)
}

func (d PostsByDate) Less(i, j int) bool {
	return d[i].Date.After(d[j].Date)
}

func (d PostsByDate) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
