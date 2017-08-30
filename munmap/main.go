// Run a small check to verify what happens when we try to access a slice
// backed by an mmap, after the file has been munmap-ed.
package main

import (
	"fmt"
	"os"

	"github.com/dgraph-io/badger/y"
)

func main() {
	fd, err := os.OpenFile("lorem.txt", os.O_RDONLY, 0666)
	defer fd.Close()
	y.Check(err)

	fi, err := fd.Stat()
	y.Check(err)

	mmap, err := y.Mmap(fd, false, fi.Size())
	y.Check(err)

	before := string(mmap[0:5])
	fmt.Printf("before mmap: %s\n", before)

	err = y.Munmap(mmap)
	y.Check(err)

	after := string(mmap[0:5])
	fmt.Printf("after mmap: %s\n", after)
}
