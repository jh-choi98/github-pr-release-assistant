package main

// GitHubPullRequestEvent: The top-level structure of the webhook
// payload
type GitHubPullRequestEvent struct {
	Action			string			`json:"action"`			// e.g. "opened", "closed"
	Number			int				`json:"number"`			// PR Number
	PullRequest		PullRequest		`json:"pull_request"`	// Nested PR details
	Repository		Repository		`json:"repository"`		// Repository info
	Installation	Installation	`json:"installation"`	// App Installation ID
}

// PullRequest: The specific details of the PR
type PullRequest struct {
	ID		int64	`json:"id"`
	Title	string	`json:"title"`
	Body	string	`json:"body"`		// The description user wrote
	State	string	`json:"state"`		// "open" or "closed"
	URL		string	`json:"url"`		// API URL
	DiffURL	string	`json:"diff_url"`	// The link to download the code changes
}

type Repository struct {
	FullName	string	`json:"full_name"`	// e.g. "jh-choi98/pr-agent-test"
	Name		string	`json:"name"`
	Owner		User	`json:"owner"`
}

type User struct {
	Login	string	`json:"login"`
}

type Installation struct {
	ID	int64	`json:"id"`
}
