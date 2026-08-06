package registry

import (
	"context"
	"strings"

	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

// FetchAllTags implements Tags function but removes semver StrictNewVersion check
// fix for issue https://github.com/devtron-labs/devtron/issues/4385, tags were not semver compatible, so they were getting filtered by StrictNewVersion check
func (c *Client) FetchAllTags(ref string) ([]string, error) {
	parsedReference, err := registry.ParseReference(ref)
	if err != nil {
		return nil, err
	}

	repository, err := remote.NewRepository(parsedReference.String())
	if err != nil {
		return nil, err
	}
	repository.PlainHTTP = c.plainHTTP
	repository.Client = c.authorizer
	tags, err := fetchTags(context.Background(), repository)
	if err == nil {
		return tags, nil
	}

	if repository.PlainHTTP || !strings.Contains(err.Error(), "server gave HTTP response") {
		return nil, err
	}

	repository.PlainHTTP = true
	return fetchTags(context.Background(), repository)

}

func fetchTags(ctx context.Context, repository *remote.Repository) ([]string, error) {
	var tags []string

	err := repository.Tags(ctx, "", func(batch []string) error {
		tags = append(tags, batch...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tags, nil
}
