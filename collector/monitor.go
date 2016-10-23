package collector

import (
	"errors"
	"strings"
	"github.com/fsouza/go-dockerclient"
	"net"
	"strconv"
	"log"
)


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
	cpuUpper 	int
	cpuLower 	int
	lastStats 	docker.Stats
}

// NewMonitor creates new monitor with specified docker client,
// container id and stat updating interval
func NewMonitor(c MonitorDockerClient, id string, interval int) (*Monitor, error) {
	container, err := c.InspectContainer(id)
	if err != nil {
		return nil, err
	}

	do_monitor := extractEnv(container, "COLLECTD_MONITOR")
	if do_monitor != "true" {
		return nil, ErrNoNeedToMonitor
	}

	namespace := container.Config.Labels["io.kubernetes.pod.namespace"]
	pod := container.Config.Labels["io.kubernetes.pod.name"]
	containerName := container.Config.Labels["io.kubernetes.container.name"]
	cpuRange := extractEnv(container, "COLLECTD_CPU_RANGE")
	cpuLower := -1
	cpuUpper := -1

	cpuLowerS, cpuUpperS, err := net.SplitHostPort(cpuRange)
	if cpuLowerS != "" && cpuUpperS != "" {
		cpuUpper, err = strconv.Atoi(cpuUpperS)
		cpuLower, err = strconv.Atoi(cpuLowerS)
	}

	if namespace == "" || pod == "" || containerName == "" {
		return nil, ErrNoNeedToMonitor
	}

	log.Printf("Monioring %d %s cpu %s:%s  =  %d:%d", interval, containerName, cpuLowerS, cpuUpperS, cpuLower, cpuUpper)

	return &Monitor{
		client:   c,
		id:       container.ID,
		namespace: sanitizeForGraphite(namespace),
		pod: sanitizeForGraphite(pod),
		container: sanitizeForGraphite(containerName),
		interval: interval,
		cpuLower: cpuLower,
		cpuUpper: cpuUpper,
	}, nil
}

func (m *Monitor) handle(ch chan<- Stats) error {
	in := make(chan *docker.Stats)

	go func() {
		i := 0
		for s := range in {
			log.Println("ccccc")

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
				cpuUpper: m.cpuUpper,
				cpuLower: m.cpuLower,
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


func extractEnv(c *docker.Container, envPrefix string) string {
	for _, e := range c.Config.Env {
		if strings.HasPrefix(e, envPrefix) {
			return strings.TrimPrefix(e, envPrefix + "=")
		}
	}

	return ""
}