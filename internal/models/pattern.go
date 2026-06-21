package models

// Topic is one of the 15 DSA categories (Arrays & Hashing, Two Pointers, etc).
type Topic struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	SortOrder int    `json:"sort_order"`
}

// Pattern is a single recognizable technique within a Topic, e.g.
// "Monotonic stack (next greater/smaller element)".
type Pattern struct {
	ID            string `json:"id"`
	TopicID       string `json:"topic_id"`
	Name          string `json:"name"`
	CoreIdea      string `json:"core_idea"`
	QuestionTitle string `json:"question_title"`
	QuestionURL   string `json:"question_url"`
	SortOrder     int    `json:"sort_order"`
}
