package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"softwareupgrade"
	"strings"
	"time"

	"github.com/twmb/algoimpl/go/graph"
)

type (
	EximchainNode struct {
		IP string
	}
)

func findPublicIP(ipMap map[string]string, nodename string) (result string) {
	res := ipMap[nodename]
	if res != "" {
		result = res
	}
	return
}

var (
	cache []*interface{}
)

// NodeToIntfAddr converts a node to a pointer to an interface{}
func (n *EximchainNode) NodeToIntfAddr() (result *interface{}) {
	var intf interface{} = *n // any struct is compatible to an interface{}
	result = &intf            // take the address of the struct
	// which is masquerading as an interface{}, so it's now *intf{}
	cache = append(cache, result)
	return
}

func IntfToEximchainNode(p *interface{}) (result *EximchainNode) {
	result1 := (*p).(EximchainNode)
	return &result1
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

type (
	NetworkBlocks struct {
		privateIPBlocks []*net.IPNet
	}
)

func WriteLn(f *os.File, line string) {
	f.WriteString(line)
	f.WriteString("\n")
}

// NewNetworkBlocks create a block of networks consisting of private IP blocks
func NewNetworkBlocks() (result *NetworkBlocks) {
	result = &NetworkBlocks{}
	for _, cidr := range []string{
		"10.0.0.0/8",     // RFC1918
		"127.0.0.0/8",    // IPv4 loopback
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
	} {
		_, block, _ := net.ParseCIDR(cidr)
		result.privateIPBlocks = append(result.privateIPBlocks, block)
	}
	return
}

// StringToIP converts a given IP string into an IP address
func StringToIP(ip string) (result net.IP) {
	return net.ParseIP(ip)
}

// IPToString converts a given IP address into a string
func IPToString(ip net.IP) (result string) {
	return ip.String()
}

// IsPrivateIP checks if a given IP is in the network block.
func (n *NetworkBlocks) IsPrivateIP(ip net.IP) bool {
	for _, block := range n.privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println()
	fmt.Println("Creates graph file for analysis of Eximchain Nodes")
	fmt.Println()

	var (
		InputFile       string
		OutputFile      string
		ListOfFiles     string
		Files           []string
		extension       string
		iterations      int
		concurrent      int
		calculateRadius bool
	)
	flag.StringVar(&InputFile, "in", "", "Name of file containing admin.peers output")
	flag.StringVar(&ListOfFiles, "list", "", "Filename of file containing list of files containing admin.peers output")
	flag.StringVar(&OutputFile, "out", "", "Name of output file for GraphViz analysis")
	flag.StringVar(&extension, "extension", "", "Extension to attach to filename when searching for files")
	flag.IntVar(&iterations, "iterations", 100, "Iterations to run for minimum cut, minimum of 1")
	flag.IntVar(&concurrent, "concurrent", 1, "Number of iterations to start concurrently, minimum of 1")
	flag.BoolVar(&calculateRadius, "radius", false, "Calculate the radius (default false)")

	flag.Parse()

	// Verify parameters
	if (ListOfFiles == "" && InputFile == "") || OutputFile == "" || concurrent < 1 || iterations < 1 {
		if OutputFile == "" {
			fmt.Printf("%s is not specified\n", "Output file")
		}
		flag.PrintDefaults()
		os.Exit(1)
	}

	if expandedOutputFile, err := softwareupgrade.Expand(OutputFile); err == nil {
		OutputFile = expandedOutputFile
	} else {
		log.Fatalf("Unable to expand %s\n", OutputFile)
	}

	if ListOfFiles != "" && InputFile != "" {
		fmt.Println("Specify either -in or -list, not both.")
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}

	var (
		BaseDir    string
		publicIPs  []string
		privateIPs []string
		IPMap      map[string]string
		IPsNode    []string
	)

	if ListOfFiles != "" {
		if expandedListOfFiles, err := softwareupgrade.Expand(ListOfFiles); err == nil {
			ListOfFiles = expandedListOfFiles
		} else {
			fmt.Printf("Unable to expand %s\n", ListOfFiles)
			os.Exit(1)
		}
		BaseDir = path.Dir(ListOfFiles)
		ListOfFilesBytes, err := ioutil.ReadFile(ListOfFiles)
		if err != nil {
			fmt.Printf("Error reading %s due to %v\n", ListOfFiles, err)
			os.Exit(1)
		}
		sListOfFilesBytes := string(ListOfFilesBytes)
		IPsNode = strings.Split(sListOfFilesBytes, "\n")
		// public
		// private
		// node name

		IPMap = make(map[string]string)
		if IPsNode[len(IPsNode)-1] == "" {
			IPsNode = IPsNode[:len(IPsNode)-1]
		}

		for i := 0; i < len(IPsNode); i++ {
			Line := strings.Split(IPsNode[i], ",")
			publicIP := Line[1]
			publicIPs = append(publicIPs, publicIP)
			privateIP := Line[2]
			privateIPs = append(privateIPs, privateIP)
			hostname := Line[0]
			Files = append(Files, hostname)
			if privateIP == "" || publicIP == "" {
				log.Fatalln("There are empty IPs in the list!")
			}
			IPMap[privateIP] = publicIP
			IPMap[hostname] = publicIP
		}

	} else {
		if InputFile == "" {
			fmt.Printf("%s is not specified\n", "Input file")
			return
		}
		Files = []string{InputFile}
	}

	if extension != "" {
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
	}

	networkBlocks := NewNetworkBlocks()
	g := NewGraphContainer(graph.Undirected)
	uniqueNodes := make(map[string]*graph.Node)

	// loop through each file, and get the IP address listed after every "remoteAddress: "
	for _, InputFile = range Files {
		if InputFile == "" {
			continue
		}
		nodename := InputFile
		InputFile = InputFile + extension
		re, err := regexp.Compile(`remoteAddress: "(([0-9]{1,3}.){3}[0-9]{1,3}):([0-9]{1,5})"`)
		if err != nil {
			fmt.Printf("%v", err)
			return
		}
		var (
			byteContent []byte
		)
		if !softwareupgrade.FileExists(InputFile) {
			InputFile = filepath.Join(BaseDir, InputFile)
			if !softwareupgrade.FileExists(InputFile) {
				log.Fatalf("Can't find file: %s\n", InputFile)
			}
		}
		byteContent, err = ioutil.ReadFile(InputFile)
		if err != nil {
			fmt.Printf("Unable to read file: %s due to %v", InputFile, err)
			return
		}
		publicIP := findPublicIP(IPMap, nodename)

		fileContent := string(byteContent)
		res := re.FindAllStringSubmatch(fileContent, -1)

		var node string
		if publicIP != "" {
			node = publicIP
		} else {
			log.Fatalln("There is an unexpected empty value where an IP is expected!")
		}
		var newNode1 *graph.Node
		// check against the cache for an existing node, or create a new one and cache it.
		if aNode := uniqueNodes[publicIP]; aNode != nil {
			newNode1 = aNode
		} else {
			newNode1 = g.MakeNode()
			uniqueNodes[publicIP] = newNode1
			eximchainNode1 := EximchainNode{node}
			newNode1.Value = eximchainNode1.NodeToIntfAddr()
		}
		for _, matches := range res {
			// 0 - "remoteAddress: "192.168.0.1:54321"
			// 1 - "192.168.0.1"
			// 2 - 0.
			// 3 - 54321
			sIP := matches[1]
			var IP string
			nIP := net.ParseIP(sIP)
			if networkBlocks.IsPrivateIP(nIP) {
				IP = findPublicIP(IPMap, sIP) // map private IP to public IP
			} else {
				IP = sIP
			}

			var newNode2 *graph.Node
			// check against the cache for an existing node with the same IP, or create a new one and cache it.
			if aNode2 := uniqueNodes[IP]; aNode2 != nil {
				newNode2 = aNode2
				oldEximchainNode := IntfToEximchainNode(newNode2.Value)
				if oldEximchainNode.IP == "" {
					log.Fatalln("Logic error! IP is empty!")
				}
			} else {
				newNode2 = g.MakeNode()
				uniqueNodes[IP] = newNode2
				eximchainNode2 := EximchainNode{IP}
				newNode2.Value = eximchainNode2.NodeToIntfAddr()
			}
			g.MakeEdge(newNode1, newNode2)
		}

	}

	var (
		f                 *os.File
		err               error
		graphContentbytes strings.Builder
	)

	f, err = os.OpenFile(OutputFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	f.Truncate(0)
	defer f.Close()

	graphContentbytes.Reset()
	f.WriteString(`graph EximchainNodes {`)
	f.WriteString("\n")
	f.WriteString("  subgraph cluster0 {\n")
	cut := g.RandMinimumCut(iterations, concurrent)
	diameter := g.Diameter()
	WriteLn(f, "    node [color=white];")
	stats := fmt.Sprintf("Eximchain Nodes: min cut=%d, IPs=%d, diameter=%d", len(cut), len(IPsNode), diameter)
	fmt.Println(stats)
	line := fmt.Sprintf(`    label="%s"`, stats)
	WriteLn(f, line)
	WriteLn(f, `    "     ";`)
	WriteLn(f, "  };")

	if calculateRadius {
		start := time.Now()
		fmt.Printf("Radius: %d\n", g.Radius())
		elapsed := time.Since(start)
		fmt.Printf("Radius took %s\n", elapsed)
	}
	subgraphs := g.StronglyConnectedComponents()
	if len(subgraphs) > 0 {
		for i, subgraph := range subgraphs {
			line := fmt.Sprintf("  subgraph cluster_%d {", i+1)
			WriteLn(f, line)
			line = fmt.Sprintf(`    label="Subgraph %d"`, i+1)
			WriteLn(f, line)
			for _, node := range subgraph {
				// node is interface{}
				aNode := IntfToEximchainNode(node.Value)
				line := fmt.Sprintf(`    "%s";`, aNode.IP)
				WriteLn(f, line)
			}
			WriteLn(f, "  };")
		}
	}

	WriteLn(f, "}")
	fmt.Printf("Output file: %s generated.\n", OutputFile)
	fmt.Println()
}
