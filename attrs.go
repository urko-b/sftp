package sftp

// ssh_FXP_ATTRS support
// see http://tools.ietf.org/html/draft-ietf-secsh-filexfer-02#section-5

import (
	"os"
	"time"
)

const (
	sshFileXferAttrSize        = 0x00000001
	sshFileXferAttrUIDGID      = 0x00000002
	sshFileXferAttrPermissions = 0x00000004
	sshFileXferAttrACmodTime   = 0x00000008
	sshFileXferAttrExtended    = 0x80000000

	sshFileXferAttrAll = sshFileXferAttrSize | sshFileXferAttrUIDGID | sshFileXferAttrPermissions |
		sshFileXferAttrACmodTime | sshFileXferAttrExtended
)

// fileInfo is an artificial type designed to satisfy os.FileInfo.
type fileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	mtime time.Time
	sys   interface{}
}

// Name returns the base name of the file.
func (fi *fileInfo) Name() string { return fi.name }

// Size returns the length in bytes for regular files; system-dependent for others.
func (fi *fileInfo) Size() int64 { return fi.size }

// Mode returns file mode bits.
func (fi *fileInfo) Mode() os.FileMode { return fi.mode }

// ModTime returns the last modification time of the file.
func (fi *fileInfo) ModTime() time.Time { return fi.mtime }

// IsDir returns true if the file is a directory.
func (fi *fileInfo) IsDir() bool { return fi.Mode().IsDir() }

func (fi *fileInfo) Sys() interface{} { return fi.sys }

// FileStat holds the original unmarshalled values from a call to READDIR or
// *STAT. It is exported for the purposes of accessing the raw values via
// os.FileInfo.Sys(). It is also used server side to store the unmarshalled
// values for SetStat.
type FileStat struct {
	Size     uint64
	Mode     uint32
	Mtime    uint32
	Atime    uint32
	UID      uint32
	GID      uint32
	Extended []StatExtended
}

// StatExtended contains additional, extended information for a FileStat.
type StatExtended struct {
	ExtType string
	ExtData string
}

func fileInfoFromStat(st *FileStat, name string) os.FileInfo {
	fs := &fileInfo{
		name:  name,
		size:  int64(st.Size),
		mode:  toFileMode(st.Mode),
		mtime: time.Unix(int64(st.Mtime), 0),
		sys:   st,
	}
	return fs
}

func fileStatFromInfo(fi os.FileInfo) (uint32, FileStat) {
	mtime := fi.ModTime().Unix()
	atime := mtime
	var flags uint32 = sshFileXferAttrSize |
		sshFileXferAttrPermissions |
		sshFileXferAttrACmodTime

	fileStat := FileStat{
		Size:  uint64(fi.Size()),
		Mode:  fromFileMode(fi.Mode()),
		Mtime: uint32(mtime),
		Atime: uint32(atime),
	}

	// os specific file stat decoding
	fileStatFromInfoOs(fi, &flags, &fileStat)

	return flags, fileStat
}
