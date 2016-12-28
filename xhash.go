// (C) 2016 by Ricardo Branco
//
// MIT License
//
// v0.1
//
// TODO:
// + Support HMAC
// + Read filenames from file
// + Support -c option like md5sum(1)
// + Use getopt
// + Use different output formats for display

package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
	"hash"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var hashes = map[string]*struct {
	sum string
	hash.Hash
	size int
}{
	"BLAKE2b256": {"", nil, 32},
	"BLAKE2b384": {"", nil, 48},
	"BLAKE2b512": {"", nil, 64},
	"BLAKE2s256": {"", nil, 32},
	"MD4":        {"", nil, 16},
	"MD5":        {"", nil, 16},
	"RIPEMD160":  {"", nil, 20},
	"SHA1":       {"", nil, 20},
	"SHA224":     {"", nil, 28},
	"SHA256":     {"", nil, 32},
	"SHA384":     {"", nil, 48},
	"SHA512":     {"", nil, 64},
	"SHA512-224": {"", nil, 28},
	"SHA512-256": {"", nil, 32},
	"SHA3-224":   {"", nil, 28},
	"SHA3-256":   {"", nil, 32},
	"SHA3-384":   {"", nil, 48},
	"SHA3-512":   {"", nil, 64},
}

// The keys to the above dictionary in sorted order
var keys []string

var progname string

func main() {
	progname = path.Base(os.Args[0])

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [-s STRING...]|[FILE... DIRECTORY...]\n\n", progname)
		flag.PrintDefaults()
	}

	chosen := make(map[string]*bool)
	for h := range hashes {
		chosen[h] = flag.Bool(strings.ToLower(h), false, fmt.Sprintf("%s algorithm", h))
	}

	var all, isString *bool
	all = flag.Bool("all", false, "all algorithms")
	isString = flag.Bool("s", false, "treat arguments as strings")

	sha3_hashes := []string{}
	sha2_hashes := []string{}
	blake2_hashes := []string{}
	var size_hashes = map[int]*struct {
		hashes []string
		set    *bool
	}{
		128: {},
		160: {},
		224: {},
		256: {},
		384: {},
		512: {},
	}

	for h := range hashes {
		if strings.HasPrefix(h, "SHA3-") {
			sha3_hashes = append(sha3_hashes, h)
		} else if strings.HasPrefix(h, "SHA2") || strings.HasPrefix(h, "SHA384") || strings.HasPrefix(h, "SHA512") {
			sha2_hashes = append(sha2_hashes, h)
		} else if strings.HasPrefix(h, "BLAKE2") {
			blake2_hashes = append(blake2_hashes, h)
		}
		size_hashes[hashes[h].size*8].hashes = append(size_hashes[hashes[h].size*8].hashes, h)
	}

	for size := range size_hashes {
		sizeStr := strconv.Itoa(size)
		size_hashes[size].set = flag.Bool(sizeStr, false, "all "+sizeStr+" algorithms")
	}
	var sha3Opt, sha2Opt, blake2Opt *bool
	if len(sha3_hashes) > 0 {
		sha3Opt = flag.Bool("sha3", false, "all SHA-3 algorithms")
	}
	if len(sha2_hashes) > 0 {
		sha2Opt = flag.Bool("sha2", false, "all SHA-2 algorithms")
	}
	if len(blake2_hashes) > 0 {
		blake2Opt = flag.Bool("blake2", false, "all BLAKE2 algorithms")
	}

	flag.Parse()

	for size := range size_hashes {
		if !*(size_hashes[size].set) {
			continue
		}
		for _, h := range size_hashes[size].hashes {
			*chosen[h] = true
		}
	}

	if *sha3Opt {
		for _, h := range sha3_hashes {
			*chosen[h] = true
		}
	}
	if *sha2Opt {
		for _, h := range sha2_hashes {
			*chosen[h] = true
		}
	}
	if *blake2Opt {
		for _, h := range blake2_hashes {
			*chosen[h] = true
		}
	}

	if !(*all || choices(chosen)) {
		*all = true
	}

	for h := range chosen {
		if (*all && *chosen[h]) || !(*all || *chosen[h]) {
			delete(hashes, h)
		}
	}

	for k := range hashes {
		keys = append(keys, k)
		switch k {
		case "BLAKE2b256":
			hashes[k].Hash = blake2_(blake2b.New256)
		case "BLAKE2b384":
			hashes[k].Hash = blake2_(blake2b.New384)
		case "BLAKE2b512":
			hashes[k].Hash = blake2_(blake2b.New512)
		case "BLAKE2s256":
			hashes[k].Hash = blake2_(blake2s.New256)
		case "MD4":
			hashes[k].Hash = md4.New()
		case "MD5":
			hashes[k].Hash = md5.New()
		case "RIPEMD160":
			hashes[k].Hash = ripemd160.New()
		case "SHA1":
			hashes[k].Hash = sha1.New()
		case "SHA224":
			hashes[k].Hash = sha256.New224()
		case "SHA256":
			hashes[k].Hash = sha256.New()
		case "SHA384":
			hashes[k].Hash = sha512.New384()
		case "SHA512":
			hashes[k].Hash = sha512.New()
		case "SHA512-224":
			hashes[k].Hash = sha512.New512_224()
		case "SHA512-256":
			hashes[k].Hash = sha512.New512_256()
		case "SHA3-224":
			hashes[k].Hash = sha3.New224()
		case "SHA3-256":
			hashes[k].Hash = sha3.New256()
		case "SHA3-384":
			hashes[k].Hash = sha3.New384()
		case "SHA3-512":
			hashes[k].Hash = sha3.New512()
		}
	}

	sort.Strings(keys)

	if len(hashes) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		hash_stdin()
		os.Exit(0)
	}

	if *isString {
		for _, s := range flag.Args() {
			hash_string(s)
		}
	} else {
		for _, pathname := range flag.Args() {
			info, err := os.Stat(pathname)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", progname, err)
				os.Exit(1)
			}
			if info.IsDir() {
				hash_dir(pathname)
			} else {
				hash_file(pathname)
			}
		}
	}
}

// Wrapper for the Blake2 New() methods that needs an optional for MAC
func blake2_(f func([]byte) (hash.Hash, error)) hash.Hash {
	h, _ := f(nil)
	return h
}

// Returns true if at least some algorithm was specified on the command line
func choices(chosen map[string]*bool) bool {
	for h := range chosen {
		if *chosen[h] {
			return true
		}
	}
	return false
}

func display() {
	for _, k := range keys {
		fmt.Printf("%s", hashes[k].sum)
	}
}

func hash_string(str string) {
	done := make(chan bool)
	defer close(done)
	for h := range hashes {
		go func(h string) {
			defer func() { done <- true }()
			hashes[h].Write([]byte(str))
			hashes[h].sum = fmt.Sprintf("%s(\"%s\") = %x\n", h, str, hashes[h].Sum(nil))
			hashes[h].Reset()
			done <- true
		}(h)
	}
	for range hashes {
		<-done
	}
	display()
}

func hash_file(filename string) {
	done := make(chan error)
	defer close(done)
	for h := range hashes {
		go func(h string) {
			f, err := os.Open(filename)
			if err != nil {
				done <- err
				return
			}
			defer f.Close()

			if _, err := io.Copy(hashes[h], f); err != nil {
				done <- err
				return
			}
			hashes[h].sum = fmt.Sprintf("%s(%s) = %x\n", h, filename, hashes[h].Sum(nil))
			hashes[h].Reset()
			done <- nil
		}(h)
	}
	for range hashes {
		err := <-done
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", progname, err)
			os.Exit(1)
		}
	}
	display()
}

func visit(path string, f os.FileInfo, err error) error {
	if f.Mode().IsRegular() {
		hash_file(path)
	}
	return nil
}

func hash_dir(dir string) {
	err := filepath.Walk(dir, visit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s: %v\n", progname, dir, err)
		os.Exit(1)
	}
}

func hash_stdin() {
	done := make(chan error)
	defer close(done)
	for h := range hashes {
		go func(h string) {
			if _, err := io.Copy(hashes[h], os.Stdin); err != nil {
				done <- err
				return
			}
			hashes[h].sum = fmt.Sprintf("%s() = %x\n", h, hashes[h].Sum(nil))
			hashes[h].Reset()
			done <- nil
		}(h)
	}
	for range hashes {
		err := <-done
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", progname, err)
			os.Exit(1)
		}
	}
	display()
}
