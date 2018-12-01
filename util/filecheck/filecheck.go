package filecheck

import (
	"crypto/sha256"
	"io"
	"math/big"
	"os"

	"github.com/Nik-U/pbc"
	util_bytes "github.com/samoslab/nebula/util/bytes"
)

const param_str = `type a
q 3002478024387064405972144620549462508470407736306576747124945076471270999639488002912145296429925462661426869942567633459180609465273495576145476905067727
h 4108757660880982117779991133481651099122489506463043933048116394881517164177595631157352416201767211244336
r 730750818665534535851578973600197997769233793023
exp2 159
exp1 116
sign1 1
sign0 -1`

func GenMetadata(filepath string, chunkSize uint32) (paramStr string, generator []byte, pubKeyBytes []byte, random []byte, phi [][]byte, er error) {
	file, err := os.Open(filepath)
	if err != nil {
		er = err
		return
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		er = err
		return
	}
	size := (fi.Size() + int64(chunkSize) - 1) / int64(chunkSize)
	phi = make([][]byte, 0, size)
	// params := pbc.GenerateA(160, 512)
	// paramStr = params.String()
	paramStr = param_str
	pairing, err := pbc.NewPairingFromString(param_str)
	if err != nil {
		panic(err)
	}
	// pairing := params.NewPairing() // TODO fix panic
	g := pairing.NewG2().Rand()
	generator = g.Bytes()
	priKey := pairing.NewZr().Rand()
	pubKey := pairing.NewG2().PowZn(g, priKey)
	pubKeyBytes = pubKey.Bytes()
	u := pairing.NewG1().Rand()
	random = u.Bytes()
	uPower := u.PreparePower()
	buf := make([]byte, chunkSize)
	i := 0
	for {
		bytesRead, err := file.Read(buf)
		if err != nil && err != io.EOF {
			er = err
			return
		}
		if bytesRead > 0 {
			i++
			e1 := pairing.NewG1().SetFromHash(hash(pubKeyBytes, uint32(i)))
			bm := new(big.Int)
			bm.SetBytes(buf[:bytesRead])
			e2 := pairing.NewG1()
			uPower.PowZn(e2, pairing.NewZr().SetBig(bm))
			e1.Mul(e1, e2)
			e1.PowZn(e1, priKey)
			phi = append(phi, e1.Bytes())
		}
		if bytesRead < int(chunkSize) {
			break
		}
	}
	return
}

func hash(pubKeyBytes []byte, i uint32) []byte {
	hasher := sha256.New()
	hasher.Write(pubKeyBytes)
	hasher.Write(util_bytes.FromUint32(i))
	return hasher.Sum(nil)
}
