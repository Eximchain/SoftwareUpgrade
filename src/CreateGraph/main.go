package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"softwareupgrade"
	"strings"
)

func findPublicIP(ipMap map[string]string, nodename string) (result string) {
	res := ipMap[nodename]
	if res != "" {
		result = res
	}
	return
}

func main() {

	fmt.Println()
	fmt.Println("Creates graph file for analysis")
	fmt.Println()

	var (
		InputFile   string
		OutputFile  string
		ListOfFiles string
		Files       []string
		suffix      string
	)
	flag.StringVar(&InputFile, "in", "", "Name of file containing admin.peers output")
	flag.StringVar(&ListOfFiles, "list", "", "Filename of file containing list of files containing admin.peers output")
	flag.StringVar(&OutputFile, "out", "", "Name of output file for GraphViz analysis")
	flag.StringVar(&suffix, "suffix", "", "Suffix to attach to filename")

	if len(os.Args) < 3 {
		flag.PrintDefaults()
		return
	}
	flag.Parse()

	if OutputFile == "" {
		fmt.Printf("%s is not specified\n", "Output file")
		return
	}

	if ListOfFiles != "" && InputFile != "" {
		fmt.Println("Specify either -in or -list, not both.")
		fmt.Println()
		flag.Usage()
		return
	}

	var (
		BaseDir    string
		publicIPs  []string
		privateIPs []string
		IPMap      map[string]string
	)

	if ListOfFiles != "" {
		if expandedListOfFiles, err := softwareupgrade.Expand(ListOfFiles); err == nil {
			ListOfFiles = expandedListOfFiles
		} else {
			fmt.Printf("Unable to expand %s\n", ListOfFiles)
			return
		}
		BaseDir = path.Dir(ListOfFiles)
		ListOfFilesBytes, err := ioutil.ReadFile(ListOfFiles)
		if err != nil {
			fmt.Printf("Error reading %s due to %v\n", ListOfFiles, err)
			return
		}
		sListOfFilesBytes := string(ListOfFilesBytes)
		IPsNode := strings.Split(sListOfFilesBytes, "\n")
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

	if suffix != "" {
		if !strings.HasPrefix(suffix, ".") {
			suffix = "." + suffix
		}
	}

	if expandedOutputFile, err := softwareupgrade.Expand(OutputFile); err == nil {
		OutputFile = expandedOutputFile
	} else {
		log.Fatalf("Unable to expand %s\n", OutputFile)
	}

	var (
		f          *os.File
		err        error
		CreateFlag int
	)
	CreateFlag = os.O_APPEND | os.O_WRONLY
	if !softwareupgrade.FileExists(OutputFile) {
		CreateFlag = CreateFlag | os.O_CREATE
	}
	f, err = os.OpenFile(OutputFile, CreateFlag, 0644)
	if err != nil {
		log.Fatalf("Unable to open file: %s due to error: %v\n", OutputFile, err)
	}
	f.WriteString("graph {\n")
	f.Close()
	var (
		graphContentbytes strings.Builder
		nodeCount         int
	)
	for nodeCount, InputFile = range Files {
		if InputFile == "" {
			continue
		}
		nodename := InputFile
		InputFile = InputFile + suffix
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
		}
		byteContent, err = ioutil.ReadFile(InputFile)
		if err != nil {
			fmt.Printf("Unable to read file: %s due to %v", InputFile, err)
			return
		}
		publicIP := findPublicIP(IPMap, nodename)
		fileContent := string(byteContent)
		res := re.FindAllStringSubmatch(fileContent, -1)

		// 0 - "remoteAddress: "192.168.0.1:54321"
		// 1 - "192.168.0.1"
		// 2 - 0.
		// 3 - 54321
		var node string
		if publicIP != "" {
			node = publicIP
		} else {
			node = fmt.Sprintf("%04d", nodeCount+1)
		}
		graphContentbytes.Reset()
		fmt.Fprintf(&graphContentbytes, `  "%s" -- {%s`, node, "\n")
		for _, matches := range res {
			IP := matches[1]
			fmt.Fprintf(&graphContentbytes, `   "%s" %s`, IP, "\n")
		}
		fmt.Fprintf(&graphContentbytes, `  };%s`, "\n")

		var (
			f *os.File
		)
		CreateFlag = os.O_APPEND | os.O_WRONLY
		if !softwareupgrade.FileExists(OutputFile) {
			CreateFlag = CreateFlag | os.O_CREATE
		}
		f, err = os.OpenFile(OutputFile, CreateFlag, 0644)
		if err != nil {
			log.Fatal(err)
		}
		s := graphContentbytes.String()
		f.WriteString(s)
		f.Close()
	}

	f, err = os.OpenFile(OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString("}\n")
	f.Close()
}
