package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"meow"
	"os"
	"time"
)

// Similar to the meow_example program from the upstream repo,
// this program hashes either a 16,000 byte buffer, a single file,
// or 2 files for comparison, depending on the number of args.

func main() {
	flag.Usage = func() {
		fmt.Println("Usage:")
		fmt.Printf("%s - hash a test buffer\n", os.Args[0])
		fmt.Printf("%s [filename] - hash the contents of [filename]\n", os.Args[0])
		fmt.Printf("%s [filename0] [filename1] - hash the contents of [filename0] and [filename1] and compare them\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	start := time.Now()
	switch len(os.Args) {
	case 1:
		hashBuffer()
	case 2:
		hashFile(os.Args[1])
	case 3:
		compareTwoFiles(os.Args[1], os.Args[2])
	default:
		flag.Usage()
	}
	fmt.Printf("\ntook %s\n", time.Since(start))
}

// hashBuffer create and hashes a repeating 16,000 byte buffer.
func hashBuffer() {
	const size = 16000
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i)
	}
	hash := meow.Hash(data)
	fmt.Printf("Hash of a test buffer:\n\t%s\n", meow.String(hash[:]))
}

// hashFile hashes a single file's contents.
func hashFile(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	hash := meow.Hash(data)
	fmt.Printf("Hash of \"%s\":\n\t%s\n", filename, meow.String(hash[:]))
}

// compareTwoFiles hashes and compares the contents of two files.
func compareTwoFiles(filenameA, filenameB string) {
	dataA, err := ioutil.ReadFile(filenameA)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dataB, err := ioutil.ReadFile(filenameB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hashA := meow.Hash(dataA)
	hashB := meow.Hash(dataB)

	filesMatch := bytes.Equal(dataA, dataB)
	hashesMatch := bytes.Equal(hashA, hashB)

	switch {
	case filesMatch && hashesMatch:
		fmt.Printf("Files \"%s\" and \"%s\" are the same:\n\t%s\n",
			filenameA, filenameB, meow.String(hashA))

	case filesMatch:
		fmt.Println("MEOW HASH FAILURE: Files match but hashes don't!")
		fmt.Printf("\tHash of \"%s\":\n\t %s\n", filenameA, meow.String(hashA))
		fmt.Printf("\tHash of \"%s\":\n\t %s\n", filenameB, meow.String(hashB))

	case hashesMatch:
		fmt.Println("MEOW HASH FAILURE: Hashes match but files don't!")
		fmt.Printf("\tHash of both \"%s\" and \"%s\":\n\t%s\n",
			filenameA, filenameB, meow.String(hashA))

	default:
		fmt.Printf("Files \"%s\" and \"%s\" are different:\n", filenameA, filenameB)
		fmt.Printf("\tHash of \"%s\":\n\t %s\n", filenameA, meow.String(hashA))
		fmt.Printf("\tHash of \"%s\":\n\t %s\n", filenameB, meow.String(hashB))

	}
}
