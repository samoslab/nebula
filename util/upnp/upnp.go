package upnp

import (
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway1"
	"github.com/huin/goupnp/dcps/internetgateway2"
)

// An IGD provides an interface to the most commonly used functions of an
// Internet Gateway Device: discovering the external IP, and forwarding ports.
type IGD struct {
	// This interface is satisfied by the internetgateway1.WANIPConnection1
	// and internetgateway1.WANPPPConnection1 types.
	client interface {
		GetExternalIPAddress() (string, error)
		AddPortMapping(string, uint16, string, uint16, string, bool, string, uint32) error
		DeletePortMapping(string, uint16, string) error
		GetServiceClient() *goupnp.ServiceClient
	}
}

// ExternalIP returns the router's external IP.
func (d *IGD) ExternalIP() (string, error) {
	return d.client.GetExternalIPAddress()
}

// Forward forwards the specified port, and adds its description to the
// router's port mapping table.
func (d *IGD) Forward(port uint16, desc string) error {
	ip, err := d.getInternalIP()
	if err != nil {
		return err
	}

	time.Sleep(time.Millisecond)
	err = d.client.AddPortMapping("", port, "TCP", port, ip, true, desc, 0)
	if err != nil {
		return err
	}

	time.Sleep(time.Millisecond)
	return d.client.AddPortMapping("", port, "UDP", port, ip, true, desc, 0)
}

// Clear un-forwards a port, removing it from the router's port mapping table.
func (d *IGD) Clear(port uint16) error {
	time.Sleep(time.Millisecond)
	tcpErr := d.client.DeletePortMapping("", port, "TCP")
	time.Sleep(time.Millisecond)
	udpErr := d.client.DeletePortMapping("", port, "UDP")

	// only return an error if both deletions failed
	if tcpErr != nil && udpErr != nil {
		return tcpErr
	}
	return nil
}

// Location returns the URL of the router, for future lookups (see Load).
func (d *IGD) Location() string {
	return d.client.GetServiceClient().Location.String()
}

// getInternalIP returns the user's local IP.
func (d *IGD) getInternalIP() (string, error) {
	host, _, _ := net.SplitHostPort(d.client.GetServiceClient().RootDevice.URLBase.Host)
	devIP := net.ParseIP(host)
	if devIP == nil {
		return "", errors.New("could not determine router's internal IP")
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			if x, ok := addr.(*net.IPNet); ok && x.Contains(devIP) {
				return x.IP.String(), nil
			}
		}
	}

	return "", errors.New("could not determine internal IP")
}

// // Discover is deprecated; use DiscoverCtx instead.
// func Discover() (*IGD, error) {
// 	return DiscoverCtx(context.Background())
// }

// DiscoverCtx scans the local network for routers and returns the first
// UPnP-enabled router it encounters.  It will try up to 3 times to find a
// router, sleeping a random duration between each attempt.  This is to
// mitigate a race condition with many callers attempting to discover
// simultaneously.
func Discover() (*IGD, error) {
	// TODO: if more than one client is found, only return those on the same
	// subnet as the user?
	maxTries := 3
	// sleepTime := time.Millisecond * time.Duration(Intn(5000))
	for try := 0; try < maxTries; try++ {
		pppclients, _, _ := internetgateway1.NewWANPPPConnection1Clients()
		if len(pppclients) > 0 {
			return &IGD{pppclients[0]}, nil
		}
		ipclients, _, _ := internetgateway1.NewWANIPConnection1Clients()
		if len(ipclients) > 0 {
			return &IGD{ipclients[0]}, nil
		}
		ppp2clients, _, _ := internetgateway2.NewWANPPPConnection1Clients()
		if len(ppp2clients) > 0 {
			return &IGD{ppp2clients[0]}, nil
		}
		ipclientsArr, _, _ := internetgateway2.NewWANIPConnection2Clients()
		if len(ipclientsArr) > 0 {
			return &IGD{ipclientsArr[0]}, nil
		}
		ip2clients, _, _ := internetgateway2.NewWANIPConnection1Clients()
		if len(ip2clients) > 0 {
			return &IGD{ip2clients[0]}, nil
		}

		// select {
		// case <-ctx.Done():
		// 	return nil, context.Canceled
		// case <-time.After(sleepTime):
		// }
		// sleepTime *= 2
	}
	return nil, errors.New("no UPnP-enabled gateway found")
}

// Load connects to the router service specified by rawurl. This is much
// faster than Discover. Generally, Load should only be called with values
// returned by the IGD's Location method.
func Load(rawurl string) (*IGD, error) {
	loc, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	pppclients, _ := internetgateway1.NewWANPPPConnection1ClientsByURL(loc)
	if len(pppclients) > 0 {
		return &IGD{pppclients[0]}, nil
	}
	ipclients, _ := internetgateway1.NewWANIPConnection1ClientsByURL(loc)
	if len(ipclients) > 0 {
		return &IGD{ipclients[0]}, nil
	}
	return nil, errors.New("no UPnP-enabled gateway found at URL " + rawurl)
}
