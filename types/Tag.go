package types

import "fmt"

type Tag struct {
	Name       string
	Slug       string
	Posts      []Post
	CountPosts int
}

type Tags []Tag

type TagsByName []Tag

func (tag *Tag) Permalink() string {
	return fmt.Sprintf("/tags/%s.html", tag.Slug)
}

func (tag *Tag) AddPost(post Post) []Post {
	tag.Posts = append(tag.Posts, post)
	return tag.Posts
}

func (tags Tags) Find(key string) Tag {
	for i := range tags {
		if tags[i].Name == key {
			return tags[i]
		}
	}

	return Tag{}
}

func (tags TagsByName) Len() int {
	return len(tags)
}

func (tags TagsByName) Less(i, j int) bool {
	return tags[i].Name < tags[j].Name
}

func (tags TagsByName) Swap(i, j int) {
	tags[i], tags[j] = tags[j], tags[i]
}
