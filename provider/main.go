package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/spolabs/nebula/provider/config"
	"github.com/spolabs/nebula/provider/node"
	pb "github.com/spolabs/nebula/provider/pb"
	"github.com/spolabs/nebula/provider/server"
	"google.golang.org/grpc"
)

const home_config_folder = ".spo-nebula-provider"

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	daemonCommand := flag.NewFlagSet("daemon", flag.ExitOnError)
	daemonConfigDir := daemonCommand.String("configDir", usr.HomeDir+string(os.PathSeparator)+home_config_folder, "config director")

	registerCommand := flag.NewFlagSet("register", flag.ExitOnError)
	registerConfigDir := registerCommand.String("configDir", usr.HomeDir+string(os.PathSeparator)+home_config_folder, "config director")
	walletAddressFlag := registerCommand.String("walletAddress", "", "wallet address to accept earnings")
	billEmailFlag := registerCommand.String("billEmail", "", "email where send bill to")
	availabilityFlag := registerCommand.String("availability", "", "promise availability: 98%, 99%, 99.9%")
	upBandwidthFlag := registerCommand.Uint("upBandwidth", 0, "upload bandwidth, unit: Mbps, eg: 100, 20, 8, 4")
	downBandwidthFlag := registerCommand.Uint("downBandwidth", 0, "download bandwidth, unit: Mbps, eg: 100, 20")
	mainStoragePathFlag := registerCommand.String("mainStoragePath", "", "main storage path")
	mainStorageVolumeFlag := registerCommand.String("mainStorageVolume", "", "main storage volume size, unit TB or GB, eg: 2TB or 500GB")
	extraStorageFlag := registerCommand.String("extraStorage", "", "extra storage, format:path1:volume1;path2:volume2, eg: /mnt/sde1:1T;/mnt/sdf1:800G;/mnt/sdg1:500G")

	addStorageCommand := flag.NewFlagSet("addStorage", flag.ExitOnError)
	addStorageConfigDir := addStorageCommand.String("configDir", usr.HomeDir+string(os.PathSeparator)+home_config_folder, "config director")
	pathFlag := addStorageCommand.String("path", "", "add storage path")
	volumeFlag := addStorageCommand.String("volume", "", "add storage volume size, unit T or G, eg: 2T or 500G")
	if len(os.Args) == 1 {
		fmt.Printf("usage: %s <command> [<args>]\n", os.Args[0])
		fmt.Println("The most commonly used commands are: ")
		fmt.Println(" daemon [-configDir]")
		fmt.Println(" register [-configDir] -billEmail")
		fmt.Println(" addStorage [-configDir] -path storage-path -volume storage-volume")
		return
	}

	switch os.Args[1] {
	case "daemon":
		daemonCommand.Parse(os.Args[2:])
		daemon(*daemonConfigDir)
	case "register":
		registerCommand.Parse(os.Args[2:])
		register(*registerConfigDir, *walletAddressFlag, *billEmailFlag, *availabilityFlag,
			*upBandwidthFlag, *downBandwidthFlag, *mainStoragePathFlag, *mainStorageVolumeFlag, *extraStorageFlag)
	case "addStorage":
		addStorageCommand.Parse(os.Args[2:])
		addStorage(*addStorageConfigDir, *pathFlag, *volumeFlag)
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}
}

func daemon(configDir string) {
	err := config.LoadConfig(configDir)
	if err != nil {
		if err == config.NoConfErr {
			fmt.Printf("Config file is not ready, please run \"%s register\" first\n", os.Args[0])
			return
		} else if err == config.ConfVerifyErr {
			fmt.Println("Config file error, can not start.")
			return
		}
		log.Fatalf("failed to LoadConfig: %v", err)
	}
	config.StartAutoCheck()
	defer config.StopAutoCheck()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 6666))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	providerServer := server.NewProviderServer()
	defer providerServer.Close()
	pb.RegisterProviderServiceServer(grpcServer, providerServer)
	grpcServer.Serve(lis)
}

func register(configDir string, walletAddress string, billEmail string,
	availability string, upBandwidth uint, downBandwidth uint,
	mainStoragePath string, mainStorageVolume string, extraStorageFlag string) {
	if len(walletAddress) == 0 {
		fmt.Println("walletAddress is required.")
		os.Exit(3)
	}
	if _, err := cipher.DecodeBase58Address(walletAddress); err != nil {
		fmt.Printf("walletAddress is not valid:%s\n", walletAddress)
		os.Exit(4)
	}
	if len(billEmail) == 0 {
		fmt.Println("billEmail is required.")
		os.Exit(5)
	}
	email_re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !email_re.MatchString(billEmail) {
		fmt.Printf("billEmail is not valid:%s\n", billEmail)
		os.Exit(6)
	}
	if len(availability) == 0 {
		fmt.Println("availability is required.")
		os.Exit(7)
	}
	if availability[len(availability)-1] != '%' {
		fmt.Printf("availability should be a percentage: %s\n", availability)
		os.Exit(8)
	}
	availFloat, err := strconv.ParseFloat(availability[:len(availability)-1], 64)
	if err != nil {
		fmt.Printf("availability: %s is not valid:%s\n", availability, err)
		os.Exit(9)
	}
	if availFloat < 0.98 {
		fmt.Println("availability must more than 98%.")
		os.Exit(10)
	}
	if upBandwidth == 0 {
		fmt.Println("upBandwidth is required.")
		os.Exit(11)
	}
	upBandwidthBps := uint64(upBandwidth) * 1000 * 1000
	if downBandwidth == 0 {
		fmt.Println("downBandwidth is required.")
		os.Exit(12)
	}
	downBandwidthBps := uint64(downBandwidth) * 1000 * 1000
	if len(mainStorageVolume) == 0 {
		fmt.Println("mainStorageVolume is required.")
		os.Exit(13)
	}
	origMainStorageVolume := mainStorageVolume
	mainStorageVolume = strings.ToUpper(mainStorageVolume)
	if mainStorageVolume[len(mainStorageVolume)-1] == 'B' {
		mainStorageVolume = mainStorageVolume[:len(mainStorageVolume)-1]
	}
	var mainStorageVolumeByte uint64
	if mainStorageVolume[len(mainStorageVolume)-1] == 'G' {
		val, err := strconv.ParseInt(mainStorageVolume[:len(mainStorageVolume)-1], 10, 64)
		if err != nil {
			fmt.Printf("mainStorageVolume: %s is not valid:%s\n", origMainStorageVolume, err)
			os.Exit(14)
		}
		mainStorageVolumeByte = uint64(val) * 1000 * 1000 * 1000
	} else if mainStorageVolume[len(mainStorageVolume)-1] == 'T' {
		val, err := strconv.ParseInt(mainStorageVolume[:len(mainStorageVolume)-1], 10, 64)
		if err != nil {
			fmt.Printf("mainStorageVolume: %s is not valid:%s\n", origMainStorageVolume, err)
			os.Exit(15)
		}
		mainStorageVolumeByte = uint64(val) * 1000 * 1000 * 1000 * 1000
	} else {
		fmt.Printf("mainStorageVolume: %s is not valid\n", origMainStorageVolume)
		os.Exit(16)
	}
	// TODO call Tracker provider register api
	doRegister(configDir, walletAddress, billEmail, availFloat, upBandwidthBps, downBandwidthBps, mainStoragePath, mainStorageVolumeByte)
}

func doRegister(configDir string, walletAddress string, billEmail string,
	availability float64, upBandwidth uint64, downBandwidth uint64,
	mainStoragePath string, mainStorageVolume uint64) {
	no := node.NewNode(10)
	pc := newProviderConfig(no, walletAddress, billEmail, availability, upBandwidth, downBandwidth, mainStoragePath, mainStorageVolume)
	config.CreateProviderConfig(configDir, pc)
}

func addStorage(configDir string, path string, volume string) {
	fmt.Printf("addStorage path:%s, volume:%s\n", path, volume)
	//TODO
}

func newProviderConfig(no *node.Node, walletAddress string, billEmail string,
	availability float64, upBandwidth uint64, downBandwidth uint64,
	mainStoragePath string, mainStorageVolume uint64) *config.ProviderConfig {
	pc := &config.ProviderConfig{
		NodeId:            no.NodeIdStr(),
		WalletAddress:     walletAddress,
		BillEmail:         billEmail,
		PublicKey:         no.PublicKeyStr(),
		PrivateKey:        no.PrivateKeyStr(),
		Availability:      availability,
		MainStoragePath:   mainStoragePath,
		MainStorageVolume: mainStorageVolume,
		UpBandwidth:       upBandwidth,
		DownBandwidth:     downBandwidth,
	}
	m := make(map[string]string, len(no.EncryptKey))
	for k, v := range no.EncryptKey {
		m[k] = hex.EncodeToString(v)
	}
	pc.EncryptKey = m
	return pc
}
