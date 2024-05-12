package object

import (
	"time"

	"github.com/bastienvty/netsecfs/utils"
)

var logger = utils.GetLogger("juicefs")

type Object interface {
	Key() string
	Size() int64
	Mtime() time.Time
	IsDir() bool
	IsSymlink() bool
	StorageClass() string
}

type obj struct {
	key   string
	size  int64
	mtime time.Time
	isDir bool
	sc    string
}

func (o *obj) Key() string          { return o.key }
func (o *obj) Size() int64          { return o.size }
func (o *obj) Mtime() time.Time     { return o.mtime }
func (o *obj) IsDir() bool          { return o.isDir }
func (o *obj) IsSymlink() bool      { return false }
func (o *obj) StorageClass() string { return o.sc }

// ObjectStorage is the interface for object storage.
// all of these API should be idempotent.
type ObjectStorage interface {
	// Description of the object storage.
	String() string
	// Get the data for the given object specified by key.
	/*Get(key string, off, limit int64, getters ...AttrGetter) (io.ReadCloser, error)
	// Put data read from a reader to an object specified by key.
	Put(key string, in io.Reader, getters ...AttrGetter) error
	// Copy an object from src to dst.
	Copy(dst, src string) error
	// Delete a object.
	Delete(key string, getters ...AttrGetter) error*/
}

type Shutdownable interface {
	Shutdown()
}

func Shutdown(o ObjectStorage) {
	if s, ok := o.(Shutdownable); ok {
		s.Shutdown()
	}
}
