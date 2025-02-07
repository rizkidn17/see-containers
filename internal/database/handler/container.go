package handler

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"log"
	"net"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

type ContainerInfo struct {
	ID         string             // Container ID
	Names      []string           // Container names
	Image      string             // Image name
	ImageID    string             // Image ID
	Command    string             // Command used to start the container
	Created    time.Time          // Timestamp of container creation
	State      string             // Container state (e.g., running, exited)
	Status     string             // Human-readable status
	Ports      []types.Port       // List of exposed ports
	Labels     map[string]string  // Key-value labels assigned to the container
	URL        string             // URL to access the container
	NetworkIPs map[string]string  // Network name -> IP address mapping
	Mounts     []types.MountPoint // Mounted volumes
}

func ListContainersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	
	containers, err := cli.ContainerList(ctx, containertypes.ListOptions{})
	if err != nil {
		panic(err)
	}
	
	// Convert Docker SDK container list to structured data
	containersData := make([]ContainerInfo, len(containers))
	
	for i, c := range containers {
		// Extract network information (ensure NetworkSettings exists)
		networkIPs := make(map[string]string)
		if c.NetworkSettings != nil {
			for netName, netSettings := range c.NetworkSettings.Networks {
				networkIPs[netName] = netSettings.IPAddress
			}
		}
		
		// Convert unix timestamp to human-readable format
		unixTime, err := strconv.ParseInt(strconv.FormatInt(c.Created, 10), 10, 64)
		if err != nil {
			panic(err)
		}
		cleanedTime := time.Unix(unixTime, 0)
		
		var publicPort int
		for _, port := range c.Ports {
			if port.PublicPort != 0 { // PublicPort exists
				publicPort = int(port.PublicPort)
				break
			}
		}
		
		// Populate structured container data
		containersData[i] = ContainerInfo{
			ID:         c.ID,
			Names:      c.Names,
			Image:      c.Image,
			ImageID:    c.ImageID,
			Command:    c.Command,
			Created:    cleanedTime,
			State:      c.State,
			Status:     c.Status,
			Ports:      c.Ports,
			Labels:     c.Labels,
			URL:        fmt.Sprintf("http://%s:%d", GetLocalIP(), publicPort), // To be updated
			NetworkIPs: networkIPs,                                            // Ensure the struct contains this field
			Mounts:     c.Mounts,
		}
	}
	
	tmplPath := "./web/templates/index.html"
	
	// Parse the HTML template
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Printf("error parsing template: %v", err)
		return
	}
	
	// Render the template with data
	err = tmpl.Execute(w, containersData)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("error rendering template: %v", err)
	}
}

func StartContainerByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Start the container with the specified ID
}

func StopContainerByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Stop the container with the specified ID
}

// GetLocalIP finds the first non-loopback local IP address
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1" // Fallback if unable to detect
	}
	
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil { // Ensure it's IPv4
				return ipNet.IP.String()
			}
		}
	}
	return "127.0.0.1" // Default if no local IP is found
}
