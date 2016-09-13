package collector

import (
	"errors"
	"strings"
	"github.com/fsouza/go-dockerclient"
)


var namespaceLabel = "io.kubernetes.pod.namespace"
var podLabel = "io.kubernetes.pod.name"
var containerNameLabel = "io.kubernetes.container.name"
var containerHashLabel = "io.kubernetes.container.hash"


// ErrNoNeedToMonitor is used to skip containers
// that shouldn't be monitored by collectd
var ErrNoNeedToMonitor = errors.New("container is not supposed to be monitored")

// MonitorDockerClient represents restricted interface for docker client
// that is used in monitor, docker.Client is a subset of this interface
type MonitorDockerClient interface {
	InspectContainer(id string) (*docker.Container, error)
	Stats(opts docker.StatsOptions) error
}

// Monitor is responsible for monitoring of a single container (task)
type Monitor struct {
	client   	MonitorDockerClient
	id       	string
	namespace	string
	pod		string
	container	string
	interval 	int
	lastStats 	docker.Stats
}

// NewMonitor creates new monitor with specified docker client,
// container id and stat updating interval
func NewMonitor(c MonitorDockerClient, id string, interval int) (*Monitor, error) {
	container, err := c.InspectContainer(id)
	if err != nil {
		return nil, err
	}
	namespace := container.Config.Labels[namespaceLabel]
	pod := container.Config.Labels[podLabel]
	containerName := container.Config.Labels[containerNameLabel] + "-" + container.Config.Labels[containerHashLabel]

	if namespace == "" || pod == "" || containerName == "" {
		return nil, ErrNoNeedToMonitor
	}

	return &Monitor{
		client:   c,
		id:       container.ID,
		namespace: sanitizeForGraphite(namespace),
		pod: sanitizeForGraphite(pod),
		container: sanitizeForGraphite(containerName),
		interval: interval,
	}, nil
}

func (m *Monitor) handle(ch chan<- Stats) error {
	in := make(chan *docker.Stats)

	go func() {
		i := 0
		for s := range in {
			if i%m.interval != 0 {
				i++
				continue
			}

			ch <- Stats{
				Namespace: m.namespace,
				Pod: m.pod,
				Container:   m.container,
				Stats: *s,
				PrevStats: m.lastStats,
			}

			m.lastStats = *s

			i++
		}
	}()

	return m.client.Stats(docker.StatsOptions{
		ID:     m.id,
		Stats:  in,
		Stream: true,
	})
}


func sanitizeForGraphite(s string) string {
	r := strings.Replace(s, ".", "_", -1)

  // strip leading / and santize any other / for mesos ids  
  return strings.Replace(strings.TrimPrefix(r, "/"), "/", "_", -1)
}
