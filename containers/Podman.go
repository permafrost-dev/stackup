package containers

import (
	"encoding/json"
	"fmt"

	"github.com/stackup-app/stackup/utils"
)

type PodmanContainers []struct {
	ID      string   `json:"Id"`
	Names   []string `json:"Names"`
	Image   string   `json:"Image"`
	ImageID string   `json:"ImageID"`
	Command string   `json:"Command"`
	Created int      `json:"Created"`
	Ports   []struct {
		PrivatePort int    `json:"PrivatePort"`
		PublicPort  int    `json:"PublicPort"`
		Type        string `json:"Type"`
	} `json:"Ports"`
	Labels struct {
		PODMANSYSTEMDUNIT                  string `json:"PODMAN_SYSTEMD_UNIT"`
		ComDockerComposeContainerNumber    string `json:"com.docker.compose.container-number"`
		ComDockerComposeProject            string `json:"com.docker.compose.project"`
		ComDockerComposeProjectConfigFiles string `json:"com.docker.compose.project.config_files"`
		ComDockerComposeProjectWorkingDir  string `json:"com.docker.compose.project.working_dir"`
		ComDockerComposeService            string `json:"com.docker.compose.service"`
		ComProjectGroup                    string `json:"com.project.group,omitempty"`
		ComProjectName                     string `json:"com.project.name,omitempty"`
		ComProjectStack                    string `json:"com.project.stack,omitempty"`
		IoPodmanComposeConfigHash          string `json:"io.podman.compose.config-hash"`
		IoPodmanComposeProject             string `json:"io.podman.compose.project"`
		IoPodmanComposeVersion             string `json:"io.podman.compose.version"`
	} `json:"Labels"`
	State           string `json:"State"`
	Status          string `json:"Status"`
	NetworkSettings struct {
		Networks struct {
			AcdPosBackendAcdposnet struct {
				IPAMConfig          interface{} `json:"IPAMConfig"`
				Links               interface{} `json:"Links"`
				Aliases             []string    `json:"Aliases"`
				NetworkID           string      `json:"NetworkID"`
				EndpointID          string      `json:"EndpointID"`
				Gateway             string      `json:"Gateway"`
				IPAddress           string      `json:"IPAddress"`
				IPPrefixLen         int         `json:"IPPrefixLen"`
				IPv6Gateway         string      `json:"IPv6Gateway"`
				GlobalIPv6Address   string      `json:"GlobalIPv6Address"`
				GlobalIPv6PrefixLen int         `json:"GlobalIPv6PrefixLen"`
				MacAddress          string      `json:"MacAddress"`
				DriverOpts          interface{} `json:"DriverOpts"`
			} `json:"acd-pos-backend_acdposnet"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
	Mounts []struct {
		Type        string `json:"Type"`
		Name        string `json:"Name"`
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		Driver      string `json:"Driver"`
		Mode        string `json:"Mode"`
		RW          bool   `json:"RW"`
		Propagation string `json:"Propagation"`
	} `json:"Mounts"`
	Name             string      `json:"Name"`
	Config           interface{} `json:"Config"`
	NetworkingConfig interface{} `json:"NetworkingConfig"`
	Platform         interface{} `json:"Platform"`
	AdjustCPUShares  bool        `json:"AdjustCPUShares"`
}

type PodmanContainerDetail struct {
	ID      string `json:"Id"`
	Created struct {
	} `json:"Created"`
	Path  string   `json:"Path"`
	Args  []string `json:"Args"`
	State struct {
		Status     string `json:"Status"`
		Running    bool   `json:"Running"`
		Paused     bool   `json:"Paused"`
		Restarting bool   `json:"Restarting"`
		OOMKilled  bool   `json:"OOMKilled"`
		Dead       bool   `json:"Dead"`
		Pid        int    `json:"Pid"`
		ExitCode   int    `json:"ExitCode"`
		Error      string `json:"Error"`
		StartedAt  struct {
		} `json:"StartedAt"`
		FinishedAt struct {
		} `json:"FinishedAt"`
		Health struct {
			Status        string      `json:"Status"`
			FailingStreak int         `json:"FailingStreak"`
			Log           interface{} `json:"Log"`
		} `json:"Health"`
	} `json:"State"`
	Image           string        `json:"Image"`
	ResolvConfPath  string        `json:"ResolvConfPath"`
	HostnamePath    string        `json:"HostnamePath"`
	HostsPath       string        `json:"HostsPath"`
	LogPath         string        `json:"LogPath"`
	Name            string        `json:"Name"`
	RestartCount    int           `json:"RestartCount"`
	Driver          string        `json:"Driver"`
	Platform        string        `json:"Platform"`
	MountLabel      string        `json:"MountLabel"`
	ProcessLabel    string        `json:"ProcessLabel"`
	AppArmorProfile string        `json:"AppArmorProfile"`
	ExecIDs         []interface{} `json:"ExecIDs"`
	HostConfig      struct {
		Binds           []string `json:"Binds"`
		ContainerIDFile string   `json:"ContainerIDFile"`
		LogConfig       struct {
			Type   string      `json:"Type"`
			Config interface{} `json:"Config"`
		} `json:"LogConfig"`
		NetworkMode  string `json:"NetworkMode"`
		PortBindings struct {
			Six379TCP []struct {
				HostIP   string `json:"HostIp"`
				HostPort string `json:"HostPort"`
			} `json:"6379/tcp"`
		} `json:"PortBindings"`
		RestartPolicy struct {
			Name              string `json:"Name"`
			MaximumRetryCount int    `json:"MaximumRetryCount"`
		} `json:"RestartPolicy"`
		AutoRemove           bool          `json:"AutoRemove"`
		VolumeDriver         string        `json:"VolumeDriver"`
		VolumesFrom          interface{}   `json:"VolumesFrom"`
		ConsoleSize          []int         `json:"ConsoleSize"`
		CapAdd               []interface{} `json:"CapAdd"`
		CapDrop              []interface{} `json:"CapDrop"`
		CgroupnsMode         string        `json:"CgroupnsMode"`
		DNS                  []interface{} `json:"Dns"`
		DNSOptions           []interface{} `json:"DnsOptions"`
		DNSSearch            []interface{} `json:"DnsSearch"`
		ExtraHosts           []interface{} `json:"ExtraHosts"`
		GroupAdd             []interface{} `json:"GroupAdd"`
		IpcMode              string        `json:"IpcMode"`
		Cgroup               string        `json:"Cgroup"`
		Links                interface{}   `json:"Links"`
		OomScoreAdj          int           `json:"OomScoreAdj"`
		PidMode              string        `json:"PidMode"`
		Privileged           bool          `json:"Privileged"`
		PublishAllPorts      bool          `json:"PublishAllPorts"`
		ReadonlyRootfs       bool          `json:"ReadonlyRootfs"`
		SecurityOpt          []interface{} `json:"SecurityOpt"`
		UTSMode              string        `json:"UTSMode"`
		UsernsMode           string        `json:"UsernsMode"`
		ShmSize              int           `json:"ShmSize"`
		Runtime              string        `json:"Runtime"`
		Isolation            string        `json:"Isolation"`
		CPUShares            int           `json:"CpuShares"`
		Memory               int           `json:"Memory"`
		NanoCpus             int           `json:"NanoCpus"`
		CgroupParent         string        `json:"CgroupParent"`
		BlkioWeight          int           `json:"BlkioWeight"`
		BlkioWeightDevice    interface{}   `json:"BlkioWeightDevice"`
		BlkioDeviceReadBps   interface{}   `json:"BlkioDeviceReadBps"`
		BlkioDeviceWriteBps  interface{}   `json:"BlkioDeviceWriteBps"`
		BlkioDeviceReadIOps  interface{}   `json:"BlkioDeviceReadIOps"`
		BlkioDeviceWriteIOps interface{}   `json:"BlkioDeviceWriteIOps"`
		CPUPeriod            int           `json:"CpuPeriod"`
		CPUQuota             int           `json:"CpuQuota"`
		CPURealtimePeriod    int           `json:"CpuRealtimePeriod"`
		CPURealtimeRuntime   int           `json:"CpuRealtimeRuntime"`
		CpusetCpus           string        `json:"CpusetCpus"`
		CpusetMems           string        `json:"CpusetMems"`
		Devices              []interface{} `json:"Devices"`
		DeviceCgroupRules    interface{}   `json:"DeviceCgroupRules"`
		DeviceRequests       interface{}   `json:"DeviceRequests"`
		MemoryReservation    int           `json:"MemoryReservation"`
		MemorySwap           int           `json:"MemorySwap"`
		MemorySwappiness     int           `json:"MemorySwappiness"`
		OomKillDisable       bool          `json:"OomKillDisable"`
		PidsLimit            int           `json:"PidsLimit"`
		Ulimits              []struct {
			Name string `json:"Name"`
			Hard int    `json:"Hard"`
			Soft int    `json:"Soft"`
		} `json:"Ulimits"`
		CPUCount           int         `json:"CpuCount"`
		CPUPercent         int         `json:"CpuPercent"`
		IOMaximumIOps      int         `json:"IOMaximumIOps"`
		IOMaximumBandwidth int         `json:"IOMaximumBandwidth"`
		MaskedPaths        interface{} `json:"MaskedPaths"`
		ReadonlyPaths      interface{} `json:"ReadonlyPaths"`
	} `json:"HostConfig"`
	GraphDriver struct {
		Data struct {
			LowerDir  string `json:"LowerDir"`
			MergedDir string `json:"MergedDir"`
			UpperDir  string `json:"UpperDir"`
			WorkDir   string `json:"WorkDir"`
		} `json:"Data"`
		Name string `json:"Name"`
	} `json:"GraphDriver"`
	SizeRootFs int `json:"SizeRootFs"`
	Mounts     []struct {
		Type        string `json:"Type"`
		Name        string `json:"Name"`
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		Driver      string `json:"Driver"`
		Mode        string `json:"Mode"`
		RW          bool   `json:"RW"`
		Propagation string `json:"Propagation"`
	} `json:"Mounts"`
	Config struct {
		Hostname     string `json:"Hostname"`
		Domainname   string `json:"Domainname"`
		User         string `json:"User"`
		AttachStdin  bool   `json:"AttachStdin"`
		AttachStdout bool   `json:"AttachStdout"`
		AttachStderr bool   `json:"AttachStderr"`
		ExposedPorts []struct {
			Port struct {
			} `json:",omitempty"`
		} `json:"ExposedPorts"`
		Tty        bool        `json:"Tty"`
		OpenStdin  bool        `json:"OpenStdin"`
		StdinOnce  bool        `json:"StdinOnce"`
		Env        []string    `json:"Env"`
		Cmd        []string    `json:"Cmd"`
		Image      string      `json:"Image"`
		Volumes    interface{} `json:"Volumes"`
		WorkingDir string      `json:"WorkingDir"`
		Entrypoint []string    `json:"Entrypoint"`
		OnBuild    interface{} `json:"OnBuild"`
		Labels     struct {
			PODMANSYSTEMDUNIT                  string `json:"PODMAN_SYSTEMD_UNIT"`
			ComDockerComposeContainerNumber    string `json:"com.docker.compose.container-number"`
			ComDockerComposeProject            string `json:"com.docker.compose.project"`
			ComDockerComposeProjectConfigFiles string `json:"com.docker.compose.project.config_files"`
			ComDockerComposeProjectWorkingDir  string `json:"com.docker.compose.project.working_dir"`
			ComDockerComposeService            string `json:"com.docker.compose.service"`
			IoPodmanComposeConfigHash          string `json:"io.podman.compose.config-hash"`
			IoPodmanComposeProject             string `json:"io.podman.compose.project"`
			IoPodmanComposeVersion             string `json:"io.podman.compose.version"`
		} `json:"Labels"`
		StopSignal  string `json:"StopSignal"`
		StopTimeout int    `json:"StopTimeout"`
	} `json:"Config"`
	NetworkSettings struct {
		Bridge                 string `json:"Bridge"`
		SandboxID              string `json:"SandboxID"`
		HairpinMode            bool   `json:"HairpinMode"`
		LinkLocalIPv6Address   string `json:"LinkLocalIPv6Address"`
		LinkLocalIPv6PrefixLen int    `json:"LinkLocalIPv6PrefixLen"`
		Ports                  []struct {
			Port []struct {
				HostIP   string `json:"HostIp"`
				HostPort string `json:"HostPort"`
			} `json:",omitempty"`
		} `json:"Ports"`
		SandboxKey             string      `json:"SandboxKey"`
		SecondaryIPAddresses   interface{} `json:"SecondaryIPAddresses"`
		SecondaryIPv6Addresses interface{} `json:"SecondaryIPv6Addresses"`
		EndpointID             string      `json:"EndpointID"`
		Gateway                string      `json:"Gateway"`
		GlobalIPv6Address      string      `json:"GlobalIPv6Address"`
		GlobalIPv6PrefixLen    int         `json:"GlobalIPv6PrefixLen"`
		IPAddress              string      `json:"IPAddress"`
		IPPrefixLen            int         `json:"IPPrefixLen"`
		IPv6Gateway            string      `json:"IPv6Gateway"`
		MacAddress             string      `json:"MacAddress"`
		Networks               []struct {
			Network struct {
				IPAMConfig          interface{} `json:"IPAMConfig"`
				Links               interface{} `json:"Links"`
				Aliases             []string    `json:"Aliases"`
				NetworkID           string      `json:"NetworkID"`
				EndpointID          string      `json:"EndpointID"`
				Gateway             string      `json:"Gateway"`
				IPAddress           string      `json:"IPAddress"`
				IPPrefixLen         int         `json:"IPPrefixLen"`
				IPv6Gateway         string      `json:"IPv6Gateway"`
				GlobalIPv6Address   string      `json:"GlobalIPv6Address"`
				GlobalIPv6PrefixLen int         `json:"GlobalIPv6PrefixLen"`
				MacAddress          string      `json:"MacAddress"`
				DriverOpts          interface{} `json:"DriverOpts"`
			} `json:",omitempty"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
}

func GetActivePodmanContainers() *PodmanContainers {
	http := utils.NewPodmanSocketClient()
	output, _ := http.Get("http://d/v4.0.0/containers/json")

	var items PodmanContainers
	json.Unmarshal([]byte(output), &items)

	return &items
}

func InspectPodmanContainer(containerId string) PodmanContainerDetail {
	http := utils.NewPodmanSocketClient()
	output, _ := http.Get(fmt.Sprintf("http://d/v4.0.0/containers/%s/json", containerId))

	var items PodmanContainerDetail
	json.Unmarshal([]byte(output), &items)

	return items
}

func InspectPodmanContainers(containerIds []string) []PodmanContainerDetail {
	var items []PodmanContainerDetail

	for _, containerId := range containerIds {
		items = append(items, InspectPodmanContainer(containerId))
	}

	return items
}
