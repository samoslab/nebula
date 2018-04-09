package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/samoslab/nebula/provider/config"
	"github.com/samoslab/nebula/provider/impl"
	"github.com/samoslab/nebula/provider/node"
	pb "github.com/samoslab/nebula/provider/pb"
	client "github.com/samoslab/nebula/provider/register_client"
	trp_pb "github.com/samoslab/nebula/tracker/register/provider/pb"
	util_rsa "github.com/samoslab/nebula/util/rsa"
	"google.golang.org/grpc"
)

const home_config_folder = ".spo-nebula-provider"

func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Get OS current user failed: ", err.Error())
		return
	}

	daemonCommand := flag.NewFlagSet("daemon", flag.ExitOnError)
	daemonConfigDirFlag := daemonCommand.String("configDir", usr.HomeDir+string(os.PathSeparator)+home_config_folder, "config director")
	daemonTrackerServerFlag := daemonCommand.String("trackerServer", "127.0.0.1:6677", "tracker server address, eg: 127.0.0.1:6677")
	listenFlag := daemonCommand.String("listen", ":6666", "listen address and port, eg: 111.111.111.111:6666 or :6666")

	registerCommand := flag.NewFlagSet("register", flag.ExitOnError)
	registerConfigDirFlag := registerCommand.String("configDir", usr.HomeDir+string(os.PathSeparator)+home_config_folder, "config director")
	registerTrackerServerFlag := registerCommand.String("trackerServer", "127.0.0.1:6677", "tracker server address, eg: 127.0.0.1:6677")
	walletAddressFlag := registerCommand.String("walletAddress", "", "SPO wallet address to accept earnings")
	billEmailFlag := registerCommand.String("billEmail", "", "email where send bill to")
	availabilityFlag := registerCommand.String("availability", "", "promise availability, must more than 98%, eg: 98%, 99%, 99.9%")
	upBandwidthFlag := registerCommand.Uint("upBandwidth", 0, "upload bandwidth, unit: Mbps, eg: 100, 20, 8, 4")
	downBandwidthFlag := registerCommand.Uint("downBandwidth", 0, "download bandwidth, unit: Mbps, eg: 100, 20")
	mainStoragePathFlag := registerCommand.String("mainStoragePath", "", "main storage path")
	mainStorageVolumeFlag := registerCommand.String("mainStorageVolume", "", "main storage volume size, unit TB or GB, eg: 2TB or 500GB")
	extraStorageFlag := registerCommand.String("extraStorage", "", "extra storage, format:path1:volume1,path2:volume2, eg: /mnt/sde1:1TB,/mnt/sdf1:800GB,/mnt/sdg1:500GB")
	portFlag := registerCommand.Uint("port", 6666, "outer network port for client to connect, eg:6666")
	hostFlag := registerCommand.String("host", "", "outer ip or domain for client to connect, eg: 123.123.123.123")
	dynamicDomainFlag := registerCommand.String("dynamicDomain", "", "dynamic domain for client to connect, eg: mydomain.xicp.net")

	verifyEmailCommand := flag.NewFlagSet("verifyEmail", flag.ExitOnError)
	verifyEmailConfigDirFlag := verifyEmailCommand.String("configDir", usr.HomeDir+string(os.PathSeparator)+home_config_folder, "config director")
	verifyEmailTrackerServerFlag := verifyEmailCommand.String("trackerServer", "127.0.0.1:6677", "tracker server address, eg: 127.0.0.1:6677")
	verifyCodeFlag := verifyEmailCommand.String("verifyCode", "", "verify code from verify email")

	resendVerifyCodeCommand := flag.NewFlagSet("resendVerifyCode", flag.ExitOnError)
	resendVerifyCodeConfigDirFlag := resendVerifyCodeCommand.String("configDir", usr.HomeDir+string(os.PathSeparator)+home_config_folder, "config director")
	resendVerifyCodeTrackerServerFlag := resendVerifyCodeCommand.String("trackerServer", "127.0.0.1:6677", "tracker server address, eg: 127.0.0.1:6677")

	addStorageCommand := flag.NewFlagSet("addStorage", flag.ExitOnError)
	addStorageConfigDirFlag := addStorageCommand.String("configDir", usr.HomeDir+string(os.PathSeparator)+home_config_folder, "config director")
	addStorageTrackerServerFlag := addStorageCommand.String("trackerServer", "127.0.0.1:6677", "tracker server address, eg: 127.0.0.1:6677")
	pathFlag := addStorageCommand.String("path", "", "add storage path")
	volumeFlag := addStorageCommand.String("volume", "", "add storage volume size, unit TB or GB, eg: 2TB or 500GB")
	if len(os.Args) == 1 {
		fmt.Printf("usage: %s <command> [<args>]\n", os.Args[0])
		fmt.Println("The most commonly used commands are: ")
		fmt.Println(" register [-configDir config-dir] [-trackerServer tracker-server-and-port] [-host outer-host] [-dynamicDomain dynamic-domain] [-port outer-port] -walletAddress wallet-address -billEmail bill-email -downBandwidth down-bandwidth -upBandwidth up-bandwidth -availability availability-percentage -mainStoragePath storage-path -mainStorageVolume storage-volume -extraStorage extra-storage-string")
		registerCommand.PrintDefaults()
		fmt.Println(" verifyEmail [-configDir config-dir] [-trackerServer tracker-server-and-port] -verifyCode verify-code")
		verifyEmailCommand.PrintDefaults()
		fmt.Println(" resendVerifyCode [-configDir config-dir] [-trackerServer tracker-server-and-port]")
		resendVerifyCodeCommand.PrintDefaults()
		fmt.Println(" daemon [-configDir config-dir] [-trackerServer tracker-server-and-port] [-listen listen-address-and-port]")
		daemonCommand.PrintDefaults()
		fmt.Println(" addStorage [-configDir config-dir] [-trackerServer tracker-server-and-port] -path storage-path -volume storage-volume")
		addStorageCommand.PrintDefaults()
		return
	}

	switch os.Args[1] {
	case "daemon":
		daemonCommand.Parse(os.Args[2:])
		daemon(*daemonConfigDirFlag, *daemonTrackerServerFlag, *listenFlag)
	case "register":
		registerCommand.Parse(os.Args[2:])
		register(*registerConfigDirFlag, *registerTrackerServerFlag, *walletAddressFlag, *billEmailFlag, *availabilityFlag,
			*upBandwidthFlag, *downBandwidthFlag, *portFlag, *hostFlag, *dynamicDomainFlag, *mainStoragePathFlag, *mainStorageVolumeFlag, *extraStorageFlag)
	case "addStorage":
		addStorageCommand.Parse(os.Args[2:])
		addStorage(*addStorageConfigDirFlag, *addStorageTrackerServerFlag, *pathFlag, *volumeFlag)
	case "verifyEmail":
		verifyEmailCommand.Parse(os.Args[2:])
		verifyEmail(*verifyEmailConfigDirFlag, *verifyEmailTrackerServerFlag, *verifyCodeFlag)
	case "resendVerifyCode":
		resendVerifyCodeCommand.Parse(os.Args[2:])
		resendVerifyCode(*resendVerifyCodeConfigDirFlag, *resendVerifyCodeTrackerServerFlag)
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}
}
func verifyEmail(configDir string, trackerServer string, verifyCode string) {
	err := config.LoadConfig(configDir)
	if err != nil {
		if err == config.NoConfErr {
			fmt.Printf("Config file is not ready, please run \"%s register\" to register first\n", os.Args[0])
			return
		} else if err == config.ConfVerifyErr {
			fmt.Println("Config file wrong, can not verify email.")
			return
		}
		fmt.Println("failed to load config, can not verify email: " + err.Error())
		return
	}
	if verifyCode == "" {
		fmt.Printf("verifyCode is required.\n")
		os.Exit(9)
	}
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s\n", err.Error())
		return
	}
	defer conn.Close()
	prsc := trp_pb.NewProviderRegisterServiceClient(conn)
	code, errMsg, err := client.VerifyBillEmail(prsc, verifyCode)
	if err != nil {
		fmt.Printf("verifyEmail failed: %s\n", err.Error())
		return
	}
	if code != 0 {
		fmt.Println(errMsg)
		return
	}
	fmt.Println("verifyEmail success, you can start daemon now.")
}

func resendVerifyCode(configDir string, trackerServer string) {
	err := config.LoadConfig(configDir)
	if err != nil {
		if err == config.NoConfErr {
			fmt.Printf("Config file is not ready, please run \"%s register\" to register first\n", os.Args[0])
			return
		} else if err == config.ConfVerifyErr {
			fmt.Println("Config file wrong, can not resend verify code email.")
			return
		}
		fmt.Println("failed to load config, can not resend verify code email: " + err.Error())
		return
	}
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s\n", err.Error())
		return
	}
	defer conn.Close()
	prsc := trp_pb.NewProviderRegisterServiceClient(conn)
	success, err := client.ResendVerifyCode(prsc)
	if err != nil {
		fmt.Printf("resendVerifyCode failed: %s\n", err.Error())
		return
	}
	if !success {
		fmt.Println("resendVerifyCode failed, please retry")
		return
	}
	fmt.Println("resendVerifyCode success, you can verify bill email.")
}

func daemon(configDir string, trackerServer string, listen string) {
	err := config.LoadConfig(configDir)
	if err != nil {
		if err == config.NoConfErr {
			fmt.Printf("Config file is not ready, please run \"%s register\" to register first\n", os.Args[0])
			return
		} else if err == config.ConfVerifyErr {
			fmt.Println("Config file wrong, can not start daemon.")
			return
		}
		fmt.Println("failed to load config, can not start daemon: " + err.Error())
		return
	}
	config.StartAutoCheck()
	defer config.StopAutoCheck()
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		fmt.Printf("failed to listen: %s, error: %s\n", listen, err.Error())
		return
	}
	grpcServer := grpc.NewServer()
	providerServer := impl.NewProviderService()
	defer providerServer.Close()
	pb.RegisterProviderServiceServer(grpcServer, providerServer)
	grpcServer.Serve(lis)
}

func register(configDir string, trackerServer string, walletAddress string, billEmail string,
	availability string, upBandwidth uint, downBandwidth uint, port uint, host string, dynamicDomain string,
	mainStoragePath string, mainStorageVolume string, extraStorageFlag string) {
	err := config.LoadConfig(configDir)
	if err == nil {
		fmt.Println("config file is adready exsits: " + configDir)
		os.Exit(2)
	}
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
	availFloat = availFloat / 100
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
	if port < 1 || port > 65535 {
		fmt.Println("port must between 1 to 65535.")
		os.Exit(14)
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
			os.Exit(21)
		}
		mainStorageVolumeByte = uint64(val) * 1000 * 1000 * 1000
	} else if mainStorageVolume[len(mainStorageVolume)-1] == 'T' {
		val, err := strconv.ParseInt(mainStorageVolume[:len(mainStorageVolume)-1], 10, 64)
		if err != nil {
			fmt.Printf("mainStorageVolume: %s is not valid:%s\n", origMainStorageVolume, err)
			os.Exit(22)
		}
		mainStorageVolumeByte = uint64(val) * 1000 * 1000 * 1000 * 1000
	} else {
		fmt.Printf("mainStorageVolume: %s is not valid\n", origMainStorageVolume)
		os.Exit(23)
	}
	// TODO support extra storage
	// TODO speed test
	testUpBandwidthBps := upBandwidthBps
	testDownBandwidthBps := downBandwidthBps
	doRegister(configDir, trackerServer, walletAddress, billEmail, availFloat, upBandwidthBps, downBandwidthBps, testUpBandwidthBps, testDownBandwidthBps, uint32(port), host, dynamicDomain, mainStoragePath, mainStorageVolumeByte, nil)
}

func encrypt(pubKey *rsa.PublicKey, data []byte) []byte {
	res, err := util_rsa.EncryptLong(pubKey, data, node.RSA_KEY_BYTES)
	if err != nil {
		fmt.Println("public key encrypt error: " + err.Error())
		os.Exit(16)
	}
	return res
}

func doRegister(configDir string, trackerServer string, walletAddress string, billEmail string,
	availability float64, upBandwidth uint64, downBandwidth uint64,
	testUpBandwidth uint64, testDownBandwidth uint64, port uint32, host string,
	dynamicDomain string, mainStoragePath string, mainStorageVolume uint64, extraStorage map[string]uint64) {
	no := node.NewNode(10)
	pc := newProviderConfig(no, walletAddress, billEmail, availability, upBandwidth, downBandwidth, mainStoragePath, mainStorageVolume)
	fmt.Println(trackerServer)
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s\n", err.Error())
		return
	}
	defer conn.Close()
	prsc := trp_pb.NewProviderRegisterServiceClient(conn)
	pubKeyBytes, clientIp, err := client.GetPublicKey(prsc)
	if err != nil {
		fmt.Printf("GetPublicKey failed: %s\n", err.Error())
		return
	}
	pubKey, err := x509.ParsePKCS1PublicKey(pubKeyBytes)
	if err != nil {
		fmt.Printf("Parse PublicKey failed: %s\n", err.Error())
		return
	}
	extraStorageSlice := make([]uint64, 0, len(extraStorage))
	for _, v := range extraStorage {
		extraStorageSlice = append(extraStorageSlice, v)
	}
	if host == "" && dynamicDomain == "" {
		fmt.Println("not specify host and dynamic domain, will use: " + clientIp)
		host = clientIp
	}
	code, errMsg, err := client.Register(prsc, encrypt(pubKey, no.NodeId),
		encrypt(pubKey, no.PubKeyBytes), encrypt(pubKey, no.EncryptKey["0"]), encrypt(pubKey, []byte(pc.WalletAddress)),
		encrypt(pubKey, []byte(pc.BillEmail)), mainStorageVolume, upBandwidth, downBandwidth,
		testUpBandwidth, testDownBandwidth, availability, port, encrypt(pubKey, []byte(host)), encrypt(pubKey, []byte(dynamicDomain)), extraStorageSlice, no.PriKey)
	if err != nil {
		fmt.Println("Register failed: " + err.Error())
		return
	}
	if code != 0 {
		fmt.Println(errMsg)
		return
	}
	path := config.CreateProviderConfig(configDir, pc)
	fmt.Println("Register success, please recieve verify code email to verify bill email and backup your config file: " + path)
}

func addStorage(configDir string, trackerServer string, path string, volume string) {
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
