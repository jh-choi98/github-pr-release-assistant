package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v66/github"
	sashabaranov "github.com/sashabaranov/go-openai"
)

func HandleNewPR(event GitHubPullRequestEvent) {
	ctx := context.Background()

	gitHubAppIDStr := os.Getenv("GITHUB_APP_ID")
	if gitHubAppIDStr == "" {
		fmt.Println("GITHUB_APP_ID not set")
		return
	}
	
	gitHubAppID, err := strconv.ParseInt(gitHubAppIDStr, 10, 64)
	if err != nil {
		fmt.Println("Invalid GitHub App ID")
		return
	}

	gitHubPrivateKeyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH")
	if gitHubPrivateKeyPath == "" {
		fmt.Println("GITHUB_PRIVATE_KEY_PATH not set")
		return
	}

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, gitHubAppID, event.Installation.ID, gitHubPrivateKeyPath)
	if err != nil {
		fmt.Printf("Auth Error: Could not load private key: %v\n", err)
		return
	}

	ghClient := github.NewClient(&http.Client{Transport: itr})

	fmt.Println("	Authenticated with GitHub App successfully!")

	diffContent, err := getDiff(ctx, ghClient, event)
	if err != nil {
		fmt.Printf("Failed to fetch diff: %v\n", err)
			return
	}
	fmt.Printf("	Diff fetched! Length: %d characters\n", len(diffContent))
 
	summary, err := getAISummary(ctx, diffContent)
	if err != nil {
		fmt.Printf("OpenAI Error: %v\n", err)
		return
	}
	fmt.Println("	AI Summary generated!")

	comment := &github.IssueComment{
		Body: github.String(summary),
	}

	_, _, err = ghClient.Issues.CreateComment(ctx, event.Repository.Owner.Login, event.Repository.Name, event.Number, comment)
	if err != nil {
		fmt.Printf("Failed to post comment: %v\n", err)
		return
	}
	fmt.Printf("	Success! Comment posted to PR #%d\n", event.Number)
}

func getDiff(ctx context.Context, client *github.Client, event GitHubPullRequestEvent) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", event.PullRequest.DiffURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Client().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	return string(bodyBytes), nil
}

func getAISummary(ctx context.Context, diff string) (string, error) {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return "", fmt.Errorf("LLM Key not set")
	}

	client := sashabaranov.NewClient(openAIKey)

	if len(diff) > 10000 {
		diff = diff[:10000] + "\n...(truncated)..."
	}
	prompt := fmt.Sprintf(`You are a skilled Senior Software Engineer acting as a code reviewer.
Analyze the following code diff and provide a concise summary in Markdown format.

Structure:
## 🔍 Summary
(One sentence summary)

## 🛠 Key Changes
- (Bullet points of what changed)

## ⚠️ Potential Risks
- (Any bugs, security issues, or performance concerns? If none, say "None detected".)

Here is the diff:
%s`, diff)
	resp, err := client.CreateChatCompletion(
		ctx,
		sashabaranov.ChatCompletionRequest{
			Model: sashabaranov.GPT4oMini,
			Messages: []sashabaranov.ChatCompletionMessage{
				{
					Role: sashabaranov.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
