package registry

import (
	"context"

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

	ctx := context.Background()
	repository, err := remote.NewRepository(parsedReference.String())
	if err != nil {
		return nil, err
	}
	repository.PlainHTTP = c.plainHTTP
	repository.Client = c.authorizer
	var registryTags []string
	err = repository.Tags(ctx, "", func(tags []string) error {
		registryTags = append(registryTags, tags...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return registryTags, nil
}
