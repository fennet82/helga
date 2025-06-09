package errors

import "fmt"

type ErrArtifactoryAPI struct {
	DerivedFromErr error
	Repo           string
	Path           string
}

func (e ErrArtifactoryAPI) Error() string {
	return fmt.Sprintf("general Artifactory API error, repo name:%s, artifact path:%s, error: %s",
		e.Repo, e.Path, e.DerivedFromErr.Error())
}

type ErrPkgsDoNotMatch struct {
	ErrMsg string
}

func (e ErrPkgsDoNotMatch) Error() string {
	return e.ErrMsg
}
