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
	"os"
	"sort"
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
	
	containers, err := cli.ContainerList(ctx, containertypes.ListOptions{All: true})
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
		
		// Get the host IP address
		hostIP, err := getHostIP()
		if err != nil {
			panic(err)
		}
		
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
			URL:        fmt.Sprintf("http://%s:%d", hostIP, publicPort), // To be updated
			NetworkIPs: networkIPs,                                      // Ensure the struct contains this field
			Mounts:     c.Mounts,
		}
	}
	
	sort.Slice(containersData, func(i, j int) bool {
		// "running" containers should appear before others
		if containersData[i].State == "running" && containersData[j].State != "running" {
			return true
		}
		if containersData[i].State != "running" && containersData[j].State == "running" {
			return false
		}
		// If both have the same state, maintain original order
		return false
	})
	
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

func GetContainerLogsByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve logs for the container with the specified ID
}

// getHostIP retrieves the host IP address
func getHostIP() (string, error) {
	// Check if running inside a Docker container
	if os.Getenv("RUNNING_IN_DOCKER") == "true" {
		// Use host.docker.internal (supported on Docker for Windows and macOS)
		hostIP := os.Getenv("HOST_IP")
		if hostIP != "" {
			return hostIP, nil
		}
		// Fallback to resolving host.docker.internal
		addrs, err := net.LookupHost("host.docker.internal")
		if err == nil && len(addrs) > 0 {
			return addrs[0], nil
		}
	}
	
	// If not running in Docker, get the LAN IP address of the host
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	
	for _, iface := range interfaces {
		// Skip loopback and Docker interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	
	return "", fmt.Errorf("no valid LAN IP address found")
}
