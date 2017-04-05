package extra

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgraph-io/badger/y"
)

// BenchmarkWrite gives us write speed to some drive.
func BenchmarkWrite(b *testing.B) {
	f, err := ioutil.TempFile("/tmp/ramdisk", "table_")
	defer f.Close()
	defer os.Remove(f.Name())

	y.Check(err)
	buf := make([]byte, b.N*1000)
	_, err = f.Write(buf)
}
