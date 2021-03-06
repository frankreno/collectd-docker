package collector

import "github.com/fsouza/go-dockerclient"

// Stats represents singe stat from docker stats api for specific task
type Stats struct {
	Namespace   string
	Pod   string
	Container   string
	Stats docker.Stats
	PrevStats docker.Stats
}
