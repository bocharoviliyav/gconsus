package entity

import "time"

type UserActivityInfo struct {
	Login                   string `json:"login"`
	ContributionsCollection struct {
		TotalCommitContributions        int `json:"totalCommitContributions"`
		CommitContributionsByRepository []struct {
			Repository struct {
				Name  string `json:"name"`
				Owner struct {
					Login string `json:"login"`
				} `json:"owner"`
			} `json:"repository"`
			Contributions struct {
				Nodes []struct {
					CommitCount int       `json:"commitCount"`
					OccurredAt  time.Time `json:"occurredAt"`
				} `json:"nodes"`
			} `json:"contributions" graphql:"contributions(first: 10)"`
		} `json:"commitContributionsByRepository"`
		TotalPullRequestContributions        int `json:"totalPullRequestContributions"`
		PullRequestContributionsByRepository []struct {
			Repository struct {
				Name  string `json:"name"`
				Owner struct {
					Login string `json:"login"`
				} `json:"owner"`
			} `json:"repository"`
			Contributions struct {
				Nodes []struct {
					PullRequest struct {
						Title  string `json:"title"`
						Url    string `json:"url"`
						Merged bool   `json:"merged"`
					} `json:"pullRequest"`
					OccurredAt time.Time `json:"occurredAt"`
				} `json:"nodes"`
			} `json:"contributions" graphql:"contributions(first: 10)"`
		} `json:"pullRequestContributionsByRepository"`
		TotalIssueContributions        int `json:"totalIssueContributions"`
		IssueContributionsByRepository []struct {
			Repository struct {
				Name  string `json:"name"`
				Owner struct {
					Login string `json:"login"`
				} `json:"owner"`
			} `json:"repository"`
			Contributions struct {
				Nodes []struct {
					Issue struct {
						Title string `json:"title"`
						Url   string `json:"url"`
						State string `json:"state"`
					} `json:"issue"`
					OccurredAt time.Time `json:"occurredAt"`
				} `json:"nodes"`
			} `json:"contributions" graphql:"contributions(first: 10)"`
		} `json:"issueContributionsByRepository"`
		TotalPullRequestReviewContributions int `json:"totalPullRequestReviewContributions"`
		PullRequestReviewContributions      struct {
			Nodes []struct {
				PullRequest struct {
					Title      string `json:"title"`
					Url        string `json:"url"`
					Repository struct {
						Name  string `json:"name"`
						Owner struct {
							Login string `json:"login"`
						} `json:"owner"`
					} `json:"repository"`
				} `json:"pullRequest"`
				OccurredAt time.Time `json:"occurredAt"`
			} `json:"nodes"`
		} `json:"pullRequestReviewContributions" graphql:"pullRequestReviewContributions(first: 10)"`
	} `json:"contributionsCollection" graphql:"contributionsCollection(from: $from, to: $to)"`
}

type UserActivityQuery struct {
	User UserActivityInfo `graphql:"user(login: $login)"`
}
