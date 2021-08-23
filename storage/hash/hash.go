package hash

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"strconv"
	"strings"

	errorpkg "github.com/peak/s5cmd/error"
	"github.com/peak/s5cmd/storage"
)

// ObjectHash defines hash related properties of storage.Object structure
type ObjectHash struct {
	object    *storage.Object
	multipart int
	isLocal   bool
}

// New returns a new ObjectHash object.
func New(object *storage.Object) *ObjectHash {
	// fmt.Printf("object %#v\n", object.URL)
	return &ObjectHash{
		object:    object,
		multipart: checkMultipart(object.Etag),
		isLocal:   !object.URL.IsRemote(),
	}
}

// Different checks is given objecthash is different than source hash.
func (o *ObjectHash) Different(target *ObjectHash) error {
	if o.multipart != 0 { // source is multipart uploaded.
		return errorpkg.ErrorMultipartUpload
	}

	if target.multipart != 0 { // target is multipart uploaded.
		return errorpkg.ErrorMultipartUpload
	}

	if o.isLocal { // local -> remote
		localHash, err := fileHash(o.object.URL.Path)
		if err != nil {
			return err
		}
		if localHash == target.object.Etag {
			return errorpkg.ErrorSameHash
		} else {
			return nil
		}
	} else {
		if target.isLocal { // remote -> local
			localHash, err := fileHash(target.object.URL.Path)
			if err != nil {
				return err
			}
			if o.object.Etag == localHash {
				return errorpkg.ErrorSameHash
			} else {
				return nil
			}
		} else { // remote -> remote
			if o.object.Etag == target.object.Etag {
				return errorpkg.ErrorSameHash
			} else {
				return nil
			}
		}
	}
}

// checkMultipart checks if given object is uploaded using multipart.
func checkMultipart(hashValue string) int {
	splits := strings.Split(hashValue, "-")
	if len(splits) != 2 {
		return 0
	}
	multipart, _ := strconv.Atoi(splits[1])
	return multipart
}

// fileHash computes hash of local file.
func fileHash(path string) (string, error) {
	file, err := os.Open(path)

	if err != nil {
		return "", err
	}
	defer file.Close()
	return fileToHash(file)
}

// fileToHash converts file to hash.
func fileToHash(r io.Reader) (string, error) {
	var MD5String string
	hash := md5.New()

	if _, err := io.Copy(hash, r); err != nil {
		return MD5String, err
	}

	hashInBytes := hash.Sum(nil)[:16]
	MD5String = hex.EncodeToString(hashInBytes)

	return MD5String, nil
}