package meta

import (
	"fmt"
)

const GitMetaName = ".gitmeta"

type Metadata struct {
	data     map[string]string
	filename string
	err      error
}

func Load() *Metadata {
	Init()

	meta, ok := value.Load().(*Metadata)
	if !ok {
		return &Metadata{
			data: map[string]string{},
			err:  fmt.Errorf("invalid metadata type(%v) or has not been initialized", value.Load()),
		}
	}

	return meta
}

func (meta *Metadata) Err() error {
	return meta.err
}

// App returns project name of the service
func (meta *Metadata) App() string {
	return meta.data["app_id"]
}

// Git returns git url of the service source
func (meta *Metadata) Git() string {
	return meta.data["repo"]
}

// Version returns the git tag of the service build from
func (meta *Metadata) Version() string {
	return meta.data["ref"]
}

// GitCommit returns the git commit log of the service build from
func (meta *Metadata) GitCommit() string {
	return meta.data["hash"]
}

// GitAuthor returns the author of the service
func (meta *Metadata) GitAuthor() string {
	return meta.data["author"]
}

// GitDate returns the date of the service build from
func (meta *Metadata) GitDate() string {
	return meta.data["committer_date"]
}

func (meta *Metadata) CandidateVersion() string {
	return meta.data["candidates_tag"]
}
