package daemon

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/reedsolomon"
	util_hash "github.com/samoslab/nebula/util/hash"
)

// RsEncoder reedsolomon stream encoder file
func RsEncoder(outDir, fname string, dataShards, parShards int) ([]HashFile, error) {
	enc, err := reedsolomon.NewStream(dataShards, parShards)
	if err != nil {
		return nil, err
	}

	fmt.Println("Opening", fname)
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	instat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	shards := dataShards + parShards
	out := make([]*os.File, shards)

	// Create the resulting files.
	dir, file := filepath.Split(fname)
	if outDir != "" {
		dir = outDir
	}
	for i := range out {
		outfn := fmt.Sprintf("%s.%d", file, i)
		fmt.Println("Creating", outfn)
		out[i], err = os.Create(filepath.Join(dir, outfn))
		if err != nil {
			return nil, err
		}
	}

	// Split into files.
	data := make([]io.Writer, dataShards)
	for i := range data {
		data[i] = out[i]
	}
	// Do the split
	err = enc.Split(f, data, instat.Size())

	// Close and re-open the files.
	input := make([]io.Reader, dataShards)

	for i := range data {
		out[i].Close()
		f, err := os.Open(out[i].Name())
		if err != nil {
			return nil, err
		}
		input[i] = f
		defer f.Close()
	}

	// Create parity output writers
	parity := make([]io.Writer, parShards)
	for i := range parity {
		parity[i] = out[dataShards+i]
		defer out[dataShards+i].Close()
	}

	// Encode parity
	err = enc.Encode(input, parity)
	if err != nil {
		return nil, err
	}
	fmt.Printf("File split into %d data + %d parity shards.\n", dataShards, parShards)
	result := []HashFile{}
	for i := range out {
		outfn := filepath.Join(dir, fmt.Sprintf("%s.%d", file, i))
		hash, err := util_hash.Sha1File(outfn)
		if err != nil {
			return nil, err
		}
		fileInfo, err := os.Stat(outfn)
		if err != nil {
			return nil, err
		}
		//fmt.Printf("filename %s, hash %+v ,size %d\n", outfn, hash, fileInfo.Size())
		hf := HashFile{}
		hf.FileHash = hash
		hf.FileName = outfn
		hf.FileSize = fileInfo.Size()
		hf.SliceIndex = i
		result = append(result, hf)
	}
	return result, nil
}

// RsDecoder reedsolomon stream decoder file
func RsDecoder(fname, outfname string, dataShards, parShards int) error {
	// Create matrix
	enc, err := reedsolomon.NewStream(dataShards, parShards)
	if err != nil {
		return err
	}

	// Open the inputs
	shards, size, err := openInput(dataShards, parShards, fname)
	if err != nil {
		return err
	}

	// Verify the shards
	ok, err := enc.Verify(shards)
	if ok {
		fmt.Println("No reconstruction needed")
	} else {
		fmt.Println("Verification failed. Reconstructing data")
		shards, size, err = openInput(dataShards, parShards, fname)
		if err != nil {
			return err
		}
		// Create out destination writers
		out := make([]io.Writer, len(shards))
		for i := range out {
			if shards[i] == nil {
				outfn := fmt.Sprintf("%s.%d", fname, i)
				fmt.Println("Creating", outfn)
				out[i], err = os.Create(outfn)
				if err != nil {
					return err
				}
			}
		}
		err = enc.Reconstruct(shards, out)
		if err != nil {
			fmt.Println("Reconstruct failed -", err)
			return err
		}
		// Close output.
		for i := range out {
			if out[i] != nil {
				err := out[i].(*os.File).Close()
				return err
			}
			outfn := fmt.Sprintf("%s.%d", fname, i)
			if len(outfn) > 3 {
				fmt.Printf("remove file %s\n", outfn)
				//if err := os.Remove(outfn); err != nil {
				//fmt.Printf("remove file %s err %v", outfn, err)
				//}
			}
		}
		shards, size, err = openInput(dataShards, parShards, fname)
		ok, err = enc.Verify(shards)
		if !ok {
			fmt.Println("Verification failed after reconstruction, data likely corrupted:", err)
		}
		if err != nil {
			return err
		}
	}

	// Join the shards and write them
	outfn := outfname
	if outfn == "" {
		outfn = fname
	}

	fmt.Println("Writing data to", outfn)
	f, err := os.Create(outfn)
	if err != nil {
		return err
	}
	defer f.Close()

	shards, size, err = openInput(dataShards, parShards, fname)
	if err != nil {
		return err
	}

	// We don't know the exact filesize.
	err = enc.Join(f, shards, int64(dataShards)*size)
	if err != nil {
		return err
	}

	return nil
}

func openInput(dataShards, parShards int, fname string) (r []io.Reader, size int64, err error) {
	// Create shards and load the data.
	shards := make([]io.Reader, dataShards+parShards)
	for i := range shards {
		infn := fmt.Sprintf("%s.%d", fname, i)
		fmt.Println("Opening", infn)
		f, err := os.Open(infn)
		if err != nil {
			fmt.Println("Error reading file", err)
			shards[i] = nil
			continue
		} else {
			shards[i] = f
		}
		stat, err := f.Stat()
		checkErr(err)
		if stat.Size() > 0 {
			size = stat.Size()
		} else {
			shards[i] = nil
		}
	}
	return shards, size, nil
}

func checkErr(err error) error {
	if err != nil {
		return err
	}
	return nil
}
