package lib

type PodmanContainers []PodmanContainer

type PodmanContainer struct {
	AutoRemove bool        `yaml:"AutoRemove"`
	Command    []string    `yaml:"Command"`
	CreatedAt  string      `yaml:"CreatedAt"`
	Exited     bool        `yaml:"Exited"`
	ExitedAt   int64       `yaml:"ExitedAt"`
	ExitCode   int         `yaml:"ExitCode"`
	ID         string      `yaml:"Id"`
	Image      string      `yaml:"Image"`
	ImageID    string      `yaml:"ImageID"`
	IsInfra    bool        `yaml:"IsInfra"`
	Labels     Labels      `yaml:"Labels,omitempty"`
	Mounts     []string    `yaml:"Mounts"`
	Names      []string    `yaml:"Names"`
	Namespaces Namespaces  `yaml:"Namespaces"`
	Networks   []string    `yaml:"Networks"`
	Pid        int         `yaml:"Pid"`
	Pod        string      `yaml:"Pod"`
	PodName    string      `yaml:"PodName"`
	Ports      []Ports     `yaml:"Ports"`
	Size       interface{} `yaml:"Size"`
	StartedAt  int64       `yaml:"StartedAt"`
	State      string      `yaml:"State"`
	Status     string      `yaml:"Status"`
	Created    int64       `yaml:"Created"`
}

type Labels struct {
	PODMANSYSTEMDUNIT                   string `yaml:"PODMAN_SYSTEMD_UNIT"`
	ComDockerComposeContainerNumber     string `yaml:"com.docker.compose.container-number"`
	ComDockerComposeProject             string `yaml:"com.docker.compose.project"`
	ComDockerComposeProjectConfigFiles  string `yaml:"com.docker.compose.project.config_files"`
	ComDockerComposeProjectWorkingDir   string `yaml:"com.docker.compose.project.working_dir"`
	ComDockerComposeService             string `yaml:"com.docker.compose.service"`
	ComProjectGroup                     string `yaml:"com.project.group,omitempty"`
	ComProjectName                      string `yaml:"com.project.name,omitempty"`
	ComProjectStack                     string `yaml:"com.project.stack,omitempty"`
	IoPodmanComposeConfigHash           string `yaml:"io.podman.compose.config-hash"`
	IoPodmanComposeProject              string `yaml:"io.podman.compose.project"`
	IoPodmanComposeVersion              string `yaml:"io.podman.compose.version"`
	OrgOpencontainersImageAuthors       string `yaml:"org.opencontainers.image.authors"`
	OrgOpencontainersImageBaseName      string `yaml:"org.opencontainers.image.base.name"`
	OrgOpencontainersImageDescription   string `yaml:"org.opencontainers.image.description"`
	OrgOpencontainersImageDocumentation string `yaml:"org.opencontainers.image.documentation"`
	OrgOpencontainersImageLicenses      string `yaml:"org.opencontainers.image.licenses"`
	OrgOpencontainersImageRefName       string `yaml:"org.opencontainers.image.ref.name"`
	OrgOpencontainersImageSource        string `yaml:"org.opencontainers.image.source"`
	OrgOpencontainersImageTitle         string `yaml:"org.opencontainers.image.title"`
	OrgOpencontainersImageURL           string `yaml:"org.opencontainers.image.url"`
	OrgOpencontainersImageVendor        string `yaml:"org.opencontainers.image.vendor"`
	OrgOpencontainersImageVersion       string `yaml:"org.opencontainers.image.version"`
}
type Namespaces struct {
}
type Ports struct {
	HostIP        string `yaml:"host_ip"`
	ContainerPort int    `yaml:"container_port"`
	HostPort      int    `yaml:"host_port"`
	Range         int    `yaml:"range"`
	Protocol      string `yaml:"protocol"`
}
