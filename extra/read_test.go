package extra

import (
	"bytes"
	"flag"
	"fmt"

	//	"io/ioutil"
	"math/rand"
	"os"
	"syscall"
	"testing"

	"github.com/dgraph-io/badger/v2/y"
)

func createFile(writeBuf []byte) string {
	filename := fmt.Sprintf("/tmp/rwbench_%16x", rand.Int63())
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|syscall.O_DSYNC, 0666)
	y.Check(err)
	defer f.Close()
	_, err = f.Write(writeBuf)
	y.Check(err)
	return filename
}

const (
	modeControl = iota
	modeDisk
	modeMmap
	modeRAM
)

func BenchmarkRead(b *testing.B) {
	n := 64 << 20
	writeBuf := make([]byte, n)

	//	for _, mode := range []int{modeControl, modeDisk, modeMmap, modeRAM} {
	for _, mode := range []int{modeMmap, modeRAM} {
		var mmap []byte
		if mode == modeRAM {
			mmap = make([]byte, n) // Don't count the time to make this.
		}
		for _, m := range []int{1 << 2, 1 << 4, 1 << 6} {
			b.Run(fmt.Sprintf("mode=%d,m=%d", mode, m), func(b *testing.B) {
				y.AssertTruef((n%m) == 0, "%d %d", n, m)

				b.ResetTimer()
				for j := 0; j < b.N; j++ {
					func() {
						filename := createFile(writeBuf)
						defer os.Remove(filename)
						if mode == modeControl {
							return
						}

						// Measure time to open and read the whole file.
						f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|syscall.O_DSYNC, 0666)
						y.Check(err)
						defer f.Close()

						if mode == modeMmap {
							mmap, err = syscall.Mmap(int(f.Fd()), 0, n,
								syscall.PROT_READ, syscall.MAP_PRIVATE|syscall.MAP_POPULATE)
							y.Check(err)
						} else if mode == modeRAM {
							f.ReadAt(mmap, 0)
						}

						readBuf := make([]byte, m)
						numIters := n / m
						var written int
						if mode == modeDisk {
							for i := 0; i < numIters; i++ {
								k, err := f.Read(readBuf)
								y.Check(err)
								written += k
							}
						} else {
							in := bytes.NewBuffer(mmap)
							for i := 0; i < numIters; i++ {
								k, err := in.Read(readBuf)
								y.Check(err)
								written += k
							}
						}
						y.AssertTruef(written == n, "%d %d", written, n)

						if mode == modeMmap {
							y.Check(syscall.Munmap(mmap))
						}
					}()
				}

			})
		}
	}

}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}
