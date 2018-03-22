package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"

	"github.com/spolabs/nebula/provider/config"
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
	volumeFlag := addStorageCommand.String("volume", "", "add storage volume size, unit TB or GB, eg: 2TB or 500GB")
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
		daemon(daemonConfigDir)
	case "register":
		registerCommand.Parse(os.Args[2:])
		register(registerConfigDir, walletAddressFlag, billEmailFlag, availabilityFlag,
			upBandwidthFlag, downBandwidthFlag, mainStoragePathFlag, mainStorageVolumeFlag, extraStorageFlag)
	case "addStorage":
		addStorageCommand.Parse(os.Args[2:])
		addStorage(addStorageConfigDir, pathFlag, volumeFlag)
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}
}

func daemon(configDir *string) {
	err := config.LoadConfig(configDir)
	if err != nil {
		log.Fatalf("failed to LoadConfig: %v", err)
	}
	config.StartAutoReload()
	defer config.StopAutoReload()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 6666))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterProviderServiceServer(grpcServer, server.NewProviderServer(*configDir))
	grpcServer.Serve(lis)
}

func register(configDir *string, walletAddress *string, billEmail *string,
	availability *string, upBandwidth *uint, downBandwidth *uint,
	mainStoragePath *string, mainStorageVolume *string, extraStorageFlag *string) {
	fmt.Printf("register billEmail:%s\n", *billEmail)
}

func addStorage(configDir *string, path *string, volume *string) {
	fmt.Printf("addStorage path:%s, volume:%s\n", *path, *volume)
}
