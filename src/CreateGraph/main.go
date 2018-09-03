package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// type (
// 	gethPeer struct {
// 		Caps    []string `json:"caps"`
// 		ID      string   `json:"id"`
// 		Name    string   `json:"name"`
// 		Network struct {
// 			LocalAddress  string `json:"localAddress"`
// 			RemoteAddress string `json:"remoteAddress"`
// 		} `json:"network"`
// 		Protocols struct {
// 			Eth struct {
// 				Difficulty int64  `json:"difficulty"`
// 				Head       string `json:"head"`
// 				Version    int    `json:"version"`
// 			} `json:"eth"`
// 		} `json:"protocols"`
// 	}
// )

func main() {
	// var (
	// 	peer1 gethPeer
	// 	peers []gethPeer
	// )

	// peer1.Caps = []string{"eth/62", "eth/63"}
	// peer1.ID = "06d80fcd6313486407d434ed25d5a36d553da39eeacbf11672e5968339f1a0498c9848dd805e117e287573f20c50e7a814e89d9cca031a7064f4a69dd50ee9f4"
	// peer1.Name = "Geth/v1.5.0-unstable-ac24c024/linux/go1.7.3"
	// peer1.Network.LocalAddress = "10.0.1.214:33988"
	// peer1.Network.RemoteAddress = "52.36.105.244:21000"
	// peer1.Protocols.Eth.Difficulty = 2974154752
	// peer1.Protocols.Eth.Head = "0x305014bdf390e8c26aa0ddef2e7c549877c8737eae268611955098b7f974030b"
	// peer1.Protocols.Eth.Version = 63

	// peers = append(peers, peer1)
	// data, _ := json.Marshal(peers)

	// fmt.Println(string(data))

	fmt.Println()
	fmt.Println("Creates graph file for analysis")
	fmt.Println()

	var InputFile, OutputFile string
	flag.StringVar(&InputFile, "in", "", "Name of file containing admin.peers output")
	flag.StringVar(&OutputFile, "out", "", "Name of output file for GraphViz analysis")

	if len(os.Args) < 4 {
		fmt.Println()
		flag.PrintDefaults()
		return
	}
	flag.Parse()

	if InputFile == "" {
		InputFile = "/Users/chuacw/35.172.215.9.txt"
	}
	if OutputFile == "" {
		OutputFile = "/Users/chuacw/Documents/Graphs/sample0003.dot"
	}

	re, err := regexp.Compile(`remoteAddress: "(([0-9]{1,3}.){3}[0-9]{1,3}):([0-9]{1,5})"`)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	byteContent, err := ioutil.ReadFile(InputFile)
	if err != nil {
		fmt.Printf("Unable to read file: %s due to %v", InputFile, err)
		return
	}
	fileContent := string(byteContent)
	res := re.FindAllStringSubmatch(fileContent, -1)

	// 0 - "remoteAddress: "192.168.0.1:54321"
	// 1 - "192.168.0.1"
	// 2 - 0.
	// 3 - 54321

	node := "blah"
	var graphContentbytes strings.Builder
	fmt.Fprintf(&graphContentbytes, "graph {\n")
	fmt.Fprintf(&graphContentbytes, `  "%s" -- {%s`, node, "\n")
	for _, matches := range res {
		IP := matches[1]
		fmt.Fprintf(&graphContentbytes, `   "%s" %s`, IP, "\n")
	}
	fmt.Fprintf(&graphContentbytes, `  };%s`, "\n")
	fmt.Fprint(&graphContentbytes, "}\n")
	fmt.Print(graphContentbytes.String())
	ioutil.WriteFile(OutputFile, []byte(graphContentbytes.String()), 0644)
}
