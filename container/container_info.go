package container

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	CreatedTime string `json:"created_time"`
	Status      string `json:"status"`
}

var (
	RUNNING               = "running"
	STOP                  = "stppped"
	EXIT                  = "exited"
	DEFAULT_INFO_LOCATION = "/var/run/mydocker/%s/"
	CONFIG_NAME           = "config.json"
)
