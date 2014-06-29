package cwmp

import "github.com/oleiade/lane"

type CPE struct {
	SerialNumber         string
	Manufacturer         string
	OUI                  string
	ConnectionRequestURL string
	SoftwareVersion      string
	ExternalIPAddress    string
	State                string
	Queue 				 *lane.Queue
}
